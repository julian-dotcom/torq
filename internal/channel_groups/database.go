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
	"github.com/lncapital/torq/pkg/commons"
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

func getChannelGroupsByChannelId(db *sqlx.DB, channelId int, include commons.ChannelGroupInclude) ([]commons.ChannelGroup, error) {
	cachedData := commons.GetChannelGroupsByChannelId(channelId, include)
	if cachedData != nil {
		return cachedData.ChannelGroups, nil
	}
	var cgs []commons.ChannelGroup
	err := db.Select(&cgs, `
		SELECT
		    c.category_id, c.name AS category_name, c.style AS category_style,
		    t.tag_id, t.name AS tag_name, t.style AS tag_style
		FROM channel_group cg
		LEFT JOIN category c ON c.category_id=cg.category_id
		LEFT JOIN tag t ON t.tag_id=cg.tag_id
		WHERE cg.channel_id = $1;`, channelId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			commons.SetChannelGroupsByChannelId(channelId, []commons.ChannelGroup{})
			return []commons.ChannelGroup{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	commons.SetChannelGroupsByChannelId(channelId, cgs)
	return commons.GetChannelGroupsByChannelId(channelId, include).ChannelGroups, nil
}

func GenerateChannelGroupsByOrigin(db *sqlx.DB, origin groupOrigin, deleteChannelGroups bool) error {
	chans, err := channels.GetAllChannels(db)
	if err != nil {
		return errors.Wrap(err, "Obtaining channels for channel group generation.")
	}
	tx := db.MustBegin()
	if deleteChannelGroups {
		_, err = tx.Exec("DELETE FROM channel_group WHERE tag_origin_id=$1;", origin)
		if err != nil {
			if rb := tx.Rollback(); rb != nil {
				log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
			}
			return errors.Wrap(err, database.SqlExecutionError)
		}
	}
	if origin == categoryCorridor {
		cts, err := categories.GetCategories(db)
		if err != nil {
			return errors.Wrap(err, "Obtaining categories for channel group generation.")
		}
		for _, channel := range chans {
			for _, category := range cts {
				key := corridors.CorridorKey{
					CorridorType: corridors.Category(), ReferenceId: category.CategoryId, FromNodeId: channel.FirstNodeId,
					ToNodeId: channel.SecondNodeId, ChannelId: channel.ChannelID}
				if corridors.GetBestCorridorStatus(key) {
					_, err = tx.Exec(`
					INSERT INTO channel_group(node_id,channel_id,tag_origin_id,category_id,created_on)
					VALUES ($1, $2, $3, $4, $5);`,
						channel.FirstNodeId, channel.ChannelID, origin, category.CategoryId, time.Now().UTC())
					if err != nil {
						if rb := tx.Rollback(); rb != nil {
							log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
						}
						return errors.Wrap(err, database.SqlExecutionError)
					}
				}
				key = corridors.CorridorKey{
					CorridorType: corridors.Category(), ReferenceId: category.CategoryId, FromNodeId: channel.SecondNodeId,
					ToNodeId: channel.FirstNodeId, ChannelId: channel.ChannelID}
				if corridors.GetBestCorridorStatus(key) {
					_, err = tx.Exec(`
					INSERT INTO channel_group(node_id,channel_id,tag_origin_id,category_id,created_on)
					VALUES ($1, $2, $3, $4, $5);`,
						channel.SecondNodeId, channel.ChannelID, origin, category.CategoryId, time.Now().UTC())
					if err != nil {
						if rb := tx.Rollback(); rb != nil {
							log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
						}
						return errors.Wrap(err, database.SqlExecutionError)
					}
				}
			}
		}
	}
	if origin == tagCorridor {
		tgs, err := tags.GetTags(db)
		if err != nil {
			return errors.Wrap(err, "Obtaining tags for channel group generation.")
		}
		for _, channel := range chans {
			for _, tag := range tgs {
				key := corridors.CorridorKey{
					CorridorType: corridors.Tag(), ReferenceId: tag.TagId, FromNodeId: channel.FirstNodeId,
					ToNodeId: channel.SecondNodeId, ChannelId: channel.ChannelID}
				if tag.CategoryId != nil {
					key.FromCategoryId = *tag.CategoryId
				}
				if corridors.GetBestCorridorStatus(key) {
					_, err = tx.Exec(`
					INSERT INTO channel_group(node_id,channel_id,tag_origin_id,tag_id,category_id,created_on)
					VALUES ($1, $2, $3, $4, $5, $6);`,
						channel.FirstNodeId, channel.ChannelID, origin, tag.TagId, tag.CategoryId,
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
					key.FromCategoryId = *tag.CategoryId
				}
				if corridors.GetBestCorridorStatus(key) {
					_, err = tx.Exec(`
					INSERT INTO channel_group(node_id,channel_id,tag_origin_id,tag_id,category_id,created_on)
					VALUES ($1, $2, $3, $4, $5, $6);`,
						channel.SecondNodeId, channel.ChannelID, origin, tag.TagId, tag.CategoryId,
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

func RemoveChannelGroupByTag(db *sqlx.DB, tagId int) (int64, error) {
	res, err := db.Exec("DELETE FROM channel_group WHERE tag_id = $1;", tagId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
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
	tgs, err := tags.GetTagsByCategoryId(db, categoryId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	if len(tgs) > 0 {
		return 0, errors.New(fmt.Sprintf("Could not remove category since it's in use. %v", tgs))
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
