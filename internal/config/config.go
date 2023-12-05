package config

// Config is a struct to setup the service with.
type Config struct {
	RunAddress           string // env: RUN_ADDRESS, flag: -a
	DatabaseDSN          string // env: DATABASE_URI, flag: -d
	AccrualSystemAddress string // env: ACCRUAL_SYSTEM_ADDRESS, flag: -r
}

// New creates config with default values set
func New() *Config {
	return &Config{
		RunAddress: "localhost:8080",
	}
}
