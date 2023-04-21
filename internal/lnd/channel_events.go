package lnd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/internal/vector"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var ChannelChanges = make(chan core.ChannelEvent) //nolint:gochecknoglobals

func chanPointFromByte(cb []byte, oi uint32) (string, error) {
	ch, err := chainhash.NewHash(cb)
	if err != nil {
		return "", errors.Wrap(err, "Chainhash newhash")
	}
	return fmt.Sprintf("%s:%d", ch.String(), oi), nil
}

// storeChannelEvent extracts the timestamp, channel ID and PubKey from the
// ChannelEvent and converts the original struct to json.
// Then it's stored in the database in the channel_event table.
func storeChannelEvent(ctx context.Context,
	db *sqlx.DB,
	ce *lnrpc.ChannelEventUpdate,
	nodeSettings cache.NodeSettingsCache) error {

	timestampMs := time.Now().UTC()

	channelEvent := core.ChannelEvent{
		EventData: core.EventData{
			EventTime: timestampMs,
			NodeId:    nodeSettings.NodeId,
		},
		Type: ce.GetType(),
	}

	switch ce.Type {
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		c := ce.GetOpenChannel()
		remotePublicKey := c.RemotePubkey
		remoteNodeId, err := addNodeWhenNew(db, remotePublicKey, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "OPEN_CHANNEL: Add Node When New")
		}
		channelStatus := core.Open
		var initiatingNodeId *int
		if c.Initiator {
			initiatingNodeId = &nodeSettings.NodeId
		}
		channel, err := addChannelOrUpdateStatus(c.ChannelPoint, c.ChanId, &channelStatus, c.Capacity, &c.Private,
			nil, nil, nodeSettings, remoteNodeId, initiatingNodeId, nil, db)
		if err != nil {
			return errors.Wrap(err, "OPEN_CHANNEL: Add Channel Or Update Status")
		}

		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "OPEN_CHANNEL: JSON Marshall")
		}

		// This allows torq to listen to the graph for node updates
		cache.SetChannelPeerNode(remoteNodeId, remotePublicKey, nodeSettings.Chain, nodeSettings.Network, channel.Status)

		stateSettings := cache.ChannelStateSettingsCache{
			NodeId:                nodeSettings.NodeId,
			RemoteNodeId:          remoteNodeId,
			ChannelId:             channel.ChannelID,
			LocalBalance:          c.LocalBalance,
			LocalDisabled:         c.Active,
			RemoteBalance:         c.RemoteBalance,
			CommitFee:             c.CommitFee,
			CommitWeight:          c.CommitWeight,
			FeePerKw:              c.FeePerKw,
			NumUpdates:            c.NumUpdates,
			ChanStatusFlags:       c.ChanStatusFlags,
			CommitmentType:        c.CommitmentType,
			Lifetime:              c.Lifetime,
			TotalSatoshisReceived: c.TotalSatoshisReceived,
			TotalSatoshisSent:     c.TotalSatoshisSent,
		}
		if c.LocalConstraints != nil {
			stateSettings.LocalMinHtlcMsat = c.LocalConstraints.MinHtlcMsat
			stateSettings.LocalMaxHtlcMsat = c.LocalConstraints.MaxPendingAmtMsat
			stateSettings.LocalTimeLockDelta = c.LocalConstraints.CsvDelay
		}
		if c.RemoteConstraints != nil {
			stateSettings.RemoteMinHtlcMsat = c.RemoteConstraints.MinHtlcMsat
			stateSettings.RemoteMaxHtlcMsat = c.RemoteConstraints.MaxPendingAmtMsat
			stateSettings.RemoteTimeLockDelta = c.RemoteConstraints.CsvDelay
		}
		cache.SetChannelState(nodeSettings.NodeId, stateSettings)

		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channel.ChannelID, false, jsonByteArray,
			channelEvent)
		if err != nil {
			return errors.Wrap(err, "Insert Open Channel Event")
		}
		return nil
	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
		c := ce.GetClosedChannel()
		remotePublicKey := c.RemotePubkey
		remoteNodeId, err := addNodeWhenNew(db, remotePublicKey, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "CLOSED_CHANNEL: Add Node When New")
		}
		var initiatingNodeId *int
		if c.OpenInitiator == lnrpc.Initiator_INITIATOR_LOCAL {
			initiatingNodeId = &nodeSettings.NodeId
		} else if c.OpenInitiator == lnrpc.Initiator_INITIATOR_REMOTE {
			initiatingNodeId = &remoteNodeId
		}
		var closingNodeId *int
		if c.CloseInitiator == lnrpc.Initiator_INITIATOR_LOCAL {
			closingNodeId = &nodeSettings.NodeId
		} else if c.CloseInitiator == lnrpc.Initiator_INITIATOR_REMOTE {
			closingNodeId = &remoteNodeId
		}
		channel, err := addChannelOrUpdateStatus(c.ChannelPoint, c.ChanId, nil, c.Capacity, nil,
			&c.CloseType, &c.ClosingTxHash, nodeSettings, remoteNodeId, initiatingNodeId, closingNodeId, db)
		if err != nil {
			return errors.Wrap(err, "CLOSED_CHANNEL: Add Channel Or Update Status")
		}

		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "CLOSED_CHANNEL: JSON Marshall")
		}

		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channel.ChannelID, false, jsonByteArray,
			channelEvent)
		if err != nil {
			return errors.Wrap(err, "Insert Closed Channel Event")
		}

		// This stops the graph from listening to node updates
		chans, err := channels.GetOpenChannelsForNodeId(db, remoteNodeId)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to verify if remote node still has open channels: %v", remoteNodeId)
		}

		// This stops the graph from listening to channel updates
		cache.SetChannelStatus(channel.ChannelID, channel.Status)
		if len(chans) == 0 {
			cache.SetInactiveChannelPeerNode(remoteNodeId, remotePublicKey, nodeSettings.Chain, nodeSettings.Network)
		}
		return nil
	case lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL:
		c := ce.GetActiveChannel()
		channelPoint, err := chanPointFromByte(c.GetFundingTxidBytes(), c.GetOutputIndex())
		if err != nil {
			return errors.Wrap(err, "ACTIVE_CHANNEL: Get channelPoint from bytes")
		}
		channelId := cache.GetChannelIdByChannelPoint(channelPoint)
		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "ACTIVE_CHANNEL: JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jsonByteArray,
			channelEvent)
		if err != nil {
			return errors.Wrap(err, "Insert Active Channel Event")
		}
		return nil
	case lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL:
		// We receive this event in case of a closure. So let's ask LND for a fresh copy of the pending channels.
		err := importPendingChannels(db, false, nodeSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to import pending channels for nodeId: %v", nodeSettings.NodeId)
		}
		c := ce.GetInactiveChannel()
		channelPoint, err := chanPointFromByte(c.GetFundingTxidBytes(), c.GetOutputIndex())
		if err != nil {
			return errors.Wrap(err, "INACTIVE_CHANNEL: Get channelPoint from bytes")
		}
		channelId := cache.GetChannelIdByChannelPoint(channelPoint)
		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "INACTIVE_CHANNEL: JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jsonByteArray,
			channelEvent)
		if err != nil {
			return errors.Wrap(err, "Insert Inactive Channel Event")
		}
		return nil
	case lnrpc.ChannelEventUpdate_FULLY_RESOLVED_CHANNEL:
		err := importPendingChannels(db, true, nodeSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to import pending channels for nodeId: %v", nodeSettings.NodeId)
		}
		c := ce.GetFullyResolvedChannel()
		channelPoint, err := chanPointFromByte(c.GetFundingTxidBytes(), c.GetOutputIndex())
		if err != nil {
			return errors.Wrap(err, "FULLY_RESOLVED_CHANNEL: Get channelPoint from bytes")
		}
		channelId := cache.GetChannelIdByChannelPoint(channelPoint)
		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "FULLY_RESOLVED_CHANNEL: JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jsonByteArray,
			channelEvent)
		if err != nil {
			return errors.Wrap(err, "Insert Fully Resolved Channel Event")
		}
		return nil
	case lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL:
		err := importPendingChannels(db, true, nodeSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to import pending channels for nodeId: %v", nodeSettings.NodeId)
		}
		c := ce.GetPendingOpenChannel()
		channelPoint, err := chanPointFromByte(c.GetTxid(), c.GetOutputIndex())
		if err != nil {
			return errors.Wrap(err, "PENDING_OPEN_CHANNEL: Get channelPoint from bytes")
		}
		channelId := cache.GetChannelIdByChannelPoint(channelPoint)
		if channelId == 0 {
			log.Debug().Msgf("Could not store channel event since we only received funding transaction and output index (%v) and nothing more.",
				channelPoint)
			return nil
		}
		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "PENDING_OPEN_CHANNEL: JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jsonByteArray,
			channelEvent)
		if err != nil {
			return errors.Wrap(err, "Insert Pending Open Channel Event")
		}
		return nil
	}
	return nil
}

