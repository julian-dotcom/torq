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
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/pkg/commons"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

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
func storeChannelEvent(ctx context.Context, db *sqlx.DB, client lndClientSubscribeChannelEvent,
	ce *lnrpc.ChannelEventUpdate, nodeSettings commons.ManagedNodeSettings,
	channelEventChannel chan commons.ChannelEvent) error {

	timestampMs := time.Now().UTC()

	channelEvent := commons.ChannelEvent{
		EventData: commons.EventData{
			EventTime: timestampMs,
			NodeId:    nodeSettings.NodeId,
		},
		Type: ce.GetType(),
	}

	switch ce.Type {
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		c := ce.GetOpenChannel()
		remotePublicKey := c.RemotePubkey
		remoteNodeId, err := addNodeWhenNew(remotePublicKey, nodeSettings, db)
		if err != nil {
			return errors.Wrap(err, "OPEN_CHANNEL: Add Node When New")
		}
		channelStatus := commons.Open
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
		commons.SetChannelNode(remoteNodeId, remotePublicKey, nodeSettings.Chain, nodeSettings.Network, channel.Status)

		// This allows torq to listen to the graph for channel updates
		commons.SetChannel(channel.ChannelID, channel.ShortChannelID, channel.LNDShortChannelID, channel.Status,
			channel.FundingTransactionHash, channel.FundingOutputIndex,
			channel.FundingBlockHeight, channel.FundedOn,
			channel.Capacity, channel.Private, channel.FirstNodeId, channel.SecondNodeId,
			channel.InitiatingNodeId, channel.AcceptingNodeId,
			channel.ClosingTransactionHash, channel.ClosingNodeId, channel.ClosingBlockHeight, channel.ClosedOn)

		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channel.ChannelID, false, jsonByteArray,
			channelEvent, channelEventChannel)
		if err != nil {
			return errors.Wrap(err, "Insert Open Channel Event")
		}
		return nil
	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
		c := ce.GetClosedChannel()
		remotePublicKey := c.RemotePubkey
		remoteNodeId, err := addNodeWhenNew(remotePublicKey, nodeSettings, db)
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
			channelEvent, channelEventChannel)
		if err != nil {
			return errors.Wrap(err, "Insert Closed Channel Event")
		}

		// This stops the graph from listening to node updates
		chans, err := channels.GetOpenChannelsForNodeId(db, remoteNodeId)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to verify if remote node still has open channels: %v", remoteNodeId)
		}

		// This stops the graph from listening to channel updates
		commons.SetChannelStatus(channel.ChannelID, channel.Status)
		if len(chans) == 0 {
			commons.InactivateChannelNode(remotePublicKey, nodeSettings.Chain, nodeSettings.Network)
		}
		return nil
	case lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL:
		c := ce.GetActiveChannel()
		channelPoint, err := chanPointFromByte(c.GetFundingTxidBytes(), c.GetOutputIndex())
		if err != nil {
			return errors.Wrap(err, "ACTIVE_CHANNEL: Get channelPoint from bytes")
		}
		channelId := commons.GetChannelIdByChannelPoint(channelPoint)
		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "ACTIVE_CHANNEL: JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jsonByteArray,
			channelEvent, channelEventChannel)
		if err != nil {
			return errors.Wrap(err, "Insert Active Channel Event")
		}
		return nil
	case lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL:
		c := ce.GetInactiveChannel()
		channelPoint, err := chanPointFromByte(c.GetFundingTxidBytes(), c.GetOutputIndex())
		if err != nil {
			return errors.Wrap(err, "INACTIVE_CHANNEL: Get channelPoint from bytes")
		}
		channelId := commons.GetChannelIdByChannelPoint(channelPoint)
		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "INACTIVE_CHANNEL: JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jsonByteArray,
			channelEvent, channelEventChannel)
		if err != nil {
			return errors.Wrap(err, "Insert Inactive Channel Event")
		}
		return nil
	case lnrpc.ChannelEventUpdate_FULLY_RESOLVED_CHANNEL:
		c := ce.GetFullyResolvedChannel()
		channelId, err := processPendingOpenChannel(ctx, db, client, c.GetFundingTxidBytes(), c.GetOutputIndex(), nodeSettings)
		if err != nil {
			return errors.Wrap(err, "FULLY_RESOLVED_CHANNEL: Process Pending Open Channel")
		}
		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "FULLY_RESOLVED_CHANNEL: JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jsonByteArray,
			channelEvent, channelEventChannel)
		if err != nil {
			return errors.Wrap(err, "Insert Fully Resolved Channel Event")
		}
		return nil
	case lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL:
		c := ce.GetPendingOpenChannel()
		channelId, err := processPendingOpenChannel(ctx, db, client, c.GetTxid(), c.GetOutputIndex(), nodeSettings)
		if err != nil {
			return errors.Wrap(err, "PENDING_OPEN_CHANNEL: Process Pending Open Channel")
		}
		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "PENDING_OPEN_CHANNEL: JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jsonByteArray,
			channelEvent, channelEventChannel)
		if err != nil {
			return errors.Wrap(err, "Insert Pending Open Channel Event")
		}
		return nil
	}
	return nil
}

