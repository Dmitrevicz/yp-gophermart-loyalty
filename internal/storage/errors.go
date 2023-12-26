package storage

import "errors"

var (
	ErrNotFound       = errors.New("nothing found")
	ErrDuplicateEntry = errors.New("duplicate entry") // or Unique Violation
	ErrCheckViolation = errors.New("check violation") // check constraint failed

	// insufficient funds or negative balance set attempt
	ErrNegativeBalance = errors.New("points balance value can't be negative")
)
