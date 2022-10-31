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

func GetChannelsForNodeId(db *sqlx.DB, nodeId int) (channels []Channel, err error) {
	err = db.Select(&channels, `
		SELECT *
		FROM channel
		WHERE status_id IN ($1,$2,$3) AND ( first_node_id=$4 OR second_node_id=$4 );`, Opening, Open, Closing, nodeId)
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
		SELECT channel_id, short_channel_id, lnd_channel_point, status_id
		FROM channel
		WHERE status_id IN ($1,$2);`, Open, Opening)
	if err != nil {
		return errors.Wrap(err, "Obtaining channelIds and shortChannelIds")
	}
	for rows.Next() {
		var channelId int
		var shortChannelId string
		var lndChannelPoint string
		var statusId int
		err = rows.Scan(&channelId, &shortChannelId, &lndChannelPoint, &statusId)
		if err != nil {
			return errors.Wrap(err, "Obtaining channelId and shortChannelId from the resultSet")
		}
		commons.SetChannel(channelId, shortChannelId, statusId, lndChannelPoint)
	}
	return nil
}

func getChannelIdByShortChannelId(db *sqlx.DB, shortChannelId string) (int, error) {
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

func getChannelIdByLndChannelPoint(db *sqlx.DB, lndChannelPoint string) (int, error) {
	var channelId int
	err := db.Get(&channelId, "SELECT channel_id FROM channel WHERE lnd_channel_point = $1 LIMIT 1;", lndChannelPoint)
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
	err := db.QueryRowx(`
		INSERT INTO channel (
		  short_channel_id,
		  lnd_channel_point,
		  lnd_short_channel_id,
		  first_node_id,
		  second_node_id,
		  status_id,
		  created_on,
		  updated_on
		) values (
		  $1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING channel_id;`,
		channel.ShortChannelID, channel.LNDChannelPoint, channel.LNDShortChannelID, channel.FirstNodeId,
		channel.SecondNodeId, channel.Status, channel.CreatedOn, channel.UpdateOn).Scan(&channel.ChannelDBID)
	if err != nil {
		return Channel{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return channel, nil
}
