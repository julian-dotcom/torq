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
	"github.com/lncapital/torq/internal/channels"
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

	version, _, err := m.Version()
	if err != nil {
		return errors.Wrap(err, "Getting database migration version")
	}

	// three migrations for converting between c-lightning and lnd short channel id formats
	if version < 37 {
		err = m.Migrate(37)
		if err != nil {
			return errors.Wrap(err, "Migrating to version 37")
		}
		if err = convertShortChannelIds(db); err != nil {
			return errors.Wrap(err, "Converting short channel ids")
		}
		m.Force(38)
		if err != nil {
			return errors.Wrap(err, "Setting database migration version to 38")
		}
	}

	// this state should be impossible but could happen if the migration process is interrupted
	if version == 37 {
		if err = convertShortChannelIds(db); err != nil {
			return errors.Wrap(err, "Converting short channel ids")
		}
	}

	err = m.Up()
	if err != nil {
		return errors.Wrap(err, "Migrating database up")
	}

	return nil
}

func convertShortChannelIds(db *sqlx.DB) error {

	log.Println("Running short channel id conversions")
	log.Println("WARNING: This process can take 10+ minutes")
	log.Println("Please do not interrupt")

	// channel table
	{
		rows, err := db.Query("SELECT short_channel_id FROM channel;")
		if err != nil {
			return errors.Wrap(err, "Selecting short_channel_id from channel")
		}

		for rows.Next() {
			var shortChannelId string
			err = rows.Scan(&shortChannelId)
			if err != nil {
				return errors.Wrap(err, "Scanning short channel id from channel table")
			}
			lndShortChannelId, err := channels.ConvertShortChannelIDToLND(shortChannelId)
			if err != nil {
				return errors.Wrap(err, "Converting short channel id to LND format")
			}
			updateStatement := "UPDATE channel SET lnd_short_channel_id = $1 WHERE short_channel_id = $2"
			if _, err := db.Exec(updateStatement, lndShortChannelId, shortChannelId); err != nil {
				return errors.Wrap(err, "Updating lnd_short_channel_id on channel table")
			}
		}
		err = rows.Err()
		if err != nil {
			return errors.Wrap(err, "Iterating over each channel row")
		}
	}

	// channel_event table
	{
		rows, err := db.Query("SELECT lnd_short_channel_id FROM channel_event;")
		if err != nil {
			return errors.Wrap(err, "Selecting lnd_short_channel_id from channel_event")
		}

		for rows.Next() {
			var lndShortChannelId uint64
			err = rows.Scan(&lndShortChannelId)
			if err != nil {
				return errors.Wrap(err, "Scanning lnd short channel id from channel_event table")
			}
			shortChannelId := channels.ConvertLNDShortChannelID(lndShortChannelId)
			updateStatement := "UPDATE channel_event SET short_channel_id = $1 WHERE lnd_short_channel_id = $2"
			if _, err := db.Exec(updateStatement, shortChannelId, lndShortChannelId); err != nil {
				return errors.Wrap(err, "Updating short_channel_id on channel_event table")
			}
		}
		err = rows.Err()
		if err != nil {
			return errors.Wrap(err, "Iterating over each channel_event row")
		}
	}

	// forward table
	{
		rows, err := db.Query("SELECT lnd_outgoing_short_channel_id, lnd_incoming_short_channel_id FROM forward;")
		if err != nil {
			return errors.Wrap(err, "Selecting lnd_outgoing_short_channel_id and lnd_incoming_short_channel_id from forward")
		}

		for rows.Next() {
			var lndOutgoingShortChannelId uint64
			var lndIncomingShortChannelId uint64
			err = rows.Scan(&lndOutgoingShortChannelId, &lndIncomingShortChannelId)
			if err != nil {
				return errors.Wrap(err, "Scanning lnd_outgoing_short_channel_id and lnd_incoming_short_channel_id from forward table")
			}
			outgoingShortChannelId := channels.ConvertLNDShortChannelID(lndOutgoingShortChannelId)
			updateStatement := "UPDATE forward SET outgoing_short_channel_id = $1 WHERE lnd_outgoing_short_channel_id = $2"
			if _, err := db.Exec(updateStatement, outgoingShortChannelId, lndOutgoingShortChannelId); err != nil {
				return errors.Wrap(err, "Updating outgoing_short_channel_id on forward table")
			}
			incomingShortChannelId := channels.ConvertLNDShortChannelID(lndIncomingShortChannelId)
			updateStatement = "UPDATE forward SET incoming_short_channel_id = $1 WHERE lnd_incoming_short_channel_id = $2"
			if _, err := db.Exec(updateStatement, incomingShortChannelId, lndIncomingShortChannelId); err != nil {
				return errors.Wrap(err, "Updating incoming_short_channel_id on forward table")
			}
		}
		err = rows.Err()
		if err != nil {
			return errors.Wrap(err, "Iterating over each forward table row")
		}
	}

	return nil
}
