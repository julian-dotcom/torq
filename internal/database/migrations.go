package database

import (
	"database/sql"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/database/migrations"
	"log"
	"net/http"
)

// newMigrationInstance fetches sql files and creates a new migration instance.
func newMigrationInstance(db *sql.DB) (*migrate.Migrate, error) {
	sourceInstance, err := httpfs.New(http.FS(migrations.MigrationFiles), ".")
	if err != nil {
		return nil, fmt.Errorf("invalid source instance, %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err := migrate.NewWithInstance("httpfs", sourceInstance, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("could not create migration instance: %v", err)
	}

	return m, nil
}

// MigrateUp migrates up to the latest migration version. It should be used when the version number changes.
func MigrateUp(db *sqlx.DB) error {
	m, err := newMigrationInstance(db.DB)
	if err != nil {
		return err
	}

	log.Println("Migrations might take a while. Please be patient.")

	err = m.Up()
	if err != nil {
		return errors.Wrap(err, "Migrating database up")
	}

	log.Println("Migration done.")

	return nil
}
