package channels

import (
	"github.com/gin-gonic/gin"

	// "gopkg.in/guregu/null.v4"
	"net/http"
	"time"

	"github.com/lncapital/torq/pkg/server_errors"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
)

func getChannelsHandler(c *gin.Context, db *sqlx.DB) {
	from, err := time.Parse("2006-01-02", c.Query("from"))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	to, err := time.Parse("2006-01-02", c.Query("to"))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	r, err := getAggForwardsByChanIds(db, from, to)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

type channelData struct {
	// Database primary key of channel
	ChannelDBID int `json:"channelDbId"`
	// Short channel id in c-lightning / BOLT format
	ShortChannelID string `json:"shortChannelId"`
	// The channel ID
	ChanId uint64 `json:"chan_id"`
	// Alias of remote peer
	Alias string `json:"alias"`
	// The remote public key
	PubKey string `json:"pub_key"`
	// The inbound amount in sats (Satoshis)
	AmountIn uint64 `json:"amount_in"`
	// The inbound revenue in sats. This is what the channel has indirectly produced.
	// This revenue are not really earned by this channel/peer/group, but represents
	// the channel/peer/group contribution to revenue earned by other channels.
	RevenueIn uint64 `json:"revenue_in"`
	// Number of inbound forwards.
	CountIn uint64 `json:"count_in"`
	// The  utbound amount in sats (Satoshis)
	AmountOut uint64 `json:"amount_out"`
	// The outbound revenue in sats. This is what the channel has directly produced.
	RevenueOut uint64 `json:"revenue_out"`
	// Number of outbound forwards.
	CountOut uint64 `json:"count_out"`
	Capacity uint64 `json:"capacity"`
}

func getAggForwardsByChanIds(db *sqlx.DB, fromTime time.Time, toTime time.Time) (r []*channelData, err error) {
	var sql = `
select
    channel.channel_db_id,
	channel.short_channel_id,
    fwr.chan_id,
    coalesce(ne.alias, '') as alias,
    fwr.pub_key,
    fwr.amount_out,
    fwr.amount_in,
    fwr.revenue_out,
    fwr.revenue_in,
    fwr.count_out,
    fwr.count_in,
    fwr.capacity
from (
    select ce.pub_key,
	    ce.chan_point,
        fw.chan_id,
        ce.closed,
        ce.capacity,
        amount_out,
        amount_in,
        revenue_out,
        revenue_in,
        count_out,
        count_in
    from (
        select coalesce(o.chan_id, i.chan_id) as chan_id,
               coalesce(o.amount,0) as amount_out,
               coalesce(o.revenue,0) as revenue_out,
               coalesce(o.count,0) as count_out,
               coalesce(i.amount,0) as amount_in,
               coalesce(i.revenue,0) as revenue_in,
               coalesce(i.count,0) as count_in
        from (
            select outgoing_channel_id chan_id,
                   floor(sum(outgoing_amount_msat)/1000) as amount,
                   floor(sum(fee_msat)/1000) as revenue,
                   count(time) as count
            from forward
            where time >= $1
                and time <= $2
            group by outgoing_channel_id
            ) as o
        full outer join (
            select incoming_channel_id as chan_id,
                   floor(sum(incoming_amount_msat)/1000) as amount,
                   floor(sum(fee_msat)/1000) as revenue,
                   count(time) as count
            from forward
            where time >= $1
                and time <= $2
            group by incoming_channel_id) as i
        on i.chan_id = o.chan_id
    ) as fw
    left join (
        select chan_id,
               pub_key,
		       chan_point,
               last(event->'capacity', time)::numeric as capacity,
               last(event_type, time) = 1 as closed
        from channel_event
        where event_type in (0,1)
        group by chan_id, pub_key, chan_point
    ) as ce
    on ce.chan_id = fw.chan_id
) as fwr
left join (
    select pub_key, last(alias, timestamp) as alias from node_event group by pub_key) as ne
on ne.pub_key = fwr.pub_key
left join channel on channel.channel_point = fwr.chan_point
order by revenue_out desc;
`

	rows, err := db.Query(sql, fromTime, toTime)
	if err != nil {
		return nil, errors.Wrapf(err, "Error running aggregated forwards query")
	}

	for rows.Next() {
		c := &channelData{}
		err = rows.Scan(&c.ChannelDBID,
			&c.ShortChannelID,
			&c.ChanId,
			&c.Alias,
			&c.PubKey,
			&c.AmountOut,
			&c.AmountIn,
			&c.RevenueOut,
			&c.RevenueIn,
			&c.CountOut,
			&c.CountIn,
			&c.Capacity)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, c)

	}

	return r, nil
}
