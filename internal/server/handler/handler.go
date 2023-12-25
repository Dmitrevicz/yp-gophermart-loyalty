package handler

import (
	"net/http"
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service/accrual"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service/auth"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/gin-gonic/gin"
)

type handlers struct {
	cfg     *config.Config
	auth    service.AuthService
	accrual service.AccrualService
	Mids    *middlewares
	storage storage.Storage
}

func New(cfg *config.Config, s storage.Storage) *handlers {
	auther := auth.New(cfg.AuthSecretKey, time.Second*time.Duration(cfg.AuthTokenLifetimeSec))

	return &handlers{
		cfg:     cfg,
		auth:    auther,
		accrual: accrual.New(cfg.AccrualSystemAddress),
		Mids:    NewMiddlewares(cfg, auther),
		storage: s,
	}
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
