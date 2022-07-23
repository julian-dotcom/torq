package channel_history

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type Balance struct {
	Date             time.Time `db:"date" json:"date"`
	InboundCapacity  int64     `db:"inbound_capacity" json:"inbound_capacity"`
	OutboundCapacity int64     `db:"outbound_capacity" json:"outbound_capacity"`
	CapacityDiff     *int64    `db:"capacity_diff" json:"capacity_diff"`
}

type ChannelBalance struct {
	ChanId   string
	Balances []*Balance
}

func getChannelBalance(db *sqlx.DB, chanId string, from time.Time, to time.Time) (ChannelBalance, error) {
	cb := ChannelBalance{ChanId: chanId}
	q := `WITH
    		tz AS (select preferred_timezone from settings),
		   initial_balance as (
				select coalesce((-amount)-total_fees, 0) as initial_balance
				from channel_event
				left join tx on split_part(chan_point, ':', 1) = tx_hash
				where event_type in (0,1) and
					  chan_id = $1
				limit 1
			),
			chan_id as (select $1::text)
		select time as date,
		       outbound_capacity,
		       outbound_capacity - lag(outbound_capacity) over (order by time) as capacity_diff
		from (
			select time,
			       floor((table initial_balance) + sum(amt/1000) over(order by time)) as outbound_capacity
			from (
				(select time,
				   -outgoing_amount_msat as amt
				from forward
				where outgoing_channel_id = ((table chan_id)::numeric)
				order by time)
				UNION
				(select time,
					   incoming_amount_msat as amt
				from forward
				where incoming_channel_id = ((table chan_id)::numeric)
				order by time)
				UNION
				(select creation_timestamp as time,
					-(value_msat + fee_msat) as amt
				from payment p
				where status = 'SUCCEEDED'
				and (htlcs->-1->'route'->'hops'->0->>'chan_id')::text = (table chan_id)::text
				order by time)
				UNION
				(select settle_date as time,
					   -- We need to fetch the amount paid to a channel using MPP.
				   (select sum(a) from UNNEST(ARRAY(SELECT jsonb_array_elements_text(jsonb_path_query_array(htlcs, ('$?(@.chan_id=='|| (table chan_id)::text ||' && @.state==1).amt_msat')::jsonpath)))::numeric[]) as a) amt
				from invoice
				where invoice_state = 'SETTLED'
					and jsonb_path_query_array(htlcs, '$[*].chan_id') @> (table chan_id)::jsonb
				order by time)
			) a
		) b
	where time::timestamp AT TIME ZONE (table tz) between $2 and $3
;`

	rows, err := db.Queryx(q, chanId, from, to)
	if err != nil {
		return cb, err
	}
	for rows.Next() {
		b := Balance{}

		err = rows.StructScan(&b)
		if err != nil {
			return cb, err
		}

		if len(cb.Balances) == 0 {
			cb.Balances = []*Balance{&b}
		} else {
			cb.Balances = append(cb.Balances, &b)
		}

	}

	return cb, nil

}