func addChannelOrUpdateStatus(channelPoint string, lndShortChannelId uint64, channelStatus *commons.ChannelStatus,
	capacity int64, private *bool, closeType *lnrpc.ChannelCloseSummary_ClosureType, closingTxHash *string,
	nodeSettings commons.ManagedNodeSettings, remoteNodeId int, initiatingNodeId *int, closingNodeId *int,
	db *sqlx.DB) (channels.Channel, error) {

	fundingTransactionHash, fundingOutputIndex := commons.ParseChannelPoint(channelPoint)
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
	shortChannelId := commons.ConvertLNDShortChannelID(lndShortChannelId)
	if lndShortChannelId != 0 {
		channel.LNDShortChannelID = &lndShortChannelId
		channel.ShortChannelID = &shortChannelId
	}
	var err error
	channel.ChannelID, err = channels.AddChannelOrUpdateChannelStatus(db, channel)
	if err != nil {
		return channels.Channel{}, errors.Wrapf(err, "Adding or updating channel (channelId: %v, shortChannelId: %v)", channel.ChannelID, shortChannelId)
	}
	return channel, nil
}

func addNodeWhenNew(remotePublicKey string, nodeSettings commons.ManagedNodeSettings, db *sqlx.DB) (int, error) {
	remoteNodeId := commons.GetNodeIdByPublicKey(remotePublicKey, nodeSettings.Chain, nodeSettings.Network)
	if remoteNodeId == 0 {
		newNode := nodes.Node{
			PublicKey: remotePublicKey,
			Chain:     nodeSettings.Chain,
			Network:   nodeSettings.Network,
		}
		var err error
		remoteNodeId, err = nodes.AddNodeWhenNew(db, newNode)
		if err != nil {
			return 0, errors.Wrapf(err, "Adding node with public key: %v", remotePublicKey)
		}
	}
	return remoteNodeId, nil
}

