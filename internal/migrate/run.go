package migrate

import (
	"fmt"
	"io/fs"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Run applies all pending up migrations from the given fs (e.g. os.DirFS("migrations")).
// Database URL must be a postgres connection string.
func Run(dbURL string, migrationsFS fs.FS, dir string) error {
	source, err := iofs.New(migrationsFS, dir)
	if err != nil {
		return fmt.Errorf("migrate source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, dbURL)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	if err == migrate.ErrNoChange {
		log.Println("migrate: no pending migrations")
		return nil
	}
	log.Println("migrate: up applied")
	return nil
}
