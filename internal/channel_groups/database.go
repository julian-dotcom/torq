package channel_groups

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/categories"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/tags"
)

func getChannelGroup(db *sqlx.DB, channelGroupId int) (channelGroup, error) {
	cg := channelGroup{}
	err := db.Get(&cg, "SELECT * FROM channel_group WHERE channel_group_id = $1;", channelGroupId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return channelGroup{}, nil
		}
		return channelGroup{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return cg, nil
}

func GenerateChannelGroups(db *sqlx.DB) error {
	chans, err := channels.GetAllChannels(db)
	if err != nil {
		return errors.Wrap(err, "Obtaining channels for channel group generation.")
	}
	tgs, err := tags.GetTags(db)
	if err != nil {
		return errors.Wrap(err, "Obtaining tags for channel group generation.")
	}
	cts, err := categories.GetCategories(db)
	if err != nil {
		return errors.Wrap(err, "Obtaining categories for channel group generation.")
	}
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, database.SqlBeginTransactionError)
	}

	_, err = tx.Exec("DELETE FROM channel_group WHERE tag_origin_id=$1;", corridorOrigin)
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
		}
		return errors.Wrap(err, database.SqlExecutionError)
	}
	for _, channel := range chans {
		for _, category := range cts {
			if corridors.GetBestCorridorStatus(
				corridors.CorridorKey{
					CorridorType: corridors.Category(), ReferenceId: category.CategoryId, FromNodeId: channel.FirstNodeId,
					ToNodeId: channel.SecondNodeId, ChannelId: channel.ChannelID}) {
				_, err = tx.Exec(`
					INSERT INTO channel_tag(from_node_id,to_node_id,channel_id,tag_origin_id,tag_id,created_on)
					VALUES ($1, $2, $3, $4, $5, $6);`,
					channel.FirstNodeId, channel.SecondNodeId, channel.ChannelID, corridorOrigin, tag.TagId,
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
					CorridorType: corridors.Category(), ReferenceId: category.CategoryId, FromNodeId: channel.SecondNodeId,
					ToNodeId: channel.FirstNodeId, ChannelId: channel.ChannelID}) {
				_, err = tx.Exec(`
					INSERT INTO channel_tag(from_node_id,to_node_id,channel_id,tag_origin_id,tag_id,created_on)
					VALUES ($1, $2, $3, $4, $5, $6);`,
					channel.SecondNodeId, channel.FirstNodeId, channel.ChannelID, corridorOrigin, tag.TagId,
					time.Now().UTC())
				if err != nil {
					if rb := tx.Rollback(); rb != nil {
						log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
					}
					return errors.Wrap(err, database.SqlExecutionError)
				}
			}
		}
		for _, tag := range tgs {
			key := corridors.CorridorKey{
				CorridorType: corridors.Tag(), ReferenceId: tag.TagId, FromNodeId: channel.FirstNodeId,
				ToNodeId: channel.SecondNodeId, ChannelId: channel.ChannelID}
			if tag.CategoryId != nil {
				key.FromCategoryId = tag.CategoryId
			}
			if corridors.GetBestCorridorStatus(key) {
				_, err = tx.Exec(`
					INSERT INTO channel_group(from_node_id,to_node_id,channel_id,tag_origin_id,tag_id,created_on)
					VALUES ($1, $2, $3, $4, $5, $6);`,
					channel.FirstNodeId, channel.SecondNodeId, channel.ChannelID, corridorOrigin, tag.TagId,
					time.Now().UTC())
				if err != nil {
					if rb := tx.Rollback(); rb != nil {
						log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
					}
					return errors.Wrap(err, database.SqlExecutionError)
				}
			}
			key = corridors.CorridorKey{
				CorridorType: corridors.Tag(), ReferenceId: tag.TagId, FromNodeId: channel.SecondNodeId,
				ToNodeId: channel.FirstNodeId, ChannelId: channel.ChannelID}
			if tag.CategoryId != nil {
				key.FromCategoryId = tag.CategoryId
			}
			if corridors.GetBestCorridorStatus(key) {
				_, err = tx.Exec(`
					INSERT INTO channel_group(from_node_id,to_node_id,channel_id,tag_origin_id,tag_id,created_on)
					VALUES ($1, $2, $3, $4, $5, $6);`,
					channel.SecondNodeId, channel.FirstNodeId, channel.ChannelID, corridorOrigin, tag.TagId,
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

func removeTag(db *sqlx.DB, tagId int) (int64, error) {
	referencingCorridors, err := corridors.GetCorridorsReferencingTag(db, tagId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	if len(referencingCorridors) > 0 {
		return 0, errors.New(fmt.Sprintf("Could not remove tag since it's in use. %v", referencingCorridors))
	}
	res, err := db.Exec(`DELETE FROM tag WHERE tag_id = $1;`, tagId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}

func removeCategory(db *sqlx.DB, categoryId int) (int64, error) {
	if categoryId < 0 {
		return 0, errors.New("Could not remove system category.")
	}
	referencingCorridors, err := corridors.GetCorridorsReferencingCategory(db, categoryId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	if len(referencingCorridors) > 0 {
		return 0, errors.New(fmt.Sprintf("Could not remove category since it's in use. %v", referencingCorridors))
	}
	tags, err := tags.GetTagsByCategoryId(db, categoryId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	if len(tags) > 0 {
		return 0, errors.New(fmt.Sprintf("Could not remove category since it's in use. %v", tags))
	}
	channelCategories, err := GetChannelCategories(db, categoryId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	if len(channelCategories) > 0 {
		return 0, errors.New(fmt.Sprintf("Could not remove category since it's in use. %v", channelCategories))
	}
	res, err := db.Exec(`DELETE FROM category WHERE category_id = $1;`, categoryId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}