func processPendingOpenChannel(ctx context.Context, db *sqlx.DB, client lndClientSubscribeChannelEvent,
	txId []byte, outputIndex uint32, nodeSettings commons.ManagedNodeSettings) (int, error) {

	channelPoint, err := chanPointFromByte(txId, outputIndex)
	if err != nil {
		return 0, err
	}

	fundingTransactionHash, fundingOutputIndex := commons.ParseChannelPoint(channelPoint)
	channelId := commons.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
	if channelId == 0 {
		pendingChannelsRequest := lnrpc.PendingChannelsRequest{}
		pendingChannelsResponse, err := client.PendingChannels(ctx, &pendingChannelsRequest)
		if err != nil {
			return 0, errors.Wrap(err, "Obtaining more information from LND about the new channel")
		}
		for _, pendingChannel := range pendingChannelsResponse.PendingOpenChannels {
			if pendingChannel.Channel.ChannelPoint == channelPoint {
				remoteNodeId := commons.GetNodeIdByPublicKey(pendingChannel.Channel.RemoteNodePub, nodeSettings.Chain, nodeSettings.Network)
				if remoteNodeId == 0 {
					remoteNode := nodes.Node{
						PublicKey: pendingChannel.Channel.RemoteNodePub,
						Chain:     nodeSettings.Chain,
						Network:   nodeSettings.Network,
					}
					remoteNodeId, err = nodes.AddNodeWhenNew(db, remoteNode)
					if err != nil {
						return 0, errors.Wrap(err, "Registering new node for new channel")
					}
				}
				newChannel := channels.Channel{
					FundingTransactionHash: fundingTransactionHash,
					FundingOutputIndex:     fundingOutputIndex,
					FirstNodeId:            nodeSettings.NodeId,
					SecondNodeId:           remoteNodeId,
					Capacity:               pendingChannel.Channel.Capacity,
					Status:                 commons.Opening,
				}
				channelId, err = channels.AddChannelOrUpdateChannelStatus(db, newChannel)
				if err != nil {
					return 0, errors.Wrap(err, "Registering new channel")
				}
			}
		}
	}
	return channelId, nil
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

// SubscribeAndStoreChannelEvents Subscribes to channel events from LND and stores them in the
// database as a time series
func SubscribeAndStoreChannelEvents(ctx context.Context, client lndClientSubscribeChannelEvent, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, channelEventChannel chan commons.ChannelEvent,
	importRequestChannel chan commons.ImportRequest,
	serviceEventChannel chan commons.ServiceEvent) {

	defer log.Info().Msgf("SubscribeAndStoreChannelEvents terminated for nodeId: %v", nodeSettings.NodeId)

	var stream lnrpc.Lightning_SubscribeChannelEventsClient
	var err error
	var chanEvent *lnrpc.ChannelEventUpdate
	serviceStatus := commons.Inactive
	subscriptionStream := commons.ChannelEventStream

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if stream == nil {
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
			stream, err = client.SubscribeChannelEvents(ctx, &lnrpc.ChannelEventSubscription{})
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				log.Error().Err(err).Msgf("Obtaining stream (SubscribeChannelEvents) from LND failed, will retry in %v seconds", commons.STREAM_ERROR_SLEEP_SECONDS)
				stream = nil
				time.Sleep(commons.STREAM_ERROR_SLEEP_SECONDS * time.Second)
				continue
			}
			// HACK to know if the context is a testcase.
			if importRequestChannel != nil {
				responseChannel := make(chan error)
				importRequestChannel <- commons.ImportRequest{
					ImportType: commons.ImportChannelAndRoutingPolicies,
					Out:        responseChannel,
				}
				err = <-responseChannel
				if err != nil {
					log.Error().Err(err).Msgf("Obtaining RoutingPolicies (SubscribeChannelGraph) from LND failed, will retry in %v seconds", commons.STREAM_ERROR_SLEEP_SECONDS)
					stream = nil
					time.Sleep(commons.STREAM_ERROR_SLEEP_SECONDS * time.Second)
					continue
				}
			}
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
		}

		chanEvent, err = stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
			log.Error().Err(err).Msgf("Receiving channel events from the stream failed, will retry in %v seconds", commons.STREAM_ERROR_SLEEP_SECONDS)
			stream = nil
			time.Sleep(commons.STREAM_ERROR_SLEEP_SECONDS * time.Second)
			continue
		}

		err = storeChannelEvent(ctx, db, client, chanEvent, nodeSettings, channelEventChannel)
		if err != nil {
			// TODO FIXME STORE THIS SOMEWHERE??? CHANNELEVENT IS NOW IGNORED???
			log.Error().Err(err).Msg("Storing channel event failed")
		}
	}
}