func importPendingChannels(db *sqlx.DB, force bool, nodeSettings cache.NodeSettingsCache) error {
	err := ImportPendingChannels(db, force, nodeSettings.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain pending channels for nodeId: %v", nodeSettings.NodeId)
		return errors.Wrapf(err, "Obtaining pending channels for nodeId: %v", nodeSettings.NodeId)
	}
	return nil
}

func addChannelOrUpdateStatus(channelPoint string, lndShortChannelId uint64, channelStatus *core.ChannelStatus,
	capacity int64, private *bool, closeType *lnrpc.ChannelCloseSummary_ClosureType, closingTxHash *string,
	nodeSettings cache.NodeSettingsCache, remoteNodeId int, initiatingNodeId *int, closingNodeId *int,
	db *sqlx.DB) (channels.Channel, error) {

	fundingTransactionHash, fundingOutputIndex := core.ParseChannelPoint(channelPoint)
	channel := channels.Channel{
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		FirstNodeId:            nodeSettings.NodeId,
		SecondNodeId:           remoteNodeId,
		InitiatingNodeId:       initiatingNodeId,
		ClosingNodeId:          closingNodeId,
		Capacity:               capacity,
	}
	if private != nil {
		channel.Private = *private
	}
	if initiatingNodeId != nil {
		if *initiatingNodeId == remoteNodeId {
			channel.AcceptingNodeId = &nodeSettings.NodeId
		}
		if *initiatingNodeId == nodeSettings.NodeId {
			channel.AcceptingNodeId = &remoteNodeId
		}
	}
	if channelStatus != nil {
		channel.Status = *channelStatus
	}
	if closeType != nil {
		channel.Status = channels.GetClosureStatus(*closeType)
	}
	if closingTxHash != nil {
		channel.ClosingTransactionHash = closingTxHash
	}
	shortChannelId := core.ConvertLNDShortChannelID(lndShortChannelId)
	if lndShortChannelId != 0 {
		channel.LNDShortChannelID = &lndShortChannelId
		channel.ShortChannelID = &shortChannelId
	}
	var err error
	channel.ChannelID, err = channels.AddChannelOrUpdateChannelStatus(db, nodeSettings, channel)
	if err != nil {
		return channels.Channel{}, errors.Wrapf(err, "Adding or updating channel (channelId: %v, shortChannelId: %v)", channel.ChannelID, shortChannelId)
	}
	return channel, nil
}

