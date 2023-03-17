package settings

import (
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/pkg/commons"
)

func getSettings(db *sqlx.DB) (settings, error) {
	var settingsData settings
	err := db.Get(&settingsData, `
		SELECT default_date_range, default_language, preferred_timezone, week_starts_on, torq_uuid, mixpanel_opt_out
		FROM settings
		LIMIT 1;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return settings{}, nil
		}
		return settings{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return settingsData, nil
}

func InitializeManagedSettingsCache(db *sqlx.DB, vectorUrl string) error {
	settingsData, err := getSettings(db)
	if err == nil {
		log.Debug().Msg("Pushing settings to ManagedSettings cache.")
		commons.SetSettings(settingsData.DefaultDateRange, settingsData.DefaultLanguage, settingsData.WeekStartsOn,
			settingsData.PreferredTimezone, settingsData.TorqUuid, settingsData.MixpanelOptOut, vectorUrl)
	} else {
		log.Error().Err(err).Msg("Failed to obtain settings for ManagedSettings cache.")
	}
	return nil
}

func getTimeZones(db *sqlx.DB) (timeZones []timeZone, err error) {
	err = db.Select(&timeZones, "SELECT name FROM pg_timezone_names ORDER BY name;")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []timeZone{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return timeZones, nil
}

func updateSettings(db *sqlx.DB, settings settings) (err error) {
	_, err = db.Exec(`
		UPDATE settings SET
		  default_date_range = $1,
		  default_language = $2,
		  preferred_timezone = $3,
		  week_starts_on = $4,
		  mixpanel_opt_out = $5,
		  updated_on = $6;`,
		settings.DefaultDateRange, settings.DefaultLanguage, settings.PreferredTimezone, settings.WeekStartsOn,
		settings.MixpanelOptOut,
		time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	commons.SetSettings(settings.DefaultDateRange, settings.DefaultLanguage, settings.WeekStartsOn,
		settings.PreferredTimezone, settings.TorqUuid, settings.MixpanelOptOut, commons.GetVectorUrlBase())
	return nil
}

func getNodeConnectionDetails(db *sqlx.DB, nodeId int) (NodeConnectionDetails, error) {
	var nodeConnectionDetailsData NodeConnectionDetails
	err := db.Get(&nodeConnectionDetailsData, `SELECT * FROM node_connection_details WHERE node_id = $1;`, nodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NodeConnectionDetails{}, nil
		}
		return NodeConnectionDetails{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodeConnectionDetailsData, nil
}

func GetPingSystemNodeIds(db *sqlx.DB, pingSystem PingSystem) ([]int, error) {
	var nodeIds []int
	err := db.Select(&nodeIds, `
		SELECT node_id
		FROM node_connection_details
		WHERE status_id = $1 AND ping_system%$2>=$3
		ORDER BY node_id;`, commons.Active, pingSystem*2, pingSystem)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []int{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodeIds, nil
}

func getPingConnectionDetails(db *sqlx.DB, pingSystem PingSystem) ([]NodeConnectionDetails, error) {
	var ncds []NodeConnectionDetails
	err := db.Select(&ncds, `
		SELECT *
		FROM node_connection_details
		WHERE status_id = $1 AND ping_system%$2>=$3
		ORDER BY node_id;`, commons.Active, pingSystem*2, pingSystem)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []NodeConnectionDetails{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return ncds, nil
}

func GetAllNodeConnectionDetails(db *sqlx.DB, includeDeleted bool) ([]NodeConnectionDetails, error) {
	var nodeConnectionDetailsArray []NodeConnectionDetails
	var err error
	if includeDeleted {
		err = db.Select(&nodeConnectionDetailsArray, `SELECT * FROM node_connection_details ORDER BY node_id;`)
	} else {
		err = db.Select(&nodeConnectionDetailsArray, `
			SELECT *
			FROM node_connection_details
			WHERE status_id != $1
			ORDER BY node_id;`, commons.Deleted)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []NodeConnectionDetails{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodeConnectionDetailsArray, nil
}

func InitializeManagedNodeCache(db *sqlx.DB) error {
	nodeConnectionDetailsArray, err := GetAllNodeConnectionDetails(db, true)
	if err == nil {
		log.Debug().Msg("Pushing torq nodes to ManagedNodes cache.")
		for _, torqNode := range nodeConnectionDetailsArray {
			node, err := nodes.GetNodeById(db, torqNode.NodeId)
			if err == nil {
				commons.SetTorqNode(node.NodeId, torqNode.Name, torqNode.Status, node.PublicKey, node.Chain, node.Network)
			} else {
				log.Error().Err(err).Msg("Failed to obtain torq node for ManagedNodes cache.")
			}
		}
	} else {
		log.Error().Err(err).Msg("Failed to obtain torq nodes for ManagedNodes cache.")
	}

	log.Debug().Msg("Pushing channel nodes to ManagedNodes cache.")
	rows, err := db.Query(`
		SELECT DISTINCT n.public_key, n.chain, n.network, n.node_id, c.status_id
		FROM node n
		JOIN channel c ON ( c.first_node_id=n.node_id OR c.second_node_id=n.node_id );`)
	if err != nil {
		return errors.Wrap(err, "Obtaining nodeIds and publicKeys")
	}
	defer rows.Close()

	for rows.Next() {
		var publicKey string
		var nodeId int
		var chain commons.Chain
		var network commons.Network
		var channelStatus commons.ChannelStatus
		err = rows.Scan(&publicKey, &chain, &network, &nodeId, &channelStatus)
		if err != nil {
			return errors.Wrap(err, "Obtaining nodeId and publicKey from the resultSet")
		}
		commons.SetChannelNode(nodeId, publicKey, chain, network, channelStatus)
	}
	return nil
}

func InitializeManagedNodeAliasCache(db *sqlx.DB) {
	log.Debug().Msg("Pushing node aliases to ManagedNodeAlias cache.")

	torqNodeIds := commons.GetAllTorqNodeIds()
	for _, torqNodeId := range torqNodeIds {
		torqNodeAlias := getNodeAlias(db, torqNodeId)
		if torqNodeAlias != "" {
			commons.SetNodeAlias(torqNodeId, torqNodeAlias)
		}
		for _, channelId := range commons.GetChannelIdsByNodeId(torqNodeId) {
			channelSettings := commons.GetChannelSettingByChannelId(channelId)
			remoteNodeId := channelSettings.FirstNodeId
			if remoteNodeId == torqNodeId {
				remoteNodeId = channelSettings.SecondNodeId
			}
			remoteNodeAlias := getNodeAlias(db, remoteNodeId)
			if remoteNodeAlias != "" {
				commons.SetNodeAlias(remoteNodeId, remoteNodeAlias)
			}
		}
	}
}

func InitializeManagedTaggedCache(db *sqlx.DB) error {
	log.Debug().Msg("Pushing tags to ManagedTagged cache.")
	rows, err := db.Queryx(`SELECT tag_id, node_id, channel_id FROM tagged_entity;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return errors.Wrap(err, "Obtaining tags")
	}
	defer rows.Close()
	nodeTags := make(map[int][]int)
	channelTags := make(map[int][]int)
	for rows.Next() {
		var tagId int
		var nodeId *int
		var channelId *int
		err := rows.Scan(&tagId, &nodeId, &channelId)
		if err != nil {
			return errors.Wrap(err, "Obtaining nodeId and channelId from the resultSet")
		}
		if tagId != 0 {
			if nodeId != nil && *nodeId != 0 {
				nodeTags[*nodeId] = append(nodeTags[*nodeId], tagId)
			}
			if channelId != nil && *channelId != 0 {
				channelTags[*channelId] = append(channelTags[*channelId], tagId)
			}
		}
	}
	for nodeId, tagIds := range nodeTags {
		commons.SetTagIdsByNodeId(nodeId, tagIds)
	}
	for channelId, tagIds := range channelTags {
		commons.SetTagIdsByChannelId(channelId, tagIds)
	}
	return nil
}

