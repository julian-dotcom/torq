package channel_history

import (
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"gopkg.in/guregu/null.v4"
	"net/http"
	"strings"
	"time"
)

func getChannelHistoryHandler(c *gin.Context, db *sqlx.DB) {
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

	chanIds := strings.Split(c.Param("chanIds"), ",")

	chanHistory, err := getChannelHistory(db, chanIds, from, to)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	chanEventHistory, err := getChannelEventHistory(db, chanIds, from, to)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	channels, err := getChannels(db, chanIds)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	r := ChannelHistory{
		Label:    "No Label",
		Channels: channels,
		Data:     chanHistory,
		Events:   chanEventHistory,
	}

	c.JSON(http.StatusOK, r)
}

type channel struct {
	// Node Alias
	Alias null.String `json:"alias"`
	// Database primary key of channel
	ChannelDBID null.Int `json:"channelDbId"`
	// The channel point
	ChanPoint null.String `json:"channel_point"`
	// The remote public key
	PubKey null.String `json:"pub_key"`
	// Short channel id in c-lightning / BOLT format
	ShortChannelID null.String `json:"shortChannelId"`
	// The channel ID
	ChanId null.String `json:"chan_id"`
	// Is the channel open
	Open null.Bool `json:"open"`

	// The channels total capacity (as created)
	Capacity uint64 `json:"capacity"`
}

func getChannels(db *sqlx.DB, chanIds []string) (r []*channel, err error) {

	sql := `
		select ne.alias,
		       chan_id,
		       ce.channel_point,
		       ce.pub_key,
		       capacity,
		       open,
		       short_channel_id,
		       channel_db_id
		from (select
				last(chan_id, time) as chan_id,
				last(chan_point, time) as channel_point,
				last(pub_key, time) as pub_key,
				last(event->'capacity', time) as capacity,
				(last(event_type, time)) = 0 as open
			from channel_event
			where event_type in (0,1)
				and chan_id in (?)
			group by chan_id) as ce
		left join channel as c on c.channel_point = ce.channel_point
		left join (
			select pub_key,
			       last(alias, timestamp) as alias
			from node_event
			group by pub_key) as ne on ne.pub_key = ce.pub_key;
	`
	qs, args, err := sqlx.In(sql, chanIds)
	if err != nil {
		return r, errors.Wrapf(err, "sqlx.In(%s, %v)", sql, chanIds)
	}

	qsr := db.Rebind(qs)

	rows, err := db.Query(qsr, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "Error running getChannelsByPubkey query")
	}

	for rows.Next() {
		c := &channel{}
		err = rows.Scan(
			&c.Alias,
			&c.ChanId,
			&c.ChanPoint,
			&c.PubKey,
			&c.Capacity,
			&c.Open,
			&c.ShortChannelID,
			&c.ChannelDBID,
		)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, c)

	}
	return r, nil
}

type ChannelHistory struct {
	// The Label of the requested channels group,
	// this is either an alias in the case where a single channel or a single node is requested.
	// In the case where a group of channels is requested the Label will be based on the common name,
	// such as a tag.
	Label string `json:"label"`

	// A list of channels included in this response
	Channels []*channel      `json:"channels"`
	Data     []*ChannelData  `json:"data"`
	Events   []*ChannelEvent `json:"events"`
}

type ChannelData struct {
	Alias string `json:"alias"`

	Date time.Time `json:"date"`
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
}

