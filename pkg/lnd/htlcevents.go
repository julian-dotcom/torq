package lnd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"

	"github.com/lncapital/torq/pkg/commons"

	"github.com/rs/zerolog/log"
)

type HtlcEvent struct {
	Time              time.Time `json:"time" db:"time"`
	Data              string    `json:"data" db:"data"`
	EventOrigin       *string   `json:"eventOrigin" db:"event_origin"`
	EventType         *string   `json:"eventType" db:"event_type"`
	OutgoingHtlcId    *uint64   `json:"outgoingHtlcId" db:"outgoing_htlc_id"`
	IncomingHtlcId    *uint64   `json:"incomingHtlcId" db:"incoming_htlc_id"`
	TimestampNs       *uint64   `json:"timestampNs" db:"timestamp_ns"`
	IncomingAmtMsat   *uint64   `json:"incomingAmtMsat" db:"incoming_amt_msat"`
	OutgoingAmtMsat   *uint64   `json:"outgoingAmtMsat" db:"outgoing_amt_msat"`
	IncomingTimelock  *uint32   `json:"incomingTimelock" db:"incoming_timelock"`
	OutgoingTimelock  *uint32   `json:"outgoingTimelock" db:"outgoing_timelock"`
	BoltFailureCode   *string   `json:"boltFailureCode" db:"bolt_failure_code"`
	BoltFailureString *string   `json:"boltFailureString" db:"bolt_failure_string"`
	LndFailureDetail  *string   `json:"lndFailureDetail" db:"lnd_failure_detail"`
	OutgoingChannelId *int      `json:"outgoingChannelId" db:"outgoing_channel_id"`
	IncomingChannelId *int      `json:"incomingChannelId" db:"incoming_channel_id"`
	NodeId            int       `json:"nodeId" db:"node_id"`
}

func storeLinkFailEvent(db *sqlx.DB, h *routerrpc.HtlcEvent, nodeId int) (HtlcEvent, error) {
	return storeFullEvent(db, h, nodeId, "LinkFailEvent")
}

func storeSettleEvent(db *sqlx.DB, h *routerrpc.HtlcEvent, nodeId int) (HtlcEvent, error) {
	return storeFullEvent(db, h, nodeId, "SettleEvent")
}

func storeForwardFailEvent(db *sqlx.DB, h *routerrpc.HtlcEvent, nodeId int) (HtlcEvent, error) {
	return storeFullEvent(db, h, nodeId, "ForwardFailEvent")
}

func storeForwardEvent(db *sqlx.DB, h *routerrpc.HtlcEvent, nodeId int) (HtlcEvent, error) {
	return storeFullEvent(db, h, nodeId, "ForwardEvent")
}

func storeFullEvent(db *sqlx.DB, h *routerrpc.HtlcEvent, nodeId int, eventType string) (HtlcEvent, error) {
	jb, err := json.Marshal(h)
	if err != nil {
		return HtlcEvent{}, errors.Wrapf(err, "Marshalling HTLC Event (%v) %v", eventType, h)
	}
	timestampMs := convertMicro(int64(h.TimestampNs))
	incomingChannelId := getChannelIdByLndShortChannelId(h.IncomingChannelId)
	outgoingChannelId := getChannelIdByLndShortChannelId(h.OutgoingChannelId)
	eventOrigin := h.EventType.String()
	htlcEvent := HtlcEvent{
		Time:              timestampMs,
		Data:              string(jb),
		EventOrigin:       &eventOrigin,
		EventType:         &eventType,
		OutgoingHtlcId:    &h.OutgoingHtlcId,
		IncomingHtlcId:    &h.IncomingHtlcId,
		TimestampNs:       &h.TimestampNs,
		OutgoingChannelId: outgoingChannelId,
		IncomingChannelId: incomingChannelId,
		NodeId:            nodeId,
	}
	if "LinkFailEvent" == eventType {
		htlcEvent.IncomingAmtMsat = &h.GetLinkFailEvent().Info.IncomingAmtMsat
		htlcEvent.OutgoingAmtMsat = &h.GetLinkFailEvent().Info.OutgoingAmtMsat
		htlcEvent.IncomingTimelock = &h.GetLinkFailEvent().Info.IncomingTimelock
		htlcEvent.OutgoingTimelock = &h.GetLinkFailEvent().Info.OutgoingTimelock
		boltFailureCode := h.GetLinkFailEvent().WireFailure.String()
		lndFailureDetail := h.GetLinkFailEvent().FailureDetail.String()
		htlcEvent.BoltFailureCode = &boltFailureCode
		htlcEvent.BoltFailureString = &h.GetLinkFailEvent().FailureString
		htlcEvent.LndFailureDetail = &lndFailureDetail
	}
	if "ForwardEvent" == eventType {
		htlcEvent.IncomingAmtMsat = &h.GetForwardEvent().Info.IncomingAmtMsat
		htlcEvent.OutgoingAmtMsat = &h.GetForwardEvent().Info.OutgoingAmtMsat
		htlcEvent.IncomingTimelock = &h.GetForwardEvent().Info.IncomingTimelock
		htlcEvent.OutgoingTimelock = &h.GetForwardEvent().Info.OutgoingTimelock
	}
	err = addHtlcEvent(db, htlcEvent)
	if err != nil {
		return HtlcEvent{}, errors.Wrapf(err, "Storing HTLC Event (%v)", eventType)
	}
	return htlcEvent, nil
}

