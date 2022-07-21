package payments

import (
	"database/sql"
	"encoding/json"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"time"
)

type Payment struct {
	PaymentIndex            uint64    `json:"payment_index" db:"payment_index"`
	Date                    time.Time `json:"date" db:"date"`
	DestinationPubKey       *string   `json:"destination_pub_key" db:"destination_pub_key"`
	Status                  string    `json:"status" db:"status"`
	ValueMsat               uint64    `json:"value_msat" db:"value_msat"`
	FeeMsat                 uint64    `json:"fee_msat" db:"fee_msat"`
	FailureReason           string    `json:"failure_reason" db:"failure_reason"`
	PaymentHash             string    `json:"payment_hash" db:"payment_hash"`
	PaymentPreimage         string    `json:"payment_preimage" db:"payment_preimage"`
	PaymentRequest          *string   `json:"payment_request" db:"payment_request"`
	IsRebalance             *bool     `json:"is_rebalance" db:"is_rebalance"`
	IsMPP                   bool      `json:"is_mpp" db:"is_mpp"`
	CountSuccessfulAttempts int       `json:"count_successful_attempts" db:"count_successful_attempts"`
	CountFailedAttempts     int       `json:"count_failed_attempts" db:"count_failed_attempts"`
	SecondsInFlight         *float32  `json:"seconds_in_flight" db:"seconds_in_flight"`
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

type PaymentDetailsRaw struct {
	Payment
	SuccessfulRoutes []byte `json:"successful_routes" db:"successful_routes"`
	FailedRoutes     []byte `json:"failed_routes" db:"failed_routes"`
}

type PaymentDetails struct {
	Payment
	SuccessfulRoutes []*Route `json:"successful_routes" db:"successful_routes"`
	FailedRoutes     []*Route `json:"failed_routes" db:"failed_routes"`
}

func getPaymentsMaxOffset(db *sqlx.DB, filters sq.Sqlizer, from time.Time, to time.Time, limit int) (maxOffset int, err error) {
	err = db.Get(&maxOffset, `
		WITH
			pub_keys as (select array_agg(pub_key) from local_node)
			select floor(count(*)::float8/$3) from payment;`)
	if err != nil {
		return 0, err
	}

	return maxOffset, err
}

func getPayments(db *sqlx.DB, filter sq.Sqlizer, order []string, limit uint64, offset uint64) (r []*Payment,
	err error) {

	//language=PostgreSQL
	qb := sq.Select("*").
		PlaceholderFormat(sq.Dollar).
		FromSelect(
			sq.Select(`
				payment_index,
				creation_timestamp as date,
				destination_pub_key,
				status,
				value_msat,
				fee_msat,
				failure_reason,
				payment_hash,
				payment_preimage,
				payment_request,
				destination_pub_key = ANY(ARRAY[(table pub_keys)]) as is_rebalance,
				is_mpp,
				count_successful_attempts,
				count_failed_attempts,
				extract(epoch from (to_timestamp(resolved_ns/1000000000)-creation_timestamp))::numeric seconds_in_flight
			`).
				PlaceholderFormat(sq.Dollar).
				From("payment"),
			"subquery").
		Where(filter).
		OrderBy(order...).
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

	rows, err := db.Queryx(qs, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var p Payment
		err = rows.Scan(
			&p.PaymentIndex,
			&p.Date,
			&p.DestinationPubKey,
			&p.Status,
			&p.ValueMsat,
			&p.FeeMsat,
			&p.FailureReason,
			&p.PaymentHash,
			&p.PaymentPreimage,
			&p.PaymentRequest,
			&p.IsRebalance,
			&p.IsMPP,
			&p.CountSuccessfulAttempts,
			&p.CountFailedAttempts,
			&p.SecondsInFlight,
		)

		if err != nil {
			return nil, err
		}

		r = append(r, &p)

	}

	return r, nil
}

type ErrPaymentNotFound struct {
	Identifier string
}

func (e ErrPaymentNotFound) Error() string {
	return "Payment not found"
}

func getPaymentDetails(db *sqlx.DB, identifier string) (*PaymentDetails, error) {

	//language=PostgreSQL
	qb := sq.Select(`
				payment_index,
				creation_timestamp as date,
				destination_pub_key,
				status,
				value_msat,
				fee_msat,
				failure_reason,
				payment_hash,
				payment_preimage,
				payment_request,
				destination_pub_key = ANY(ARRAY[(table pub_keys)]) as is_rebalance,
				is_mpp,
				count_successful_attempts,
				count_failed_attempts,
				extract(epoch from (to_timestamp(resolved_ns/1000000000)-creation_timestamp))::numeric seconds_in_flight,
				successful_routes,
				failed_routes
			`).
		PlaceholderFormat(sq.Dollar).
		From("payment").
		Where(
			sq.Or{
				sq.Eq{"payment_hash": identifier},
				sq.Eq{"payment_request": identifier},
				sq.Eq{"payment_preimage": identifier},
			}).
		Prefix(`WITH
			pub_keys as (select array_agg(pub_key) from local_node)
		`)

	qs, args, err := qb.ToSql()
	r := PaymentDetails{}
	var sr []byte
	var fr []byte

	err = db.QueryRowx(qs, args...).Scan(
		&r.PaymentIndex,
		&r.Date,
		&r.DestinationPubKey,
		&r.Status,
		&r.ValueMsat,
		&r.FeeMsat,
		&r.FailureReason,
		&r.PaymentHash,
		&r.PaymentPreimage,
		&r.PaymentRequest,
		&r.IsRebalance,
		&r.IsMPP,
		&r.CountSuccessfulAttempts,
		&r.CountFailedAttempts,
		&r.SecondsInFlight,
		&sr,
		&fr,
	)

	switch err {
	case nil:
		break
	case sql.ErrNoRows:
		return nil, ErrPaymentNotFound{identifier}
	default:
		return nil, err
	}

	// Unmarshal the Successful routes json byte array
	err = json.Unmarshal(sr, &r.SuccessfulRoutes)
	if err != nil {
		return nil, err
	}

	// Unmarshal the Failed routes json byte array
	err = json.Unmarshal(fr, &r.FailedRoutes)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
