package payments

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"time"
)

type Payment struct {
	PaymentIndex            uint64         `json:"payment_index" db:"payment_index"`
	CreationTimestamp       time.Time      `json:"creation_timestamp" db:"creation_timestamp"`
	Status                  string         `json:"status" db:"status"`
	DestinationPubKey       *string        `json:"destination_pub_key" db:"destination_pub_key"`
	FailureReason           string         `json:"failure_reason" db:"failure_reason"`
	FeeMsat                 uint64         `json:"fee_msat" db:"fee_msat"`
	ValueMsat               uint64         `json:"value_msat" db:"value_msat"`
	PaymentHash             string         `json:"payment_hash" db:"payment_hash"`
	PaymentPreimage         string         `json:"payment_preimage" db:"payment_preimage"`
	PaymentRequest          *string        `json:"payment_request" db:"payment_request"`
	IsRebalance             *bool          `json:"is_rebalance" db:"is_rebalance"`
	CountSuccessfulAttempts int            `json:"count_successful_attempts" db:"count_successful_attempts"`
	CountFailedAttempts     int            `json:"count_failed_attempts" db:"count_failed_attempts"`
	SuccessfulAttemptTimeNs pq.StringArray `json:"successful_attempt_time_ns" db:"successful_attempt_time_ns"`
	SuccessfulResolveTimeNs pq.StringArray `json:"successful_resolve_time_ns" db:"successful_resolve_time_ns"`
}

type Hop struct {
	Expiry           uint64 `json:"expiry" db:"expiry"`
	ChanId           uint64 `json:"chan_id" db:"chan_id"`
	PubKey           string `json:"pub_key" db:"pub_key"`
	FeeMsat          uint64 `json:"fee_msat" db:"fee_msat"`
	TlvPayload       bool   `json:"tlv_payload" db:"tlv_payload"`
	ChanCapacity     uint64 `json:"chan_capacity" db:"chan_capacity"`
	AmtToForwardMsat uint64 `json:"amt_to_forward_msat" db:"amt_to_forward_msat"`
}

type Route struct {
	Hops          []*Hop `json:"hops" db:"hops"`
	TotalAmt      uint64 `json:"total_amt" db:"total_amt"`
	TotalFees     uint64 `json:"total_fees" db:"total_fees"`
	TotalAmtMsat  uint64 `json:"total_amt_msat" db:"total_amt_msat"`
	TotalFeesMsat uint64 `json:"total_fees_msat" db:"total_fees_msat"`
	TotalTimeLock uint64 `json:"total_time_lock" db:"total_time_lock"`
}

type PaymentDetails struct {
	Payment
	SuccessFulRoutes []*Route `json:"successful_routes" db:"successful_routes"`
	FailedRoutes     []*Route `json:"failed_routes" db:"failed_routes"`
}

func getPaymentsMaxOffset(db *sqlx.DB, from time.Time, to time.Time, limit int) (maxOffset int, err error) {
	err = db.Get(&maxOffset, `
		WITH
			tz AS (select preferred_timezone as tz from settings)
			select floor(count(*)::float8/$3) from payment
			WHERE creation_timestamp::timestamp AT TIME ZONE (table tz) between $1 AND $2;`, limit)
	if err != nil {
		return 0, err
	}

	return maxOffset, err
}

type QueryPaymentsParams struct {
	From                    *time.Time
	To                      *time.Time
	Destination             []string
	IsRebalance             *bool
	CountSuccessfulAttempts *QueryFilter
	CountFailedAttempts     *QueryFilter
	FailureReason           []string
	Status                  []string
}

type QueryFilter struct {
	Operator string
	Value    string
}

