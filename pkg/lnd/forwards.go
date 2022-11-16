package lnd

import (
	"context"
	"database/sql"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/pkg/commons"
)

func convMicro(ns uint64) time.Time {
	return time.Unix(0, int64(ns)).Round(time.Microsecond).UTC()
}

// storeForwardingHistory
func storeForwardingHistory(db *sqlx.DB, fwh []*lnrpc.ForwardingEvent, nodeId int) error {
	if len(fwh) > 0 {
		tx := db.MustBegin()
		stmt, err := tx.Prepare(`INSERT INTO forward(time, time_ns, fee_msat,
				incoming_amount_msat, outgoing_amount_msat, incoming_channel_id, outgoing_channel_id, node_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (time, time_ns) DO NOTHING;`)
		if err != nil {
			return err
		}
		for _, event := range fwh {
			incomingShortChannelId := channels.ConvertLNDShortChannelID(event.ChanIdIn)
			incomingChannelId := commons.GetChannelIdFromShortChannelId(incomingShortChannelId)
			incomingShortChannelIdP := &incomingChannelId
			if incomingChannelId == 0 {
				log.Error().Msgf("Forward received for a non existing channel (incomingShortChannelId: %v)",
					incomingShortChannelId)
				incomingShortChannelIdP = nil
			}
			outgoingShortChannelId := channels.ConvertLNDShortChannelID(event.ChanIdOut)
			outgoingChannelId := commons.GetChannelIdFromShortChannelId(outgoingShortChannelId)
			outgoingChannelIdP := &outgoingChannelId
			if outgoingChannelId == 0 {
				log.Error().Msgf("Forward received for a non existing channel (outgoingShortChannelId: %v)",
					outgoingShortChannelId)
				outgoingChannelIdP = nil
			}
			_, err = stmt.Exec(convMicro(event.TimestampNs), event.TimestampNs, event.FeeMsat,
				event.AmtInMsat, event.AmtOutMsat, incomingShortChannelIdP, outgoingChannelIdP, nodeId)
			if err != nil {
				return errors.Wrapf(err, "storeForwardingHistory->tx.Exec(%v)", event)
			}
		}
		err = stmt.Close()
		if err != nil {
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// MAXEVENTS is used to set the maximum events in ForwardingHistoryRequest.
// It's also used to check if we need to request more.
const MAXEVENTS int = 50000

// fetchLastForwardTime fetches the latest recorded forward, if none is set already.
// This should only run once when a server starts.
func fetchLastForwardTime(db *sqlx.DB) (uint64, error) {

	var lastNs uint64

	row := db.QueryRow("SELECT time_ns FROM forward ORDER BY time_ns DESC LIMIT 1;")
	err := row.Scan(&lastNs)

	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return lastNs, nil
}

type lightningClientForwardingHistory interface {
	ForwardingHistory(ctx context.Context, in *lnrpc.ForwardingHistoryRequest,
		opts ...grpc.CallOption) (*lnrpc.ForwardingHistoryResponse, error)
}

// FwhOptions allows the caller to adjust the number of forwarding events can be requested at a time
// and set a custom time interval between requests.
type FwhOptions struct {
	MaxEvents *int
	Tick      <-chan time.Time
}

// SubscribeForwardingEvents repeatedly requests forwarding history starting after the last
// forwarding stored in the database and stores new forwards.
func SubscribeForwardingEvents(ctx context.Context, client lightningClientForwardingHistory, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}, opt *FwhOptions) {

	maxEvents := MAXEVENTS

	// Check if maxEvents has been set and that it is bellow the hard coded maximum defined by
	// the constant MAXEVENTS.
	if (opt != nil) && ((*opt.MaxEvents > MAXEVENTS) || (*opt.MaxEvents <= 0)) {
		maxEvents = *opt.MaxEvents
	}

	// Create the default ticker used to fetch forwards at a set interval
	c := clock.New()
	ticker := c.Tick(10 * time.Second)

	// If a custom ticker is set in the options, override the default ticker.
	if (opt != nil) && (opt.Tick != nil) {
		ticker = opt.Tick
	}

	rl := ratelimit.New(1) // 1 per second maximum rate limit

	// Request the forwarding history at the requested interval.
	// NB!: This timer is slowly being shifted because of the time required to
	//fetch and store the response.
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			// Keep fetching until LND returns less than the max number of records requested.
			for {
				rl.Take() // rate limited to 1 per second, when caught up will normally be 1 every 10 seconds

				// Fetch the nanosecond timestamp of the most recent record we have.
				lastNs, err := fetchLastForwardTime(db)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to obtain last know forward, will retry in 10 seconds")
					break
				}
				lastTimestamp := lastNs / uint64(time.Second)

				fwhReq := &lnrpc.ForwardingHistoryRequest{
					StartTime:    lastTimestamp,
					NumMaxEvents: uint32(maxEvents),
				}
				fwh, err := client.ForwardingHistory(ctx, fwhReq)
				if err != nil {
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}
					log.Error().Err(err).Msgf("Failed to obtain forwards, will retry in 10 seconds")
					break
				}

				// Store the forwarding history
				err = storeForwardingHistory(db, fwh.ForwardingEvents, nodeSettings.NodeId)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to store forward event")
					rl.Take()
				}

				// Stop fetching if there are fewer forwards than max requested
				// (indicates that we have the last forwarding record)
				if len(fwh.ForwardingEvents) < maxEvents {
					break
				} else {
					log.Info().Msgf("Still running bulk import of forward events")
				}
			}
		}
	}
}
