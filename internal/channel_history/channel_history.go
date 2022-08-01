package channel_history

import (
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"time"
)

type ChannelHistoryRecords struct {
	Alias string `json:"alias"`

	Date time.Time `json:"date"`
	// The  outbound amount in sats (Satoshis)
	AmountOut *uint64 `json:"amount_out"`
	// The inbound amount in sats (Satoshis)
	AmountIn *uint64 `json:"amount_in"`
	// The total amount in sats (Satoshis) forwarded
	AmountTotal *uint64 `json:"amount_total"`

	// The outbound revenue in sats. This is what the channel has directly produced.
	RevenueOut *uint64 `json:"revenue_out"`
	// The inbound revenue in sats. This is what the channel has indirectly produced.
	// This revenue are not really earned by this channel/peer/group, but represents
	// the channel/peer/group contribution to revenue earned by other channels.
	RevenueIn *uint64 `json:"revenue_in"`
	// The total revenue in sats. This is what the channel has directly and indirectly produced.
	RevenueTotal *uint64 `json:"revenue_total"`

	// Number of outbound forwards.
	CountOut *uint64 `json:"count_out"`
	// Number of inbound forwards.
	CountIn *uint64 `json:"count_in"`
	// Number of total forwards.
	CountTotal *uint64 `json:"count_total"`
}

func getChannelHistory(db *sqlx.DB, chanIds []string, from time.Time,
	to time.Time) (r []*ChannelHistoryRecords,
	err error) {

	sql := `WITH
		fromDate AS (VALUES (?)),
		toDate AS (VALUES (?)),
		allChannels as (VALUES(?)),
		tz AS (select preferred_timezone as tz from settings)
		select
		    (coalesce(i.date, o.date)::timestamp AT TIME ZONE (table tz)) as date,

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
			select time_bucket_gapfill('1 days', time::timestamp AT TIME ZONE (table tz), ?, ?) as date,
				   lnd_outgoing_short_channel_id lnd_short_channel_id,
				   floor(sum(outgoing_amount_msat)/1000) as amount,
				   floor(sum(fee_msat)/1000) as revenue,
				   count(time) as count
			from forward, settings
			where ((table allChannels)::boolean or lnd_outgoing_short_channel_id in (?))
				and time::timestamp AT TIME ZONE (table tz) >= (table fromDate)::timestamp
				and time::timestamp AT TIME ZONE (table tz) <= (table toDate)::timestamp
			group by date, lnd_outgoing_short_channel_id
			) as o
		full outer join (
			select time_bucket_gapfill('1 days', time::timestamp AT TIME ZONE (table tz), ?, ?) as date,
				   lnd_incoming_short_channel_id as lnd_short_channel_id,
				   floor(sum(incoming_amount_msat)/1000) as amount,
				   floor(sum(fee_msat)/1000) as revenue,
				   count(time) as count
			from forward, settings
			where ((table allChannels)::boolean or lnd_incoming_short_channel_id in (?))
				and time::timestamp AT TIME ZONE (table tz) >= (table fromDate)::timestamp
				and time::timestamp AT TIME ZONE (table tz) <= (table toDate)::timestamp
			group by date, lnd_incoming_short_channel_id)  as i
		on (i.lnd_short_channel_id = o.lnd_short_channel_id) and (i.date = o.date)
		group by (coalesce(i.date, o.date)), (table tz)
		order by date;
	`

	// TODO: Clean up
	// Quick hack to simplify logic for fetching flow for all channels
	var getAll = false
	if chanIds[0] == "1" {
		getAll = true
	}

	qs, args, err := sqlx.In(sql, from, to, getAll, from, to, chanIds, from, to, chanIds)
	if err != nil {
		return r, errors.Wrapf(err, "sqlx.In(%s, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v)",
			sql, from, to, getAll, from, to, chanIds, from, to, chanIds)
	}

	qsr := db.Rebind(qs)
	rows, err := db.Query(qsr, args)
	if err != nil {
		return nil, errors.Wrapf(err, "Error running getChannelHistory query")
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
