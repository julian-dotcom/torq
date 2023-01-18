package lnd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/zpay32"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/commons"
)

type lightningClient_ListPayments interface {
	ListPayments(ctx context.Context, in *lnrpc.ListPaymentsRequest,
		opts ...grpc.CallOption) (*lnrpc.ListPaymentsResponse,
		error)
}

// PayOptions allows the caller to adjust the number of payments can be requested at a time
// and set a custom time interval between requests.
type PayOptions struct {
	Tick <-chan time.Time
}

func SubscribeAndStorePayments(ctx context.Context, client lightningClient_ListPayments, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}, opt *PayOptions) {

	defer log.Info().Msgf("SubscribeAndStorePayments terminated for nodeId: %v", nodeSettings.NodeId)

	var lastPaymentIndex uint64
	var payments *lnrpc.ListPaymentsResponse
	var err error
	serviceStatus := commons.Inactive
	bootStrapping := true
	subscriptionStream := commons.PaymentStream
	ticker := clock.New().Tick(commons.STREAM_PAYMENTS_TICKER_SECONDS * time.Second)
	includeIncomplete := commons.RunningServices[commons.LndService].GetIncludeIncomplete(nodeSettings.NodeId)

	// If a custom ticker is set in the options, override the default ticker.
	if (opt != nil) && (opt.Tick != nil) {
		ticker = opt.Tick
	}

	// Request the Payments at the requested interval.
	// NB!: This timer is slowly being shifted because of the time required to
	// fetch and store the response.
	for {
		// Exit if canceled
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			importCounter := 0

			lastPaymentIndex, err = fetchLastPaymentIndex(db, nodeSettings.NodeId)
			if err != nil {
				serviceStatus = SendStreamEvent(eventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
				log.Error().Err(err).Msgf("Failed to obtain last know forward, will retry in %v seconds", commons.STREAM_PAYMENTS_TICKER_SECONDS)
				continue
			}

			for {
				if bootStrapping {
					serviceStatus = SendStreamEvent(eventChannel, nodeSettings.NodeId, subscriptionStream, commons.Initializing, serviceStatus)
				} else {
					serviceStatus = SendStreamEvent(eventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
				}
				payments, err = fetchPayments(ctx, client, lastPaymentIndex, includeIncomplete)
				if err != nil {
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}
					serviceStatus = SendStreamEvent(eventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
					log.Error().Err(err).Msgf("Failed to obtain payments, will retry in %v seconds", commons.STREAM_PAYMENTS_TICKER_SECONDS)
					break
				}

				// Store the payments
				err = storePayments(db, payments.Payments, nodeSettings, eventChannel, bootStrapping)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to store payments, will retry in %v seconds", commons.STREAM_PAYMENTS_TICKER_SECONDS)
					break
				}

				// Stop fetching if there are fewer forwards than max requested
				// (indicates that we have the last forwarding record)
				if len(payments.Payments) == 0 || lastPaymentIndex == payments.LastIndexOffset {
					if bootStrapping {
						log.Info().Msgf("Bulk import of payments: %v", importCounter)
					}
					bootStrapping = false
					break
				}
				lastPaymentIndex = payments.LastIndexOffset
				if bootStrapping {
					importCounter++
					if importCounter%500 == 0 {
						log.Info().Msgf("Still running bulk import of payments (%v)", importCounter)
					}
				}
			}
		}
	}
}

func fetchLastPaymentIndex(db *sqlx.DB, nodeId int) (uint64, error) {
	var last uint64

	row := db.QueryRow(`select coalesce(max(payment_index), 0) as latest from payment where node_id = $1;`, nodeId)
	err := row.Scan(&last)

	if err != nil {
		return 0, errors.Wrap(err, "fetching last payment index")
	}

	return last, nil
}