func addNodeWhenNew(db *sqlx.DB, remotePublicKey string, nodeSettings cache.NodeSettingsCache) (int, error) {
	remoteNodeId := cache.GetPeerNodeIdByPublicKey(remotePublicKey, nodeSettings.Chain, nodeSettings.Network)
	if remoteNodeId == 0 {
		newNode := nodes.Node{
			PublicKey: remotePublicKey,
			Chain:     nodeSettings.Chain,
			Network:   nodeSettings.Network,
		}

		peerConnectionHistory := &nodes.NodeConnectionHistory{
			TorqNodeId:       nodeSettings.NodeId,
			ConnectionStatus: core.NodeConnectionStatusConnected,
		}

		var err error
		remoteNodeId, err = nodes.AddNodeWhenNew(db, newNode, peerConnectionHistory)
		if err != nil {
			return 0, errors.Wrapf(err, "Adding node with public key: %v", remotePublicKey)
		}
	}
	return remoteNodeId, nil
}

type lndClientChannelEvent interface {
	ListChannels(ctx context.Context, in *lnrpc.ListChannelsRequest,
		opts ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error)
	GetChanInfo(ctx context.Context, in *lnrpc.ChanInfoRequest,
		opts ...grpc.CallOption) (*lnrpc.ChannelEdge, error)
}

