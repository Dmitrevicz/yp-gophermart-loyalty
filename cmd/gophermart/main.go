package main

import (
	"fmt"
	"log"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/server"
)

func main() {
	cfg := config.New()
	fmt.Printf("config (default): %+v\n", *cfg) // XXX: printing to double-check autotests, remove in production

	if err := cfg.Parse(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("config (parsed): %+v\n", *cfg)

	if err := server.Start(cfg); err != nil {
		log.Fatal(err)
	}
}
