package model

type User struct {
	ID           int64  `json:"id"`
	Login        string `json:"login"`
	PasswordHash string `json:"-"`
}

type AccrualOrder struct {
	OrderID string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
