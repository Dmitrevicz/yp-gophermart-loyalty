// Package postgres implements storage repository using PostgreSQL as data storage.
package postgres

import (
	"database/sql"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
)

type Storage struct {
	db      *sql.DB
	users   *UsersRepo
	orders  *OrdersRepo
	balance *BalanceRepo
}

func New(db *sql.DB) *Storage {
	s := &Storage{
		db: db,
	}

	// initialize all repos once before they will be used
	s.users = NewUsersRepo(s)
	s.orders = NewOrdersRepo(s)
	s.balance = NewBalanceRepo(s)

	return s
}

func (s *Storage) Users() storage.UsersRepository {
	// if s.users == nil {
	// 	s.users = NewUsersRepo(s)
	// }

	return s.users
}

func (s *Storage) Orders() storage.OrdersRepository {
	return s.orders
}

func (s *Storage) Balance() storage.BalanceRepository {
	return s.balance
}
