package postgres

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
)

type OrdersRepo struct {
	s *Storage
}

func NewOrdersRepo(s *Storage) *OrdersRepo {
	return &OrdersRepo{
		s: s,
	}
}

const fieldsOrders = `
	id,
	user_id,
	uploaded_at,
	status,
	accrual,
	processed_at
`

const queryGetOrder = `SELECT ` + fieldsOrders + `FROM users WHERE id=$1;`

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
	FROM users WHERE id=$1;
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

func (r *OrdersRepo) Create(order model.Order) (id string, err error) {
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
