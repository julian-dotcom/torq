package flow

import (
	"net/http"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"gopkg.in/guregu/null.v4"
)

type channelFlowData struct {
	// Alias of remote peer
	Alias null.String `json:"alias"`
	// The channel point
	LNDChannelPoint null.String `json:"lndChannelPoint"`
	// The remote public key
	PubKey null.String `json:"pubKey"`
	// Short channel id in c-lightning / BOLT format
	LNDShortChannelId null.String `json:"lndShortChannelId"`

	// The  outbound amount in sats (Satoshis)
	AmountOut uint64 `json:"amountOut"`
	// The inbound amount in sats (Satoshis)
	AmountIn uint64 `json:"amountIn"`

	// The outbound revenue in sats. This is what the channel has directly produced.
	RevenueOut uint64 `json:"revenueOut"`
	// The inbound revenue in sats. This is what the channel has indirectly produced.
	// This revenue are not really earned by this channel/peer/group, but represents
	// the channel/peer/group contribution to revenue earned by other channels.
	RevenueIn uint64 `json:"revenueIn"`

	// Number of outbound forwards.
	CountOut uint64 `json:"countOut"`
	// Number of inbound forwards.
	CountIn uint64 `json:"countIn"`
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

	chanIds := strings.Split(c.Query("chanIds"), ",")

	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	r, err := getFlow(db, chanIds, from, to)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func getFlow(db *sqlx.DB, chanIds []string, fromTime time.Time,
	toTime time.Time) (r []*channelFlowData,
	err error) {

	const sql = `
		select
			ne.alias,
			fw.lnd_short_channel_id,
			ce.lnd_channel_point,
			ne.pub_key,

			coalesce(fw.amount_in, 0) as amount_in,
			coalesce(fw.revenue_in, 0) as revenue_in,
			coalesce(fw.count_in, 0) as count_in,

			coalesce(fw.amount_out, 0) as amount_out,
			coalesce(fw.revenue_out, 0) as revenue_out,
			coalesce(fw.count_out, 0) as count_out
		from (
			select
				coalesce(o.lnd_outgoing_short_channel_id, i.lnd_incoming_short_channel_id) as lnd_short_channel_id,
				i.amount as amount_in,
				o.amount as amount_out,
				i.revenue as revenue_in,
				o.revenue as revenue_out,
				i.count as count_in,
				o.count as count_out
				from
						 (select
					lnd_outgoing_short_channel_id,
					floor(sum(outgoing_amount_msat)/1000) as amount,
					floor(sum(fee_msat)/1000) as revenue,
					count(time) as count
				from forward as fw
				where time >= ?
            		and time <= ?
					and ((?) or (lnd_incoming_short_channel_id in (?)))
				group by lnd_outgoing_short_channel_id) as o
				full outer join (
				select
					lnd_incoming_short_channel_id,
					floor(sum(outgoing_amount_msat)/1000) as amount,
					floor(sum(fee_msat)/1000) as revenue,
					count(time) as count
				from forward as fw
				where time >= ?
            		and time <= ?
					and ((?) or (lnd_outgoing_short_channel_id in (?)))
				group by lnd_incoming_short_channel_id) as i on o.lnd_outgoing_short_channel_id = i.lnd_incoming_short_channel_id) as fw
			left join (
			select
				lnd_short_channel_id,
				lnd_channel_point,
				pub_key,
				last(event->'capacity', time) as capacity,
				(1-last(event_type, time)) as open
			from channel_event where event_type in (0,1)
		   group by lnd_short_channel_id, lnd_channel_point, pub_key
		) as ce on fw.lnd_short_channel_id = ce.lnd_short_channel_id
		left join (
			select
				pub_key,
				last(alias, timestamp) as alias
			from node_event
			group by pub_key
		) as ne on ce.pub_key = ne.pub_key
	`

	// TODO: Clean up
	// Quick hack to simplify logic for fetching flow for all channels
	var getAll = false
	if chanIds[0] == "1" {
		getAll = true
	}

	qs, args, err := sqlx.In(sql, fromTime, toTime, getAll, chanIds, fromTime, toTime, getAll, chanIds)
	if err != nil {
		return r, errors.Wrapf(err, "sqlx.In(%s, %v, %v, %v, %v, %v, %v, %v, %v)",
			sql, fromTime, toTime, getAll, chanIds, fromTime, toTime, getAll, chanIds)
	}

	qsr := db.Rebind(qs)
	rows, err := db.Query(qsr, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "Error running flow query")
	}

	for rows.Next() {
		c := &channelFlowData{}
		err = rows.Scan(
			&c.Alias,
			&c.LNDShortChannelId,
			&c.LNDChannelPoint,
			&c.PubKey,

			&c.AmountOut,
			&c.RevenueOut,
			&c.CountOut,

			&c.AmountIn,
			&c.RevenueIn,
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