func getPayments(db *sqlx.DB, qpp QueryPaymentsParams, limit uint64, offset uint64) (r []*Payment, err error) {

	//language=PostgreSQL
	q := `(select creation_timestamp, payment_index, payment_hash, payment_preimage, payment_request, status, value_msat, fee_msat, failure_reason,
		    destination_pub_key,
			(htlcs->-1->'route'->'hops'->-1->>'pub_key' = ANY(ARRAY[(table pub_keys)])) as is_rebalance,
			jsonb_array_length(jsonb_path_query_array(htlcs, '$??(@.status==1).route')) count_successful_attempts,
			jsonb_array_length(jsonb_path_query_array(htlcs, '$??(@.status!=1).route')) count_failed_attempts,
    		ARRAY(SELECT jsonb_array_elements(jsonb_path_query_array(htlcs,'$??(@.status==1).attempt_time_ns'))) as successful_attempt_time_ns,
    		ARRAY(SELECT jsonb_array_elements(jsonb_path_query_array(htlcs,'$??(@.status==1).resolve_time_ns'))) as successful_resolve_time_ns
		from (
			SELECT creation_timestamp, created_on, payment_index, payment_hash, payment_preimage, payment_request,
status, value_msat, fee_msat, failure_reason,htlcs,
			htlcs->-1->'route'->'hops'->-1->>'pub_key' as destination_pub_key
			from payment
			order by creation_timestamp desc
		) a) b
	`

	qbs := sq.And{}

	if qpp.From != nil {
		qbs = append(qbs, sq.GtOrEq{"creation_timestamp::timestamp AT TIME ZONE (table tz)": qpp.From})
	}

	if qpp.To != nil {
		qbs = append(qbs, sq.LtOrEq{"creation_timestamp::timestamp AT TIME ZONE (table tz)": qpp.To})
	}

	if qpp.Destination != nil {
		qbs = append(qbs, sq.Eq{"b.destination_pub_key": qpp.Destination})
	}

	if qpp.IsRebalance != nil {
		qbs = append(qbs, sq.Eq{"b.is_rebalance": qpp.IsRebalance})
	}

	if qpp.FailureReason != nil {
		qbs = append(qbs, sq.Eq{"b.failure_reason": qpp.FailureReason})
	}

	if qpp.Status != nil {
		qbs = append(qbs, sq.Eq{"status": qpp.Status})
	}

	qb := sq.Select("*").
		PlaceholderFormat(sq.Dollar).
		From(q).
		Where(qbs).
		Limit(limit).
		Offset(offset).
		Prefix(`WITH
			tz AS (select preferred_timezone as tz from settings),
			pub_keys as (select array_agg(pub_key) from local_node)
		`)

	// Compile the query
	qs, args, err := qb.ToSql()

	if err != nil {
		return nil, err
	}

	// Log for debugging
	log.Debug().Msgf("Query: %s, \n Args: %v", qs, args)

	rows, err := db.DB.Query(qs, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var p Payment
		err = rows.Scan(
			&p.CreationTimestamp,
			&p.PaymentIndex,
			&p.PaymentHash,
			&p.PaymentPreimage,
			&p.PaymentRequest,
			&p.Status,
			&p.ValueMsat,
			&p.FeeMsat,
			&p.FailureReason,
			&p.DestinationPubKey,
			&p.IsRebalance,
			&p.CountSuccessfulAttempts,
			&p.CountFailedAttempts,
			&p.SuccessfulAttemptTimeNs,
			&p.SuccessfulResolveTimeNs,
		)

		if err != nil {
			return nil, err
		}

		r = append(r, &p)

	}

	return r, nil
}

func getPaymentDetails(db *sqlx.DB, paymentHash string) (r *PaymentDetails, err error) {

	q := `
		WITH
			pub_keys as (select array_agg(pub_key) from local_node)
		select creation_timestamp, payment_index, payment_hash, payment_preimage, payment_request, status, value_msat, fee_msat, failure_reason,
			(htlcs->-1->'route'->'hops'->-1->>'pub_key' = ANY(ARRAY[(table pub_keys)])) as is_rebalance,
			jsonb_array_length(jsonb_path_query_array(htlcs, '$?(@.status==1).route')) count_successful_attempts,
			jsonb_array_length(jsonb_path_query_array(htlcs, '$?(@.status!=1).route')) count_failed_attempts,
			jsonb_path_query_array(htlcs, '$?(@.status==1).attempt_time_ns') as successful_attempt_time_ns,
			jsonb_path_query_array(htlcs, '$?(@.status==1).resolve_time_ns') as successful_resolve_time_ns,
		    jsonb_path_query_array(htlcs, '$?(@.status!=1).route') failed_routes,
   			jsonb_path_query_array(htlcs, '$?(@.status==1).route') successful_routes
		from payment
			WHERE payment_hash = $1;
	`

	err = db.Get(r, q, paymentHash)

	return r, nil
}
