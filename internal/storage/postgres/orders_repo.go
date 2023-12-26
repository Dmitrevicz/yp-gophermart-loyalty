package postgres

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/logger"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/util/generator"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type OrdersRepo struct {
	s      *Storage
	numgen *generator.OrderNumberGenerator
}

// NewOrdersRepo may panic
func NewOrdersRepo(s *Storage) (repo *OrdersRepo) {
	repo = &OrdersRepo{
		s: s,
	}

	if err := initOrderNumbersGenerator(repo); err != nil {
		logger.Log.Fatal("can't initialize order number generator", zap.Error(err))
		return
	}

	return
}

func initOrderNumbersGenerator(repo *OrdersRepo) (err error) {
	lastNum, err := repo.LastOrderNumber()
	if err != nil {
		return err
	}

	repo.numgen, err = generator.NewOrderNumberGenerator(string(lastNum))
	if err != nil {
		return err
	}

	return nil
}

const fieldsOrders = `
	id,
	user_id,
	uploaded_at,
	status,
	accrual,
	processed_at
`

const queryGetOrder = `SELECT ` + fieldsOrders + `FROM orders WHERE id=$1;`

// Get returns nil order when wasn't found and storage.ErrNotFound error.
func (r *OrdersRepo) Get(id model.OrderNumber) (order *model.Order, err error) {
	stmt, err := r.s.db.Prepare(queryGetOrder)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// order.ProcessedAt is nullable
	var nsProcessedAt sql.NullTime
	var tsUploadedAt time.Time

	order = new(model.Order)
	if err = stmt.QueryRow(id).Scan(
		&order.ID,
		&order.UserID,
		&tsUploadedAt,
		&order.Status,
		&order.Accrual,
		&nsProcessedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = storage.ErrNotFound
		}
		return nil, err
	}

	if nsProcessedAt.Valid {
		order.ProcessedAt = nsProcessedAt.Time.Format(model.LayoutTimestamps)
	}

	order.UploadedAt = tsUploadedAt.Format(model.LayoutTimestamps)

	return order, nil
}

const queryGetOrdersByUserID = `SELECT ` +
	fieldsOrders + `
	FROM orders WHERE user_id = $1
	ORDER BY uploaded_at ASC;
`

func (r *OrdersRepo) GetByUserID(userID int64) (orders []model.Order, err error) {
	orders = make([]model.Order, 0)

	stmt, err := r.s.db.Prepare(queryGetOrdersByUserID)
	if err != nil {
		return orders, err
	}
	defer stmt.Close()

	// order.ProcessedAt is nullable
	var nsProcessedAt sql.NullTime
	var tsUploadedAt time.Time

	rows, err := stmt.Query(userID)
	if err != nil {
		return orders, err
	}
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		if err = rows.Scan(
			&order.ID,
			&order.UserID,
			&tsUploadedAt,
			&order.Status,
			&order.Accrual,
			&nsProcessedAt,
		); err != nil {
			return orders, err
		}

		if nsProcessedAt.Valid {
			order.ProcessedAt = nsProcessedAt.Time.Format(model.LayoutTimestamps)
		}

		order.UploadedAt = tsUploadedAt.Format(model.LayoutTimestamps)

		orders = append(orders, order)
	}

	return orders, rows.Err()
}

const queryGetOrdersByStatus = `SELECT ` +
	fieldsOrders + `
	FROM orders WHERE status = $1
	ORDER BY uploaded_at ASC;
`

func (r *OrdersRepo) GetByStatus(status string) (orders []model.Order, err error) {
	orders = make([]model.Order, 0)

	stmt, err := r.s.db.Prepare(queryGetOrdersByStatus)
	if err != nil {
		return orders, err
	}
	defer stmt.Close()

	// order.ProcessedAt is nullable
	var nsProcessedAt sql.NullTime
	var tsUploadedAt time.Time

	rows, err := stmt.Query(status)
	if err != nil {
		return orders, err
	}
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		if err = rows.Scan(
			&order.ID,
			&order.UserID,
			&tsUploadedAt,
			&order.Status,
			&order.Accrual,
			&nsProcessedAt,
		); err != nil {
			return orders, err
		}

		if nsProcessedAt.Valid {
			order.ProcessedAt = nsProcessedAt.Time.Format(model.LayoutTimestamps)
		}

		order.UploadedAt = tsUploadedAt.Format(model.LayoutTimestamps)

		orders = append(orders, order)
	}

	return orders, rows.Err()
}

// newOrderNumber generates new order number.
//
// Почему-то сначала подумал, что номер заказа надо генерить самому.
// Не нужно, но пока оставил.
func (r *OrdersRepo) newOrderNumber() (number model.OrderNumber, err error) {
	num, err := r.numgen.New()
	if err != nil {
		if errors.Is(err, generator.ErrOrderGeneratorLimitReached) {
			// when this happen - generator logic might be reworked
			logger.Log.Warn("order number generator error", zap.Error(err),
				zap.String("tip", "generator logic might be reworked"),
			)
		} else {
			return number, err
		}
	}

	return model.OrderNumber(num), nil
}

const queryCreateOrder = `
	INSERT INTO orders (
		id,
		user_id,
		status
	) 
	VALUES ($1, $2, $3) RETURNING id;
`

func (r *OrdersRepo) Create(order model.Order) (id string, err error) {
	if order.ID == "" {
		// Почему-то сначала подумал, что номер заказа надо генерить самому.
		// Не нужно, но пока оставил.
		order.ID, err = r.newOrderNumber()
		if err != nil {
			return id, err
		}
	}

	stmt, err := r.s.db.Prepare(queryCreateOrder)
	if err != nil {
		return id, err
	}
	defer stmt.Close()

	if err = stmt.QueryRow(
		order.ID,
		order.UserID,
		order.Status,
	).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				// might not use it anywhere, but let it be...
				// (there is already handler-level check on order creation)
				return id, storage.ErrDuplicateEntry
			}
		}

		return id, err
	}

	return id, err
}

const querySetProcessedOrder = `
	UPDATE orders
	SET
		status = $2,
		accrual = $3,
		processed_at = $4
	WHERE id = $1;
`

func (r *OrdersRepo) SetProcessedStatus(orderID model.OrderNumber, status string, accrual float64) (processedAt time.Time, err error) {
	processedAt = time.Now()

	_, err = r.s.db.Exec(querySetProcessedOrder,
		orderID,
		status,
		accrual,
		processedAt,
	)
	if err != nil {
		return
	}

	return
}

const queryGetLastOrderNum = `SELECT id FROM orders ORDER BY uploaded_at DESC LIMIT 1;`

func (r *OrdersRepo) LastOrderNumber() (orderNumber model.OrderNumber, err error) {
	if err = r.s.db.QueryRow(queryGetLastOrderNum).Scan(
		&orderNumber,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			orderNumber = "0"
			return orderNumber, nil
		}
		return "0", err
	}

	return orderNumber, nil
}
