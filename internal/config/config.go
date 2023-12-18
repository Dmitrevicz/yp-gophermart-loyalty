package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/caarlos0/env/v10"
)

// Config is a struct to setup the service with.
type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`            // flag: -a
	DatabaseDSN          string `env:"DATABASE_URI"`           // flag: -d
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"` // flag: -r
	GinMode              string `env:"GIN_MODE"`               // flag: --gin_mode
	AuthSecretKey        string `env:"SECRET"`                 // flag: -s
	LogLevel             string `env:"LOG_LVL"`                // flag: --log_lvl
}

// New creates config with default values set
func New() *Config {
	return &Config{
		RunAddress: "localhost:8080",
		LogLevel:   "info",
	}
}

// parseFlags defines and parses command-line flags.
func (cfg *Config) parseFlags() {
	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "TCP address for the server to listen on")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "data source name to connect to database")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "bonuses calculator service address")
	flag.StringVar(&cfg.GinMode, "gin_mode", cfg.GinMode, "gin mode")
	flag.StringVar(&cfg.AuthSecretKey, "s", cfg.AuthSecretKey, "secret key")
	flag.StringVar(&cfg.LogLevel, "log_lvl", cfg.LogLevel, "logger level")

	flag.Parse()
}

// Parse parses config from both command-line flags and env.
func (cfg *Config) Parse() (err error) {
	cfg.parseFlags()

	if err = env.Parse(cfg); err != nil {
		return parseError(err)
	}

	if err = cfg.Validate(); err != nil {
		return parseError(err)
	}

	return nil
}

func (cfg *Config) Validate() error {
	if strings.TrimSpace(cfg.AccrualSystemAddress) == "" {
		return validationError("accrual system address is required")
	}

	// XXX: Might move such checks to proper services initialization funcs
	// instead of making config package to be responsible of it as it is now.
	if strings.TrimSpace(cfg.AuthSecretKey) == "" {
		// return validationError("auth secret key is empty") // must be an error in actual real-world production
		log.Println("[Warning] auth secret key is empty")
	}

	return nil
}

func parseError(err error) error {
	return fmt.Errorf("config parse failed: %w", err)
}

func validationError(msg string) error {
	return errors.New("invalid config: " + msg)
}
