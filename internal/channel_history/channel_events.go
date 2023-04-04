package channel_history

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/lncapital/torq/pkg/cache"
)

type ChannelEvent struct {
	// The date used by the chart library to place roughly in the timeline. Usually bucketed pr day
	Date time.Time `json:"date"`
	// Specific time the event happened.
	Datetime  time.Time `json:"datetime"`
	ChannelId int       `json:"channelId"`
	// The type of event. E.g. disable/enable, change in fee rate, base fee, min/max htlc amount
	Type *string `json:"type"`
	// Was this changed by our node (outbound) or their node (inbound)
	Outbound *bool `json:"outbound"`
	// The value, in cases where there is a value change,
	//like with fee rate etc. Not used by disable/enable and channel open/close
	Value *uint64 `json:"value"`
	// The previous value
	PreviousValue *uint64 `json:"previousValue"`
}

func getChannelEventHistory(db *sqlx.DB, nodeIds []int, channelIds []int, from time.Time, to time.Time) (r []*ChannelEvent, err error) {

	sql := `select * from (
-- disabled changes
select date(ts)::timestamp AT TIME ZONE ($1) as date,
        ts::timestamp AT TIME ZONE ($1) as datetime,
       channel_id,
       outbound,
       case when disabled = true then 'disabled' else 'enabled' end as type,
       null as value,
 	   null as prev
from (SELECT ts,
             channel_id,
             CASE
                WHEN announcing_node_id = ANY($5) THEN True
            	ELSE False
			 END AS outbound,
             disabled,
             lag(disabled, 1, null) OVER (PARTITION BY channel_id ORDER BY ts) AS prev
      FROM routing_policy
      where channel_id = ANY ($4)
        and ts::timestamp AT TIME ZONE ($1) >= ($2)::timestamp
        and ts::timestamp AT TIME ZONE ($1) <= ($3)::timestamp
) as o
where prev != disabled

UNION
-- fee rate changes
select date(ts)::timestamp AT TIME ZONE ($1) as date,
       ts::timestamp AT TIME ZONE ($1) as datetime,
       channel_id,
       outbound,
       'fee_rate' as type,
       fee_rate as value,
       prev
from (SELECT ts as ts,
       		 channel_id,
             CASE
                WHEN announcing_node_id = ANY($5) THEN True
            	ELSE False
			 END AS outbound,
             fee_rate_mill_msat as fee_rate,
             lag(fee_rate_mill_msat, 1, null) OVER (PARTITION BY channel_id ORDER BY ts) AS prev
      FROM routing_policy
      where channel_id = ANY($4)
        and ts::timestamp AT TIME ZONE ($1) >= ($2)::timestamp
        and ts::timestamp AT TIME ZONE ($1) <= ($3)::timestamp
) as o
where prev != fee_rate

UNION
-- base fee changes
select date(ts)::timestamp AT TIME ZONE ($1) as date,
       ts::timestamp AT TIME ZONE ($1) as datetime,
       channel_id,
       outbound,
       'base_fee' as type,
       round(fee_base / 1000) as value,
       round(prev / 1000) as prev
from (SELECT ts as ts,
             channel_id,
             CASE
                WHEN announcing_node_id = ANY($5) THEN True
            	ELSE False
			 END AS outbound,
             fee_base_msat as fee_base,
             lag(fee_base_msat, 1, null) OVER (PARTITION BY channel_id ORDER BY ts) AS prev
      FROM routing_policy
      where channel_id = ANY($4)
        and ts::timestamp AT TIME ZONE ($1) >= ($2)::timestamp
        and ts::timestamp AT TIME ZONE ($1) <= ($3)::timestamp
) as o
where prev != fee_base

UNION
-- max_htlc changes
select date(ts)::timestamp AT TIME ZONE ($1) as date,
       ts::timestamp AT TIME ZONE ($1) as datetime,
       channel_id,
       outbound,
       'max_htlc' as type,
       round(max_htlc_msat / 1000) as value,
       round(prev / 1000) as prev
from (SELECT ts as ts,
             channel_id,
             CASE
                WHEN announcing_node_id = ANY($5) THEN True
            	ELSE False
			 END AS outbound,
             max_htlc_msat,
             lag(max_htlc_msat, 1, null) OVER (PARTITION BY channel_id ORDER BY ts) AS prev
      FROM routing_policy
      where channel_id = ANY($4)
        and ts::timestamp AT TIME ZONE ($1) >= ($2)::timestamp
        and ts::timestamp AT TIME ZONE ($1) <= ($3)::timestamp
) as o
where prev != max_htlc_msat

UNION
-- min_htlc changes
select date(ts)::timestamp AT TIME ZONE ($1) as date,
       ts::timestamp AT TIME ZONE ($1) as datetime,
       channel_id,
       outbound,
       'min_htlc' as type,
       round(min_htlc / 1000) as value,
       round(prev / 1000) as prev
from (SELECT ts as ts,
             channel_id,
             CASE
                WHEN announcing_node_id = ANY($5) THEN True
            	ELSE False
			 END AS outbound,
             min_htlc,
             lag(min_htlc, 1, null) OVER (PARTITION BY channel_id ORDER BY ts) AS prev
      FROM routing_policy
      where channel_id = ANY($4)
        and ts::timestamp AT TIME ZONE ($1) >= ($2)::timestamp
        and ts::timestamp AT TIME ZONE ($1) <= ($3)::timestamp
) as o
where prev != min_htlc
order by datetime desc) as events where prev is not null and value is not null`

	preferredTimeZone := cache.GetSettings().PreferredTimeZone
	rows, err := db.Queryx(sql, preferredTimeZone, from, to, pq.Array(channelIds), pq.Array(nodeIds))
	if err != nil {
		return r, errors.Wrapf(err, "sqlx.In(%s, %v, %v, %v, %v, %v)",
			sql, preferredTimeZone, from, to, channelIds, nodeIds)
	}
	defer rows.Close()
	for rows.Next() {
		c := &ChannelEvent{}
		err = rows.Scan(
			&c.Date,
			&c.Datetime,
			&c.ChannelId,
			&c.Outbound,
			&c.Type,
			&c.Value,
			&c.PreviousValue,
		)
		if err != nil {
			return r, errors.Wrap(err, "SQL row scan")
		}

		// Append to the result
		r = append(r, c)

	}
	return r, nil
}