type lndClientSubscribeChannelEvent interface {
	lndClientChannelEvent
	SubscribeChannelEvents(ctx context.Context, in *lnrpc.ChannelEventSubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error)
	PendingChannels(ctx context.Context, in *lnrpc.PendingChannelsRequest,
		opts ...grpc.CallOption) (*lnrpc.PendingChannelsResponse, error)
}

// SubscribeAndStoreChannelEvents Subscribes to channel events from LND and
// stores them in the database as a time series
func SubscribeAndStoreChannelEvents(ctx context.Context,
	client lndClientSubscribeChannelEvent,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.LndServiceChannelEventStream

	stream, err := client.SubscribeChannelEvents(ctx, &lnrpc.ChannelEventSubscription{})
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		}
		log.Error().Err(err).Msgf(
			"%v failure to obtain a stream from LND for nodeId: %v", serviceType.String(), nodeSettings.NodeId)
		cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
		return
	}

	cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		default:
		}

		chanEvent, err := stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
				return
			}
			log.Error().Err(err).Msgf(
				"Receiving channel events from the stream failed for nodeId: %v", nodeSettings.NodeId)
			cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		}

		err = storeChannelEvent(ctx, db, chanEvent, nodeSettings)
		if err != nil {
			// TODO FIXME STORE THIS SOMEWHERE??? CHANNELEVENT IS NOW IGNORED???
			log.Error().Err(err).Msgf("Storing channel event failed for nodeId: %v", nodeSettings.NodeId)
		}
	}
}

func ImportPendingChannelsFromLnd(ctx context.Context, db *sqlx.DB, client lnrpc.LightningClient,
	nodeSettings cache.NodeSettingsCache) error {
	r, err := client.PendingChannels(ctx, &lnrpc.PendingChannelsRequest{})
	if err != nil {
		return errors.Wrap(err, "LND: Get pending channels")
	}

	err = storeImportedWaitingCloseChannels(db, r.WaitingCloseChannels, nodeSettings)
	if err != nil {
		return errors.Wrap(err, "Store imported waiting close channels")
	}
	err = storeImportedPendingOpenChannels(db, r.PendingOpenChannels, nodeSettings)
	if err != nil {
		return errors.Wrap(err, "Store imported pending open channels")
	}
	err = storeImportedPendingForceClosingChannels(db, r.PendingForceClosingChannels, nodeSettings)
	if err != nil {
		return errors.Wrap(err, "Store imported pending force closing channels")
	}
	return nil
}

func ImportOpenChannels(ctx context.Context, db *sqlx.DB, client lnrpc.LightningClient,
	nodeSettings cache.NodeSettingsCache) error {
	r, err := client.ListChannels(ctx, &lnrpc.ListChannelsRequest{})
	if err != nil {
		return errors.Wrap(err, "LND: List channels")
	}

	err = storeImportedOpenChannels(db, r.Channels, nodeSettings)
	if err != nil {
		return errors.Wrap(err, "Store imported open channels")
	}
	return nil
}

func ImportClosedChannels(ctx context.Context, db *sqlx.DB, client lnrpc.LightningClient,
	nodeSettings cache.NodeSettingsCache) error {
	r, err := client.ClosedChannels(ctx, &lnrpc.ClosedChannelsRequest{})
	if err != nil {
		return errors.Wrap(err, "LND: Get closed channels")
	}

	err = storeImportedClosedChannels(db, r.Channels, nodeSettings)
	if err != nil {
		return errors.Wrap(err, "Store imported closed channels")
	}
	return nil
}

