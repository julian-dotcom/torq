package channel_history

import (
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/pkg/commons"
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

func getChannelEventHistory(db *sqlx.DB, lndShortChannelIds []string, from time.Time, to time.Time) (r []*ChannelEvent, err error) {
	channelIds := make([]int, len(lndShortChannelIds))
	for _, lndShortChannelIdString := range lndShortChannelIds {
		lndShortChannelId, err := strconv.ParseUint(lndShortChannelIdString, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "Converting LND short channel id %v", lndShortChannelId)
		}
		shortChannelId := channels.ConvertLNDShortChannelID(lndShortChannelId)
		channelIds = append(channelIds, commons.GetChannelIdFromShortChannelId(shortChannelId))
	}

	sql := `
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
       		 outbound,
             disabled,
             lag(disabled, 1, false) OVER (PARTITION BY channel_id ORDER BY ts) AS prev
      FROM routing_policy
      where channel_id in ($4)
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
       		 outbound,
             fee_rate_mill_msat as fee_rate,
             lag(fee_rate_mill_msat, 1, 0) OVER (PARTITION BY channel_id ORDER BY ts) AS prev
      FROM routing_policy
      where channel_id in ($4)
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
       		 outbound,
             fee_base_msat as fee_base,
             lag(fee_base_msat, 1, 0) OVER (PARTITION BY channel_id ORDER BY ts) AS prev
      FROM routing_policy
      where channel_id in ($4)
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
       		 outbound,
             max_htlc_msat,
             lag(max_htlc_msat, 1, 0) OVER (PARTITION BY channel_id ORDER BY ts) AS prev
      FROM routing_policy
      where channel_id in ($4)
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
       		 outbound,
             min_htlc,
             lag(min_htlc, 1, 0) OVER (PARTITION BY channel_id ORDER BY ts) AS prev
      FROM routing_policy
      where channel_id in ($4)
        and ts::timestamp AT TIME ZONE ($1) >= ($2)::timestamp
        and ts::timestamp AT TIME ZONE ($1) <= ($3)::timestamp
) as o
where prev  != min_htlc
order by datetime desc;
`
	preferredTimeZone := commons.GetSettings().PreferredTimeZone
	qs, args, err := sqlx.In(sql, preferredTimeZone, from, to, channelIds)
	if err != nil {
		return r, errors.Wrapf(err, "sqlx.In(%s, %v, %v, %v, %v, %v, %v, %v, %v)",
			sql, preferredTimeZone, from, to, channelIds)
	}

	qsr := db.Rebind(qs)
	rows, err := db.Query(qsr, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "Error running getChannelEventHistory query")
	}

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
			return r, err
		}

		// Append to the result
		r = append(r, c)

	}
	return r, nil
}
