package forwards

import (
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

func getForwardsTableHandler(c *gin.Context, db *sqlx.DB) {
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
	network := c.Query("network")
	chain := c.Query("chain")

	r, err := getForwardsTableData(db, commons.GetAllTorqNodeIds(commons.GetChain(chain), commons.GetNetwork(network)), from, to)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

type forwardsTableRow struct {
	// Alias of remote peer
	Alias null.String `json:"alias"`
	// Semicolon seperated list of tagIds
	TagIds       null.String `json:"tagIds"`
	FirstNodeId  int         `json:"firstNodeId"`
	SecondNodeId int         `json:"secondNodeId"`
	// Database primary key of channel
	ChannelID              null.Int `json:"channelId"`
	FundingTransactionHash string   `json:"fundingTransactionHash"`
	FundingOutputIndex     string   `json:"fundingOutputIndex"`
	// The remote public key
	PubKey null.String `json:"pub_key"`
	// Short channel id in c-lightning / BOLT format
	ShortChannelID null.String `json:"shortChannelId"`
	// The channel ID
	LNDShortChannelId null.String `json:"chan_id"`
	// Color of remote peer (Vanity)
	Color null.String `json:"color"`
	// Is the channel open
	StatusId null.Int `json:"statusId"`

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
}

func getForwardsTableData(db *sqlx.DB, nodeIds []int,
	fromTime time.Time, toTime time.Time) (r []*forwardsTableRow, err error) {

	var sqlString = `
		select
			coalesce(scne.node_alias, LEFT(scn.public_key, 20)) as alias,
			coalesce(ct.tag_ids, '') as tag_ids,
			coalesce(c.first_node_id, 0) as first_node_id,
			coalesce(c.second_node_id, 0) as second_node_id,
			coalesce(c.channel_id, 0) as channel_id,
			coalesce(c.funding_transaction_hash, 'Funding tansaction missing') as funding_transaction_hash,
			coalesce(c.funding_output_index, 0) as funding_output_index,
			coalesce(scn.public_key, '') as pub_key,
			coalesce(c.short_channel_id, 'Short channel ID missing') as short_channel_id,
			coalesce(c.lnd_short_channel_id::text, 'LND short channel id missing') as lnd_short_channel_id,
			coalesce(scne.node_color, 'Color missing') as color,
			coalesce(c.status_id, 0) as status_id,


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
			coalesce(round((fw.amount_in + fw.amount_out) / ce.capacity::numeric, 2), 0) as turnover_total

		from channel as c
		left join (
			select channel_id, string_agg(tag_id::text, ';') AS tag_ids
			from channel_tag
			group by channel_id
		) as ct on c.channel_id = ct.channel_id
		left join (
			select channel_id, last(event->'capacity', time) as capacity
			from channel_event
			where event_type in (0,1)
		    group by channel_id
		) as ce on c.channel_id = ce.channel_id
		left join (
			select event_node_id, last(alias, timestamp) as node_alias, last(color, timestamp) as node_color
			from node_event
			group by event_node_id
		) as fcne on c.first_node_id = fcne.event_node_id
		left join (
			select node_id, public_key
			from node
		) as fcn on c.first_node_id = fcn.node_id
		left join (
			select event_node_id, last(alias, timestamp) as node_alias, last(color, timestamp) as node_color
			from node_event
			group by event_node_id
		) as scne on c.second_node_id = scne.event_node_id
		left join (
			select node_id, public_key
			from node
		) as scn on c.second_node_id = scn.node_id
		left join (
			select coalesce(o.channel_id, i.channel_id, 0) as channel_id,
				coalesce(o.amount,0) as amount_out,
				coalesce(o.revenue,0) as revenue_out,
				coalesce(o.count,0) as count_out,
				coalesce(i.amount,0) as amount_in,
				coalesce(i.revenue,0) as revenue_in,
				coalesce(i.count,0) as count_in
			from (
				select outgoing_channel_id channel_id,
					   floor(sum(outgoing_amount_msat)/1000) as amount,
					   floor(sum(fee_msat)/1000) as revenue,
					   count(time) as count
				from forward
				where time::timestamp AT TIME ZONE $3 >= $1::timestamp AT TIME ZONE $3
					and time::timestamp AT TIME ZONE $3 <= $2::timestamp AT TIME ZONE $3
				group by outgoing_channel_id
			) as o
			full outer join (
				select incoming_channel_id as channel_id,
					   floor(sum(incoming_amount_msat)/1000) as amount,
					   floor(sum(fee_msat)/1000) as revenue,
					   count(time) as count
				from forward
				where time::timestamp AT TIME ZONE $3 >= $1::timestamp AT TIME ZONE $3
					and time::timestamp AT TIME ZONE $3 <= $2::timestamp AT TIME ZONE $3
				group by incoming_channel_id
			) as i
			on i.channel_id = o.channel_id
		) as fw on fw.channel_id = c.channel_id
		WHERE ( c.first_node_id = ANY($4) OR c.second_node_id = ANY($4) )
`

	rows, err := db.Queryx(sqlString, fromTime, toTime, commons.GetSettings().PreferredTimeZone, pq.Array(nodeIds))
	if err != nil {
		return nil, errors.Wrapf(err, "Running aggregated forwards query")
	}

	for rows.Next() {
		c := &forwardsTableRow{}
		err = rows.Scan(
			&c.Alias,
			&c.TagIds,
			&c.FirstNodeId,
			&c.SecondNodeId,
			&c.ChannelID,
			&c.FundingTransactionHash,
			&c.FundingOutputIndex,
			&c.PubKey,
			&c.ShortChannelID,
			&c.LNDShortChannelId,
			&c.Color,
			&c.StatusId,

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
		)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, c)

	}

	return r, nil
}
