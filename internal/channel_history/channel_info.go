package channel_history

import (
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"

	"github.com/lncapital/torq/internal/channels"
)

type channel struct {
	FirstNodeId  null.String `json:"firstNodeId" db:"first_node_id"`
	SecondNodeId null.String `json:"secondNodeId" db:"second_node_id"`
	// Database primary key of channel
	ChannelDBID null.Int `json:"channelDbId" db:"channel_id"`
	// The channel point
	LNDChannelPoint null.String `json:"channelPoint" db:"lnd_channel_point"`
	// Short channel id in c-lightning / BOLT format
	ShortChannelID null.String `json:"shortChannelId" db:"short_channel_id"`
	// The channel ID
	LNDShortChannelId null.String     `json:"chanId" db:"lnd_short_channel_id"`
	Status            channels.Status `json:"status" db:"status_id"`

	// TODO FIXME Capacity shouldn't require a JOIN and should be included in channel table
	// Capacity The channels total capacity (as created). Obtained via join from channel_event
	Capacity *uint64 `json:"capacity" db:"capacity"`
}

func getChannels(db *sqlx.DB, chanIds []string) (r []*channel, err error) {

	sql := `
		select ce.capacity, c.*
		from (
			select last(event->'capacity', time) as capacity
			from channel_event ce
			where event_type in (0,1) and ($1 or channel_id in ($1))
			group by channel_id
		) as ce
		join channel as c on c.channel_id = ce.channel_id;
	`

	// TODO: Clean up
	// Quick hack to simplify logic for fetching all channels
	var getAll = false
	if chanIds[0] == "1" {
		getAll = true
	}

	qs, args, err := sqlx.In(sql, getAll, chanIds)
	if err != nil {
		return r, errors.Wrapf(err, "sqlx.In(%s, %v)", sql, chanIds)
	}

	qsr := db.Rebind(qs)

	rows, err := db.Queryx(qsr, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "Running getChannelsByPubkey query")
	}

	for rows.Next() {
		c := &channel{}
		err = rows.StructScan(&c)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, c)
	}
	return r, nil
}
