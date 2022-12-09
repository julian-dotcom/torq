package flow

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

type channelFlowData struct {
	// Alias of remote peer
	Alias                  null.String `json:"alias"`
	FundingTransactionHash string      `json:"fundingTransactionHash"`
	FundingOutputIndex     int         `json:"fundingOutputIndex"`
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

func getFlow(db *sqlx.DB, lndShortChannelIdStrings []string, fromTime time.Time,
	toTime time.Time) (r []*channelFlowData,
	err error) {

	var channelIds []int
	var getAll = false
	if len(lndShortChannelIdStrings) == 1 && lndShortChannelIdStrings[0] == "1" {
		// TODO: Clean up Quick hack to simplify logic for fetching all channels
		channelIds = []int{0}
		getAll = true
	} else {
		for _, lndShortChannelIdString := range lndShortChannelIdStrings {
			lndShortChannelId, err := strconv.ParseUint(lndShortChannelIdString, 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err, "Converting LND short channel id from string")
			}
			channelIds = append(channelIds, commons.GetChannelIdByShortChannelId(commons.ConvertLNDShortChannelID(lndShortChannelId)))
		}
	}

	const sql = `
		select
			ne.alias,
			fw.channel_id,
			COALESCE(c.funding_transaction_hash, ''),
			COALESCE(c.funding_output_index, 0),
			n.public_key,

			coalesce(fw.amount_in, 0) as amount_in,
			coalesce(fw.revenue_in, 0) as revenue_in,
			coalesce(fw.count_in, 0) as count_in,

			coalesce(fw.amount_out, 0) as amount_out,
			coalesce(fw.revenue_out, 0) as revenue_out,
			coalesce(fw.count_out, 0) as count_out
		from (
			select
				coalesce(o.outgoing_channel_id, i.incoming_channel_id) as channel_id,
				i.amount as amount_in,
				o.amount as amount_out,
				i.revenue as revenue_in,
				o.revenue as revenue_out,
				i.count as count_in,
				o.count as count_out
			from (
				select
					outgoing_channel_id,
					floor(sum(outgoing_amount_msat)/1000) as amount,
					floor(sum(fee_msat)/1000) as revenue,
					count(time) as count
				from forward
				where time >= $1
					and time <= $2
					and ($3 or incoming_channel_id = ANY($4))
				group by outgoing_channel_id
			) as o
			full outer join (
				select
					incoming_channel_id,
					floor(sum(outgoing_amount_msat)/1000) as amount,
					floor(sum(fee_msat)/1000) as revenue,
					count(time) as count
				from forward
				where time >= $1
					and time <= $2
					and ($3 or outgoing_channel_id = ANY($4))
				group by incoming_channel_id
			) as i
				on o.outgoing_channel_id = i.incoming_channel_id
		) as fw
		left join (
			select
				channel_id,
				last(event->'capacity', time) as capacity,
				(1-last(event_type, time)) as open
			from channel_event
			where event_type in (0,1)
			group by channel_id
		) as ce
			on fw.channel_id = ce.channel_id
		left join channel c on c.channel_id = ce.channel_id
		left join (
			select
				event_node_id,
				last(alias, timestamp) as alias
			from node_event
			group by event_node_id
		) as ne
			on c.second_node_id = ne.event_node_id
		left join node n on ne.event_node_id = n.node_id
	`

	rows, err := db.Queryx(sql, fromTime, toTime, getAll, pq.Array(channelIds))
	if err != nil {
		return nil, errors.Wrapf(err, "Error running flow query")
	}

	for rows.Next() {
		c := &channelFlowData{}
		err = rows.Scan(
			&c.Alias,
			&c.LNDShortChannelId,
			&c.FundingTransactionHash,
			&c.FundingOutputIndex,
			&c.PubKey,

			&c.AmountOut,
			&c.RevenueOut,
			&c.CountOut,

			&c.AmountIn,
			&c.RevenueIn,
			&c.CountIn,
		)
		if err != nil {
			return r, errors.Wrap(err, "SQL row scan")
		}

		// Append to the result
		r = append(r, c)

	}
	return r, nil
}
