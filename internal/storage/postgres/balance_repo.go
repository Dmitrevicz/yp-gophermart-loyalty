package postgres

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/logger"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type BalanceRepo struct {
	s *Storage
}

func NewBalanceRepo(s *Storage) *BalanceRepo {
	return &BalanceRepo{
		s: s,
	}
}

const queryGetBalance = `
	SELECT 
		balance,
		updated, 
		(
			SELECT COALESCE(SUM(w.value),0)
			FROM withdrawals w
			WHERE w.user_id = $1
		) AS total
	FROM loyalty_points 
	WHERE user_id=$1;
`

// Get returns current balance with total withdrawn value.
func (r *BalanceRepo) Get(userID int64) (balance model.Balance, err error) {
	stmt, err := r.s.db.Prepare(queryGetBalance)
	if err != nil {
		return balance, storage.WrapCaller(err)
	}
	defer stmt.Close()

	var tsUpdated time.Time

	if err = stmt.QueryRow(userID).Scan(
		&balance.Balance,
		&tsUpdated,
		&balance.TotalWithdrawn,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = storage.ErrNotFound
		}
		return balance, storage.WrapCaller(err)
	}

	balance.UserID = userID
	balance.Updated = tsUpdated.Format(model.LayoutTimestamps)

	return balance, nil
}

const querySetOrUpdateBalance = `
	WITH new_balance AS (
		INSERT INTO loyalty_points (
			user_id,
			balance,
			updated
		)
		VALUES ($1, $2, now())
		ON CONFLICT(user_id)
		DO UPDATE SET 
			balance=loyalty_points.balance + $2,
			updated=now()
		RETURNING balance, updated
	)
	SELECT 
		balance, 
		updated, 
		(
			SELECT COALESCE(SUM(w.value),0)
			FROM withdrawals w
			WHERE w.user_id = $1
		) AS total
	FROM new_balance;
`

// Add adds new accrual sum to current balance.
// Returns new updated balance and current total withdrawn value.
func (r *BalanceRepo) Add(accrual float64, userID int64) (balance model.Balance, err error) {
	if accrual < 0 {
		accrual = 0
	}

	stmt, err := r.s.db.Prepare(querySetOrUpdateBalance)
	if err != nil {
		return balance, storage.WrapCaller(err)
	}
	defer stmt.Close()

	if err = stmt.QueryRow(
		userID,
		accrual,
	).Scan(
		&balance.Balance,
		&balance.Updated,
		&balance.TotalWithdrawn,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.CheckViolation {
				// new balance value can't be negative
				return balance, storage.WrapCaller(storage.ErrNegativeBalance)
			}
		}

		return balance, storage.WrapCaller(err)
	}

	balance.UserID = userID

	return balance, storage.WrapCaller(err)
}

const queryWithdraw = `
	UPDATE loyalty_points as b
	SET 
		balance=b.balance - $1,
		updated=now()
	WHERE user_id = $2;
`

const queryAddWithdrawHistory = `
	INSERT INTO withdrawals (
		id,
		user_id,
		order_number,
		value,
		processed_at
	)
	VALUES ($1, $2, $3, $4, now())
`

// Withdraw decreases curent balance and writes entry to history.
// Parameter orderID is a hypothetical order number.
func (r *BalanceRepo) Withdraw(sum float64, userID int64, orderID model.OrderNumber) (err error) {
	tx, err := r.s.db.Begin()
	if err != nil {
		return storage.WrapCaller(err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	// 1. decrease balance
	_, err = tx.Exec(queryWithdraw, sum, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.CheckViolation {
				// new balance value can't be negative
				return storage.WrapCaller(storage.ErrNegativeBalance)
			}
		}

		return storage.WrapCaller(err)
	}

	// generate withdrawal id
	wdID, err := uuid.NewV7()
	if err != nil {
		logger.Log.Error("uuid generator failed", zap.Error(err))
		return storage.WrapCaller(err)
	}

	// 2. save withdrawal entry to history
	_, err = tx.Exec(queryAddWithdrawHistory, wdID, userID, orderID, sum)
	if err != nil {
		return storage.WrapCaller(err)
	}

	if err = tx.Commit(); err != nil {
		return storage.WrapCaller(err)
	}

	return nil
}

const queryWithdrawalsHistory = `
	SELECT
	 id,
	 order_number,
	 value,
	 processed_at
	FROM withdrawals WHERE user_id=$1
	ORDER BY processed_at ASC;
`

// Withdrawals returns all withdrawal calls for user.
func (r *BalanceRepo) Withdrawals(userID int64) (history []model.Withdrawal, err error) {
	history = make([]model.Withdrawal, 0)

	stmt, err := r.s.db.Prepare(queryWithdrawalsHistory)
	if err != nil {
		return history, storage.WrapCaller(err)
	}
	defer stmt.Close()

	var tsProcessedAt time.Time

	rows, err := stmt.Query(userID)
	if err != nil {
		return history, storage.WrapCaller(err)
	}
	defer rows.Close()

	for rows.Next() {
		var wd model.Withdrawal
		if err = rows.Scan(
			&wd.ID,
			&wd.Order,
			&wd.Value,
			&tsProcessedAt,
		); err != nil {
			return history, storage.WrapCaller(err)
		}

		wd.ProcessedAt = tsProcessedAt.Format(model.LayoutTimestamps)

		history = append(history, wd)
	}

	return history, storage.WrapCaller(rows.Err())
}
