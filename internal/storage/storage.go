// Package storage contains storage and repository interfaces definition.
package storage

import (
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
)

// Storage is a set of repositories.
type Storage interface {
	Users() UsersRepository
	Balance() BalanceRepository
	Orders() OrdersRepository
}

// UsersRepository is a set of methods to manipulate users' accounts.
//
// Note: I don't use pointers to model.User due to very small struct size.
type UsersRepository interface {
	// Get finds user by id. When requested user doesn't exist
	// storage.ErrNotFound error is returned.
	Get(id int64) (user model.User, err error)
	// FindByLogin finds user by login. When requested user doesn't exist
	// storage.ErrNotFound error is returned.
	FindByLogin(login string) (user model.User, err error)
	Create(user model.User) (id int64, err error)
	Delete(id int64) error
}

// OrdersRepository is a set of methods to manipulate users' orders.
type OrdersRepository interface {
	// Get returns nil order when wasn't found and storage.ErrNotFound error.
	Get(id model.OrderNumber) (order *model.Order, err error)
	GetByUserID(userID int64) (order []model.Order, err error)
	GetByStatus(status string) (order []model.Order, err error)
	LastOrderNumber() (orderNumber model.OrderNumber, err error)
	Create(order model.Order) (id string, err error)
	SetProcessedStatus(orderID model.OrderNumber, status string, accrual float64) (processedAt time.Time, err error)
}

// BalanceRepository is a set of methods to manipulate users' loyalty points.
type BalanceRepository interface {
	// Get returns current balance with total withdrawn value.
	// Error storage.ErrNotFound is returned when no data found.
	Get(userID int64) (balance model.Balance, err error)
	// Add adds new accrual sum to current balance.
	// Returns new updated balance and current total withdrawn value.
	Add(accrual float64, userID int64) (balance model.Balance, err error)
	// Withdraw decreases curent balance and writes entry to history.
	// Parameter orderID is a hypothetical order number.
	Withdraw(sum float64, userID int64, orderID model.OrderNumber) (err error)
	// Withdrawals returns all withdrawal calls for user.
	Withdrawals(userID int64) (history []model.Withdrawal, err error)
}