func ImportPendingChannels(db *sqlx.DB, vectorUrl string, client lnrpc.LightningClient,
	nodeSettings commons.ManagedNodeSettings, lightningRequestChannel chan interface{}) error {
	ctx := context.Background()
	r, err := client.PendingChannels(ctx, &lnrpc.PendingChannelsRequest{})
	if err != nil {
		return errors.Wrap(err, "LND: Get pending channels")
	}

	err = storeImportedWaitingCloseChannels(db, vectorUrl, r.WaitingCloseChannels, nodeSettings, lightningRequestChannel)
	if err != nil {
		return errors.Wrap(err, "Store imported waiting close channels")
	}
	err = storeImportedPendingOpenChannels(db, r.PendingOpenChannels, nodeSettings)
	if err != nil {
		return errors.Wrap(err, "Store imported pending open channels")
	}
	err = storeImportedPendingForceClosingChannels(db, vectorUrl, r.PendingForceClosingChannels, nodeSettings, lightningRequestChannel)
	if err != nil {
		return errors.Wrap(err, "Store imported pending force closing channels")
	}
	return nil
}

func ImportOpenChannels(db *sqlx.DB, vectorUrl string, client lnrpc.LightningClient,
	nodeSettings commons.ManagedNodeSettings, lightningRequestChannel chan interface{}) error {
	ctx := context.Background()
	r, err := client.ListChannels(ctx, &lnrpc.ListChannelsRequest{})
	if err != nil {
		return errors.Wrap(err, "LND: List channels")
	}

	err = storeImportedOpenChannels(db, vectorUrl, r.Channels, nodeSettings, lightningRequestChannel)
	if err != nil {
		return errors.Wrap(err, "Store imported open channels")
	}
	return nil
}

func ImportClosedChannels(db *sqlx.DB, vectorUrl string, client lnrpc.LightningClient,
	nodeSettings commons.ManagedNodeSettings, lightningRequestChannel chan interface{}) error {
	ctx := context.Background()
	r, err := client.ClosedChannels(ctx, &lnrpc.ClosedChannelsRequest{})
	if err != nil {
		return errors.Wrap(err, "LND: Get closed channels")
	}

	err = storeImportedClosedChannels(db, vectorUrl, r.Channels, nodeSettings, lightningRequestChannel)
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
	vectorUrl string,
	waitingCloseChannels []*lnrpc.PendingChannelsResponse_WaitingCloseChannel,
	nodeSettings commons.ManagedNodeSettings,
	lightningRequestChannel chan interface{}) error {

	if len(waitingCloseChannels) == 0 {
		return nil
	}

	for _, waitingCloseChannel := range waitingCloseChannels {
		lndChannel := waitingCloseChannel.GetChannel()
		closing := commons.Closing

		remoteNodeId, initiatingNodeId, closeType, closingNodeId, err := processPendingChannel(db, lndChannel, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "ImportedWaitingCloseChannels: ProcessPendingChannel")
		}

		lndShortChannelId := processEmptyChanId(vectorUrl, lndChannel.ChannelPoint, nodeSettings, lightningRequestChannel)

		var closingTransactionHash *string
		if waitingCloseChannel.ClosingTxid != "" {
			closingTransactionHash = &waitingCloseChannel.ClosingTxid
		}

		_, err = addChannelOrUpdateStatus(lndChannel.ChannelPoint, lndShortChannelId, &closing, lndChannel.Capacity, &lndChannel.Private,
			closeType, closingTransactionHash, nodeSettings, remoteNodeId, initiatingNodeId, closingNodeId, db)
		if err != nil {
			return errors.Wrap(err, "ImportedWaitingCloseChannels: Add Channel Or Update Status")
		}

		commons.SetChannelNode(remoteNodeId, lndChannel.RemoteNodePub, nodeSettings.Chain, nodeSettings.Network, closing)
	}
	return nil
}

