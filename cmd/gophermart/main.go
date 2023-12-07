package main

import (
	"fmt"
	"log"
	"net/http"

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

	s := server.New(cfg)
	if err := http.ListenAndServe(cfg.RunAddress, s); err != nil {
		log.Fatal(err)
	}
}
