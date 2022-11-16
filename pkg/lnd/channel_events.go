package lnd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func chanPointFromByte(cb []byte, oi uint32) (string, error) {
	ch, err := chainhash.NewHash(cb)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", ch.String(), oi), nil
}

// storeChannelEvent extracts the timestamp, channel ID and PubKey from the
// ChannelEvent and converts the original struct to json.
// Then it's stored in the database in the channel_event table.
func storeChannelEvent(ctx context.Context, db *sqlx.DB, client lndClientSubscribeChannelEvent,
	ce *lnrpc.ChannelEventUpdate, nodeSettings commons.ManagedNodeSettings,
	eventChannel chan interface{}) error {

	timestampMs := time.Now().UTC()

	channelEvent := broadcast.ChannelEvent{
		EventData: broadcast.EventData{
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
		channel, err := addChannelOrUpdateStatus(c.ChannelPoint, c.ChanId, &channelStatus, nil, nil, nodeSettings, remoteNodeId, db)
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
		commons.SetChannel(channel.ChannelID, channel.ShortChannelID, channel.Status, channel.FundingTransactionHash, channel.FundingOutputIndex)

		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channel.ChannelID, false, jsonByteArray,
			channelEvent, eventChannel)
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
		channel, err := addChannelOrUpdateStatus(c.ChannelPoint, c.ChanId, nil, &c.CloseType, &c.ClosingTxHash, nodeSettings, remoteNodeId, db)
		if err != nil {
			return errors.Wrap(err, "CLOSED_CHANNEL: Add Channel Or Update Status")
		}

		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "CLOSED_CHANNEL: JSON Marshall")
		}

		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channel.ChannelID, false, jsonByteArray,
			channelEvent, eventChannel)
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
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channelPoint)
		channelId := commons.GetChannelIdFromFundingTransaction(fundingTransactionHash, fundingOutputIndex)
		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "ACTIVE_CHANNEL: JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jsonByteArray,
			channelEvent, eventChannel)
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
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channelPoint)
		channelId := commons.GetChannelIdFromFundingTransaction(fundingTransactionHash, fundingOutputIndex)
		jsonByteArray, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "INACTIVE_CHANNEL: JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jsonByteArray,
			channelEvent, eventChannel)
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
			channelEvent, eventChannel)
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
			channelEvent, eventChannel)
		if err != nil {
			return errors.Wrap(err, "Insert Pending Open Channel Event")
		}
		return nil
	}
	return nil
}

func addChannelOrUpdateStatus(channelPoint string, lndShortChannelId uint64, channelStatus *commons.ChannelStatus,
	closeType *lnrpc.ChannelCloseSummary_ClosureType, closingTxHash *string,
	nodeSettings commons.ManagedNodeSettings, remoteNodeId int,
	db *sqlx.DB) (channels.Channel, error) {

	fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channelPoint)
	channel := channels.Channel{
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		FirstNodeId:            nodeSettings.NodeId,
		SecondNodeId:           remoteNodeId,
	}
	if channelStatus != nil {
		channel.Status = *channelStatus
	}
	if closingTxHash != nil {
		channel.Status = channels.GetClosureStatus(*closeType)
		channel.ClosingTransactionHash = closingTxHash
	}
	shortChannelId := channels.ConvertLNDShortChannelID(lndShortChannelId)
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
	remoteNodeId := commons.GetNodeIdFromPublicKey(remotePublicKey, nodeSettings.Chain, nodeSettings.Network)
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

	fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channelPoint)
	channelId := commons.GetChannelIdFromFundingTransaction(fundingTransactionHash, fundingOutputIndex)
	if channelId == 0 {
		pendingChannelsRequest := lnrpc.PendingChannelsRequest{}
		pendingChannelsResponse, err := client.PendingChannels(ctx, &pendingChannelsRequest)
		if err != nil {
			return 0, errors.Wrap(err, "Obtaining more information from LND about the new channel")
		}
		for _, pendingChannel := range pendingChannelsResponse.PendingOpenChannels {
			if pendingChannel.Channel.ChannelPoint == channelPoint {
				remoteNodeId := commons.GetNodeIdFromPublicKey(pendingChannel.Channel.RemoteNodePub, nodeSettings.Chain, nodeSettings.Network)
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
	nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}) {

	var stream lnrpc.Lightning_SubscribeChannelEventsClient
	var err error
	var chanEvent *lnrpc.ChannelEventUpdate

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if stream == nil {
			stream, err = client.SubscribeChannelEvents(ctx, &lnrpc.ChannelEventSubscription{})
			if err == nil {
				// HACK to know if the context is a testcase.
				if eventChannel != nil {
					// Import routing policies from open channels
					err = importRoutingPolicies(client, db, nodeSettings)
					if err != nil {
						log.Error().Err(err).Msg("Obtaining RoutingPolicies (SubscribeChannelGraph) from LND failed, will retry in 1 minute")
						stream = nil
						time.Sleep(1 * time.Minute)
						continue
					}
				}
			} else {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				log.Error().Err(err).Msg("Obtaining stream (SubscribeChannelEvents) from LND failed, will retry in 1 minute")
				stream = nil
				time.Sleep(1 * time.Minute)
				continue
			}
		}

		chanEvent, err = stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			log.Error().Err(err).Msg("Receiving channel events from the stream failed, will retry in 1 minute")
			stream = nil
			time.Sleep(1 * time.Minute)
			continue
		}

		err = storeChannelEvent(ctx, db, client, chanEvent, nodeSettings, eventChannel)
		if err != nil {
			// TODO FIXME STORE THIS SOMEWHERE??? CHANNELEVENT IS NOW IGNORED???
			log.Error().Err(err).Msg("Storing channel event failed")
		}
	}
}