func processPendingChannel(db *sqlx.DB,
	lndChannel *lnrpc.PendingChannelsResponse_PendingChannel,
	nodeSettings commons.ManagedNodeSettings) (int, *int, *lnrpc.ChannelCloseSummary_ClosureType, *int, error) {

	var initiatingNodeId *int
	var closingNodeId *int
	var closeType *lnrpc.ChannelCloseSummary_ClosureType

	remoteNodeId, err := addNodeWhenNew(lndChannel.RemoteNodePub, nodeSettings, db)
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
	nodeSettings commons.ManagedNodeSettings) error {

	if len(pendingOpenChannels) == 0 {
		return nil
	}

	var channelIds []int
	for _, channel := range pendingOpenChannels {
		fundingTransactionHash, fundingOutputIndex := commons.ParseChannelPoint(channel.GetChannel().ChannelPoint)
		channelId := commons.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
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
		opening := commons.Opening

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
			channel.ChannelID, true, jsonByteArray, commons.ChannelEvent{}, nil)
		if err != nil {
			return errors.Wrap(err, "ImportedPendingOpenChannels: Insert channel event")
		}

		commons.SetChannelNode(remoteNodeId, lndChannel.RemoteNodePub, nodeSettings.Chain, nodeSettings.Network, opening)
	}
	return nil
}

func storeImportedPendingForceClosingChannels(
	db *sqlx.DB,
	vectorUrl string,
	pendingForceClosingChannels []*lnrpc.PendingChannelsResponse_ForceClosedChannel,
	nodeSettings commons.ManagedNodeSettings,
	lightningRequestChannel chan interface{}) error {

	if len(pendingForceClosingChannels) == 0 {
		return nil
	}

	for _, pendingForceClosingChannel := range pendingForceClosingChannels {
		lndChannel := pendingForceClosingChannel.GetChannel()
		closing := commons.Closing

		remoteNodeId, initiatingNodeId, closeType, closingNodeId, err := processPendingChannel(db, lndChannel, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "ImportedPendingForceClosingChannels: ProcessPendingChannel")
		}

		lndShortChannelId := processEmptyChanId(vectorUrl, lndChannel.ChannelPoint, nodeSettings, lightningRequestChannel)

		var closingTransactionHash *string
		if pendingForceClosingChannel.ClosingTxid != "" {
			closingTransactionHash = &pendingForceClosingChannel.ClosingTxid
		}

		_, err = addChannelOrUpdateStatus(lndChannel.ChannelPoint, lndShortChannelId, &closing, lndChannel.Capacity, &lndChannel.Private,
			closeType, closingTransactionHash, nodeSettings, remoteNodeId, initiatingNodeId, closingNodeId, db)
		if err != nil {
			return errors.Wrap(err, "ImportedPendingForceClosingChannels: Add Channel Or Update Status")
		}

		commons.SetChannelNode(remoteNodeId, lndChannel.RemoteNodePub, nodeSettings.Chain, nodeSettings.Network, closing)
	}
	return nil
}

func storeImportedOpenChannels(db *sqlx.DB, vectorUrl string, c []*lnrpc.Channel,
	nodeSettings commons.ManagedNodeSettings,
	lightningRequestChannel chan interface{}) error {

	if len(c) == 0 {
		return nil
	}

	var channelIds []int
	for _, channel := range c {
		channelId := commons.GetChannelIdByChannelPoint(channel.ChannelPoint)
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

		remoteNodeId, err := addNodeWhenNew(lndChannel.RemotePubkey, nodeSettings, db)
		if err != nil {
			return errors.Wrap(err, "ImportedOpenChannels: Add Node When New")
		}
		channelStatus := commons.Open
		var initiatingNodeId *int
		if lndChannel.Initiator {
			initiatingNodeId = &nodeSettings.NodeId
		}
		if lndChannel.ChanId == 0 {
			lndChannel.ChanId = processEmptyChanId(vectorUrl, lndChannel.ChannelPoint, nodeSettings, lightningRequestChannel)
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
			channel.ChannelID, true, jsonByteArray, commons.ChannelEvent{}, nil)
		if err != nil {
			return errors.Wrap(err, "ImportedOpenChannels: Insert channel event")
		}

		commons.SetChannelNode(remoteNodeId, lndChannel.RemotePubkey, nodeSettings.Chain, nodeSettings.Network, commons.Open)
	}
	return nil
}

