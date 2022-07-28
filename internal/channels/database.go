package channels

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

func getChannel(db *sqlx.DB, shortChannelId string) (channel *Channel, err error) {
	channel = &Channel{}
	err = db.Get(channel, "SELECT * FROM channel WHERE short_channel_id = $1 LIMIT 1;", shortChannelId)
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "Unable to execute SQL query")
	}
	return channel, nil
}

func insertChannel(db *sqlx.DB, channel Channel) error {
	_, err := db.Exec(`
INSERT INTO channel (
  short_channel_id,
  channel_point,
  alias,
  destination_pub_key,
  local_node_id,
  created_on
) values (
  $1, $2, $3, $4, $5, $6
);`, channel.ShortChannelID, channel.ChannelPoint, channel.Alias, channel.DestinationPubKey, channel.LocalNodeId, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "Unable to execute SQL statement")
	}
	return nil
}
