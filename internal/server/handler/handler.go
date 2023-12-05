package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type handlers struct{}

func New() *handlers {
	return &handlers{}
}

// Register - регистрация пользователя.
func (h *handlers) Register(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// Login - аутентификация пользователя.
func (h *handlers) Login(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// PostOrders - загрузка пользователем номера заказа для расчёта.
func (h *handlers) PostOrders(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// GetOrders - получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях.
func (h *handlers) GetOrders(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// Balance - получение текущего баланса счёта баллов лояльности пользователя.
func (h *handlers) Balance(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// Withdraw - запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа.
func (h *handlers) Withdraw(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// Withdrawals - получение информации о выводе средств с накопительного счёта пользователем.
func (h *handlers) Withdrawals(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}