func storeImportedClosedChannels(db *sqlx.DB, vectorUrl string, c []*lnrpc.ChannelCloseSummary, nodeSettings commons.ManagedNodeSettings,
	lightningRequestChannel chan interface{}) error {

	if len(c) == 0 {
		return nil
	}

	var channelIds []int
	for _, channel := range c {
		fundingTransactionHash, fundingOutputIndex := commons.ParseChannelPoint(channel.ChannelPoint)
		channelId := commons.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
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

		remoteNodeId, err := addNodeWhenNew(lndChannel.RemotePubkey, nodeSettings, db)
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
			lndChannel.ChanId = processEmptyChanId(vectorUrl, lndChannel.ChannelPoint, nodeSettings, lightningRequestChannel)
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
			channel.ChannelID, true, jsonByteArray, commons.ChannelEvent{}, nil)
		if err != nil {
			return errors.Wrap(err, "ImportedClosedChannels: Insert channel event")
		}

		commons.SetChannelNode(remoteNodeId, lndChannel.RemotePubkey, nodeSettings.Chain, nodeSettings.Network, channels.GetClosureStatus(lndChannel.CloseType))
	}
	return nil
}

func processEmptyChanId(vectorUrl string, channelPoint string, nodeSettings commons.ManagedNodeSettings,
	lightningRequestChannel chan interface{}) uint64 {

	fundingTransactionHash, fundingOutputIndex := commons.ParseChannelPoint(channelPoint)
	channelId := commons.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
	if channelId != 0 {
		channelSettings := commons.GetChannelSettingByChannelId(channelId)
		if channelSettings.LndShortChannelId != 0 {
			return channelSettings.LndShortChannelId
		}
		if channelSettings.Status == commons.AbandonedClosed || channelSettings.Status == commons.FundingCancelledClosed {
			return 0
		}
	}

	if vectorUrl == commons.VECTOR_URL && (nodeSettings.Chain != commons.Bitcoin || nodeSettings.Network != commons.MainNet) {
		log.Info().Msgf("Skipping obtaining short channel id from vector for nodeId: %v", nodeSettings.NodeId)
		return 0
	}

	shortChannelId := commons.GetShortChannelIdFromVector(vectorUrl, fundingTransactionHash, fundingOutputIndex, nodeSettings, lightningRequestChannel)
	if shortChannelId == "" {
		log.Error().Msgf("Failed to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
		return 0
	}
	lndShortChannelId, err := commons.ConvertShortChannelIDToLND(shortChannelId)
	if err != nil {
		log.Error().Msgf("Failed (ConvertShortChannelIDToLND) to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
	}
	return lndShortChannelId
}

func insertChannelEvent(db *sqlx.DB, eventTime time.Time, eventType lnrpc.ChannelEventUpdate_UpdateType,
	nodeId, channelId int, imported bool, jsonByteArray []byte,
	channelEvent commons.ChannelEvent, channelEventChannel chan commons.ChannelEvent) error {

	var sqlStm = `INSERT INTO channel_event (time, event_type, channel_id, imported, event, node_id)
		VALUES($1, $2, $3, $4, $5, $6);`

	_, err := db.Exec(sqlStm, eventTime, eventType, channelId, imported, jsonByteArray, nodeId)
	if err != nil {
		return errors.Wrap(err, "DB Exec")
	}

	if channelEventChannel != nil {
		channelEvent.ChannelId = channelId
		channelEventChannel <- channelEvent
	}

	return nil
}