func getExistingChannelEvents(t lnrpc.ChannelEventUpdate_UpdateType, db *sqlx.DB, channelIds []int) ([]int, error) {
	if len(channelIds) == 0 {
		return []int{}, nil
	}
	// Prepare the query with an array of channel points
	q := `SELECT channel_id FROM channel_event WHERE channel_id IN (?) AND event_type = ?;`
	qs, args, err := sqlx.In(q, channelIds, t)
	if err != nil {
		return []int{}, errors.Wrap(err, "SQLX In")
	}

	// Query and create the list of existing channel points (ecp)
	var existingChannelIds []int
	qsr := db.Rebind(qs)
	rows, err := db.Query(qsr, args...)
	if err != nil {
		return []int{}, errors.Wrap(err, "DB Query")
	}
	defer rows.Close()
	for rows.Next() {
		var channelId int
		err = rows.Scan(&channelId)
		if err != nil {
			return nil, errors.Wrap(err, "SQL row scan for channel id")
		}
		existingChannelIds = append(existingChannelIds, channelId)
	}
	return existingChannelIds, nil
}

func storeImportedWaitingCloseChannels(
	db *sqlx.DB,
	waitingCloseChannels []*lnrpc.PendingChannelsResponse_WaitingCloseChannel,
	nodeSettings cache.NodeSettingsCache) error {

	if len(waitingCloseChannels) == 0 {
		return nil
	}

	for _, waitingCloseChannel := range waitingCloseChannels {
		lndChannel := waitingCloseChannel.GetChannel()
		closing := core.Closing

		remoteNodeId, initiatingNodeId, closeType, closingNodeId, err := processPendingChannel(db, lndChannel, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "ImportedWaitingCloseChannels: ProcessPendingChannel")
		}

		lndShortChannelId := processEmptyChanId(lndChannel.ChannelPoint, nodeSettings)

		var closingTransactionHash *string
		if waitingCloseChannel.ClosingTxid != "" {
			closingTransactionHash = &waitingCloseChannel.ClosingTxid
		}

		_, err = addChannelOrUpdateStatus(lndChannel.ChannelPoint, lndShortChannelId, &closing, lndChannel.Capacity, &lndChannel.Private,
			closeType, closingTransactionHash, nodeSettings, remoteNodeId, initiatingNodeId, closingNodeId, db)
		if err != nil {
			return errors.Wrap(err, "ImportedWaitingCloseChannels: Add Channel Or Update Status")
		}

		cache.SetChannelPeerNode(remoteNodeId, lndChannel.RemoteNodePub, nodeSettings.Chain, nodeSettings.Network, closing)
	}
	return nil
}

func processPendingChannel(db *sqlx.DB,
	lndChannel *lnrpc.PendingChannelsResponse_PendingChannel,
	nodeSettings cache.NodeSettingsCache) (int, *int, *lnrpc.ChannelCloseSummary_ClosureType, *int, error) {

	var initiatingNodeId *int
	var closingNodeId *int
	var closeType *lnrpc.ChannelCloseSummary_ClosureType

	remoteNodeId, err := addNodeWhenNew(db, lndChannel.RemoteNodePub, nodeSettings)
	if err != nil {
		return 0, nil, nil, nil, errors.Wrap(err, "Add Node When New")
	}

	if lndChannel.Initiator == lnrpc.Initiator_INITIATOR_LOCAL {
		initiatingNodeId = &nodeSettings.NodeId
	} else if lndChannel.Initiator == lnrpc.Initiator_INITIATOR_REMOTE {
		initiatingNodeId = &remoteNodeId
	}

	if strings.Contains(lndChannel.ChanStatusFlags, "ChanStatusLocalCloseInitiator") {
		closingNodeId = &nodeSettings.NodeId
	}
	if strings.Contains(lndChannel.ChanStatusFlags, "ChanStatusRemoteCloseInitiator") {
		closingNodeId = &remoteNodeId
	}
	if strings.Contains(lndChannel.ChanStatusFlags, "ChanStatusCoopBroadcasted") {
		cooperativeClose := lnrpc.ChannelCloseSummary_COOPERATIVE_CLOSE
		closeType = &cooperativeClose
	}

	// ChanStatusDefault is the normal state of an open channel.

	// ChanStatusCommitBroadcasted indicates that a commitment for this channel has been broadcasted.

	// ChanStatusBorked indicates that the channel has entered an
	// irreconcilable state, triggered by a state desynchronization or
	// channel breach. Channels in this state should never be added to the
	// htlc switch.

	// ChanStatusLocalDataLoss indicates that we have lost channel state
	// for this channel, and broadcasting our latest commitment might be
	// considered a breach.

	// ChanStatusRestored is a status flag that signals that the channel
	// has been restored, and doesn't have all the fields a typical channel
	// will have.

	return remoteNodeId, initiatingNodeId, closeType, closingNodeId, nil
}

