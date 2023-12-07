package handler

import (
	"net/http"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/gin-gonic/gin"
)

type handlers struct {
	cfg  *config.Config
	Mids *middlewares
}

func New(cfg *config.Config) *handlers {
	return &handlers{
		cfg:  cfg,
		Mids: NewMiddlewares(cfg),
	}
}

// Register - регистрация пользователя.
//
// Route: POST /api/user/register
func (h *handlers) Register(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// Login - аутентификация пользователя.
//
// Route: POST /api/user/login
func (h *handlers) Login(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// PostOrders - загрузка пользователем номера заказа для расчёта.
//
// Route: POST /api/user/orders
func (h *handlers) PostOrders(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// GetOrders - получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях.
//
// Route: GET /api/user/orders
func (h *handlers) GetOrders(c *gin.Context) {
	_ = readContextUserID(c)
	c.AbortWithStatus(http.StatusNotImplemented)
}

// Balance - получение текущего баланса счёта баллов лояльности пользователя.
//
// Route: GET /api/user/balance
func (h *handlers) Balance(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// Withdraw - запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа.
//
// Route: POST /api/user/balance/withdraw
func (h *handlers) Withdraw(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

// Withdrawals - получение информации о выводе средств с накопительного счёта пользователем.
//
// Route: GET /api/user/withdrawals
func (h *handlers) Withdrawals(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}
