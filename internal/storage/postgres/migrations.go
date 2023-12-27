package postgres

import (
	"embed"
	"errors"
	"fmt"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

const logPrefixMigrate = "db migrations: "

//go:embed migrations/*.sql
var migrationsDir embed.FS

func RunMigrations(databaseURL string, logVerbose bool) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return migrationsErr(err)
	}
	defer d.Close()

	m, err := migrate.NewWithSourceInstance("iofs", d, databaseURL)
	if err != nil {
		return migrationsErr(err)
	}
	defer m.Close()

	m.Log = logger.NewMigrationLogger(logger.Log, logVerbose, logPrefixMigrate)

	checkMigrationsVersion(m)
	skipping := false
	defer func() {
		if !skipping {
			checkMigrationsVersion(m)
		}
	}()

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			skipping = true
			logger.Log.Info(logPrefixMigrate + "no new migrations found, skipping")
			return nil
		}
		return migrationsErr(err)
	}

	logger.Log.Info(logPrefixMigrate + "successfuly updated")

	return nil
}

func migrationsErr(err error) error {
	return fmt.Errorf("bad migrations run attempt: %w", err)
}

func checkMigrationsVersion(m *migrate.Migrate) {
	version, dirty, vErr := m.Version()
	logger.Log.Info(logPrefixMigrate+"version check",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty),
		zap.Error(vErr),
	)
}