func getChannelIdByLndShortChannelId(lndShortChannelId uint64) *int {
	var channelId *int
	shortChannelId := commons.ConvertLNDShortChannelID(lndShortChannelId)
	tempChannelId := commons.GetChannelIdByShortChannelId(shortChannelId)
	if tempChannelId != 0 {
		channelId = &tempChannelId
	}
	return channelId
}

func addHtlcEvent(db *sqlx.DB, htlcEvent HtlcEvent) error {
	_, err := db.Exec(`
		INSERT INTO htlc_event (
			time,
			event_origin,
			timestamp_ns,
			data,
			event_type,
			incoming_amt_msat,
			outgoing_amt_msat,
			incoming_timelock,
			Outgoing_timelock,
			outgoing_htlc_id,
			incoming_htlc_id,
			bolt_failure_code,
			bolt_failure_string,
			lnd_failure_detail,
			incoming_channel_id,
			outgoing_channel_id,
			node_id
		) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		htlcEvent.Time,
		htlcEvent.EventOrigin,
		htlcEvent.TimestampNs,
		htlcEvent.Data,
		htlcEvent.EventType,
		htlcEvent.IncomingAmtMsat,
		htlcEvent.OutgoingAmtMsat,
		htlcEvent.IncomingTimelock,
		htlcEvent.OutgoingTimelock,
		htlcEvent.OutgoingHtlcId,
		htlcEvent.IncomingHtlcId,
		htlcEvent.BoltFailureCode,
		htlcEvent.BoltFailureString,
		htlcEvent.LndFailureDetail,
		htlcEvent.IncomingChannelId,
		htlcEvent.OutgoingChannelId,
		htlcEvent.NodeId,
	)
	return errors.Wrap(err, "DB Exec adding HTLC Event")
}

// SubscribeAndStoreHtlcEvents subscribes to HTLC events from LND and stores them in the database as time series.
// NB: LND has marked HTLC event streaming as experimental. Delivery is not guaranteed, so dataset might not be complete
// HTLC events is primarily used to diagnose how good a channel / node is. And if the channel allocation should change.
func SubscribeAndStoreHtlcEvents(ctx context.Context, router routerrpc.RouterClient, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, htlcEventChannel chan commons.HtlcEvent,
	serviceEventChannel chan commons.ServiceEvent) {
	var stream routerrpc.Router_SubscribeHtlcEventsClient
	var err error
	var htlcEvent *routerrpc.HtlcEvent
	var storedHtlcEvent HtlcEvent
	serviceStatus := commons.Inactive
	subscriptionStream := commons.HtlcEventStream

	defer log.Info().Msgf("SubscribeAndStoreHtlcEvents terminated for nodeId: %v", nodeSettings.NodeId)

	importHtlcEvents := commons.RunningServices[commons.LndService].HasCustomSetting(nodeSettings.NodeId, commons.ImportHtlcEvents)
	if !importHtlcEvents {
		log.Info().Msgf("Import of HTLC events is disabled for nodeId: %v", nodeSettings.NodeId)
		SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Deleted, serviceStatus)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if stream == nil {
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
			stream, err = router.SubscribeHtlcEvents(ctx, &routerrpc.SubscribeHtlcEventsRequest{})
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				log.Error().Err(err).Msgf("Obtaining stream (SubscribeTransactions) from LND failed, will retry in %v seconds", commons.STREAM_ERROR_SLEEP_SECONDS)
				stream = nil
				time.Sleep(commons.STREAM_ERROR_SLEEP_SECONDS * time.Second)
				continue
			}
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
		}

		htlcEvent, err = stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
			log.Error().Err(err).Msgf("Receiving htlc events from the stream failed, will retry in %v seconds", commons.STREAM_ERROR_SLEEP_SECONDS)
			stream = nil
			time.Sleep(commons.STREAM_ERROR_SLEEP_SECONDS * time.Second)
			continue
		}

		switch htlcEvent.Event.(type) {
		case *routerrpc.HtlcEvent_ForwardEvent:
			storedHtlcEvent, err = storeForwardEvent(db, htlcEvent, nodeSettings.NodeId)
			if err != nil {
				// TODO FIXME STORE THIS SOMEWHERE??? TRANSACTION IS NOW IGNORED???
				log.Error().Err(err).Msgf("Failed to store forward event of type HtlcEvent_ForwardEvent")
			}
		case *routerrpc.HtlcEvent_ForwardFailEvent:
			storedHtlcEvent, err = storeForwardFailEvent(db, htlcEvent, nodeSettings.NodeId)
			if err != nil {
				// TODO FIXME STORE THIS SOMEWHERE??? TRANSACTION IS NOW IGNORED???
				log.Error().Err(err).Msgf("Failed to store forward event of type HtlcEvent_ForwardFailEvent")
			}
		case *routerrpc.HtlcEvent_LinkFailEvent:
			storedHtlcEvent, err = storeLinkFailEvent(db, htlcEvent, nodeSettings.NodeId)
			if err != nil {
				// TODO FIXME STORE THIS SOMEWHERE??? TRANSACTION IS NOW IGNORED???
				log.Error().Err(err).Msgf("Failed to store forward event of type HtlcEvent_LinkFailEvent")
			}
		case *routerrpc.HtlcEvent_SettleEvent:
			storedHtlcEvent, err = storeSettleEvent(db, htlcEvent, nodeSettings.NodeId)
			if err != nil {
				// TODO FIXME STORE THIS SOMEWHERE??? TRANSACTION IS NOW IGNORED???
				log.Error().Err(err).Msgf("Failed to store forward event of type HtlcEvent_SettleEvent")
			}
		}

		if htlcEventChannel != nil {
			htlcEventChannel <- commons.HtlcEvent{
				EventData: commons.EventData{
					EventTime: time.Now().UTC(),
					NodeId:    nodeSettings.NodeId,
				},
				Timestamp:         storedHtlcEvent.Time,
				EventOrigin:       storedHtlcEvent.EventOrigin,
				EventType:         storedHtlcEvent.EventType,
				OutgoingHtlcId:    storedHtlcEvent.OutgoingHtlcId,
				IncomingHtlcId:    storedHtlcEvent.IncomingHtlcId,
				TimestampNs:       storedHtlcEvent.TimestampNs,
				IncomingAmtMsat:   storedHtlcEvent.IncomingAmtMsat,
				OutgoingAmtMsat:   storedHtlcEvent.OutgoingAmtMsat,
				IncomingTimelock:  storedHtlcEvent.IncomingTimelock,
				OutgoingTimelock:  storedHtlcEvent.OutgoingTimelock,
				BoltFailureCode:   storedHtlcEvent.BoltFailureCode,
				BoltFailureString: storedHtlcEvent.BoltFailureString,
				LndFailureDetail:  storedHtlcEvent.LndFailureDetail,
				OutgoingChannelId: storedHtlcEvent.OutgoingChannelId,
				IncomingChannelId: storedHtlcEvent.IncomingChannelId,
			}
		}
	}
}
