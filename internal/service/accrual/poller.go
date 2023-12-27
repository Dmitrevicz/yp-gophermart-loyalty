package accrual

import (
	"errors"
	"fmt"
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/logger"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/util/retry"
	"go.uber.org/zap"
)

type Poller struct {
	client  service.AccrualClient
	storage storage.Storage

	// currently tracked orders
	orders *model.OrdersMap

	// order accrual results channel
	accrualResults chan model.AccrualOrder
}

func NewPoller(accrual service.AccrualClient, storage storage.Storage) *Poller {
	return &Poller{
		client:         accrual,
		storage:        storage,
		accrualResults: make(chan model.AccrualOrder, 32),
	}
}

func (p *Poller) Start() error {
	logger.Log.Info("Starting accrual poller")

	// fetch all new orders
	orders, err := p.storage.Orders().GetByStatus(StatusOrderNew)
	if err != nil {
		return fmt.Errorf("error starting accrual poller: %w", err)
	}

	p.orders = model.NewOrdersMap(len(orders))
	for _, order := range orders {
		p.orders.Set(order)
	}

	go p.process(p.accrualResults)

	// ask accrual service
	for _, order := range p.orders.GetAll() {
		go p.askAccrualService(order.ID, p.accrualResults)
	}

	go p.checkFailedOrdersTicker()

	return nil
}

func (p *Poller) RegisterNewOrder(orderNumber model.OrderNumber) error {
	order, err := p.storage.Orders().Get(orderNumber)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			// must never happen
			logger.Log.Warn("Attempt to register order that does not exist",
				zap.String("order", string(orderNumber)),
			)
			return nil
		}
		return err
	}

	if order.Status != StatusOrderNew {
		return nil
	}

	p.orders.Set(*order)

	go p.askAccrualService(order.ID, p.accrualResults)

	return nil
}

// process handles final response state from accrual service.
// Orders only must go in channel with StatusProcessed or StatusInvalid.
func (p *Poller) process(accrualOrders <-chan model.AccrualOrder) {
	for accrual := range accrualOrders {
		order, ok := p.orders.Get(accrual.OrderID)
		if !ok {
			continue
		}

		// update info in tracker map
		order.Status = accrual.Status
		order.Accrual = accrual.Accrual
		p.orders.Set(order)

		processedAt, ok := p.updateProcessedOrders(order)
		if !ok {
			continue
		}

		// set order status and accrual value in db
		/* ts, err := p.storage.Orders().SetProcessedStatus(order.ID, order.Status, order.Accrual)
		if err != nil {
			logger.Log.Error("Error changing order status", zap.Error(err),
				zap.String("order", string(order.ID)),
				zap.String("status", order.Status),
				zap.Float64("accrual", order.Accrual),
			)
			// failed orders will be tried again in checkFailedOrdersTicker()
			continue
		}

		if order.Status == StatusProcessed {
			// add earned points to user's balance
			_, err = p.storage.Balance().Add(order.Accrual, order.UserID)
			if err != nil {
				logger.Log.Error("Error changing user balance", zap.Error(err),
					zap.String("order", string(order.ID)),
					zap.String("status", order.Status),
					zap.Float64("accrual", order.Accrual),
				)
				// failed orders will be tried again in checkFailedOrdersTicker()
				continue
			}
		} */

		// stop tracking order
		p.orders.Delete(order.ID)

		logger.Log.Info("Order processed successfuly",
			zap.String("order", string(order.ID)),
			zap.String("status", order.Status),
			zap.Float64("accrual", order.Accrual),
			zap.Time("processed_at", processedAt),
		)
	}
}

// updateProcessedOrders sets order status and accrual value in db, and also
// updates user's balance if approved.
func (p *Poller) updateProcessedOrders(order model.Order) (processedAt time.Time, ok bool) {
	// set order status and accrual value in db
	processedAt, err := p.storage.Orders().SetProcessedStatus(order.ID, order.Status, order.Accrual)
	if err != nil {
		logger.Log.Error("Error changing order status", zap.Error(err),
			zap.String("order", string(order.ID)),
			zap.String("status", order.Status),
			zap.Float64("accrual", order.Accrual),
		)
		// failed orders will be tried again in checkFailedOrdersTicker()
		return processedAt, false
	}

	if order.Status == StatusProcessed {
		// add earned points to user's balance
		_, err = p.storage.Balance().Add(order.Accrual, order.UserID)
		if err != nil {
			logger.Log.Error("Error changing user balance", zap.Error(err),
				zap.String("order", string(order.ID)),
				zap.String("status", order.Status),
				zap.Float64("accrual", order.Accrual),
			)
			// failed orders will be tried again in checkFailedOrdersTicker()
			return processedAt, false
		}
	}

	return processedAt, true
}

func (p *Poller) checkFailedOrdersTicker() {
	ticker := time.NewTicker(time.Second * 3)
	for range ticker.C {
		for _, order := range p.orders.GetAll() {
			// get only processed orders that failed
			if order.Status == StatusOrderNew {
				continue
			}

			processedAt, ok := p.updateProcessedOrders(order)
			if !ok {
				// try again later on next tick
				continue
			}

			// stop tracking order
			p.orders.Delete(order.ID)

			logger.Log.Info("Order processed successfuly (after retry on ticker checker)",
				zap.String("order", string(order.ID)),
				zap.String("status", order.Status),
				zap.Float64("accrual", order.Accrual),
				zap.Time("processed_at", processedAt),
			)
		}
	}
}

// askAccrualService registers new order in accrual service and starts asking it
// waiting for final accrual status.
func (p *Poller) askAccrualService(order model.OrderNumber, accruals chan<- model.AccrualOrder) {
	var (
		err    error
		result model.AccrualOrder
	)

	retrier := retry.NewRetrier(retry.RetrierOptions{
		RetryAny: true,
		Infinite: true,
	})

	// retrier will run till final status retrieved
	if err = retrier.Do("ask accrual", func() (cErr error) {
		result, cErr = p.client.Order(order)
		if cErr != nil {
			return cErr
		}

		if p.isStatusFinal(result.Status) {
			return nil
		}

		return model.NewRetriableError(fmt.Errorf("got retriable order accrual status: %s", result.Status))
	}); err != nil {
		logger.Log.Error("Retry finished with error", zap.Error(err))
		return
	}

	accruals <- result
}

// isStatusFinal returns true when retry calls must be stopped.
func (p *Poller) isStatusFinal(s string) bool {
	if s == StatusProcessed || s == StatusInvalid {
		return true
	}

	return false
}
