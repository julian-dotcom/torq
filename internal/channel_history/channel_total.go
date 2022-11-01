package channel_history

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func getChannelTotal(db *sqlx.DB, all bool, channelIds []int, from time.Time, to time.Time) (r ChannelHistory, err error) {
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
			select outgoing_channel_id,
				   floor(sum(outgoing_amount_msat)/1000) as amount,
				   floor(sum(fee_msat)/1000) as revenue,
				   count(time) as count
			from forward
			where ($1 or outgoing_channel_id = ANY($2))
			and time >= $3::timestamp
			and time <= $4::timestamp
			group by outgoing_channel_id
			) as o
		full outer join (
			select incoming_channel_id,
				   floor(sum(outgoing_amount_msat)/1000) as amount,
				   floor(sum(fee_msat)/1000) as revenue,
				   count(time) as count
			from forward
			where ($1 or incoming_channel_id = ANY($2))
			and time >= $3::timestamp
			and time <= $4::timestamp
			group by incoming_channel_id
			) as i
		on (i.incoming_channel_id = o.outgoing_channel_id);
`

	rows, err := db.Queryx(sql, all, pq.Array(channelIds), from, to)
	if err != nil {
		return r, errors.Wrapf(err, "db.Queryx(%s, %v, %v, %v, %v)", sql, all, channelIds, from, to)
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
