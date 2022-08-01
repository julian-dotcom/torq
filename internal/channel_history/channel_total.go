package channel_history

import (
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"time"
)

func getChannelTotal(db *sqlx.DB, chanIds []string, from time.Time, to time.Time) (r ChannelHistory, err error) {

	sql := `
		select
			sum(coalesce(i.amount,0)) as amount_in,
			sum(coalesce(o.amount,0)) as amount_out,
			sum(coalesce((i.amount + o.amount), 0)) as amount_total,

			sum(coalesce(i.revenue,0)) as revenue_in,
			sum(coalesce(o.revenue,0)) as revenue_out,
			sum(coalesce((i.revenue + o.revenue), 0)) as revenue_total,

			sum(coalesce(i.count,0)) as count_in,
			sum(coalesce(o.count,0)) as count_out,
			sum(coalesce((i.count + o.count), 0)) as count_total
		from (
			select lnd_outgoing_short_channel_id,
				   floor(sum(outgoing_amount_msat)/1000) as amount,
				   floor(sum(fee_msat)/1000) as revenue,
				   count(time) as count
			from forward, settings
			where (? or lnd_outgoing_short_channel_id in (?))
			and time >= ?::timestamp
			and time <= ?::timestamp
			group by lnd_outgoing_short_channel_id
			) as o
		full outer join (
			select lnd_incoming_short_channel_id,
				   floor(sum(outgoing_amount_msat)/1000) as amount,
				   floor(sum(fee_msat)/1000) as revenue,
				   count(time) as count
			from forward, settings
			where (? or lnd_incoming_short_channel_id in (?))
			and time >= ?::timestamp
			and time <= ?::timestamp
			group by lnd_incoming_short_channel_id
			) as i
		on (i.lnd_incoming_short_channel_id = o.lnd_outgoing_short_channel_id);
`

	// TODO: Clean up
	// Quick hack to simplify logic for fetching all channels
	var getAll = false
	if chanIds[0] == "1" {
		getAll = true
	}

	qs, args, err := sqlx.In(sql, getAll, chanIds, from, to, getAll, chanIds, from, to)
	if err != nil {
		return r, errors.Wrapf(err, "sqlx.In(%s, %v, %v, %v, %v, %v, %v)", sql, from, to, chanIds, from, to, chanIds)
	}

	qsr := db.Rebind(qs)

	rows, err := db.Query(qsr, args)
	if err != nil {
		return ChannelHistory{}, errors.Wrapf(err, "Error running getChannelTotal query")
	}

	for rows.Next() {
		err = rows.Scan(
			&r.AmountIn,
			&r.AmountOut,
			&r.AmountTotal,

			&r.RevenueIn,
			&r.RevenueOut,
			&r.RevenueTotal,

			&r.CountIn,
			&r.CountOut,
			&r.CountTotal,
		)
		if err != nil {
			return r, err
		}
	}

	return r, nil
}
