package handler

import (
	"net/http"
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/gin-gonic/gin"
)

type handlers struct {
	cfg     *config.Config
	auth    *authService
	Mids    *middlewares
	storage storage.Storage
}

func New(cfg *config.Config, s storage.Storage) *handlers {
	auther := NewAuthService(cfg.AuthSecretKey, time.Hour) // XXX: might put expiration into config

	return &handlers{
		cfg:     cfg,
		auth:    auther,
		Mids:    NewMiddlewares(cfg, auther),
		storage: s,
	}
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