func storeImportedPendingOpenChannels(
	db *sqlx.DB,
	pendingOpenChannels []*lnrpc.PendingChannelsResponse_PendingOpenChannel,
	nodeSettings cache.NodeSettingsCache) error {

	if len(pendingOpenChannels) == 0 {
		return nil
	}

	var channelIds []int
	for _, channel := range pendingOpenChannels {
		fundingTransactionHash, fundingOutputIndex := core.ParseChannelPoint(channel.GetChannel().ChannelPoint)
		channelId := cache.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
		if channelId != 0 {
			channelIds = append(channelIds, channelId)
		}
	}

	existingChannelIds, err := getExistingChannelEvents(lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL, db, channelIds)
	if err != nil {
		return err
	}

icoLoop:
	for _, pendingOpenChannel := range pendingOpenChannels {
		lndChannel := pendingOpenChannel.GetChannel()
		opening := core.Opening

		remoteNodeId, initiatingNodeId, _, _, err := processPendingChannel(db, lndChannel, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "ImportedPendingOpenChannels: ProcessPendingChannel")
		}

		channel, err := addChannelOrUpdateStatus(lndChannel.ChannelPoint, 0, &opening, lndChannel.Capacity, &lndChannel.Private,
			nil, nil, nodeSettings, remoteNodeId, initiatingNodeId, nil, db)
		if err != nil {
			return errors.Wrap(err, "ImportedPendingOpenChannels: Add Channel Or Update Status")
		}

		// skip if we already have channel close channel event for this channel
		for _, existingChannelId := range existingChannelIds {
			if channel.ChannelID == existingChannelId {
				continue icoLoop
			}
		}

		jsonByteArray, err := json.Marshal(lndChannel)
		if err != nil {
			return errors.Wrap(err, "ImportedPendingOpenChannels: JSON Marshal")
		}

		err = insertChannelEvent(db, time.Now().UTC(), lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL, nodeSettings.NodeId,
			channel.ChannelID, true, jsonByteArray, core.ChannelEvent{})
		if err != nil {
			return errors.Wrap(err, "ImportedPendingOpenChannels: Insert channel event")
		}

		cache.SetChannelPeerNode(remoteNodeId, lndChannel.RemoteNodePub, nodeSettings.Chain, nodeSettings.Network, opening)
	}
	return nil
}

func storeImportedPendingForceClosingChannels(
	db *sqlx.DB,
	pendingForceClosingChannels []*lnrpc.PendingChannelsResponse_ForceClosedChannel,
	nodeSettings cache.NodeSettingsCache) error {

	if len(pendingForceClosingChannels) == 0 {
		return nil
	}

	for _, pendingForceClosingChannel := range pendingForceClosingChannels {
		lndChannel := pendingForceClosingChannel.GetChannel()
		closing := core.Closing

		remoteNodeId, initiatingNodeId, closeType, closingNodeId, err := processPendingChannel(db, lndChannel, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "ImportedPendingForceClosingChannels: ProcessPendingChannel")
		}

		lndShortChannelId := processEmptyChanId(lndChannel.ChannelPoint, nodeSettings)

		var closingTransactionHash *string
		if pendingForceClosingChannel.ClosingTxid != "" {
			closingTransactionHash = &pendingForceClosingChannel.ClosingTxid
		}

		_, err = addChannelOrUpdateStatus(lndChannel.ChannelPoint, lndShortChannelId, &closing, lndChannel.Capacity, &lndChannel.Private,
			closeType, closingTransactionHash, nodeSettings, remoteNodeId, initiatingNodeId, closingNodeId, db)
		if err != nil {
			return errors.Wrap(err, "ImportedPendingForceClosingChannels: Add Channel Or Update Status")
		}

		cache.SetChannelPeerNode(remoteNodeId, lndChannel.RemoteNodePub, nodeSettings.Chain, nodeSettings.Network, closing)
	}
	return nil
}

