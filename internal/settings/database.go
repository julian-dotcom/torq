package settings

import (
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/database"
)

func getSettings(db *sqlx.DB) (settings, error) {
	var settingsData settings
	err := db.Get(&settingsData, `SELECT * FROM settings LIMIT 1;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return settings{}, nil
		}
		return settings{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return settingsData, nil
}

func InitializeSettingsCache(db *sqlx.DB) error {
	settingsData, err := getSettings(db)
	if err == nil {
		log.Debug().Msg("Pushing settings to SettingsCache cache.")
		cache.SetSettings(settingsData.DefaultDateRange, settingsData.DefaultLanguage, settingsData.WeekStartsOn,
			settingsData.PreferredTimezone, settingsData.TorqUuid, settingsData.MixpanelOptOut,
			settingsData.SlackOAuthToken, settingsData.SlackBotAppToken,
			settingsData.TelegramHighPriorityCredentials, settingsData.TelegramLowPriorityCredentials)
	} else {
		log.Error().Err(err).Msg("Failed to obtain settings for SettingsCache cache.")
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
		  slack_oauth_token = $6,
		  slack_bot_app_token = $7,
		  telegram_high_priority_credentials = $8,
		  telegram_low_priority_credentials = $9,
		  updated_on = $10;`,
		settings.DefaultDateRange, settings.DefaultLanguage, settings.PreferredTimezone, settings.WeekStartsOn,
		settings.MixpanelOptOut, settings.SlackOAuthToken, settings.SlackBotAppToken,
		settings.TelegramHighPriorityCredentials, settings.TelegramLowPriorityCredentials,
		time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	cache.SetSettings(settings.DefaultDateRange, settings.DefaultLanguage, settings.WeekStartsOn,
		settings.PreferredTimezone, settings.TorqUuid, settings.MixpanelOptOut,
		settings.SlackOAuthToken, settings.SlackBotAppToken,
		settings.TelegramHighPriorityCredentials, settings.TelegramLowPriorityCredentials)
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

func GetPingSystemNodeIds(db *sqlx.DB, pingSystem core.PingSystem) ([]int, error) {
	var nodeIds []int
	err := db.Select(&nodeIds, `
		SELECT node_id
		FROM node_connection_details
		WHERE status_id = $1 AND ping_system%$2>=$3
		ORDER BY node_id;`, core.Active, pingSystem*2, pingSystem)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []int{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodeIds, nil
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
			ORDER BY node_id;`, core.Deleted)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []NodeConnectionDetails{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodeConnectionDetailsArray, nil
}

func GetNodeDetailsById(db *sqlx.DB, nodeId int) (string, core.Chain, core.Network, error) {
	var publicKey string
	var chain core.Chain
	var network core.Network
	err := db.QueryRowx(`SELECT public_key, chain, network FROM node WHERE node_id=$1;`, nodeId).
		Scan(&publicKey, &chain, &network)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", core.Bitcoin, core.MainNet, nil
		}
		return "", core.Bitcoin, core.MainNet, errors.Wrap(err, database.SqlExecutionError)
	}
	return publicKey, chain, network, nil
}

func AddNodeWhenNew(db *sqlx.DB, publicKey string, chain core.Chain, network core.Network) (int, error) {
	nodeId := cache.GetNodeIdByPublicKey(publicKey, chain, network)
	if nodeId == 0 {
		createdOn := time.Now().UTC()
		err := db.QueryRowx(`INSERT INTO node (public_key, chain, network, created_on)
			VALUES ($1, $2, $3, $4) RETURNING node_id;`,
			publicKey, chain, network, createdOn).Scan(&nodeId)
		if err != nil {
			if err, ok := err.(*pq.Error); ok {
				if err.Code == "23505" {
					err := db.Get(&nodeId, `SELECT node_id FROM node WHERE public_key=$1;`, publicKey)
					if err != nil {
						return 0, errors.Wrapf(err, "Obtaining existing nodeId for publicKey: %v", publicKey)
					}
					return nodeId, nil
				}
			}
			return 0, errors.Wrap(err, database.SqlExecutionError)
		}
		return nodeId, nil
	}
	return nodeId, nil
}

func AddNodeConnectionHistory(db *sqlx.DB,
	torqNodeId int,
	nodeId int,
	address *string,
	setting *core.NodeConnectionSetting,
	connectionStatus *core.NodeConnectionStatus) error {

	createdOn := time.Now().UTC()
	_, err := db.Exec(
		`
				INSERT INTO node_connection_history
				    (node_id, torq_node_id, connection_status, address, setting, created_on)
				VALUES
					($1, $2, $3, $4, $5, $6);`,
		nodeId, torqNodeId, connectionStatus, address, setting, createdOn)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func GetNodeConnectionHistoryWithDetail(db *sqlx.DB,
	torqNodeId int,
	peerNodeId int) (
	address *string,
	setting *core.NodeConnectionSetting,
	connectionStatus *core.NodeConnectionStatus,
	err error) {

	err = db.QueryRowx(
		`SELECT nch.connection_status, nch.address, nch.setting
				FROM node n
    			LEFT JOIN (
					SELECT LAST(node_id, created_on) as node_id,
					       LAST(connection_status, created_on) as connection_status,
					       LAST(address, created_on) as address,
					       LAST(setting, created_on) as setting
			  		FROM node_connection_history
			  		WHERE node_id = $1 AND torq_node_id = $2
			  		GROUP BY node_id, torq_node_id) nch ON nch.node_id = n.node_id
				WHERE n.node_id = $1;`, peerNodeId, torqNodeId).Scan(&connectionStatus, &address, &setting)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
			return
		}
		err = errors.Wrap(err, database.SqlExecutionError)
	}
	return
}

func InitializeNodesCache(db *sqlx.DB) error {
	nodeConnectionDetailsArray, err := GetAllNodeConnectionDetails(db, true)
	if err == nil {
		log.Debug().Msg("Pushing torq nodes to NodesCache.")
		for _, torqNode := range nodeConnectionDetailsArray {
			publicKey, chain, network, err := GetNodeDetailsById(db, torqNode.NodeId)
			if err == nil {
				cache.SetTorqNode(torqNode.NodeId, torqNode.Name, torqNode.Status, publicKey, chain, network)
			} else {
				log.Error().Err(err).Msg("Failed to obtain torq node for NodesCache.")
			}
		}
	} else {
		log.Error().Err(err).Msg("Failed to obtain torq nodes for NodesCache.")
	}

	log.Debug().Msg("Pushing channel nodes to NodesCache.")
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
		var chain core.Chain
		var network core.Network
		var channelStatus core.ChannelStatus
		err = rows.Scan(&publicKey, &chain, &network, &nodeId, &channelStatus)
		if err != nil {
			return errors.Wrap(err, "Obtaining nodeId and publicKey from the resultSet")
		}
		cache.SetChannelNode(nodeId, publicKey, chain, network, channelStatus)
	}
	return nil
}

func InitializeNodeAliasesCache(db *sqlx.DB) {
	log.Debug().Msg("Pushing node aliases to NodeAliasesCache.")

	torqNodeIds := cache.GetAllTorqNodeIds()
	for _, torqNodeId := range torqNodeIds {
		torqNodeAlias := getNodeAlias(db, torqNodeId)
		if torqNodeAlias != "" {
			cache.SetNodeAlias(torqNodeId, torqNodeAlias)
		}
		for _, channelId := range cache.GetChannelIdsByNodeId(torqNodeId) {
			channelSettings := cache.GetChannelSettingByChannelId(channelId)
			remoteNodeId := channelSettings.FirstNodeId
			if remoteNodeId == torqNodeId {
				remoteNodeId = channelSettings.SecondNodeId
			}
			remoteNodeAlias := getNodeAlias(db, remoteNodeId)
			if remoteNodeAlias != "" {
				cache.SetNodeAlias(remoteNodeId, remoteNodeAlias)
			}
		}
	}
}

func InitializeTaggedCache(db *sqlx.DB) error {
	log.Debug().Msg("Pushing tags to TaggedCache.")
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
		cache.SetTagIdsByNodeId(nodeId, tagIds)
	}
	for channelId, tagIds := range channelTags {
		cache.SetTagIdsByChannelId(channelId, tagIds)
	}
	return nil
}

func InitializeChannelsCache(db *sqlx.DB) error {
	log.Debug().Msg("Pushing channels to ChannelsCache.")
	rows, err := db.Query(`
		SELECT channel_id, short_channel_id, lnd_short_channel_id,
		       funding_transaction_hash, funding_output_index,
		       funding_block_height, funded_on,
		       status_id, capacity, private,
		       first_node_id, second_node_id, initiating_node_id, accepting_node_id,
		       closing_transaction_hash, closing_node_id,
		       closing_block_height, closed_on,
		       flags
		FROM channel;`)
	if err != nil {
		return errors.Wrap(err, "Obtaining channelIds and shortChannelIds")
	}
	defer rows.Close()
	for rows.Next() {
		var channelId int
		var shortChannelId *string
		var lndShortChannelId *uint64
		var fundingTransactionHash string
		var fundingOutputIndex int
		var fundingBlockHeight *uint32
		var fundedOn *time.Time
		var capacity int64
		var private bool
		var firstNodeId int
		var secondNodeId int
		var initiatingNodeId *int
		var acceptingNodeId *int
		var status core.ChannelStatus
		var closingTransactionHash *string
		var closingNodeId *int
		var closingBlockHeight *uint32
		var closedOn *time.Time
		var flags core.ChannelFlags
		err = rows.Scan(&channelId, &shortChannelId, &lndShortChannelId,
			&fundingTransactionHash, &fundingOutputIndex,
			&fundingBlockHeight, &fundedOn,
			&status, &capacity, &private,
			&firstNodeId, &secondNodeId, &initiatingNodeId, &acceptingNodeId,
			&closingTransactionHash, &closingNodeId,
			&closingBlockHeight, &closedOn, &flags)
		if err != nil {
			return errors.Wrap(err, "Obtaining channelId and shortChannelId from the resultSet")
		}
		cache.SetChannel(channelId, shortChannelId, lndShortChannelId, status,
			fundingTransactionHash, fundingOutputIndex, fundingBlockHeight, fundedOn,
			capacity, private, firstNodeId, secondNodeId, initiatingNodeId, acceptingNodeId,
			closingTransactionHash, closingNodeId, closingBlockHeight, closedOn,
			flags)
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
		log.Info().Msgf("Tried to obtain node alias for NodeAliasCache cache with nodeId: %v.", nodeId)
	}
	return alias
}

func setNodeConnectionDetailsStatus(db *sqlx.DB, nodeId int, status core.Status) (int64, error) {
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
	nodeId int,
	pingSystem core.PingSystem,
	status core.Status) (int64, error) {

	var err error
	var res sql.Result
	if status == core.Active {
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

func setNodeConnectionDetailsCustomSettingStatus(db *sqlx.DB,
	nodeId int,
	customSettings core.NodeConnectionDetailCustomSettings,
	status core.Status) (int64, error) {

	connectionDetails := cache.GetLndNodeConnectionDetails(nodeId)
	var err error
	var res sql.Result
	if status == core.Active {
		res, err = db.Exec(`
			UPDATE node_connection_details
			SET custom_settings = custom_settings+$1, updated_on = $2
			WHERE node_id = $3 AND custom_settings%$4 < $5;`,
			customSettings, time.Now().UTC(), nodeId, customSettings*2, customSettings)
		connectionDetails.CustomSettings = connectionDetails.CustomSettings.AddNodeConnectionDetailCustomSettings(customSettings)
	} else {
		res, err = db.Exec(`
			UPDATE node_connection_details
			SET custom_settings = custom_settings-$1, updated_on = $2
			WHERE node_id = $3 AND custom_settings%$4 >= $5;`,
			customSettings, time.Now().UTC(), nodeId, customSettings*2, customSettings)
		connectionDetails.CustomSettings = connectionDetails.CustomSettings.RemoveNodeConnectionDetailCustomSettings(customSettings)
	}
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	cache.SetLndNodeConnectionDetails(nodeId, connectionDetails)
	return rowsAffected, nil
}

func setCustomSettings(db *sqlx.DB,
	nodeId int,
	customSettings core.NodeConnectionDetailCustomSettings,
	pingSystems core.PingSystem) (int64, error) {

	res, err := db.Exec(`
		UPDATE node_connection_details SET custom_settings=$1, ping_system=$2, updated_on=$3 WHERE node_id=$4;`,
		customSettings, pingSystems, time.Now().UTC(), nodeId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	connectionDetails := cache.GetLndNodeConnectionDetails(nodeId)
	connectionDetails.CustomSettings = customSettings
	cache.SetLndNodeConnectionDetails(nodeId, connectionDetails)
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
	if ncd.GRPCAddress != nil && len(ncd.TLSDataBytes) != 0 && len(ncd.MacaroonDataBytes) != 0 {
		cache.SetLndNodeConnectionDetails(ncd.NodeId, cache.LndNodeConnectionDetails{
			GRPCAddress:       *ncd.GRPCAddress,
			TLSFileBytes:      ncd.TLSDataBytes,
			MacaroonFileBytes: ncd.MacaroonDataBytes,
			CustomSettings:    ncd.CustomSettings,
		})
	}
	return ncd, nil
}

func SetNodeConnectionDetailsByConnectionDetails(
	db *sqlx.DB,
	nodeId int,
	status core.Status,
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
	if ncd.GRPCAddress != nil && len(ncd.TLSDataBytes) != 0 && len(ncd.MacaroonDataBytes) != 0 {
		cache.SetLndNodeConnectionDetails(ncd.NodeId, cache.LndNodeConnectionDetails{
			GRPCAddress:       *ncd.GRPCAddress,
			TLSFileBytes:      ncd.TLSDataBytes,
			MacaroonFileBytes: ncd.MacaroonDataBytes,
			CustomSettings:    ncd.CustomSettings,
		})
	}
	return ncd, nil
}
