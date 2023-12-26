package server

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/server/handler"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/service/accrual"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/storage/postgres"
	"github.com/gin-gonic/gin"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type server struct {
	cfg     *config.Config
	router  *gin.Engine
	storage storage.Storage
	accrual service.AccrualService
}

func New(cfg *config.Config, storage storage.Storage, accrual service.AccrualService) *server {
	s := &server{
		cfg:     cfg,
		storage: storage,
		accrual: accrual,
	}

	s.configureRouter()

	return s
}

func (s *server) configureRouter() {
	h := handler.New(s.cfg, s.storage, s.accrual)

	gin.SetMode(s.cfg.GinMode)
	s.router = gin.New()

	s.router.Use(gin.Logger())       // writes requests logs to stdout
	s.router.Use(h.Mids.LogErrors()) // writes errors to stderr using zap logger
	s.router.Use(h.Mids.Recovery())

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

func newDB(dsn string) (db *sql.DB, err error) {
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// ConfigureStorage creates new storage instance for
// provided data source name (database url).
func ConfigureStorage(dsn string) (storage.Storage, error) {
	if dsn == "" {
		return nil, errors.New("can't configure storage: empty data source name (database url)")
	}

	db, err := newDB(dsn)
	if err != nil {
		return nil, errors.New("can't configure storage: " + err.Error())
	}

	return postgres.New(db), nil
}

func Start(cfg *config.Config) (err error) {
	if cfg.DatabaseDSN != "" {
		err = postgres.RunMigrations(cfg.DatabaseDSN, cfg.VerboseMigrateLogger)
		if err != nil {
			return err
		}
	}

	storage, err := ConfigureStorage(cfg.DatabaseDSN)
	if err != nil {
		return err
	}

	accrualService := accrual.New(cfg.AccrualSystemAddress, storage)
	if err = accrualService.Poller().Start(); err != nil {
		return err
	}

	server := New(cfg, storage, accrualService)

	return http.ListenAndServe(cfg.RunAddress, server)
}