// fetchPayments fetches completed payments from LND.
func fetchPayments(ctx context.Context, client lightningClient_ListPayments, last uint64, includeIncomplete bool) (
	r *lnrpc.ListPaymentsResponse, err error) {

	//retry:
	req := &lnrpc.ListPaymentsRequest{
		IncludeIncomplete: includeIncomplete,
		IndexOffset:       last,
		MaxPayments:       1, // Only fetch one at a time due to the size of failed payments
		Reversed:          false,
	}
	r, err = client.ListPayments(ctx, req)

	if err != nil {
		return nil, errors.Wrap(err, "fetch payments: lnrpc list payments")
	}

	return r, nil
}

func storePayments(db *sqlx.DB, p []*lnrpc.Payment, nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}, bootStrapping bool) error {
	const q = `INSERT INTO payment(
				  payment_hash,
				  creation_timestamp,
				  payment_preimage,
				  value_msat,
				  payment_request,
				  status,
				  fee_msat,
				  creation_time_ns,
				  htlcs,
				  payment_index,
				  failure_reason,
                  incoming_channel_id,
                  outgoing_channel_id,
                  rebalance_amount_msat,
				  node_id,
				  created_on)
			  VALUES ($1, $2, $3, $4, $5,$6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			  ON CONFLICT (creation_timestamp, payment_index) DO NOTHING;`

	var paymentEvents []commons.PaymentEvent
	if len(p) > 0 {
		tx := db.MustBegin()

		for _, payment := range p {
			htlcJson, err := json.Marshal(payment.Htlcs)
			if err != nil {
				return errors.Wrap(err, "JSON Marshal the payment HTLCs")
			}

			var incomingChannelId *int
			var outgoingChannelId *int
			var rebalanceAmountMsat *uint64
			if len(payment.Htlcs) == 0 || len(payment.Htlcs[0].Route.Hops) == 0 {
				if payment.Status == lnrpc.Payment_SUCCEEDED {
					log.Error().Msgf("The payment HTLCs and/or Hops are unknown for paymentHash: %v", payment.PaymentHash)
				}
			} else {
				incomingChannelId = getChannelIdByLndShortChannelId(payment.Htlcs[0].Route.Hops[len(payment.Htlcs[0].Route.Hops)-1].ChanId)
				outgoingChannelId = getChannelIdByLndShortChannelId(payment.Htlcs[0].Route.Hops[0].ChanId)
				if outgoingChannelId == nil {
					if payment.Status != lnrpc.Payment_FAILED {
						log.Error().Msgf("The payment HTLCs has an unknown outgoingChannel for paymentHash: %v", payment.PaymentHash)
					}
				}
				if incomingChannelId != nil && *incomingChannelId != 0 {
					channelSettings := commons.GetChannelSettingByChannelId(*incomingChannelId)
					if channelSettings.FirstNodeId == nodeSettings.NodeId ||
						channelSettings.SecondNodeId == nodeSettings.NodeId {
						rebalanceAmountMsatV := uint64(payment.Htlcs[0].Route.Hops[0].AmtToForwardMsat)
						rebalanceAmountMsat = &rebalanceAmountMsatV
					}
				}
			}
			if _, err := tx.Exec(q,
				payment.PaymentHash,
				convertMicro(payment.CreationTimeNs),
				payment.PaymentPreimage,
				payment.ValueMsat,
				payment.PaymentRequest,
				payment.Status.String(),
				payment.FeeMsat,
				payment.CreationTimeNs,
				htlcJson,
				payment.PaymentIndex,
				payment.FailureReason.String(),
				incomingChannelId,
				outgoingChannelId,
				rebalanceAmountMsat,
				nodeSettings.NodeId,
				time.Now().UTC(),
			); err != nil {
				return errors.Wrap(err, "store payments: db exec")
			}
			paymentEvent := commons.PaymentEvent{
				EventData: commons.EventData{
					EventTime: time.Now(),
					NodeId:    nodeSettings.NodeId,
				},
				AmountPaid:           payment.ValueSat,
				FeeMsat:              uint64(payment.FeeMsat),
				PaymentStatus:        payment.Status,
				PaymentFailureReason: payment.FailureReason,
			}
			paymentEvent.IncomingChannelId = incomingChannelId
			paymentEvent.OutgoingChannelId = outgoingChannelId
			paymentEvent.RebalanceAmountMsat = rebalanceAmountMsat
			paymentEvents = append(paymentEvents, paymentEvent)
		}
		err := tx.Commit()
		if err != nil {
			return errors.Wrap(err, "Transaction commit")
		}
	}

	if eventChannel != nil && !bootStrapping {
		for _, paymentEvent := range paymentEvents {
			eventChannel <- paymentEvent
		}
	}

	return nil
}

