package postgres

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsDir embed.FS

func RunMigrations(databaseURL string) error {
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

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		// if errors.Is(err, migrate.ErrNoChange) {
		// 	// TODO: write logs: "no new migrations found"
		// 	return nil
		// }
		return migrationsErr(err)
	}

	return nil
}

func migrationsErr(err error) error {
	return fmt.Errorf("bad migrations run attempt: %w", err)
}
