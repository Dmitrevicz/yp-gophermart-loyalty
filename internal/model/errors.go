package model

import "errors"

// Order number validation error
var (
	ErrOrderNumberBadChars  = errors.New("order number must contain only arabic numbers")
	ErrOrderNumberLuhnCheck = errors.New("order number must be a valid sequence of Luhn algorithm")
)
