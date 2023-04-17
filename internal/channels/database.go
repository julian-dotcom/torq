package channels

import (
	"database/sql"
	"time"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/tags"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/database"
)

func GetAllChannels(db *sqlx.DB) (channels []Channel, err error) {
	err = db.Select(&channels, "SELECT * FROM channel;")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return channels, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return channels, nil
}

func GetChannels(db *sqlx.DB, nodeIds []int, all bool, channelIds []int) ([]*Channel, error) {
	sql := `SELECT * FROM channel WHERE ($1 OR channel_id = ANY ($2)) AND (first_node_id = ANY($3) OR second_node_id = ANY($3));`
	rows, err := db.Queryx(sql, all, pq.Array(channelIds), pq.Array(nodeIds))
	if err != nil {
		return nil, errors.Wrapf(err, "Running getChannels query all: %v, channelIds: %v", all, channelIds)
	}
	defer rows.Close()
	var r []*Channel
	for rows.Next() {
		c := &Channel{}
		err = rows.StructScan(&c)
		if err != nil {
			return nil, errors.Wrapf(err, "Running getChannels query StructScan all: %v, channelIds: %v", all, channelIds)
		}
		c.Tags = tags.GetTagsByTagIds(cache.GetTagIdsByChannelId(c.SecondNodeId, c.ChannelID))
		r = append(r, c)
	}
	return r, nil
}

