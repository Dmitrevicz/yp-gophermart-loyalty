package model

import (
	"errors"
	"fmt"
)

// Order number validation error
var (
	ErrOrderNumberBadChars  = errors.New("order number must contain only arabic numbers")
	ErrOrderNumberLuhnCheck = errors.New("order number must be a valid sequence of Luhn algorithm")
)

type RetriableError error

func NewRetriableError(err error) error {
	return fmt.Errorf("%w", err)
}
