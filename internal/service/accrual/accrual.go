// Package accrual contains methods to communicate with accrual service.
// Implements AccrualService interface.
package accrual

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/util/client"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/util/sync"
)

const (
	StatusRegistered = "REGISTERED" // заказ зарегистрирован, но начисление не рассчитано
	StatusInvalid    = "INVALID"    // заказ не принят к расчёту, и вознаграждение не будет начислено
	StatusProcessing = "PROCESSING" // расчёт начисления в процессе
	StatusProcessed  = "PROCESSED"  // расчёт начисления окончен
	StatusOrderNew   = "NEW"
)

var (
	pathGetOrderAccrual string = "/api/orders/"
)

const DefaultMaxReq = 32

// AccrualService implements AccrualService interface.
type AccrualService struct {
	client    *http.Client
	semaphore *sync.Semaphore
	poller    *Poller
}

func New(addr string, storage storage.Storage) *AccrualService {
	pathGetOrderAccrual = addr + pathGetOrderAccrual

	accrualService := &AccrualService{
		client:    client.NewClientDefault(),
		semaphore: sync.NewSemaphore(DefaultMaxReq),
	}

	accrualService.poller = NewPoller(accrualService, storage)

	return accrualService
}

func (a *AccrualService) Poller() service.AccrualPoller {
	return a.poller
}

// Order - получение информации о расчёте начислений баллов лояльности.
//
// GET {accrual_service}/api/orders/{number}
func (a *AccrualService) Order(id model.OrderNumber) (accrual model.AccrualOrder, err error) {
	a.semaphore.Acquire()
	defer a.semaphore.Release()

	url := pathGetOrderAccrual + string(id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return accrual, fmt.Errorf("error preparing request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return accrual, model.NewRetriableError(fmt.Errorf("error while doing the request: %w", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return accrual, fmt.Errorf("error reading response bytes: %w", err)
	}

	if len(body) > 0 {
		if err = json.Unmarshal(body, &accrual); err != nil {
			return accrual, fmt.Errorf("error decoding response body: %w", err)
		}
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		// 204 - заказ не зарегистрирован в системе расчета
		case http.StatusNoContent:
			err = model.NewRetriableError(fmt.Errorf(
				"order was not registered - status code: %s, order: %s",
				resp.Status,
				string(id),
			))
		// 429 - превышено количество запросов к сервису
		case http.StatusTooManyRequests:
			headerRetryAfter := resp.Header.Get("Retry-After")
			err = model.NewRetriableError(fmt.Errorf(
				"too many requests - status code: %s, order: %s, retry-after: %s, body: %s",
				resp.Status, string(id), headerRetryAfter, string(body),
			))
		// 500 - внутренняя ошибка сервера
		case http.StatusInternalServerError:
			err = model.NewRetriableError(fmt.Errorf(
				"internal server error - status code: %s, order: %s, body: %s",
				resp.Status, string(id), string(body),
			))
		default:
			err = fmt.Errorf("unexpected response status code: %s, order: %s, body: %s",
				resp.Status, string(id), string(body),
			)
		}
	}

	return accrual, err
}
