// Package storage contains storage and repository interfaces definition.
package storage

import "github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"

// Storage is a set of repositories.
type Storage interface {
	Users() UsersRepository
}

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
