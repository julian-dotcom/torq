package channel_history

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type RebalancingTotal struct {
	AmountMsat    uint64 `db:"amount_msat" json:"amount_msat"`
	TotalCostMsat uint64 `db:"total_cost_msat" json:"total_cost_msat"`
	SplitCostMsat uint64 `db:"split_cost_msat" json:"split_cost_msat"`
	Count         uint64 `db:"count" json:"count"`
}

func getChannelRebalancingTotal(db *sqlx.DB, chanIds []string, pubKeys []string) (cost RebalancingTotal,
	err error) {

	q := `WITH
			tz AS (select preferred_timezone from settings),
			chan_ids as (select $1::text[]),
			pub_keys as (select $2::text[])
		select coalesce(round(sum(amount_msat)),0) as amount_msat,
			   coalesce(round(sum(total_fee_msat)),0) as total_cost_msat,
			   coalesce(round(sum(split_fee_msat)),0) as split_cost_msat,
			   coalesce(count(*), 0) as count
		from (
			select creation_timestamp at time zone (table tz),
				   value_msat as amount_msat,
				   fee_msat as total_fee_msat,
				   case
				   when
					   -- When two channels in the same group is involved, return the full rebalancing cost.
					   htlcs->-1->'route'->'hops'->0->>'chan_id' = ANY(ARRAY[(table chan_ids)]) and
					   htlcs->-1->'route'->'hops'->-1->>'chan_id' = ANY(ARRAY[(table chan_ids)])
					   then fee_msat
				   when
					   -- When only one channel in the group is involved, return half the rebalancing cost.
					   htlcs->-1->'route'->'hops'->0->>'chan_id' = ANY(ARRAY[(table chan_ids)]) or
					   htlcs->-1->'route'->'hops'->-1->>'chan_id' = ANY(ARRAY[(table chan_ids)])
					   then fee_msat/2
				   end as split_fee_msat
			from payment p
			where status = 'SUCCEEDED'
			and (
				htlcs->-1->'route'->'hops'->0->>'chan_id' = ANY(ARRAY[(table chan_ids)])
				or htlcs->-1->'route'->'hops'->-1->>'chan_id' = ANY(ARRAY[(table chan_ids)])
			)
			and htlcs->-1->'route'->'hops'->-1->>'pub_key' = ANY(ARRAY[(table pub_keys)])
		) as a;`

	row := db.QueryRow(q, pq.Array(chanIds), pq.Array(pubKeys))
	err = row.Scan(
		&cost.AmountMsat,
		&cost.TotalCostMsat,
		&cost.SplitCostMsat,
		&cost.Count,
	)

	if err == sql.ErrNoRows {
		return cost, nil
	}

	if err != nil {
		return cost, err
	}

	return cost, nil

}

//
//func getChannelRebalancingHisotry(db *sqlx.DB, chanIds []string) (cost *uint64,
//	err error) {
//
//	q := `WITH
//			tz AS (select preferred_timezone from settings)
//		select creation_timestamp at time zone (table tz),
//			   round((value_msat + fee_msat)/1000) as amount,
//			   round(fee_msat/1000) as fee,
//			   htlcs->-1->'route'->'hops'->0->'chan_id' as outbound_chan_id,
//			   htlcs->-1->'route'->'hops'->-1->'chan_id' as destination_chan_id,
//			   htlcs->-1->'route'->'hops'->-1->>'pub_key' as destination_pub_key,
//			   jsonb_array_length(htlcs->-1->'route'->'hops') as hops,
//			   htlcs->-1->'route'->'hops' as hops,
//			   jsonb_path_query_array(htlcs->-1->'route'->'hops', '$[*].pub_key') as hops_pub_keys,
//			   jsonb_path_query_array(htlcs->-1->'route'->'hops', '$[*].chan_id') as hops_chan_ids,
//			   to_timestamp(((htlcs->-1->>'resolve_time_ns')::numeric)/1000000000) at time zone (table tz) as resolve_time_ns,
//			   to_timestamp(((htlcs->-1->>'attempt_time_ns')::numeric)/1000000000) at time zone (table tz) as attempt_time_ns,
//			   -- Resolved duration is the duration it took for the entire payment to succeed (all attempts)
//			   to_timestamp(((htlcs->-1->>'resolve_time_ns')::numeric)/1000000000) at time zone (table tz) - creation_timestamp at time zone (table tz) as resolved_duration,
//			   -- Attempt duration is the duration it took for the successfull (last) HTLC attempt to succeed
//			   to_timestamp(((htlcs->-1->>'resolve_time_ns')::numeric)/1000000000) at time zone (table tz) - to_timestamp(((htlcs->-1->>'attempt_time_ns')::numeric)/1000000000) at time zone (table tz) as attempt_duration
//		from payment p
//		where status = 'SUCCEEDED'
//		and (
//		    htlcs->-1->'route'->'hops'->0->>'chan_id' in ('111', '111')
//			or htlcs->-1->'route'->'hops'->-1->>'chan_id' in ('111', '111')
//		)
//		and ARRAY[htlcs->-1->'route'->'hops'->-1->>'pub_key'] in (ARRAY['sfdsf']);`
//
//	qs, args, err := sqlx.In(q, chanIds)
//	if err != nil {
//		return nil, errors.Wrapf(err, "sqlx.In(%s, %v)", q, chanIds)
//	}
//
//	qsr := db.Rebind(qs)
//
//	row := db.QueryRow(qsr, args...)
//	err = row.Scan(&cost)
//
//	if err == sql.ErrNoRows {
//		return cost, nil
//	}
//
//	if err != nil {
//		return cost, err
//	}
//
//	return cost, nil
//
//}
