package main

import (
	"fmt"
	"log"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/logger"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg := config.New()
	fmt.Printf("config (default): %+v\n", *cfg) // XXX: printing to double-check autotests, remove in production

	if err := cfg.Parse(); err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("config (parsed): %+v\n", *cfg)

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalln("failed initializing logger:", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting Server",
		zap.String("addr", cfg.RunAddress),
		zap.String("loglvl", logger.Log.Level().String()),
	)

	if err := server.Start(cfg); err != nil {
		logger.Log.Fatal("Server Start returned error", zap.Error(err))
	}
}