func ImportChannelList(t lnrpc.ChannelEventUpdate_UpdateType, db *sqlx.DB, client lnrpc.LightningClient,
	nodeSettings commons.ManagedNodeSettings) error {
	ctx := context.Background()
	switch t {
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		req := lnrpc.ListChannelsRequest{}
		r, err := client.ListChannels(ctx, &req)
		if err != nil {
			return errors.Wrap(err, "LND: List channels")
		}

		err = storeImportedOpenChannels(db, r.Channels, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "Store imported open channels")
		}

	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
		req := lnrpc.ClosedChannelsRequest{}
		r, err := client.ClosedChannels(ctx, &req)
		if err != nil {
			return errors.Wrap(err, "LND: Get closed channels")
		}

		err = storeImportedClosedChannels(db, r.Channels, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "Store imported closed channels")
		}

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
	for rows.Next() {
		var channelId int
		err = rows.Scan(&channelId)
		if err != nil {
			return nil, err
		}
		existingChannelIds = append(existingChannelIds, channelId)
	}

	return existingChannelIds, nil
}

func storeImportedOpenChannels(db *sqlx.DB, c []*lnrpc.Channel, nodeSettings commons.ManagedNodeSettings) error {

	if len(c) == 0 {
		return nil
	}

	var channelIds []int
	for _, channel := range c {
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channel.ChannelPoint)
		channelId := commons.GetChannelIdFromFundingTransaction(fundingTransactionHash, fundingOutputIndex)
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
		channel, err := addChannelOrUpdateStatus(lndChannel.ChannelPoint, lndChannel.ChanId,
			&channelStatus, nil, nil, nodeSettings, remoteNodeId, db)
		if err != nil {
			return errors.Wrap(err, "ImportedOpenChannels: Add Channel Or Update Status")
		}

		if lndChannel.ChanId == 0 {
			fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(lndChannel.ChannelPoint)
			log.Error().Msgf("Failed to obtain shortChannelId for open channel with channel point %v:%v",
				fundingTransactionHash, fundingOutputIndex)
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
			channel.ChannelID, true, jsonByteArray, broadcast.ChannelEvent{}, nil)
		if err != nil {
			return errors.Wrap(err, "ImportedOpenChannels: Insert channel event")
		}
	}
	return nil
}

func storeImportedClosedChannels(db *sqlx.DB, c []*lnrpc.ChannelCloseSummary,
	nodeSettings commons.ManagedNodeSettings) error {

	if len(c) == 0 {
		return nil
	}

	var channelIds []int
	for _, channel := range c {
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channel.ChannelPoint)
		channelId := commons.GetChannelIdFromFundingTransaction(fundingTransactionHash, fundingOutputIndex)
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
		channel, err := addChannelOrUpdateStatus(lndChannel.ChannelPoint, lndChannel.ChanId,
			nil, &lndChannel.CloseType, &lndChannel.ClosingTxHash, nodeSettings, remoteNodeId, db)
		if err != nil {
			return errors.Wrap(err, "ImportedClosedChannels: Add Channel Or Update Status")
		}

		if lndChannel.ChanId == 0 {
			fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(lndChannel.ChannelPoint)
			log.Error().Msgf("Failed to obtain shortChannelId for closed channel with channel point %v:%v",
				fundingTransactionHash, fundingOutputIndex)
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
			channel.ChannelID, true, jsonByteArray, broadcast.ChannelEvent{}, nil)
		if err != nil {
			return errors.Wrap(err, "ImportedClosedChannels: Insert channel event")
		}
	}
	return nil
}

func insertChannelEvent(db *sqlx.DB, eventTime time.Time, eventType lnrpc.ChannelEventUpdate_UpdateType,
	nodeId, channelId int, imported bool, jsonByteArray []byte,
	channelEvent broadcast.ChannelEvent, eventChannel chan interface{}) error {

	var sqlStm = `INSERT INTO channel_event (time, event_type, channel_id, imported, event, node_id)
		VALUES($1, $2, $3, $4, $5, $6);`

	_, err := db.Exec(sqlStm, eventTime, eventType, channelId, imported, jsonByteArray, nodeId)
	if err != nil {
		return errors.Wrap(err, "DB Exec")
	}

	if eventChannel != nil {
		channelEvent.ChannelId = channelId
		eventChannel <- channelEvent
	}

	return nil
}
