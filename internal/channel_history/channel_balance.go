package channel_history

import (
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/cache"
)

type Balance struct {
	Date             time.Time `db:"date" json:"date"`
	InboundCapacity  int64     `db:"inbound_capacity" json:"inboundCapacity"`
	OutboundCapacity int64     `db:"outbound_capacity" json:"outboundCapacity"`
	CapacityDiff     *int64    `db:"capacity_diff" json:"capacityDiff"`
}

type ChannelBalance struct {
	ChannelId      string     `json:"channelId"`
	ShortChannelId string     `json:"shortChannelId"`
	Balances       []*Balance `json:"balances"`
}

func getChannelBalance(db *sqlx.DB, channelIdString string, from time.Time, to time.Time) (ChannelBalance, error) {
	channelId, err := strconv.ParseUint(channelIdString, 10, 64)
	if err != nil {
		return ChannelBalance{}, errors.Wrapf(err, "Converting LND short channel id %v", channelIdString)
	}

	cb := ChannelBalance{ChannelId: channelIdString}
	q := `WITH
		   initial_balance as (
				SELECT (event->>'local_balance')::bigint
				FROM channel_event
				WHERE channel_id = $1
				AND event_type = 0
				LIMIT 1
			),
			lnd_channel_id as (
    				select lnd_short_channel_id::text
 						from channel
					where channel_id = $1)
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
				where outgoing_channel_id = $1
				order by time)
				UNION
				(select time,
					   incoming_amount_msat as amt
				from forward
				where incoming_channel_id = $1
				order by time)
				UNION
				(select creation_timestamp as time,
					-(select sum(a) from UNNEST(ARRAY(SELECT jsonb_array_elements_text(jsonb_path_query_array(htlcs, ('$.route[*].hops[0]?(@.chan_id=='|| (table lnd_channel_id) ||').amt_to_forward_msat')::jsonpath)))::numeric[]) as a) amt
				from payment p
				where (status = 'SUCCEEDED')
					and jsonb_path_query_array(htlcs, ('$.route[*].hops[0].chan_id')::jsonpath) @> (table lnd_channel_id)::jsonb
				order by time)
				UNION
				(select settle_date as time,
					   -- We need to fetch the amount paid to a channel using MPP.
				   (select sum(a) from UNNEST(ARRAY(SELECT jsonb_array_elements_text(jsonb_path_query_array(htlcs, ('$?(@.chan_id=='|| (table lnd_channel_id) ||' && @.state==1).amt_msat')::jsonpath)))::numeric[]) as a) amt
				from invoice
				where invoice_state = 'SETTLED'
					and jsonb_path_query_array(htlcs, '$[*].chan_id') @> (table lnd_channel_id)::jsonb
				order by time)
			) a
		) b
	where time::timestamp AT TIME ZONE ($4) between $2::timestamp and $3::timestamp
;`

	rows, err := db.Queryx(q, channelId, from, to, cache.GetSettings().PreferredTimeZone)
	if err != nil {
		return cb, errors.Wrap(err, "SQL run query")
	}
	defer rows.Close()
	for rows.Next() {
		b := Balance{}

		err = rows.StructScan(&b)
		if err != nil {
			return cb, errors.Wrap(err, "SQL struct scan")
		}

		if len(cb.Balances) == 0 {
			cb.Balances = []*Balance{&b}
		} else {
			cb.Balances = append(cb.Balances, &b)
		}

	}

	// To integer
	channelIdInt := int(channelId)
	// Get the short channel id using the cache
	channel := cache.GetChannelSettingByChannelId(channelIdInt)
	if channel.ShortChannelId != nil {
		cb.ShortChannelId = *channel.ShortChannelId
	}

	return cb, nil

}
