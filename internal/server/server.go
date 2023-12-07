package server

import (
	"net/http"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/server/handler"
	"github.com/gin-gonic/gin"
)

type server struct {
	cfg    *config.Config
	router *gin.Engine
}

func New(cfg *config.Config) *server {
	s := &server{
		cfg: cfg,
	}

	s.configureRouter()

	return s
}

func (s *server) configureRouter() {
	h := handler.New(s.cfg)

	gin.SetMode(s.cfg.GinMode)
	s.router = gin.New()

	s.router.Use(gin.Logger(), gin.Recovery()) // might use ~WithWriter() funcs to write to custom logger

	// Switched from this definition to only user's router group,
	// so provided exclusion list is no more needed:
	// r.Use(h.Mids.CheckAuth(
	// 	// excluded paths:
	// 	"/api/user/register",
	// 	"/api/user/login",
	// ))

	api := s.router.Group("/api")
	{
		// auth routes
		api.POST("/user/register", h.Register)
		api.POST("/user/login", h.Login)

		// other routes which require auth token
		user := api.Group("/user")
		user.Use(h.Mids.CheckAuth())
		{
			user.POST("/orders", h.PostOrders)
			user.GET("/orders", h.GetOrders)
			user.GET("/balance", h.Balance)
			user.POST("/balance/withdraw", h.Withdraw)
			user.GET("/withdrawals", h.Withdrawals)
		}
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
