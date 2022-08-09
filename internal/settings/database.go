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

func getLocalNode(db *sqlx.DB, localNodeId int) (localNodeData localNode, err error) {
	err = db.Get(&localNodeData, `
SELECT
  local_node_id,
  implementation,
  grpc_address,
  tls_file_name,
  macaroon_file_name,
  disabled,
  deleted
FROM local_node WHERE local_node_id = $1;`, localNodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return localNode{}, nil
		}
		return localNode{}, errors.Wrap(err, "Unable to execute SQL query")
	}
	return localNodeData, nil
}

func getLocalNodes(db *sqlx.DB) (localNodeData []localNode, err error) {
	err = db.Select(&localNodeData, `
SELECT
  local_node_id,
  implementation,
  grpc_address,
  tls_file_name,
  macaroon_file_name,
  disabled,
  deleted
FROM local_node;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []localNode{}, nil
		}
		return []localNode{}, errors.Wrap(err, "Unable to execute SQL query")
	}
	return localNodeData, nil
}

func getLocalNodeConnectionDetails(db *sqlx.DB) (localNodeData localNode, err error) {
	err = db.Get(&localNodeData, `
SELECT
  grpc_address,
  tls_data,
  macaroon_data
FROM local_node LIMIT 1;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return localNode{}, nil
		}
		return localNode{}, errors.Wrap(err, "Unable to execute SQL query")
	}
	return localNodeData, nil
}

func updateLocalNodeDisabledFlag(db *sqlx.DB, localNodeId int, disabled bool) (err error) {
	_, err = db.Exec(` UPDATE local_node SET disabled = $1`, disabled, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "Unable to execute SQL statement")
	}
	return nil
}

func updateLocalNodeDeletedFlag(db *sqlx.DB, localNodeId int, deleted bool) (err error) {
	_, err = db.Exec(` UPDATE local_node SET deleted = $1`, deleted, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "Unable to execute SQL statement")
	}
	return nil
}

func updateLocalNodeDetails(db *sqlx.DB, localNode localNode) (err error) {
	_, err = db.Exec(`
UPDATE local_node SET
  implementation = $1,
  grpc_address = $2,
  updated_on = $3;
`, localNode.Implementation, localNode.GRPCAddress, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "Unable to execute SQL statement")
	}
	return nil
}

func updateLocalNodeTLS(db *sqlx.DB, localNode localNode) (err error) {
	_, err = db.Exec(`
UPDATE local_node SET
  tls_file_name = $1,
  tls_data = $2,
  updated_on = $3;
`, localNode.TLSFileName, localNode.TLSDataBytes, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "Unable to execute SQL statement")
	}
	return nil
}

func updateLocalNodeMacaroon(db *sqlx.DB, localNode localNode) (err error) {
	_, err = db.Exec(`
UPDATE local_node SET
  macaroon_file_name = $1,
  macaroon_data = $2,
  updated_on = $3;
`, localNode.MacaroonFileName, localNode.MacaroonDataBytes, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "Unable to execute SQL statement")
	}
	return nil
}