func getNodeAlias(db *sqlx.DB, nodeId int) string {
	var alias string
	err := db.Get(&alias, `SELECT alias FROM node_event WHERE event_node_id = $1 ORDER BY timestamp DESC LIMIT 1;`,
		nodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ""
		}
		log.Info().Msgf("Tried to obtain node alias for ManagedNodeAlias cache with nodeId: %v.", nodeId)
	}
	return alias
}

func getNodeConnectionDetailsByStatus(db *sqlx.DB, status commons.Status) ([]NodeConnectionDetails, error) {
	var nodeConnectionDetailsArray []NodeConnectionDetails
	err := db.Select(&nodeConnectionDetailsArray, `
		SELECT * FROM node_connection_details WHERE status_id = $1 ORDER BY node_id;`, status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []NodeConnectionDetails{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodeConnectionDetailsArray, nil
}

func setNodeConnectionDetailsStatus(db *sqlx.DB, nodeId int, status commons.Status) (int64, error) {
	res, err := db.Exec(`
		UPDATE node_connection_details SET status_id = $1, updated_on = $2 WHERE node_id = $3 AND status_id != $1;`,
		status, time.Now().UTC(), nodeId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}

func setNodeConnectionDetailsPingSystemStatus(db *sqlx.DB,
	nodeId int, pingSystem PingSystem, status commons.Status) (int64, error) {

	var err error
	var res sql.Result
	if status == commons.Active {
		res, err = db.Exec(`
		UPDATE node_connection_details SET ping_system = ping_system+$1, updated_on = $2 WHERE node_id = $3 AND ping_system%$4 < $5;`,
			pingSystem, time.Now().UTC(), nodeId, pingSystem*2, pingSystem)
	} else {
		res, err = db.Exec(`
		UPDATE node_connection_details SET ping_system = ping_system-$1, updated_on = $2 WHERE node_id = $3 AND ping_system%$4 >= $5;`,
			pingSystem, time.Now().UTC(), nodeId, pingSystem*2, pingSystem)
	}
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}

func SetNodeConnectionDetails(db *sqlx.DB, ncd NodeConnectionDetails) (NodeConnectionDetails, error) {
	updatedOn := time.Now().UTC()
	ncd.UpdatedOn = &updatedOn
	_, err := db.Exec(`
		UPDATE node_connection_details
		SET implementation = $1, name = $2, grpc_address = $3, tls_file_name = $4, tls_data = $5,
		    macaroon_file_name = $6, macaroon_data = $7, status_id = $8, ping_system = $9, updated_on = $10,
			custom_settings = $11
		WHERE node_id = $12;`,
		ncd.Implementation, ncd.Name, ncd.GRPCAddress, ncd.TLSFileName, ncd.TLSDataBytes,
		ncd.MacaroonFileName, ncd.MacaroonDataBytes, ncd.Status, ncd.PingSystem, ncd.UpdatedOn,
		ncd.CustomSettings, ncd.NodeId)
	if err != nil {
		return ncd, errors.Wrap(err, database.SqlExecutionError)
	}
	return ncd, nil
}

func SetNodeConnectionDetailsByConnectionDetails(
	db *sqlx.DB,
	nodeId int,
	status commons.Status,
	grpcAddress string,
	tlsDataBytes []byte,
	macaroonDataBytes []byte) error {

	ncd, err := getNodeConnectionDetails(db, nodeId)
	if err != nil {
		return errors.Wrap(err, "Obtaining existing node connection details")
	}
	updatedOn := time.Now().UTC()
	ncd.UpdatedOn = &updatedOn
	ncd.MacaroonDataBytes = macaroonDataBytes
	ncd.TLSDataBytes = tlsDataBytes
	ncd.GRPCAddress = &grpcAddress
	ncd.Status = status
	_, err = SetNodeConnectionDetails(db, ncd)
	return err
}

func addNodeConnectionDetails(db *sqlx.DB, ncd NodeConnectionDetails) (NodeConnectionDetails, error) {
	updatedOn := time.Now().UTC()
	ncd.UpdatedOn = &updatedOn
	_, err := db.Exec(`
		INSERT INTO node_connection_details
		    (node_id, name, implementation, grpc_address, tls_file_name, tls_data, macaroon_file_name, macaroon_data,
		     status_id, ping_system, custom_settings, created_on, updated_on)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13);`,
		ncd.NodeId, ncd.Name, ncd.Implementation, ncd.GRPCAddress, ncd.TLSFileName, ncd.TLSDataBytes,
		ncd.MacaroonFileName, ncd.MacaroonDataBytes, ncd.Status, ncd.PingSystem, ncd.CustomSettings,
		ncd.CreateOn, ncd.UpdatedOn)
	if err != nil {
		return ncd, errors.Wrap(err, database.SqlExecutionError)
	}
	return ncd, nil
}
