package channel_history

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/lncapital/torq/pkg/commons"
)

type ChannelHistoryRecords struct {
	Alias string `json:"alias"`

	Date time.Time `json:"date"`
	// The  outbound amount in sats (Satoshis)
	AmountOut *uint64 `json:"amountOut"`
	// The inbound amount in sats (Satoshis)
	AmountIn *uint64 `json:"amountIn"`
	// The total amount in sats (Satoshis) forwarded
	AmountTotal *uint64 `json:"amountTotal"`

	// The outbound revenue in sats. This is what the channel has directly produced.
	RevenueOut *uint64 `json:"revenueOut"`
	// The inbound revenue in sats. This is what the channel has indirectly produced.
	// This revenue are not really earned by this channel/peer/group, but represents
	// the channel/peer/group contribution to revenue earned by other channels.
	RevenueIn *uint64 `json:"revenueIn"`
	// The total revenue in sats. This is what the channel has directly and indirectly produced.
	RevenueTotal *uint64 `json:"revenueTotal"`

	// Number of outbound forwards.
	CountOut *uint64 `json:"countOut"`
	// Number of inbound forwards.
	CountIn *uint64 `json:"countIn"`
	// Number of total forwards.
	CountTotal *uint64 `json:"countTotal"`
}

func getChannelHistory(db *sqlx.DB, all bool, channelIds []int, from time.Time,
	to time.Time) (r []*ChannelHistoryRecords,
	err error) {

	sql := `
		select
		    (coalesce(i.date, o.date)::timestamp AT TIME ZONE ($5)) as date,

			sum(coalesce(i.amount,0)) as amount_in,
			sum(coalesce(o.amount,0)) as amount_out,
			sum(coalesce((coalesce(i.amount,0) + coalesce(o.amount,0)), 0)) as amount_total,
			sum(coalesce(i.revenue,0)) as revenue_in,
			sum(coalesce(o.revenue,0)) as revenue_out,
			sum(coalesce((coalesce(i.revenue,0) + coalesce(o.revenue,0)), 0)) as revenue_total,
			sum(coalesce(i.count,0)) as count_in,
			sum(coalesce(o.count,0)) as count_out,
			sum(coalesce((coalesce(i.count,0) + coalesce(o.count,0)), 0)) as count_total
		from settings, (
			select time_bucket_gapfill('1 days', time::timestamp AT TIME ZONE ($5), $1, $2) as date,
				   outgoing_channel_id channel_id,
				   floor(sum(outgoing_amount_msat)/1000) as amount,
				   floor(sum(fee_msat)/1000) as revenue,
				   count(time) as count
			from forward
			where ($3 or outgoing_channel_id = ANY ($4))
				and time::timestamp AT TIME ZONE ($5) >= $1::timestamp
				and time::timestamp AT TIME ZONE ($5) <= $2::timestamp
			group by date, outgoing_channel_id
			) as o
		full outer join (
			select time_bucket_gapfill('1 days', time::timestamp AT TIME ZONE ($5), $1, $2) as date,
				   incoming_channel_id as channel_id,
				   floor(sum(incoming_amount_msat)/1000) as amount,
				   floor(sum(fee_msat)/1000) as revenue,
				   count(time) as count
			from forward
			where ($3 or incoming_channel_id = ANY ($4))
				and time::timestamp AT TIME ZONE ($5) >= $1::timestamp
				and time::timestamp AT TIME ZONE ($5) <= $2::timestamp
			group by date, incoming_channel_id)  as i
		on (i.channel_id = o.channel_id) and (i.date = o.date)
		group by (coalesce(i.date, o.date)), ($5)
		order by date;
	`

	rows, err := db.Queryx(sql, from, to, all, pq.Array(channelIds), commons.GetSettings().PreferredTimeZone)
	if err != nil {
		return nil, errors.Wrapf(err, "Getting channel history")
	}
	for rows.Next() {
		c := &ChannelHistoryRecords{}
		err = rows.Scan(
			&c.Date,

			&c.AmountIn,
			&c.AmountOut,
			&c.AmountTotal,

			&c.RevenueIn,
			&c.RevenueOut,
			&c.RevenueTotal,

			&c.CountIn,
			&c.CountOut,
			&c.CountTotal,
		)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, c)

	}
	return r, nil
}
