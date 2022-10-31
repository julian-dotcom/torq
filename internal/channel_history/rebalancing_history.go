package channel_history

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/lncapital/torq/pkg/commons"
)

type RebalancingDetails struct {
	AmountMsat    uint64 `db:"amount_msat" json:"amountMsat"`
	TotalCostMsat uint64 `db:"total_cost_msat" json:"totalCostMsat"`
	SplitCostMsat uint64 `db:"split_cost_msat" json:"splitCostMsat"`
	Count         uint64 `db:"count" json:"count"`
}

func getRebalancingCost(db *sqlx.DB, nodeIds []int, from time.Time, to time.Time) (RebalancingDetails, error) {
	settings := commons.GetSettings()

	publicKeys := make([]string, len(nodeIds))
	for nodeId := range nodeIds {
		publicKeys = append(publicKeys, commons.GetNodeSettingsByNodeId(nodeId).PublicKey)
	}

	row := db.QueryRow(`
		SELECT COALESCE(ROUND(SUM(amount_msat)),0) AS amount_msat,
			   COALESCE(ROUND(SUM(total_fee_msat)),0) AS total_cost_msat,
			   COALESCE(COUNT(*), 0) AS count
		FROM (
			SELECT creation_timestamp at time zone ($4),
				   value_msat as amount_msat,
				   fee_msat as total_fee_msat
			FROM payment p
			WHERE status = 'SUCCEEDED' AND
				htlcs->-1->'route'->'hops'->-1->>'pub_key' = ANY($1) AND
				creation_timestamp::timestamp AT TIME ZONE ($4) >= $2::timestamp AND
				creation_timestamp::timestamp AT TIME ZONE ($4) <= $3::timestamp
		) AS a;`, pq.Array(publicKeys), from, to, settings.PreferredTimeZone)
	var cost RebalancingDetails
	err := row.Scan(
		&cost.AmountMsat,
		&cost.TotalCostMsat,
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
//func getChannelRebalancingAllTime(db *sqlx.DB, chanIds []string) (RebalancingDetails, error) {
//	allTorqPublicKeys := commons.GetAllTorqPublicKeys()
//	settings := commons.GetSettings()
//
//	row := db.QueryRow(`
//		SELECT COALESCE(ROUND(SUM(amount_msat)),0) AS amount_msat,
//			   COALESCE(ROUND(SUM(total_fee_msat)),0) AS total_cost_msat,
//			   COALESCE(ROUND(SUM(split_fee_msat)),0) AS split_cost_msat,
//			   COALESCE(COUNT(*), 0) AS count
//		FROM (
//			select creation_timestamp at time zone ($3),
//				   value_msat as amount_msat,
//				   fee_msat as total_fee_msat,
//				   case
//				   when
//					   -- When two channels in the same group is involved, return the full rebalancing cost.
//					   htlcs->-1->'route'->'hops'->0->>'chan_id' = ANY($1) and
//					   htlcs->-1->'route'->'hops'->-1->>'chan_id' = ANY($1)
//					   then fee_msat
//				   when
//					   -- When only one channel in the group is involved, return half the rebalancing cost.
//					   htlcs->-1->'route'->'hops'->0->>'chan_id' = ANY($1) or
//					   htlcs->-1->'route'->'hops'->-1->>'chan_id' = ANY($1)
//					   then fee_msat/2
//				   end as split_fee_msat
//			from payment p
//			where status = 'SUCCEEDED'
//			and (
//				htlcs->-1->'route'->'hops'->0->>'chan_id' = ANY($1)
//				or htlcs->-1->'route'->'hops'->-1->>'chan_id' = ANY($1)
//			)
//			and htlcs->-1->'route'->'hops'->-1->>'pub_key' = ANY($2)
//		) as a;`, pq.Array(chanIds), pq.Array(allTorqPublicKeys), settings.PreferredTimeZone)
//	var cost RebalancingDetails
//	err := row.Scan(
//		&cost.AmountMsat,
//		&cost.TotalCostMsat,
//		&cost.SplitCostMsat,
//		&cost.Count,
//	)
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

func getChannelRebalancing(db *sqlx.DB, nodeIds []int, chanIds []string,
	from time.Time, to time.Time) (RebalancingDetails, error) {

	publicKeys := make([]string, len(nodeIds))
	for nodeId := range nodeIds {
		publicKeys = append(publicKeys, commons.GetNodeSettingsByNodeId(nodeId).PublicKey)
	}

	settings := commons.GetSettings()

	row := db.QueryRow(`
		SELECT COALESCE(ROUND(SUM(amount_msat)),0) AS amount_msat,
			   COALESCE(ROUND(SUM(total_fee_msat)),0) AS total_cost_msat,
			   COALESCE(ROUND(SUM(split_fee_msat)),0) AS split_cost_msat,
			   COALESCE(COUNT(*), 0) AS count
		from (
			select creation_timestamp at time zone ($5),
				   value_msat as amount_msat,
				   fee_msat as total_fee_msat,
				   case
				   when
					   -- When two channels in the same group is involved, return the full rebalancing cost.
					   htlcs->-1->'route'->'hops'->0->>'chan_id' = ANY($1) and
					   htlcs->-1->'route'->'hops'->-1->>'chan_id' = ANY($1)
					   then fee_msat
				   when
					   -- When only one channel in the group is involved, return half the rebalancing cost.
					   htlcs->-1->'route'->'hops'->0->>'chan_id' = ANY($1) or
					   htlcs->-1->'route'->'hops'->-1->>'chan_id' = ANY($1)
					   then fee_msat/2
				   end as split_fee_msat
			from payment p
			where status = 'SUCCEEDED'
			and (
				htlcs->-1->'route'->'hops'->0->>'chan_id' = ANY($1)
				or htlcs->-1->'route'->'hops'->-1->>'chan_id' = ANY($1)
			)
			and htlcs->-1->'route'->'hops'->-1->>'pub_key' = ANY($4)
			and creation_timestamp::timestamp AT TIME ZONE ($5) >= ($2)::timestamp
			and creation_timestamp::timestamp AT TIME ZONE ($5) <= ($3)::timestamp
			and node_id IN ($6)
		) AS a;`, pq.Array(chanIds), from, to, pq.Array(publicKeys), settings.PreferredTimeZone, pq.Array(nodeIds))

	var cost RebalancingDetails
	err := row.Scan(
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