func UpdateInFlightPayments(ctx context.Context, client lightningClient_ListPayments, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}, opt *PayOptions) {

	defer log.Info().Msgf("UpdateInFlightPayments terminated for nodeId: %v", nodeSettings.NodeId)

	var listPaymentsResponse *lnrpc.ListPaymentsResponse
	serviceStatus := commons.Inactive
	bootStrapping := true
	subscriptionStream := commons.InFlightPaymentStream
	ticker := clock.New().Tick(commons.STREAM_INFLIGHT_PAYMENTS_TICKER_SECONDS * time.Second)

	// If a custom ticker is set in the options, override the default ticker.
	if (opt != nil) && (opt.Tick != nil) {
		ticker = opt.Tick
	}

	// Request the in flight payments at the requested interval.
	// NB!: This timer is slowly being shifted because of the time required to
	// fetch and store the response.
	for {
		// Exit if canceled
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			inFlightIndexes, err := fetchInFlightPaymentIndexes(db, nodeSettings.NodeId)
			if err != nil {
				serviceStatus = SendStreamEvent(eventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
				log.Error().Err(err).Msgf("Failed to obtain in-flight payment indexes, will retry in %v seconds", commons.STREAM_INFLIGHT_PAYMENTS_TICKER_SECONDS)
				continue
			}
			if bootStrapping {
				serviceStatus = SendStreamEvent(eventChannel, nodeSettings.NodeId, subscriptionStream, commons.Initializing, serviceStatus)
			} else {
				serviceStatus = SendStreamEvent(eventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
			}
			for _, i := range inFlightIndexes {
				ifPayIndex := i - 1 // Subtract one to get that index, otherwise we would get the one after.
				// we will only get one payment back. Might not be the right one.
				listPaymentsResponse, err = fetchPayments(ctx, client, ifPayIndex, true)
				if err != nil {
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}
					log.Error().Err(err).Msg("Error with subscribe and update payments")
					continue
				}
				if len(listPaymentsResponse.Payments) == 0 {
					log.Info().Msgf("We had an inflight payment but nothing from LND: %v", i)
					if err = setPaymentToFailedDetailsUnavailable(db, i); err != nil {
						log.Error().Err(err).Msg("Error with Setting payment to failed details unavailable")
					}
					continue
				}

				if listPaymentsResponse.Payments[0].PaymentIndex != i {
					log.Warn().Msgf("Payment data missing from LND for payment index: %v", i)
					if err = setPaymentToFailedDetailsUnavailable(db, i); err != nil {
						log.Error().Err(err).Msg("Error with Setting payment to failed details unavailable")
					}
					continue
				}

				// Store the payments
				err = updatePayments(db, listPaymentsResponse.Payments, nodeSettings.NodeId)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to store update payments")
				}
			}
			bootStrapping = false
		}
	}
}

type Payment struct {
	PaymentIndex      uint64 `json:"payment_index" db:"payment_index"`
	PaymentHash       string `json:"payment_hash" db:"payment_hash"`
	PaymentPreimage   string `json:"payment_preimage" db:"payment_preimage"`
	PaymentRequest    string `json:"payment_request" db:"payment_request"`
	Status            string `json:"status" db:"status"`
	ValueMsat         int64  `json:"value_msat" db:"value_msat"`
	FeeMsat           int64  `json:"fee_msat" db:"fee_msat"`
	Htlcs             []byte `json:"htlcs" db:"htlcs"`
	FailureReason     string `json:"failure_reason" db:"failure_reason"`
	CreationTimeNs    int64  `json:"creation_time_ns" db:"creation_time_ns"`
	CreationTimestamp int64  `json:"creation_timestamp" db:"creation_timestamp"`
}

