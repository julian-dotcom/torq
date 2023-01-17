package channel_groups

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/rs/zerolog/log"
)

type channelGroup struct {
	ChannelGroupId int       `json:"channelGroupId" db:"channel_group_id"`
	GroupOriginId  int       `json:"groupOriginId" db:"group_origin_id"`
	NodeId         int       `json:"nodeId" db:"node_id"`
	CategoryId     *int      `json:"categoryId" db:"category_id"`
	TagId          *int      `json:"tagId" db:"tag_id"`
	ChannelId      int       `json:"channelId" db:"channel_id"`
	CreatedOn      time.Time `json:"createdOn" db:"created_on"`
	// No UpdatedOn as there will never be an update always create/delete.
}

type groupOrigin int

const (
	categoryCorridor = groupOrigin(iota)
	tagCorridor
)

func RemoveChannelGroupByTags(db *sqlx.DB, tags []int) (int64, error) {
	origin := tagCorridor
	var total int64

	finished := make(chan bool)
	for _, tag := range tags {
		count, err := corridors.RemoveCorridorByTag(db, tag)
		if err != nil {
			return total, errors.Wrapf(err, "deleting corridors for the tags: (%v %v)", tags, db)
		}
		total += count
	}
	if total > 0 {
		go func() {
			err := GenerateChannelGroupsByOrigin(db, origin, true)
			if err != nil {
				log.Error().Err(err).Msg("Failed to generate channel groups:RemoveChannelGroupByTags.")
			}
			finished <- true
		}()
		<-finished
	}
	return total, nil
}

func AddChannelGroupByTags(db *sqlx.DB, tags []int) error {
	//TODO to remove once we get the real targets
	nodes, err := nodes.GetPeerNodes(db)
	var duplicateCorridors int
	var CorridorsToInsert int
	if err != nil {
		return errors.Wrapf(err, "gettings the peer nodes (%v %v)", tags, db)
	}

	for _, node := range nodes {
		for i := range tags {
			CorridorsToInsert++
			corridor := corridors.Corridor{CorridorTypeId: corridors.Tag().CorridorTypeId, Flag: 1}

			corridor.ReferenceId = &tags[i]
			corridor.FromNodeId = &node.NodeId

			_, err = corridors.AddCorridor(db, corridor)
			if err != nil {
				if err.Error() != "Corridor already exists." {
					return errors.Wrapf(err, "saving the corridor (tag:%v node:%v)", &tags[i], node)
				} else {
					duplicateCorridors++
				}
			}
		}
	}
	if CorridorsToInsert != duplicateCorridors {
		origin := tagCorridor
		finished := make(chan bool)
		go func() {
			err = GenerateChannelGroupsByOrigin(db, origin, true)
			if err != nil {
				log.Error().Err(err).Msg("Failed to generate channel groups:AddChannelGroupByTags.")
			}
			finished <- true
		}()
	}

	return nil
}
