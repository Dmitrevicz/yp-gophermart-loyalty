// Package postgres implements storage repository using PostgreSQL as data storage.
package postgres

import (
	"database/sql"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
)

type Storage struct {
	db    *sql.DB
	users *UsersRepo
}

func New(db *sql.DB) *Storage {
	s := &Storage{
		db: db,
	}

	s.users = NewUsersRepo(s)

	return s
}

func (s *Storage) Users() storage.UsersRepository {
	// if s.users == nil {
	// 	s.users = NewUsersRepo(s)
	// }

	return s.users
}
