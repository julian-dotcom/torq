package flow

import (
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"gopkg.in/guregu/null.v4"
	"net/http"
	"strconv"
	"time"
)

type channelFlowData struct {
	// Alias of remote peer
	Alias null.String `json:"alias"`
	// The channel point
	ChanPoint null.String `json:"channel_point"`
	// The remote public key
	PubKey null.String `json:"pub_key"`
	// Short channel id in c-lightning / BOLT format
	ChanId null.String `json:"chan_id"`

	// The  outbound amount in sats (Satoshis)
	AmountOut uint64 `json:"amount_out"`
	// The inbound amount in sats (Satoshis)
	AmountIn uint64 `json:"amount_in"`

	// The outbound revenue in sats. This is what the channel has directly produced.
	RevenueOut uint64 `json:"revenue_out"`
	// The inbound revenue in sats. This is what the channel has indirectly produced.
	// This revenue are not really earned by this channel/peer/group, but represents
	// the channel/peer/group contribution to revenue earned by other channels.
	RevenueIn uint64 `json:"revenue_in"`

	// Number of outbound forwards.
	CountOut uint64 `json:"count_out"`
	// Number of inbound forwards.
	CountIn uint64 `json:"count_in"`
}

func getFlowHandler(c *gin.Context, db *sqlx.DB) {
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
	chanId, err := strconv.ParseUint(c.Query("chan_id"), 10, 64)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	r, err := getFlowByChannelId(db, chanId, from, to)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func getFlowByChannelId(db *sqlx.DB, chanId uint64, fromTime time.Time, toTime time.Time) (r []*channelFlowData,
	err error) {

	const sql = `
		select
			ne.alias,
			fw.chan_id,
			ce.chan_point,
			ne.pub_key,
			coalesce(fw.amount_in, 0) as amount_in,
			coalesce(fw.revenue_in, 0) as revenue_in,
			coalesce(fw.count_in, 0) as count_in,
			coalesce(fw.revenue_out, 0) as revenue_out,
			coalesce(fw.amount_out, 0) as amount_out,
			coalesce(fw.count_out, 0) as count_out
		from (
			select
				coalesce(o.outgoing_channel_id, i.incoming_channel_id) as chan_id,
				i.amount as amount_in,
				o.amount as amount_out,
				i.revenue as revenue_in,
				o.revenue as revenue_out,
				i.count as count_in,
				o.count as count_out
				from
						 (select
					incoming_channel_id,
					outgoing_channel_id,
					floor(sum(outgoing_amount_msat)/1000) as amount,
					floor(sum(fee_msat)/1000) as revenue,
					count(time) as count
				from forward as fw
				where time >= $1
            		and time <= $2
					and incoming_channel_id = $3
				group by incoming_channel_id, outgoing_channel_id) as o
				full outer join (
				select
					incoming_channel_id,
					outgoing_channel_id,
					floor(sum(outgoing_amount_msat)/1000) as amount,
					floor(sum(fee_msat)/1000) as revenue,
					count(time) as count
				from forward as fw
				where time >= $1
            		and time <= $2
					and outgoing_channel_id = $3
				group by incoming_channel_id, outgoing_channel_id
							 ) as i on o.outgoing_channel_id = i.incoming_channel_id) as fw
			left join (
			select
				chan_id,
				chan_point,
				pub_key,
				last(event->'capacity', time) as capacity,
				(1-last(event_type, time)) as open
			from channel_event where event_type in (0,1)
		   group by chan_id, chan_point, pub_key
		) as ce on fw.chan_id = ce.chan_id
		left join (
			select
				pub_key,
				last(alias, timestamp) as alias
			from node_event
			group by pub_key
		) as ne on ce.pub_key = ne.pub_key
	`

	rows, err := db.Query(sql, fromTime, toTime, chanId)
	if err != nil {
		return nil, errors.Wrapf(err, "Error running aggregated forwards query")
	}

	for rows.Next() {
		c := &channelFlowData{}
		err = rows.Scan(
			&c.Alias,
			&c.ChanPoint,
			&c.PubKey,
			&c.ChanId,

			&c.AmountOut,
			&c.AmountIn,

			&c.RevenueOut,
			&c.RevenueIn,
			&c.CountOut,
			&c.CountIn,
		)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, c)

	}
	return r, nil
}
