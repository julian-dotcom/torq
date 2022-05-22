package settings

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

func getSettings(db *sqlx.DB) (settingsData settings, err error) {
	err = db.Get(&settingsData, "SELECT default_date_range, preferred_timezone, week_starts_on FROM settings LIMIT 1;")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return settings{}, nil
		}
		return settings{}, errors.Wrap(err, "Unable to execute SQL query")
	}
	return settingsData, nil
}

func getTimeZones(db *sqlx.DB) (timeZones []timeZone, err error) {
	err = db.Select(&timeZones, "SELECT name FROM pg_timezone_names ORDER BY name;")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]timeZone, 0), nil
		}
		return make([]timeZone, 0), errors.Wrap(err, "Unable to execute SQL query")
	}
	return timeZones, nil
}

func updateSettings(db *sqlx.DB, settings settings) (err error) {
	_, err = db.Exec(`
UPDATE settings SET
  default_date_range = $1,
  preferred_timezone = $2,
  week_starts_on = $3,
  updated_on = $4;
`, settings.DefaultDateRange, settings.PreferredTimezone, settings.WeekStartsOn, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "Unable to execute SQL statement")
	}
	return nil
}