func getChannelHistory(db *sqlx.DB, chanIds []string, from time.Time, to time.Time) (r []*ChannelData, err error) {

	sql := `
		select
		    (coalesce(i.date, o.date)::timestamp AT TIME ZONE settings.preferred_timezone ) as date,

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
			select time_bucket_gapfill('1 days', time, ?, ?) as date,
				   outgoing_channel_id chan_id,
				   floor(sum(outgoing_amount_msat)/1000) as amount,
				   floor(sum(fee_msat)/1000) as revenue,
				   count(time) as count
			from forward, settings
			where outgoing_channel_id in (?)
				and time >= ?::timestamp AT TIME ZONE settings.preferred_timezone
				and time <= ?::timestamp AT TIME ZONE settings.preferred_timezone
			group by date, outgoing_channel_id
			) as o
		full outer join (
			select time_bucket_gapfill('1 days', time, ? , ?) as date,
				   incoming_channel_id as chan_id,
				   floor(sum(incoming_amount_msat)/1000) as amount,
				   floor(sum(fee_msat)/1000) as revenue,
				   count(time) as count
			from forward, settings
			where incoming_channel_id in (?)
				and time >= ?::timestamp AT TIME ZONE settings.preferred_timezone
				and time <= ?::timestamp AT TIME ZONE settings.preferred_timezone
			group by date, incoming_channel_id) as i
		on (i.chan_id = o.chan_id) and (i.date = o.date)
		group by (coalesce(i.date, o.date)), settings.preferred_timezone
		order by date;
	`

	qs, args, err := sqlx.In(sql, from, to, chanIds, from, to, from, to, chanIds, from, to)
	if err != nil {
		return r, errors.Wrapf(err, "sqlx.In(%s, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v)",
			sql, from, to, chanIds, from, to, from, to, chanIds, from, to)
	}

	qsr := db.Rebind(qs)
	rows, err := db.Query(qsr, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "Error running getChannelHistory query")
	}

	for rows.Next() {
		c := &ChannelData{}
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

type ChannelEvent struct {
	// The date used by the chart library to place roughly in the timeline. Usually bucketed pr day
	Date time.Time `json:"date"`
	// Spesific time the event happened.
	Datetime time.Time `json:"datetime"`
	// The channel point
	ChanPoint *string `json:"channel_point"`
	// The channel ID
	ChanId *string `json:"chan_id"`
	// The type of event. E.g. disable/enable, change in fee rate, base fee, min/max htlc amount
	Type *string `json:"type"`
	// Was this changed by our node (outbound) or their node (inbound)
	Outbound *bool `json:"outbound"`
	// The value, in cases where there is a value change,
	//like with fee rate etc. Not used by disable/enable and channel open/close
	Value *uint64 `json:"value"`
	// The previous value
	PreviousValue *uint64 `json:"previous_value"`
}

func getChannelEventHistory(db *sqlx.DB, chanIds []string, from time.Time, to time.Time) (r []*ChannelEvent, err error) {

	sql := `WITH
    fromDate AS (VALUES (?)),
    toDate AS (VALUES (?))
-- disabled changes
select ts::DATE as date,
        ts as datetime,
        chan_point,
	   chan_id,
       outbound,
       case when disabled = true then 'disabled' else 'enabled' end as type,
       null as value,
 	   null as prev
from (SELECT ts::timestamp AT TIME ZONE settings.preferred_timezone as ts,
                         chan_point,
             chan_id,
             outbound,
             disabled,
             lag(disabled, 1, false) OVER (PARTITION BY chan_id, outbound ORDER BY ts) AS prev
      FROM routing_policy, settings
      where chan_id in (?)
        and ts >= (table fromDate)::timestamp AT TIME ZONE settings.preferred_timezone
        and ts <= (table toDate)::timestamp AT TIME ZONE settings.preferred_timezone
) as o
where prev != disabled

UNION
-- fee rate changes
select ts::DATE as date,
       ts as datetime,
       chan_point,
       chan_id,
       outbound,
       'fee_rate' as type,
       fee_rate as value,
       prev
from (SELECT ts::timestamp AT TIME ZONE settings.preferred_timezone as ts,
             chan_point,
             chan_id,
             outbound,
             fee_rate_mill_msat as fee_rate,
             lag(fee_rate_mill_msat, 1, 0) OVER (PARTITION BY chan_id, outbound ORDER BY ts) AS prev
      FROM routing_policy, settings
      where chan_id in (?)
        and ts >= (table fromDate)::timestamp AT TIME ZONE settings.preferred_timezone
        and ts <= (table toDate)::timestamp AT TIME ZONE settings.preferred_timezone
) as o
where prev != fee_rate

UNION
-- base fee changes
select ts::DATE as date,
       ts as datetime,
       chan_point,
       chan_id,
       outbound,
       'base_fee' as type,
       round(fee_base / 1000) as value,
       round(prev / 1000) as prev
from (SELECT ts::timestamp AT TIME ZONE settings.preferred_timezone as ts,
             chan_point,
             chan_id,
             outbound,
             fee_base_msat as fee_base,
             lag(fee_base_msat, 1, 0) OVER (PARTITION BY chan_id, outbound ORDER BY ts) AS prev
      FROM routing_policy, settings
      where chan_id in (?)
        and ts >= (table fromDate)::timestamp AT TIME ZONE settings.preferred_timezone
        and ts <= (table toDate)::timestamp AT TIME ZONE settings.preferred_timezone
) as o
where prev != fee_base

UNION
-- max_htlc changes
select ts::DATE as date,
       ts as datetime,
       chan_point,
       chan_id,
       outbound,
       'max_htlc' as type,
       round(max_htlc_msat / 1000) as value,
       round(prev / 1000) as prev
from (SELECT ts::timestamp AT TIME ZONE settings.preferred_timezone as ts,
             chan_point,
             chan_id,
             outbound,
             max_htlc_msat,
             lag(max_htlc_msat, 1, 0) OVER (PARTITION BY chan_id, outbound ORDER BY ts) AS prev
      FROM routing_policy, settings
      where chan_id in (?)
        and ts >= (table fromDate)::timestamp AT TIME ZONE settings.preferred_timezone
        and ts <= (table toDate)::timestamp AT TIME ZONE settings.preferred_timezone
) as o
where prev != max_htlc_msat

UNION
-- min_htlc changes
select ts::DATE as date,
       ts as datetime,
       chan_point,
       chan_id,
       outbound,
       'min_htlc' as type,
       round(min_htlc / 1000) as value,
       round(prev / 1000) as prev
from (SELECT ts::timestamp AT TIME ZONE settings.preferred_timezone as ts,
             chan_point,
             chan_id,
             outbound,
             min_htlc,
             lag(min_htlc, 1, 0) OVER (PARTITION BY chan_id, outbound ORDER BY ts) AS prev
      FROM routing_policy, settings
      where chan_id in (?)
        and ts >= (table fromDate)::timestamp AT TIME ZONE settings.preferred_timezone
        and ts <= (table toDate)::timestamp AT TIME ZONE settings.preferred_timezone
) as o
where prev  != min_htlc
order by datetime;
`

	qs, args, err := sqlx.In(sql, from, to, chanIds, chanIds, chanIds, chanIds, chanIds)
	if err != nil {
		return r, errors.Wrapf(err, "sqlx.In(%s, %v, %v, %v, %v, %v, %v, %v)",
			sql, from, to, chanIds, chanIds, chanIds, chanIds, chanIds)
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
			&c.ChanPoint,
			&c.ChanId,
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
