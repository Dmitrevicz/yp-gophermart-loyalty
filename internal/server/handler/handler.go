package handler

import (
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service/accrual"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service/auth"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
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
