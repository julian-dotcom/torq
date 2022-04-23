package channels

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4"

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
	// Alias of remote peer
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
	// Color of remote peer (Vanity)
	Color null.String `json:"color"`
	// Is the channel open
	Open null.Int `json:"open"`

	// The channels total capacity (as created)
	Capacity uint64 `json:"capacity"`

	// The  outbound amount in sats (Satoshis)
	AmountOut uint64 `json:"amount_out"`
	// The inbound amount in sats (Satoshis)
	AmountIn uint64 `json:"amount_in"`
	// The total amount in sats (Satoshis) forwarded
	AmountTotal uint64 `json:"amount_total"`

	// The outbound revenue in sats. This is what the channel has directly produced.
	RevenueOut uint64 `json:"revenue_out"`
	// The inbound revenue in sats. This is what the channel has indirectly produced.
	// This revenue are not really earned by this channel/peer/group, but represents
	// the channel/peer/group contribution to revenue earned by other channels.
	RevenueIn uint64 `json:"revenue_in"`
	// The total revenue in sats. This is what the channel has directly and indirectly produced.
	RevenueTotal uint64 `json:"revenue_total"`

	// Number of outbound forwards.
	CountOut uint64 `json:"count_out"`
	// Number of inbound forwards.
	CountIn uint64 `json:"count_in"`
	// Number of total forwards.
	CountTotal uint64 `json:"count_total"`

	TurnoverOut   float32 `json:"turnover_out"`
	TurnoverIn    float32 `json:"turnover_in"`
	TurnoverTotal float32 `json:"turnover_total"`

	HtlcFailAllIn        null.Int `json:"htlc_fail_all_in"`
	HtlcLinkFailIn       null.Int `json:"htlc_link_fail_in"`
	HtlcForwardFailIn    null.Int `json:"htlc_forward_fail_in"`
	HtlcFailAllOut       null.Int `json:"htlc_fail_all_out"`
	HtlcLinkFailOut      null.Int `json:"htlc_link_fail_out"`
	HtlcForwardFailOut   null.Int `json:"htlc_forward_fail_out"`
	HtlcFailAllTotal     null.Int `json:"htlc_fail_all_total"`
	HtlcLinkFailTotal    null.Int `json:"htlc_link_fail_total"`
	HtlcForwardFailTotal null.Int `json:"htlc_forward_fail_total"`
}