func storeImportedOpenChannels(db *sqlx.DB, c []*lnrpc.Channel, nodeSettings cache.NodeSettingsCache) error {

	if len(c) == 0 {
		return nil
	}

	var channelIds []int
	for _, channel := range c {
		channelId := cache.GetChannelIdByChannelPoint(channel.ChannelPoint)
		if channelId != 0 {
			channelIds = append(channelIds, channelId)
		}
	}

	existingChannelIds, err := getExistingChannelEvents(lnrpc.ChannelEventUpdate_OPEN_CHANNEL, db, channelIds)
	if err != nil {
		return err
	}

icoLoop:
	for _, lndChannel := range c {

		remoteNodeId, err := addNodeWhenNew(db, lndChannel.RemotePubkey, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "ImportedOpenChannels: Add Node When New")
		}
		channelStatus := core.Open
		var initiatingNodeId *int
		if lndChannel.Initiator {
			initiatingNodeId = &nodeSettings.NodeId
		}
		if lndChannel.ChanId == 0 {
			lndChannel.ChanId = processEmptyChanId(lndChannel.ChannelPoint, nodeSettings)
		}

		channel, err := addChannelOrUpdateStatus(lndChannel.ChannelPoint, lndChannel.ChanId,
			&channelStatus, lndChannel.Capacity, &lndChannel.Private,
			nil, nil, nodeSettings, remoteNodeId, initiatingNodeId, nil, db)
		if err != nil {
			return errors.Wrap(err, "ImportedOpenChannels: Add Channel Or Update Status")
		}

		// skip if we have an existing channel open channel event
		for _, existingChannelId := range existingChannelIds {
			if channel.ChannelID == existingChannelId {
				continue icoLoop
			}
		}

		jsonByteArray, err := json.Marshal(lndChannel)
		if err != nil {
			return errors.Wrap(err, "ImportedOpenChannels: JSON Marshal")
		}

		err = insertChannelEvent(db, time.Now().UTC(), lnrpc.ChannelEventUpdate_OPEN_CHANNEL, nodeSettings.NodeId,
			channel.ChannelID, true, jsonByteArray, core.ChannelEvent{})
		if err != nil {
			return errors.Wrap(err, "ImportedOpenChannels: Insert channel event")
		}

		cache.SetChannelPeerNode(remoteNodeId, lndChannel.RemotePubkey, nodeSettings.Chain, nodeSettings.Network, core.Open)
	}
	return nil
}

