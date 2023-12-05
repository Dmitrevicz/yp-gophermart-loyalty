package main

import (
	"log"
	"net/http"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/config"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/server"
)

func main() {
	cfg := config.New()
	s := server.New(cfg)
	if err := http.ListenAndServe(cfg.RunAddress, s); err != nil {
		log.Fatal(err)
	}
}