func GetChannel(db *sqlx.DB, channelId int) (Channel, error) {
	var c Channel
	err := db.Get(&c, `SELECT * FROM channel WHERE channel_id = $1;`, channelId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Channel{}, nil
		}
		return Channel{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return c, nil
}

func GetLocalRoutingPolicy(channelId, nodeId int, db *sqlx.DB) (ChannelPolicy, error) {
	cp := ChannelPolicy{}
	err := db.Get(&cp, `
    SELECT disabled, time_lock_delta, min_htlc, max_htlc_msat, fee_base_msat, fee_rate_mill_msat, short_channel_id,
		announcing_node_id as node_id, connecting_node_id as remote_node_id
	FROM routing_policy rp
	LEFT JOIN channel c
	ON rp.channel_id = c.channel_id
	WHERE c.channel_id=$1
	AND
	(
		c.first_node_id = $2 AND
		c.first_node_id = rp.announcing_node_id AND
		c.second_node_id = rp.connecting_node_id
	OR
		c.second_node_id = $2 AND
		c.second_node_id = rp.announcing_node_id AND
		c.first_node_id = rp.connecting_node_id
	)
	ORDER BY ts DESC
	LIMIT 1;`, channelId, nodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ChannelPolicy{}, nil
		}
		return ChannelPolicy{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return cp, nil
}

func GetRemoteRoutingPolicy(channelId, nodeId int, db *sqlx.DB) (ChannelPolicy, error) {
	cp := ChannelPolicy{}
	err := db.Get(&cp, `
    SELECT disabled, time_lock_delta, min_htlc, max_htlc_msat, fee_base_msat, fee_rate_mill_msat, short_channel_id,
		announcing_node_id as remote_node_id, connecting_node_id as node_id
	FROM routing_policy rp
	LEFT JOIN channel c
	ON rp.channel_id = c.channel_id
	WHERE c.channel_id=$1
	AND
	(
		c.first_node_id != $2 AND
		c.first_node_id = rp.announcing_node_id AND
		c.second_node_id = rp.connecting_node_id
	OR
		c.second_node_id != $2 AND
		c.second_node_id = rp.announcing_node_id AND
		c.first_node_id = rp.connecting_node_id
	)
	ORDER BY ts DESC
	LIMIT 1;`, channelId, nodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ChannelPolicy{}, nil
		}
		return ChannelPolicy{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return cp, nil
}

func GetOpenChannelsForNodeId(db *sqlx.DB, nodeId int) (channels []Channel, err error) {
	err = db.Select(&channels, `
		SELECT *
		FROM channel
		WHERE status_id IN ($1,$2,$3) AND ( first_node_id=$4 OR second_node_id=$4 );`,
		core.Opening, core.Open, core.Closing, nodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return channels, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return channels, nil
}

func GetChannelsForTag(db *sqlx.DB) ([]ChannelForTag, error) {
	var channels []ChannelForTag
	err := db.Select(&channels, `
		SELECT c.short_channel_id, c.channel_id, c.second_node_id AS node_id, 'channel' AS type
		FROM channel c
		JOIN node_connection_details ncd on ncd.node_id = c.first_node_id
		WHERE ncd.status_id != $1 AND ncd.status_id != $2
		UNION
		SELECT c.short_channel_id, c.channel_id, c.first_node_id AS node_id, 'channel' AS type
		FROM channel c
		JOIN node_connection_details ncd on ncd.node_id = c.second_node_id
		WHERE ncd.status_id != $1 AND ncd.status_id != $2
`, core.Deleted, core.Archived)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return channels, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	var response []ChannelForTag
	for _, channel := range channels {
		alias := cache.GetNodeAlias(channel.NodeId)
		channel.Alias = &alias
		response = append(response, channel)
	}
	return response, nil
}

func GetNodesForTag(db *sqlx.DB) ([]NodeForTag, error) {
	var nodes []NodeForTag
	err := db.Select(&nodes, `
		SELECT DISTINCT c.second_node_id AS node_id, 'node' AS type
		FROM channel c
		JOIN node_connection_details ncd on ncd.node_id = c.first_node_id
		WHERE ncd.status_id != $1 AND ncd.status_id != $2 AND c.status_id=$3
		UNION
		SELECT DISTINCT c.first_node_id AS node_id, 'node' AS type
		FROM channel c
		JOIN node_connection_details ncd on ncd.node_id = c.second_node_id
		WHERE ncd.status_id != $1 AND ncd.status_id != $2 AND c.status_id=$3
;`, core.Deleted, core.Archived, core.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []NodeForTag{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	var response []NodeForTag
	for _, node := range nodes {
		node.Alias = cache.GetNodeAlias(node.NodeId)
		response = append(response, node)
	}
	return response, nil
}

func getChannelIdByShortChannelId(db *sqlx.DB, shortChannelId *string) (int, error) {
	if shortChannelId == nil || *shortChannelId == "" || *shortChannelId == "0x0x0" {
		return 0, nil
	}
	var channelId int
	err := db.Get(&channelId, "SELECT channel_id FROM channel WHERE short_channel_id = $1 LIMIT 1;", shortChannelId)
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return 0, nil
		}
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	return channelId, nil
}

func getChannelIdByFundingTransaction(db *sqlx.DB, fundingTransactionHash string, fundingOutputIndex int) (int, error) {
	var channelId int
	err := db.Get(&channelId, `
		SELECT channel_id
		FROM channel
		WHERE funding_transaction_hash = $1 AND funding_output_index = $2 LIMIT 1;`, fundingTransactionHash, fundingOutputIndex)
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return 0, nil
		}
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	return channelId, nil
}

func addChannel(db *sqlx.DB, channel Channel) (Channel, error) {
	channel.CreatedOn = time.Now().UTC()
	channel.UpdateOn = &channel.CreatedOn
	if channel.ShortChannelID != nil && (*channel.ShortChannelID == "" || *channel.ShortChannelID == "0x0x0") {
		channel.ShortChannelID = nil
	}
	if channel.LNDShortChannelID != nil && *channel.LNDShortChannelID == 0 {
		channel.LNDShortChannelID = nil
	}
	err := db.QueryRowx(`
		INSERT INTO channel (
		  short_channel_id,
		  funding_transaction_hash, funding_output_index, funding_block_height, funded_on,
		  closing_transaction_hash, closing_node_id, closing_block_height, closed_on,
		  lnd_short_channel_id,
		  first_node_id, second_node_id, initiating_node_id, accepting_node_id,
		  capacity, private, status_id, flags,
		  created_on, updated_on
		) VALUES (
		  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		) RETURNING channel_id;`,
		channel.ShortChannelID,
		channel.FundingTransactionHash, channel.FundingOutputIndex, channel.FundingBlockHeight, channel.FundedOn,
		channel.ClosingTransactionHash, channel.ClosingNodeId, channel.ClosingBlockHeight, channel.ClosedOn,
		channel.LNDShortChannelID,
		channel.FirstNodeId, channel.SecondNodeId, channel.InitiatingNodeId, channel.AcceptingNodeId,
		channel.Capacity, channel.Private, channel.Status, channel.Flags,
		channel.CreatedOn, channel.UpdateOn).Scan(&channel.ChannelID)
	if err != nil {
		return Channel{}, errors.Wrap(err, database.SqlExecutionError)
	}
	cache.SetChannel(channel.ChannelID, channel.ShortChannelID, channel.LNDShortChannelID, channel.Status,
		channel.FundingTransactionHash, channel.FundingOutputIndex, channel.FundingBlockHeight, channel.FundedOn,
		channel.Capacity, channel.Private, channel.FirstNodeId, channel.SecondNodeId, channel.InitiatingNodeId, channel.AcceptingNodeId,
		channel.ClosingTransactionHash, channel.ClosingNodeId, channel.ClosingBlockHeight, channel.ClosedOn,
		channel.Flags)
	return channel, nil
}

func getChannelsWithStatus(db *sqlx.DB, network core.Network, status []core.ChannelStatus) ([]Channel, error) {
	var channels []Channel
	bitcoin := core.Bitcoin
	err := db.Select(&channels, `
		SELECT *
		FROM channel
		WHERE (first_node_id = ANY($1) OR second_node_id = ANY($1)) AND status_id = ANY($2)
		`, pq.Array(cache.GetAllActiveTorqNodeIds(&bitcoin, &network)),
		pq.Array(status))
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}

	return channels, nil
}

func updateChannelToClosingByChannelId(db *sqlx.DB, channelId int, closingTransactionHash string) error {
	currentSettings := cache.GetChannelSettingByChannelId(channelId)
	_, err := db.Exec(`
		UPDATE channel
		SET status_id=$1, closing_transaction_hash=$2, updated_on=$3
		WHERE channel_id=$4;`,
		core.Closing, closingTransactionHash, time.Now().UTC(), channelId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	cache.SetChannel(channelId, &currentSettings.ShortChannelId, &currentSettings.LndShortChannelId, core.Closing,
		currentSettings.FundingTransactionHash, currentSettings.FundingOutputIndex,
		currentSettings.FundingBlockHeight, currentSettings.FundedOn,
		currentSettings.Capacity, currentSettings.Private, currentSettings.FirstNodeId, currentSettings.SecondNodeId,
		currentSettings.InitiatingNodeId, currentSettings.AcceptingNodeId,
		&closingTransactionHash, currentSettings.ClosingNodeId,
		currentSettings.ClosingBlockHeight, currentSettings.ClosedOn,
		currentSettings.Flags)
	cache.RemoveChannelStateFromCache(channelId)
	return nil
}