func fetchInFlightPaymentIndexes(db *sqlx.DB, nodeId int) (r []uint64, err error) {

	rows, err := db.Query(`
		select payment_index
		from payment
		where status = 'IN_FLIGHT'
		and node_id = $1
		order by payment_index asc;
	`, nodeId)
	if err != nil {
		return nil, errors.Wrap(err, "DB Query of inflight payment indexes")
	}

	for rows.Next() {
		var res *uint64
		err = rows.Scan(&res)

		if err != nil {
			return r, errors.Wrap(err, "SQL row scan")
		}

		// Append to the result
		r = append(r, *res)

	}

	return r, nil
}

func setPaymentToFailedDetailsUnavailable(db *sqlx.DB, paymentIndex uint64) error {
	const q = `update payment set(
				  status,
				  failure_reason,
				  updated_on)
			  = ($1, $2, $3)
				where payment_index = $4;`

	_, err := db.Exec(q,
		"FAILED",
		"DETAILS_UNAVAILABLE",
		time.Now().UTC(),
		paymentIndex,
	)
	if err != nil {
		return errors.Wrap(err, "Database exec")
	}

	return nil
}

func updatePayments(db *sqlx.DB, p []*lnrpc.Payment, nodeId int) error {

	const q = `update payment set(
				  payment_hash,
				  payment_preimage,
				  value_msat,
				  payment_request,
				  status,
				  fee_msat,
				  htlcs,
				  failure_reason,
                  node_id,
				  updated_on)
			  = ($1, $2, $3,$4, $5,$6, $7, $8, $9, $10)
				where payment_index = $11;`

	if len(p) > 0 {
		tx := db.MustBegin()

		for _, payment := range p {

			htlcJson, err := json.Marshal(payment.Htlcs)
			if err != nil {
				return errors.Wrap(err, "JSON Marhsal of payment HTLCs")
			}

			status := payment.Status.String()
			fr := payment.FailureReason.String()

			// Payments can be stuck as IN_FLIGHT if they are interrupted.
			if status == "IN_FLIGHT" {
				// Check expiry time for IN_FLIGHT payments

				// Default expiry (1 day, just to be sure)
				expiry := time.Duration(24 * time.Hour)

				// Update the expiry time if the PaymentRequest is available
				if payment.PaymentRequest != "" {
					inva, err := zpay32.Decode(payment.PaymentRequest, &chaincfg.MainNetParams)
					if err != nil {
						return errors.Wrap(err, "zpay32 decode of payment request")
					}
					expiry = inva.Expiry()
				}

				currentTime := time.Now().UTC()
				created := time.Unix(0, payment.CreationTimeNs).UTC()
				// Add 10 minutes to the invoice expiry time to be safe.
				expiredAt := created.Add(expiry).Add(10 * time.Minute)

				// If the invoice has expired, mark the payment as failed.
				if currentTime.Unix() > expiredAt.Unix() {
					status = "FAILED"

					// Set the failure reason to timeout, if none is already set.
					if fr == "FAILURE_REASON_NONE" {
						fr = "FAILURE_REASON_TIMEOUT"
					}
				}
			}

			_, err = db.Exec(q,
				payment.PaymentHash,
				payment.PaymentPreimage,
				payment.ValueMsat,
				payment.PaymentRequest,
				status,
				payment.FeeMsat,
				htlcJson,
				fr,
				nodeId,
				time.Now().UTC(),
				payment.PaymentIndex,
			)

			if err != nil {
				return errors.Wrapf(err, "updatePayments->tx.Exec(%v)", q)
			}
		}
		err := tx.Commit()
		if err != nil {
			return errors.Wrap(err, "Transaction commit")
		}
	}

	return nil
}
