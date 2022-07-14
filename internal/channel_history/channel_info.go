package channel_history

import (
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type channel struct {
	// Node Alias
	Alias null.String `json:"alias"`
	// Database primary key of channel
	ChannelDBID null.Int `json:"channelDbId"`
	// The channel point
	ChanPoint null.String `json:"channel_point"`
	// The remote public key
	PubKey null.String `json:"pub_key"`
	// Short channel id in c-lightning / BOLT format
	ShortChannelID null.String `json:"shortChannelId"`
	// The channel ID
	ChanId null.String `json:"chan_id"`
	// Is the channel open
	Open null.Bool `json:"open"`

	// The channels total capacity (as created)
	Capacity *uint64 `json:"capacity"`
}

func getChannels(db *sqlx.DB, chanIds []string) (r []*channel, err error) {

	sql := `
		select ne.alias,
		       chan_id,
		       ce.channel_point,
		       ce.pub_key,
		       capacity,
		       open,
		       short_channel_id,
		       channel_db_id
		from (select
				last(chan_id, time) as chan_id,
				last(chan_point, time) as channel_point,
				last(pub_key, time) as pub_key,
				last(event->'capacity', time) as capacity,
				(last(event_type, time)) = 0 as open
			from channel_event
			where event_type in (0,1)
				and (? or chan_id in (?))
			group by chan_id) as ce
		left join channel as c on c.channel_point = ce.channel_point
		left join (
			select pub_key,
			       last(alias, timestamp) as alias
			from node_event
			group by pub_key) as ne on ne.pub_key = ce.pub_key;
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

	rows, err := db.Query(qsr, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "Error running getChannelsByPubkey query")
	}

	for rows.Next() {
		c := &channel{}
		err = rows.Scan(
			&c.Alias,
			&c.ChanId,
			&c.ChanPoint,
			&c.PubKey,
			&c.Capacity,
			&c.Open,
			&c.ShortChannelID,
			&c.ChannelDBID,
		)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, c)
	}
	return r, nil
}
