package storage

import (
	"errors"
	"fmt"
	"runtime"
)

var (
	ErrNotFound       = errors.New("nothing found")
	ErrDuplicateEntry = errors.New("duplicate entry") // or Unique Violation
	ErrCheckViolation = errors.New("check violation") // check constraint failed

	// insufficient funds or negative balance set attempt
	ErrNegativeBalance = errors.New("points balance value can't be negative")
)

func WrapCaller(err error) error {
	if err == nil {
		return err
	}

	_, file, line, ok := runtime.Caller(1)
	if ok {
		return fmt.Errorf("%w [%s:%d]", err, file, line)
	}

	return err
}
