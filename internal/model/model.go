package model

type User struct {
	ID           int64  `json:"id"`
	Login        string `json:"login"`
	PasswordHash string `json:"-"`
}

type Order struct {
	ID          string  `json:"number"`
	UploadedAt  string  `json:"uploaded_at"`
	ProcessedAt string  `json:"processed_at,omitempty"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
	UserID      int64   `json:"user_id"` // FIXME: might remove UserID from struct
}

type AccrualOrder struct {
	OrderID string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

// Balance shows current user's loyalty points.
type Balance struct {
	UserID         int64   `json:"user_id"` // FIXME: might remove UserID from struct
	Balance        float64 `json:"balance"`
	TotalWithdrawn float64 `json:"total_withdrawn"`
	Updated        string  `json:"updated"`
}

// Withdrawal is a single withdrawal entry to be shown in history later.
type Withdrawal struct {
	Order       string  `json:"order"` // гипотетический номер нового заказа пользователя
	ProcessedAt string  `json:"processed_at,omitempty"`
	Value       float64 `json:"value"`
	UserID      int64   `json:"user_id"` // FIXME: might remove UserID from struct
}
