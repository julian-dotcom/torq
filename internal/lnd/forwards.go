package lnd

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
)

const streamForwardsTickerSeconds = 10
const streamLndMaxForwards = 50000

// storeForwardingHistory
func storeForwardingHistory(db *sqlx.DB, fwh []*lnrpc.ForwardingEvent, nodeId int, bootStrapping bool) error {

	if len(fwh) > 0 {
		var forwardEvents []core.ForwardEvent
		tx := db.MustBegin()
		stmt, err := tx.Prepare(`INSERT INTO forward(time, time_ns, fee_msat,
				incoming_amount_msat, outgoing_amount_msat, incoming_channel_id, outgoing_channel_id, node_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (time, time_ns) DO NOTHING;`)
		if err != nil {
			return errors.Wrap(err, "SQL Statement prepare")
		}
		for _, event := range fwh {
			incomingShortChannelId := core.ConvertLNDShortChannelID(event.ChanIdIn)
			incomingChannelId := cache.GetChannelIdByShortChannelId(&incomingShortChannelId)
			incomingChannelIdP := &incomingChannelId
			if incomingChannelId == 0 {
				log.Error().Msgf("Forward received for a non existing channel (incomingChannelIdP: %v)",
					incomingShortChannelId)
				incomingChannelIdP = nil
			}
			outgoingShortChannelId := core.ConvertLNDShortChannelID(event.ChanIdOut)
			outgoingChannelId := cache.GetChannelIdByShortChannelId(&outgoingShortChannelId)
			outgoingChannelIdP := &outgoingChannelId
			if outgoingChannelId == 0 {
				log.Error().Msgf("Forward received for a non existing channel (outgoingShortChannelId: %v)",
					outgoingShortChannelId)
				outgoingChannelIdP = nil
			}
			_, err = stmt.Exec(convertMicro(int64(event.TimestampNs)), event.TimestampNs, event.FeeMsat,
				event.AmtInMsat, event.AmtOutMsat, incomingChannelIdP, outgoingChannelIdP, nodeId)
			if err != nil {
				return errors.Wrapf(err, "storeForwardingHistory->tx.Exec(%v)", event)
			}
			forwardEvents = append(forwardEvents, core.ForwardEvent{
				EventData: core.EventData{
					EventTime: time.Now().UTC(),
					NodeId:    nodeId,
				},
				Timestamp:         convertMicro(int64(event.TimestampNs)),
				FeeMsat:           event.FeeMsat,
				AmountInMsat:      event.AmtInMsat,
				AmountOutMsat:     event.AmtOutMsat,
				IncomingChannelId: incomingChannelIdP,
				OutgoingChannelId: outgoingChannelIdP,
			})
		}
		defer stmt.Close()
		err = tx.Commit()
		if err != nil {
			return errors.Wrap(err, "DB Commit")
		}
		if !bootStrapping {
			for _, forwardEvent := range forwardEvents {
				ProcessForwardEvent(forwardEvent)
			}
		}
	}

	return nil
}

// fetchLastForwardTime fetches the latest recorded forward, if none is set already.
// This should only run once when a server starts.
func fetchLastForwardTime(db *sqlx.DB, nodeId int) (uint64, error) {

	var lastNs uint64

	row := db.QueryRow("SELECT time_ns FROM forward WHERE node_id = $1 ORDER BY time_ns DESC LIMIT 1;", nodeId)
	err := row.Scan(&lastNs)

	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, errors.Wrap(err, "Query row of last forward time")
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
func SubscribeForwardingEvents(ctx context.Context,
	client lightningClientForwardingHistory,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache,
	opt *FwhOptions) {

	serviceType := services_helpers.LndServiceForwardsService

	maxEvents := streamLndMaxForwards
	bootStrapping := true

	// Check if maxEvents has been set and that it is bellow the hard coded maximum defined by
	// the constant MAXEVENTS.
	if (opt != nil) && ((*opt.MaxEvents > streamLndMaxForwards) || (*opt.MaxEvents <= 0)) {
		maxEvents = *opt.MaxEvents
	}

	ticker := time.NewTicker(streamForwardsTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerChannel := ticker.C
	// If a custom ticker is set in the options, override the default ticker.
	if (opt != nil) && (opt.Tick != nil) {
		tickerChannel = opt.Tick
	}

	rl := ratelimit.New(1) // 1 per second maximum rate limit

	var enforcedReferenceDate *time.Time
	importHistoricForwards := cache.HasCustomSetting(nodeSettings.NodeId, core.ImportHistoricForwards)
	if !importHistoricForwards {
		log.Info().Msgf("Import of historic forwards is disabled for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
		enforcedReferenceDateO := time.Now()
		enforcedReferenceDate = &enforcedReferenceDateO
		bootStrapping = false
	}

	// Request the forwarding history at the requested interval.
	// NB!: This timer is slowly being shifted because of the time required to
	//fetch and store the response.
	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		case <-tickerChannel:
			importCounter := 0
			// Keep fetching until LND returns less than the max number of records requested.
			for {
				rl.Take() // rate limited to 1 per second, when caught up will normally be 1 every 10 seconds

				// Fetch the nanosecond timestamp of the most recent record we have.
				lastNs, err := fetchLastForwardTime(db, nodeSettings.NodeId)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to obtain last know forward for nodeId: %v", nodeSettings.NodeId)
					cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
					return
				}
				if enforcedReferenceDate != nil && lastNs == 0 {
					lastNs = uint64(enforcedReferenceDate.UnixNano())
				}
				lastTimestamp := lastNs / uint64(time.Second)

				// if not bootstrapping and we have imported at least one forward, offset + 1 to prevent getting the same forward record again and again
				offset := uint32(0)
				if lastNs != 0 && !bootStrapping {
					offset = 1
				}

				fwhReq := &lnrpc.ForwardingHistoryRequest{
					StartTime:    lastTimestamp,
					IndexOffset:  offset,
					NumMaxEvents: uint32(maxEvents),
				}
				fwh, err := client.ForwardingHistory(ctx, fwhReq)
				if err != nil {
					if errors.Is(ctx.Err(), context.Canceled) {
						cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
						return
					}
					log.Error().Err(err).Msgf("Failed to obtain forwards for nodeId: %v", nodeSettings.NodeId)
					cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
					return
				}

				if bootStrapping {
					cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)
				}
				// Store the forwarding history
				err = storeForwardingHistory(db, fwh.ForwardingEvents, nodeSettings.NodeId, bootStrapping)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to store forward event for nodeId: %v", nodeSettings.NodeId)
				}

				// Stop fetching if there are fewer forwards than max requested
				// (indicates that we have the last forwarding record)
				importCounter += len(fwh.ForwardingEvents)
				if len(fwh.ForwardingEvents) < maxEvents {
					if bootStrapping {
						log.Info().Msgf("Bulk import of forward done (%v)", importCounter)
						cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
					}
					bootStrapping = false
					break
				} else {
					log.Info().Msgf("Still running bulk import of forward events (%v)", importCounter)
				}
			}
		}
	}
}
