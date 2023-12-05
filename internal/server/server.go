package server

import (
	"net/http"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/server/handler"
	"github.com/gin-gonic/gin"
)

type server struct {
	router http.Handler
}

func New(cfg *config.Config) *server {
	var s server

	h := handler.New()

	r := gin.New()
	r.POST("/api/user/register", h.Register)
	r.POST("/api/user/login", h.Login)
	r.POST("/api/user/orders", h.PostOrders)
	r.GET("/api/user/orders", h.GetOrders)
	r.GET("/api/user/balance", h.Balance)
	r.POST("/api/user/balance/withdraw", h.Withdraw)
	r.GET("/api/user/withdrawals", h.Withdrawals)

	s.router = r

	return &s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
