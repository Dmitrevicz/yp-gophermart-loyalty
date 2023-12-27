// Package service contains services interfaces definition.
package service

import "github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"

// AccrualService retrieves accrual info from external service.
type AccrualService interface {
	AccrualClient
	Poller() AccrualPoller
}

// AccrualClient retrieves accrual info from external service.
type AccrualClient interface {
	Order(id model.OrderNumber) (model.AccrualOrder, error)
}

type AccrualPoller interface {
	Start() error
	RegisterNewOrder(orderNumber model.OrderNumber) error
}

type AuthTokenProvider interface {
	// CreateToken creates new jwt token for user.
	CreateToken(userID int64) (string, error)
	// ParseToken parses and validates the token.
	ParseToken(tokenString string) (userID int64, err error)
}

type PasswordHasher interface {
	// PasswordHash calculates hash for password.
	PasswordHash(password string) (hash string, err error)
	// CheckPasswordHash compares password and its expected hash.
	CheckPasswordHash(hash, password string) (err error)
	// MaxPasswordLength returns max password length, which sometimes can be
	// limited (like in bcrypt).
	MaxPasswordLength() int
}

type AuthService interface {
	AuthTokenProvider
	PasswordHasher
}