func storeImportedClosedChannels(db *sqlx.DB, c []*lnrpc.ChannelCloseSummary,
	nodeSettings cache.NodeSettingsCache) error {

	if len(c) == 0 {
		return nil
	}

	var channelIds []int
	for _, channel := range c {
		fundingTransactionHash, fundingOutputIndex := core.ParseChannelPoint(channel.ChannelPoint)
		channelId := cache.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
		if channelId != 0 {
			channelIds = append(channelIds, channelId)
		}
	}

	existingChannelIds, err := getExistingChannelEvents(lnrpc.ChannelEventUpdate_CLOSED_CHANNEL, db, channelIds)
	if err != nil {
		return err
	}

icoLoop:
	for _, lndChannel := range c {

		remoteNodeId, err := addNodeWhenNew(db, lndChannel.RemotePubkey, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "ImportedClosedChannels: Add Node When New")
		}
		var initiatingNodeId *int
		if lndChannel.OpenInitiator == lnrpc.Initiator_INITIATOR_LOCAL {
			initiatingNodeId = &nodeSettings.NodeId
		} else if lndChannel.OpenInitiator == lnrpc.Initiator_INITIATOR_REMOTE {
			initiatingNodeId = &remoteNodeId
		}
		var closingNodeId *int
		if lndChannel.CloseInitiator == lnrpc.Initiator_INITIATOR_LOCAL {
			closingNodeId = &nodeSettings.NodeId
		} else if lndChannel.CloseInitiator == lnrpc.Initiator_INITIATOR_REMOTE {
			closingNodeId = &remoteNodeId
		}

		if lndChannel.ChanId == 0 {
			lndChannel.ChanId = processEmptyChanId(lndChannel.ChannelPoint, nodeSettings)
		}

		channel, err := addChannelOrUpdateStatus(lndChannel.ChannelPoint, lndChannel.ChanId, nil, lndChannel.Capacity, nil,
			&lndChannel.CloseType, &lndChannel.ClosingTxHash, nodeSettings, remoteNodeId, initiatingNodeId, closingNodeId, db)
		if err != nil {
			return errors.Wrap(err, "ImportedClosedChannels: Add Channel Or Update Status")
		}

		// skip if we already have channel close channel event for this channel
		for _, existingChannelId := range existingChannelIds {
			if channel.ChannelID == existingChannelId {
				continue icoLoop
			}
		}

		jsonByteArray, err := json.Marshal(lndChannel)
		if err != nil {
			return errors.Wrap(err, "ImportedClosedChannels: JSON Marshal")
		}

		err = insertChannelEvent(db, time.Now().UTC(), lnrpc.ChannelEventUpdate_CLOSED_CHANNEL, nodeSettings.NodeId,
			channel.ChannelID, true, jsonByteArray, core.ChannelEvent{})
		if err != nil {
			return errors.Wrap(err, "ImportedClosedChannels: Insert channel event")
		}

		cache.SetChannelPeerNode(remoteNodeId, lndChannel.RemotePubkey, nodeSettings.Chain, nodeSettings.Network, channels.GetClosureStatus(lndChannel.CloseType))
	}
	return nil
}

func processEmptyChanId(channelPoint string, nodeSettings cache.NodeSettingsCache) uint64 {
	fundingTransactionHash, fundingOutputIndex := core.ParseChannelPoint(channelPoint)
	channelId := cache.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
	if channelId != 0 {
		channelSettings := cache.GetChannelSettingByChannelId(channelId)
		if channelSettings.LndShortChannelId != nil && *channelSettings.LndShortChannelId != 0 {
			return *channelSettings.LndShortChannelId
		}
		if channelSettings.Status == core.AbandonedClosed || channelSettings.Status == core.FundingCancelledClosed {
			return 0
		}
	}

	if cache.GetVectorUrlBase() == vector.VectorUrl && (nodeSettings.Chain != core.Bitcoin || nodeSettings.Network != core.MainNet) {
		log.Info().Msgf("Skipping obtaining short channel id from vector for nodeId: %v", nodeSettings.NodeId)
		return 0
	}

	if fundingTransactionHash == nil || *fundingTransactionHash == "" || fundingOutputIndex == nil {
		log.Info().Msgf("No funding information for short channel id from vector for nodeId: %v", nodeSettings.NodeId)
		return 0
	}

	shortChannelId := vector.GetShortChannelIdFromVector(*fundingTransactionHash, *fundingOutputIndex, nodeSettings)
	if shortChannelId == "" {
		log.Error().Msgf("Failed to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
		return 0
	}
	lndShortChannelId, err := core.ConvertShortChannelIDToLND(shortChannelId)
	if err != nil {
		log.Error().Msgf("Failed (ConvertShortChannelIDToLND) to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
	}
	return lndShortChannelId
}

func insertChannelEvent(db *sqlx.DB, eventTime time.Time, eventType lnrpc.ChannelEventUpdate_UpdateType,
	nodeId, channelId int, imported bool, jsonByteArray []byte,
	channelEvent core.ChannelEvent) error {

	var sqlStm = `INSERT INTO channel_event (time, event_type, channel_id, imported, event, node_id)
		VALUES($1, $2, $3, $4, $5, $6);`

	_, err := db.Exec(sqlStm, eventTime, eventType, channelId, imported, jsonByteArray, nodeId)
	if err != nil {
		return errors.Wrap(err, "DB Exec")
	}

	channelEvent.ChannelId = channelId
	if !imported {
		ChannelChanges <- channelEvent
		ProcessChannelEvent(channelEvent)
	}

	return nil
}
