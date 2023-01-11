package graph_events

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/database"
)

func GetChannelEventFromGraph(db *sqlx.DB, channelId int, secondsAgo *int) ([]ChannelEventFromGraph, error) {
	var channelEventsFromGraph []ChannelEventFromGraph
	var err error
	if secondsAgo == nil {
		err = db.Select(&channelEventsFromGraph,
			`SELECT *
			FROM routing_policy
			WHERE channel_id=$1
			ORDER BY ts DESC;`,
			channelId)
	} else {
		err = db.Select(&channelEventsFromGraph,
			`SELECT *
			FROM routing_policy
			WHERE channel_id=$1 AND ts >= $2
			ORDER BY ts DESC;`,
			channelId, time.Now().UTC().Add(-time.Duration(*secondsAgo)*time.Second))
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []ChannelEventFromGraph{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return channelEventsFromGraph, nil
}
