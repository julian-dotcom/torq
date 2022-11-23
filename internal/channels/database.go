package channels

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/commons"
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

func GetRoutingPolicy(chanId uint64, db *sqlx.DB) ([]ChannelPolicy, error) {
	cp := []ChannelPolicy{}
	err := db.Select(&cp, `
	SELECT time_lock_delta, min_htlc, max_htlc_msat, fee_base_msat, fee_rate_mill_msat, short_channel_id, announcing_node_id, connecting_node_id
	FROM routing_policy rp
	LEFT JOIN channel c
	ON rp.channel_id = c.channel_id
	WHERE c.lnd_short_channel_id=$1
	ORDER BY ts DESC
	LIMIT 2;`, chanId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []ChannelPolicy{}, nil
		}
		return []ChannelPolicy{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return cp, nil
}

func GetNodePeerAlias(localNodeId int, remoteNodeId int, db *sqlx.DB) (string, error) {
	var alias string
	err := db.Get(&alias, "SELECT DISTINCT alias FROM node_event WHERE node_id = $1 AND event_node_id = $2 LIMIT 1;", localNodeId, remoteNodeId)
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return "", nil
		}
		return "", errors.Wrap(err, database.SqlExecutionError)
	}
	return alias, nil
}

func GetOpenChannelsForNodeId(db *sqlx.DB, nodeId int) (channels []Channel, err error) {
	err = db.Select(&channels, `
		SELECT *
		FROM channel
		WHERE status_id IN ($1,$2,$3) AND ( first_node_id=$4 OR second_node_id=$4 );`,
		commons.Opening, commons.Open, commons.Closing, nodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return channels, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return channels, nil
}

func InitializeManagedChannelCache(db *sqlx.DB) error {
	log.Debug().Msg("Pushing channels to ManagedChannel cache.")
	rows, err := db.Query(`
		SELECT channel_id, short_channel_id, funding_transaction_hash, funding_output_index, status_id FROM channel;`)
	if err != nil {
		return errors.Wrap(err, "Obtaining channelIds and shortChannelIds")
	}
	for rows.Next() {
		var channelId int
		var shortChannelId *string
		var fundingTransactionHash string
		var fundingOutputIndex int
		var status commons.ChannelStatus
		err = rows.Scan(&channelId, &shortChannelId, &fundingTransactionHash, &fundingOutputIndex, &status)
		if err != nil {
			return errors.Wrap(err, "Obtaining channelId and shortChannelId from the resultSet")
		}
		commons.SetChannel(channelId, shortChannelId, status, fundingTransactionHash, fundingOutputIndex)
	}
	return nil
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
	channel.UpdateOn.Time = channel.CreatedOn
	if channel.ShortChannelID != nil && (*channel.ShortChannelID == "" || *channel.ShortChannelID == "0x0x0") {
		channel.ShortChannelID = nil
	}
	if channel.LNDShortChannelID != nil && *channel.LNDShortChannelID == 0 {
		channel.LNDShortChannelID = nil
	}
	err := db.QueryRowx(`
		INSERT INTO channel (
		  short_channel_id,
		  funding_transaction_hash,
		  funding_output_index,
		  closing_transaction_hash,
		  lnd_short_channel_id,
		  first_node_id,
		  second_node_id,
		  status_id,
		  created_on,
		  updated_on
		) values (
		  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING channel_id;`,
		channel.ShortChannelID, channel.FundingTransactionHash, channel.FundingOutputIndex, channel.ClosingTransactionHash,
		channel.LNDShortChannelID, channel.FirstNodeId, channel.SecondNodeId, channel.Status, channel.CreatedOn,
		channel.UpdateOn).Scan(&channel.ChannelID)
	if err != nil {
		return Channel{}, errors.Wrap(err, database.SqlExecutionError)
	}
	commons.SetChannel(channel.ChannelID, channel.ShortChannelID,
		channel.Status, channel.FundingTransactionHash, channel.FundingOutputIndex)
	return channel, nil
}