func getAggForwardsByChanIds(db *sqlx.DB, fromTime time.Time, toTime time.Time) (r []*channelData, err error) {
	var sql = `
select
    coalesce(ne.alias, ce.pub_key, '') as alias,
    coalesce(c.channel_db_id, 0) as channel_db_id,
    coalesce(ce.chan_point, 'Channel point missing') as chan_point,
    coalesce(ce.pub_key, 'Public key missing') as pub_key,
    coalesce(c.short_channel_id, 'Short channel ID missing') as short_channel_id,
    coalesce(ce.chan_id::text, 'LND short channel id missing') as chan_id,
    coalesce(ne.color, 'Color missing') as color,
    coalesce(ce.open, 0) as open,


    coalesce(ce.capacity::numeric, 0) as capacity,

    coalesce(fw.amount_out, 0) as amount_out,
    coalesce(fw.amount_in, 0) as amount_in,
    coalesce((fw.amount_in + fw.amount_out), 0) as amount_total,

    coalesce(fw.revenue_out, 0) as revenue_out,
    coalesce(fw.revenue_in, 0) as revenue_in,
    coalesce((fw.revenue_in + fw.revenue_out), 0) as revenue_total,

    coalesce(fw.count_out, 0) as count_out,
    coalesce(fw.count_in, 0) as count_in,
    coalesce((fw.count_in + fw.count_out), 0) as count_total,

    coalesce(round(fw.amount_out / ce.capacity::numeric, 2), 0) as turnover_out,
    coalesce(round(fw.amount_in / ce.capacity::numeric, 2), 0) as turnover_in,
    coalesce(round((fw.amount_in + fw.amount_out) / ce.capacity::numeric, 2), 0) as turnover_total,

    coalesce(htlc.htlc_fail_all_in, 0) as htlc_fail_all_in,
	coalesce(htlc.htlc_Link_Fail_in, 0) as htlc_Link_Fail_in,
	coalesce(htlc.htlc_Forward_Fail_in, 0) as htlc_Forward_Fail_in,
	coalesce(htlc.htlc_fail_all_out, 0) as htlc_fail_all_out,
	coalesce(htlc.htlc_Link_Fail_out, 0) as htlc_Link_Fail_out,
	coalesce(htlc.htlc_Forward_Fail_out, 0) as htlc_Forward_Fail_out,
	coalesce(htlc.htlc_fail_all_total, 0) as htlc_fail_all_total,
	coalesce(htlc.htlc_Link_Fail_total, 0) as htlc_Link_Fail_total,
	coalesce(htlc.htlc_Forward_Fail_total, 0) as htlc_Forward_Fail_total

	-- 	htlc.htlc_fail_all_in / coalesce(fw.count_out, 0) as success

from channel as c
left join (
    select
        chan_id,
        chan_point,
        pub_key,
        last(event->'capacity', time) as capacity,
    	(1-last(event_type, time)) as open
    from channel_event where event_type in (0,1)
   group by chan_id, chan_point, pub_key
) as ce on c.channel_point = ce.chan_point
left join (
    select
        pub_key,
        last(alias, timestamp) as alias,
        last(color, timestamp) as color
    from node_event
    group by pub_key
) as ne on ce.pub_key = ne.pub_key
left join (
    select coalesce(o.chan_id, i.chan_id, 0) as chan_id,
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
) as fw on fw.chan_id = ce.chan_id
left join (
	select
		coalesce(htlc_in.chan_id, htlc_out.chan_id) as chan_id,
		sum(coalesce(htlc_in.htlc_fail_total, 0)) as htlc_fail_all_in,
		sum(coalesce(htlc_in.htlc_Link_Fail_Event, 0)) as htlc_Link_Fail_in,
		sum(coalesce(htlc_in.htlc_Forward_Fail_Event, 0)) as htlc_Forward_Fail_in,
		sum(coalesce(htlc_out.htlc_fail_total, 0)) as htlc_fail_all_out,
		sum(coalesce(htlc_out.htlc_Link_Fail_Event, 0)) as htlc_Link_Fail_out,
		sum(coalesce(htlc_out.htlc_Forward_Fail_Event, 0)) as htlc_Forward_Fail_out,
		sum(coalesce(htlc_out.htlc_fail_total + htlc_in.htlc_fail_total, 0)) as htlc_fail_all_total,
		sum(coalesce(htlc_out.htlc_Link_Fail_Event + htlc_in.htlc_Link_Fail_Event, 0)) as htlc_Link_Fail_total,
		sum(coalesce(htlc_out.htlc_Forward_Fail_Event + htlc_in.htlc_Forward_Fail_Event, 0)) as htlc_Forward_Fail_total
	from (
		select f.outgoing_channel_id as chan_id,
			   count(f.event_type) filter (where f.event_type = 'ForwardFailEvent') as htlc_Forward_Fail_Event,
			   count(f.event_type) filter (where f.event_type = 'LinkFailEvent') as htlc_Link_Fail_Event,
			   count(f.event_type) as htlc_fail_total
		from (
			select outgoing_channel_id,
				   (jsonb_object_keys(event->'Event')::text) as event_type

			from htlc_event
			where time >= $1
            	and time <= $2
		) as f
		where f.event_type not in ('ForwardEvent', 'SettleEvent')
		group by outgoing_channel_id, f.event_type) as htlc_out
	full outer join (
		select f.incoming_channel_id as chan_id,
			   count(f.event_type) filter (where f.event_type = 'ForwardFailEvent') as htlc_Forward_Fail_Event,
			   count(f.event_type) filter (where f.event_type = 'LinkFailEvent') as htlc_Link_Fail_Event,
			   count(f.event_type) as htlc_fail_total
		from (
			select incoming_channel_id,
				   (jsonb_object_keys(event->'Event')::text) as event_type
			from htlc_event
			where time >= $1
            	and time <= $2
		) as f
		where f.event_type not in ('ForwardEvent', 'SettleEvent')
		group by incoming_channel_id, f.event_type
	) as htlc_in on htlc_out.chan_id = htlc_in.chan_id
	where coalesce(htlc_in.chan_id, htlc_out.chan_id) != 0
	group by coalesce(htlc_in.chan_id, htlc_out.chan_id)
) as htlc on htlc.chan_id = ce.chan_id
`

	rows, err := db.Query(sql, fromTime, toTime)
	if err != nil {
		return nil, errors.Wrapf(err, "Error running aggregated forwards query")
	}

	for rows.Next() {
		c := &channelData{}
		err = rows.Scan(
			&c.Alias,
			&c.ChannelDBID,
			&c.ChanPoint,
			&c.PubKey,
			&c.ShortChannelID,
			&c.ChanId,
			&c.Color,
			&c.Open,

			&c.Capacity,

			&c.AmountOut,
			&c.AmountIn,
			&c.AmountTotal,

			&c.RevenueOut,
			&c.RevenueIn,
			&c.RevenueTotal,

			&c.CountOut,
			&c.CountIn,
			&c.CountTotal,

			&c.TurnoverOut,
			&c.TurnoverIn,
			&c.TurnoverTotal,

			&c.HtlcFailAllIn,
			&c.HtlcLinkFailIn,
			&c.HtlcForwardFailIn,
			&c.HtlcFailAllOut,
			&c.HtlcLinkFailOut,
			&c.HtlcForwardFailOut,
			&c.HtlcFailAllTotal,
			&c.HtlcLinkFailTotal,
			&c.HtlcForwardFailTotal,
		)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, c)

	}

	return r, nil
}
