package channel_tags

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/tags"
)

func getChannelTag(db *sqlx.DB, channelTagId int) (channelTag, error) {
	ct := channelTag{}
	err := db.Get(&ct, "SELECT * FROM channel_tag WHERE channel_tag_id = $1;", channelTagId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return channelTag{}, nil
		}
		return channelTag{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return ct, nil
}

//
//func addChannelTag(db *sqlx.DB, ct channelTag) (channelTag, error) {
//	ct.CreatedOn = time.Now().UTC()
//	// no updated_on because a channel tag is never update only insert/delete.
//	err := db.QueryRowx(`
//		INSERT INTO channel_tag (
//		  from_node_id,
//		  to_node_id,
//		  channel_id,
//		  tag_origin_id,
//		  tag_id,
//		  created_on
//		) values ($1, $2, $3, $4, $5, $6)
//		RETURNING channel_tag_id;`,
//		ct.FromNodeId, ct.ToNodeId, ct.ChannelId, ct.TagOriginId, ct.TagId, ct.CreatedOn).Scan(&ct.ChannelTagId)
//	if err != nil {
//		return channelTag{}, errors.Wrap(err, database.SqlExecutionError)
//	}
//
//	return ct, nil
//}

func GenerateChannelTag(db *sqlx.DB) error {
	chans, err := channels.GetAllChannels(db)
	if err != nil {
		return errors.Wrap(err, "Obtaining channels for channel tag generation.")
	}
	tgs, err := tags.GetTags(db)
	if err != nil {
		return errors.Wrap(err, "Obtaining tags for channel tag generation.")
	}
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, database.SqlBeginTransactionError)
	}

	_, err = tx.Exec("DELETE FROM channel_tag WHERE tag_origin_id=$1;", corridor)
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
		}
		return errors.Wrap(err, database.SqlExecutionError)
	}
	for _, channel := range chans {
		for _, tag := range tgs {
			if corridors.GetBestCorridorStatus(
				corridors.CorridorKey{
					CorridorType: corridors.Tag(), ReferenceId: tag.TagId, FromNodeId: channel.FirstNodeId,
					ToNodeId: channel.SecondNodeId, ChannelId: channel.ChannelID}) {
				_, err = tx.Exec(`
					INSERT INTO channel_tag(from_node_id,to_node_id,channel_id,tag_origin_id,tag_id,created_on)
					VALUES ($1, $2, $3, $4, $5, $6);`,
					channel.FirstNodeId, channel.SecondNodeId, channel.ChannelID, corridor, tag.TagId,
					time.Now().UTC())
				if err != nil {
					if rb := tx.Rollback(); rb != nil {
						log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
					}
					return errors.Wrap(err, database.SqlExecutionError)
				}
			}
			if corridors.GetBestCorridorStatus(
				corridors.CorridorKey{
					CorridorType: corridors.Tag(), ReferenceId: tag.TagId, FromNodeId: channel.SecondNodeId,
					ToNodeId: channel.FirstNodeId, ChannelId: channel.ChannelID}) {
				_, err = tx.Exec(`
					INSERT INTO channel_tag(from_node_id,to_node_id,channel_id,tag_origin_id,tag_id,created_on)
					VALUES ($1, $2, $3, $4, $5, $6);`,
					channel.SecondNodeId, channel.FirstNodeId, channel.ChannelID, corridor, tag.TagId,
					time.Now().UTC())
				if err != nil {
					if rb := tx.Rollback(); rb != nil {
						log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
					}
					return errors.Wrap(err, database.SqlExecutionError)
				}
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, database.SqlCommitTransactionError)
	}
	return nil
}
