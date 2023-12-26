package model

import (
	"regexp"
	"sync"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/google/uuid"
)

type User struct {
	ID           int64  `json:"id"`
	Login        string `json:"login"`
	PasswordHash string `json:"-"`
}

type OrderNumber string

var regexOrderNumber = regexp.MustCompile("^[0-9]+$")

func (num OrderNumber) Validate() (err error) {
	if !regexOrderNumber.MatchString(string(num)) {
		return ErrOrderNumberBadChars
	}

	if err = goluhn.Validate(string(num)); err != nil {
		return ErrOrderNumberLuhnCheck
	}

	return nil
}

type Order struct {
	ID          OrderNumber `json:"number"`
	UploadedAt  string      `json:"uploaded_at"`
	ProcessedAt string      `json:"processed_at,omitempty"`
	Status      string      `json:"status"`
	Accrual     float64     `json:"accrual"`
	UserID      int64       `json:"user_id"` // FIXME: might remove UserID from struct
}

type AccrualOrder struct {
	OrderID OrderNumber `json:"order"`
	Status  string      `json:"status"`
	Accrual float64     `json:"accrual"`
}

// Balance shows current user's loyalty points.
type Balance struct {
	UserID         int64   `json:"user_id"`   // FIXME: might remove UserID from struct
	Balance        float64 `json:"current"`   // current balance of loyalty points
	TotalWithdrawn float64 `json:"withdrawn"` // total withdrawn points amount
	Updated        string  `json:"updated"`
}

// Withdrawal is a single withdrawal entry to be shown in history later.
type Withdrawal struct {
	ID          uuid.UUID `json:"id"`           // withdrawal id
	Order       string    `json:"order"`        // specs: "гипотетический номер нового заказа пользователя"
	ProcessedAt string    `json:"processed_at"` // timestamp
	Value       float64   `json:"sum"`          // withdrawn points amount
	UserID      int64     `json:"user_id"`      // FIXME: might remove UserID from struct
}

type OrdersMap struct {
	orders map[OrderNumber]Order
	mu     sync.RWMutex
}

func NewOrdersMap(size int) *OrdersMap {
	return &OrdersMap{
		orders: make(map[OrderNumber]Order, size),
	}
}

func (m *OrdersMap) Get(id OrderNumber) (order Order, ok bool) {
	m.mu.RLock()
	order, ok = m.orders[id]
	m.mu.RUnlock()

	return
}

func (m *OrdersMap) GetAll() (orders []Order) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.orders) == 0 {
		return nil
	}

	orders = make([]Order, 0, len(m.orders))
	for _, order := range m.orders {
		orders = append(orders, order)
	}

	return orders
}

func (m *OrdersMap) Set(order Order) {
	m.mu.Lock()
	m.orders[order.ID] = order
	m.mu.Unlock()
}

func (m *OrdersMap) Delete(id OrderNumber) {
	m.mu.Lock()
	delete(m.orders, id)
	m.mu.Unlock()
}
