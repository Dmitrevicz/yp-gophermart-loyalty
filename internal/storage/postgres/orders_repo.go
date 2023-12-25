package postgres

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/logger"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/util/generator"
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

	repo.numgen, err = generator.NewOrderNumberGenerator(lastNum)
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

func (r *OrdersRepo) Get(id string) (order model.Order, err error) {
	stmt, err := r.s.db.Prepare(queryGetOrder)
	if err != nil {
		return order, err
	}
	defer stmt.Close()

	// order.ProcessedAt is nullable
	var nsProcessedAt sql.NullString

	if err = stmt.QueryRow(id).Scan(
		&order.ID,
		&order.UserID,
		&order.UploadedAt,
		&order.Status,
		&order.Accrual,
		&nsProcessedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = storage.ErrNotFound
		}
		return order, err
	}

	if nsProcessedAt.Valid {
		order.ProcessedAt = nsProcessedAt.String
	}

	return order, nil
}

const queryGetOrdersByUserID = `SELECT ` +
	fieldsOrders + `
	FROM orders WHERE id=$1;
`

func (r *OrdersRepo) GetByUserID(userID int64) (orders []model.Order, err error) {
	orders = make([]model.Order, 0)

	stmt, err := r.s.db.Prepare(queryGetOrdersByUserID)
	if err != nil {
		return orders, err
	}
	defer stmt.Close()

	// order.ProcessedAt is nullable
	var nsProcessedAt sql.NullString

	// TODO: timestamps layout = time.RFC3339

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
			&order.UploadedAt,
			&order.Status,
			&order.Accrual,
			&nsProcessedAt,
		); err != nil {
			return orders, err
		}

		if nsProcessedAt.Valid {
			order.ProcessedAt = nsProcessedAt.String
		}
	}

	return orders, rows.Err()
}

const queryCreateOrder = `
	INSERT INTO orders (
		id,
		user_id,
		status
	) 
	VALUES ($1, $2, $3) RETURNING id;
`

// newOrderNumber generates new order number.
func (r *OrdersRepo) newOrderNumber() (number string, err error) {
	if number, err = r.numgen.New(); err != nil {
		if errors.Is(err, generator.ErrOrderGeneratorLimitReached) {
			// when this happen - generator logic might be reworked
			logger.Log.Warn("order number generator error", zap.Error(err),
				zap.String("tip", "generator logic might be reworked"),
			)
		} else {
			return number, err
		}
	}

	return number, nil
}

func (r *OrdersRepo) Create(order model.Order) (id string, err error) {
	order.ID, err = r.newOrderNumber()
	if err != nil {
		return id, err
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

func (r *OrdersRepo) SetProcessedStatus(orderID, status string, accrual float64) (processedAt time.Time, err error) {
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

func (r *OrdersRepo) LastOrderNumber() (orderNumber string, err error) {
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
