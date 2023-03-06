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

	"github.com/lncapital/torq/pkg/commons"
)

const streamForwardsTickerSeconds = 10
const streamLndMaxForwards = 50000

// storeForwardingHistory
func storeForwardingHistory(db *sqlx.DB, fwh []*lnrpc.ForwardingEvent, nodeId int,
	forwardEventChannel chan<- commons.ForwardEvent, bootStrapping bool) error {

	if len(fwh) > 0 {
		var forwardEvents []commons.ForwardEvent
		tx := db.MustBegin()
		stmt, err := tx.Prepare(`INSERT INTO forward(time, time_ns, fee_msat,
				incoming_amount_msat, outgoing_amount_msat, incoming_channel_id, outgoing_channel_id, node_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (time, time_ns) DO NOTHING;`)
		if err != nil {
			return errors.Wrap(err, "SQL Statement prepare")
		}
		for _, event := range fwh {
			incomingShortChannelId := commons.ConvertLNDShortChannelID(event.ChanIdIn)
			incomingChannelId := commons.GetChannelIdByShortChannelId(incomingShortChannelId)
			incomingChannelIdP := &incomingChannelId
			if incomingChannelId == 0 {
				log.Error().Msgf("Forward received for a non existing channel (incomingChannelIdP: %v)",
					incomingShortChannelId)
				incomingChannelIdP = nil
			}
			outgoingShortChannelId := commons.ConvertLNDShortChannelID(event.ChanIdOut)
			outgoingChannelId := commons.GetChannelIdByShortChannelId(outgoingShortChannelId)
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
			forwardEvents = append(forwardEvents, commons.ForwardEvent{
				EventData: commons.EventData{
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
		if forwardEventChannel != nil && !bootStrapping {
			for _, forwardEvent := range forwardEvents {
				forwardEventChannel <- forwardEvent
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
func SubscribeForwardingEvents(ctx context.Context, client lightningClientForwardingHistory, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, forwardEventChannel chan<- commons.ForwardEvent,
	serviceEventChannel chan<- commons.ServiceEvent, opt *FwhOptions) {

	defer log.Info().Msgf("SubscribeForwardingEvents terminated for nodeId: %v", nodeSettings.NodeId)

	maxEvents := streamLndMaxForwards
	serviceStatus := commons.ServiceInactive
	bootStrapping := true
	subscriptionStream := commons.ForwardStream
	ticker := clock.New().Tick(streamForwardsTickerSeconds * time.Second)

	// Check if maxEvents has been set and that it is bellow the hard coded maximum defined by
	// the constant MAXEVENTS.
	if (opt != nil) && ((*opt.MaxEvents > streamLndMaxForwards) || (*opt.MaxEvents <= 0)) {
		maxEvents = *opt.MaxEvents
	}

	// If a custom ticker is set in the options, override the default ticker.
	if (opt != nil) && (opt.Tick != nil) {
		ticker = opt.Tick
	}

	rl := ratelimit.New(1) // 1 per second maximum rate limit

	importForwards := commons.RunningServices[commons.LndService].HasCustomSetting(nodeSettings.NodeId, commons.ImportForwards)
	if !importForwards {
		log.Info().Msgf("Import of forwards is disabled for nodeId: %v", nodeSettings.NodeId)
		SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.ServiceDeleted, serviceStatus)
		return
	}

	var enforcedReferenceDate *time.Time
	importHistoricForwards := commons.RunningServices[commons.LndService].HasCustomSetting(nodeSettings.NodeId, commons.ImportHistoricForwards)
	if !importHistoricForwards {
		log.Info().Msgf("Import of historic forwards is disabled for nodeId: %v", nodeSettings.NodeId)
		serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.ServiceActive, serviceStatus)
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
			return
		case <-ticker:
			importCounter := 0
			// Keep fetching until LND returns less than the max number of records requested.
			for {
				rl.Take() // rate limited to 1 per second, when caught up will normally be 1 every 10 seconds

				// Fetch the nanosecond timestamp of the most recent record we have.
				lastNs, err := fetchLastForwardTime(db, nodeSettings.NodeId)
				if err != nil {
					serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.ServicePending, serviceStatus)
					log.Error().Err(err).Msgf("Failed to obtain last know forward, will retry in %v seconds", streamForwardsTickerSeconds)
					break
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
						return
					}
					serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.ServicePending, serviceStatus)
					log.Error().Err(err).Msgf("Failed to obtain forwards, will retry in %v seconds", streamForwardsTickerSeconds)
					break
				}

				if bootStrapping {
					serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.ServiceInitializing, serviceStatus)
				} else {
					serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.ServiceActive, serviceStatus)
				}
				// Store the forwarding history
				err = storeForwardingHistory(db, fwh.ForwardingEvents, nodeSettings.NodeId, forwardEventChannel, bootStrapping)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to store forward event")
				}

				// Stop fetching if there are fewer forwards than max requested
				// (indicates that we have the last forwarding record)
				importCounter += len(fwh.ForwardingEvents)
				if len(fwh.ForwardingEvents) < maxEvents {
					if bootStrapping {
						log.Info().Msgf("Bulk import of forward done (%v)", importCounter)
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
