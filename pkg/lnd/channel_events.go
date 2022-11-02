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
	"github.com/lncapital/torq/pkg/commons"

	"github.com/rs/zerolog/log"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
)

// websocket channel event
type wsChannelEvent struct {
	Type                   string `json:"type"`
	ChannelEventType       string `json:"channelEventType"`
	ShortChannelId         string `json:"shortChannelId,omitempty"`
	FundingTransactionHash string `json:"fundingTransactionHash"`
	FundingOutputIndex     int    `json:"fundingOutputIndex"`
	PubKey                 string `json:"pubKey,omitempty"`
}

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
	wsChan chan interface{}) error {

	timestampMs := time.Now().UTC()

	var wsChanEvent wsChannelEvent

	wsChanEvent.Type = "channelEvent"
	wsChanEvent.ChannelEventType = ce.GetType().String()

	switch ce.Type {
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		c := ce.GetOpenChannel()
		chanID := c.ChanId
		channelPoint := c.ChannelPoint
		publicKey := c.RemotePubkey
		shortChannelId := channels.ConvertLNDShortChannelID(chanID)
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channelPoint)

		node := nodes.Node{
			PublicKey: c.RemotePubkey,
			Chain:     nodeSettings.Chain,
			Network:   nodeSettings.Network,
		}
		remoteNodeId, err := nodes.AddNodeWhenNew(db, node)
		if err != nil {
			return errors.Wrap(err, "Adding node")
		}

		// This allows torq to listen to the graph for node updates
		commons.SetChannelNode(remoteNodeId, c.RemotePubkey, nodeSettings.Chain, nodeSettings.Network)

		channel := channels.Channel{
			ShortChannelID:         shortChannelId,
			FundingTransactionHash: fundingTransactionHash,
			FundingOutputIndex:     fundingOutputIndex,
			FirstNodeId:            nodeSettings.NodeId,
			SecondNodeId:           remoteNodeId,
			LNDShortChannelID:      chanID,
			Status:                 channels.Open,
		}
		channelId, err := channels.AddChannelOrUpdateChannelStatus(db, channel)
		if err != nil {
			return errors.Wrap(err, "Adding channel")
		}

		// This allows torq to listen to the graph for channel updates
		commons.SetChannel(channelId, channel.ShortChannelID, int(channels.Open), fundingTransactionHash, fundingOutputIndex)

		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jb)
		if err != nil {
			return errors.Wrap(err, "Insert Open Channel Event")
		}

		if wsChan != nil {
			wsChanEvent.ShortChannelId = shortChannelId
			wsChanEvent.FundingTransactionHash = fundingTransactionHash
			wsChanEvent.FundingOutputIndex = fundingOutputIndex
			wsChanEvent.PubKey = publicKey
			wsChan <- wsChanEvent
		}

		return nil

	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
		c := ce.GetClosedChannel()
		chanID := c.ChanId
		channelPoint := c.ChannelPoint
		publicKey := c.RemotePubkey
		shortChannelId := channels.ConvertLNDShortChannelID(c.ChanId)
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channelPoint)

		var err error
		remoteNodeId := commons.GetNodeIdFromPublicKey(publicKey, nodeSettings.Chain, nodeSettings.Network)
		if remoteNodeId == 0 {
			newNode := nodes.Node{
				PublicKey: publicKey,
				Chain:     nodeSettings.Chain,
				Network:   nodeSettings.Network,
			}
			remoteNodeId, err = nodes.AddNodeWhenNew(db, newNode)
			if err != nil {
				return errors.Wrapf(err, "Adding node (publicKey: %v shortChannelId: %v)", publicKey, shortChannelId)
			}
		}

		channel := channels.Channel{
			ShortChannelID:         shortChannelId,
			FundingTransactionHash: fundingTransactionHash,
			FundingOutputIndex:     fundingOutputIndex,
			ClosingTransactionHash: &c.ClosingTxHash,
			FirstNodeId:            nodeSettings.NodeId,
			SecondNodeId:           remoteNodeId,
			LNDShortChannelID:      chanID,
			Status:                 channels.GetClosureStatus(c.CloseType),
		}
		channelId, err := channels.AddChannelOrUpdateChannelStatus(db, channel)
		if err != nil {
			return errors.Wrapf(err, "Adding or updating channel (channelId: %v, shortChannelId: %v)", channelId, shortChannelId)
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jb)
		if err != nil {
			return errors.Wrap(err, "Insert Closed Channel Event")
		}

		// This stops the graph from listening to node updates
		chans, err := channels.GetChannelsForNodeId(db, remoteNodeId)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to verify if remote node still has open channels: %v", remoteNodeId)
		}

		// This stops the graph from listening to channel updates
		commons.RemoveChannel(channelId)
		if len(chans) == 0 {
			commons.RemoveChannelNode(publicKey, nodeSettings.Chain, nodeSettings.Network)
		}

		if wsChan != nil {
			wsChanEvent.ShortChannelId = shortChannelId
			wsChanEvent.FundingTransactionHash = fundingTransactionHash
			wsChanEvent.FundingOutputIndex = fundingOutputIndex
			wsChanEvent.PubKey = publicKey
			wsChan <- wsChanEvent
		}

		return nil

	case lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL:
		c := ce.GetActiveChannel()
		channelPoint, err := chanPointFromByte(c.GetFundingTxidBytes(), c.GetOutputIndex())
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channelPoint)
		channelId := commons.GetChannelIdFromFundingTransaction(fundingTransactionHash, fundingOutputIndex)
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jb)
		if err != nil {
			return errors.Wrap(err, "Insert Active Channel Event")
		}

		if wsChan != nil {
			wsChanEvent.FundingTransactionHash = fundingTransactionHash
			wsChanEvent.FundingOutputIndex = fundingOutputIndex
			wsChan <- wsChanEvent
		}

		return nil
	case lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL:
		c := ce.GetInactiveChannel()
		channelPoint, err := chanPointFromByte(c.GetFundingTxidBytes(), c.GetOutputIndex())
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channelPoint)
		channelId := commons.GetChannelIdFromFundingTransaction(fundingTransactionHash, fundingOutputIndex)
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jb)
		if err != nil {
			return errors.Wrap(err, "Insert Inactive Channel Event")
		}

		if wsChan != nil {
			wsChanEvent.FundingTransactionHash = fundingTransactionHash
			wsChanEvent.FundingOutputIndex = fundingOutputIndex
			wsChan <- wsChanEvent
		}

		return nil

	case lnrpc.ChannelEventUpdate_FULLY_RESOLVED_CHANNEL:
		c := ce.GetFullyResolvedChannel()
		channelId, err := processPendingOpenChannel(ctx, db, client, c.GetFundingTxidBytes(), c.GetOutputIndex(), nodeSettings)
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jb)
		if err != nil {
			return errors.Wrap(err, "Insert Fully Resolved Channel Event")
		}

		//if wsChan != nil {
		//	wsChanEvent.LNDChannelPoint = channelPoint
		//	wsChan <- wsChanEvent
		//}

	case lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL:
		c := ce.GetPendingOpenChannel()
		channelId, err := processPendingOpenChannel(ctx, db, client, c.GetTxid(), c.GetOutputIndex(), nodeSettings)
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, nodeSettings.NodeId, channelId, false, jb)
		if err != nil {
			return errors.Wrap(err, "Insert Pending Open Channel Event")
		}

		//if wsChan != nil {
		//	wsChanEvent.LNDChannelPoint = channelPoint
		//	wsChan <- wsChanEvent
		//}

		return nil
	default:
	}

	return nil
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
				// TODO FIXME ShortChannelID and LNDShortChannelID should be nullable
				newChannel := channels.Channel{
					ShortChannelID:         "",
					FundingTransactionHash: fundingTransactionHash,
					FundingOutputIndex:     fundingOutputIndex,
					FirstNodeId:            nodeSettings.NodeId,
					SecondNodeId:           remoteNodeId,
					Status:                 channels.Opening,
					LNDShortChannelID:      0,
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

type lndClientSubscribeChannelEvent interface {
	SubscribeChannelEvents(ctx context.Context, in *lnrpc.ChannelEventSubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error)
	PendingChannels(ctx context.Context, in *lnrpc.PendingChannelsRequest,
		opts ...grpc.CallOption) (*lnrpc.PendingChannelsResponse, error)
}

// SubscribeAndStoreChannelEvents Subscribes to channel events from LND and stores them in the
// database as a time series
func SubscribeAndStoreChannelEvents(ctx context.Context, client lndClientSubscribeChannelEvent, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, wsChan chan interface{}) error {

	cesr := lnrpc.ChannelEventSubscription{}
	stream, err := client.SubscribeChannelEvents(ctx, &cesr)
	if err != nil {
		return errors.Wrap(err, "Subscribe channel events")
	}

	rl := ratelimit.New(1) // 1 per second maximum rate limit
	for {

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		chanEvent, err := stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				break
			}

			log.Error().Err(err).Msg("Subscribe channel events stream receive error")
			// rate limited resubscribe
			log.Info().Msg("Attempting reconnect to channel events")
			for {
				rl.Take()
				stream, err = client.SubscribeChannelEvents(ctx, &cesr)
				if err == nil {
					log.Info().Msg("Reconnected to channel events")
					break
				}
				log.Debug().Err(err).Msg("Reconnecting to channel events error")
			}
			continue
		}

		err = storeChannelEvent(ctx, db, client, chanEvent, nodeSettings, wsChan)
		if err != nil {
			log.Error().Err(err).Msg("Subscribe channel events store event error")
			// rate limit for caution but hopefully not needed
			rl.Take()
			continue
		}

	}

	return nil
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

	channelIds := make([]int, len(c))
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
	for _, channel := range c {
		shortChannelId := channels.ConvertLNDShortChannelID(channel.ChanId)
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channel.ChannelPoint)
		remoteNodeId := commons.GetNodeIdFromPublicKey(channel.RemotePubkey, nodeSettings.Chain, nodeSettings.Network)
		if remoteNodeId == 0 {
			newNode := nodes.Node{
				PublicKey: channel.RemotePubkey,
				Chain:     nodeSettings.Chain,
				Network:   nodeSettings.Network,
			}
			remoteNodeId, err = nodes.AddNodeWhenNew(db, newNode)
			if err != nil {
				return errors.Wrapf(err, "Adding node (publicKey: %v shortChannelId: %v)", channel.RemotePubkey, shortChannelId)
			}
		}

		// check if we have seen this channel before and if not store in the channel table
		channelRecord := channels.Channel{
			ShortChannelID:         shortChannelId,
			FundingTransactionHash: fundingTransactionHash,
			FundingOutputIndex:     fundingOutputIndex,
			FirstNodeId:            nodeSettings.NodeId,
			SecondNodeId:           remoteNodeId,
			LNDShortChannelID:      channel.ChanId,
			Status:                 channels.Open,
		}
		channelId, err := channels.AddChannelOrUpdateChannelStatus(db, channelRecord)
		if err != nil {
			return errors.Wrapf(err, "Adding or updating channel status (channelId: %v shortChannelId: %v)", channelId, shortChannelId)
		}
		// skip if we have an existing channel open channel event
		for _, existingChannelId := range existingChannelIds {
			if channelId == existingChannelId {
				continue icoLoop
			}
		}
		jb, err := json.Marshal(channel)
		if err != nil {
			return errors.Wrap(err, "JSON Marshal")
		}
		err = insertChannelEvent(db, time.Now().UTC(), lnrpc.ChannelEventUpdate_OPEN_CHANNEL, nodeSettings.NodeId, channelId, true, jb)
		if err != nil {
			return errors.Wrap(err, "Insert channel event")
		}
	}
	return nil
}

func storeImportedClosedChannels(db *sqlx.DB, c []*lnrpc.ChannelCloseSummary,
	nodeSettings commons.ManagedNodeSettings) error {

	if len(c) == 0 {
		return nil
	}

	channelIds := make([]int, len(c))
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
	for _, channel := range c {
		shortChannelId := channels.ConvertLNDShortChannelID(channel.ChanId)
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(channel.ChannelPoint)
		remoteNodeId := commons.GetNodeIdFromPublicKey(channel.RemotePubkey, nodeSettings.Chain, nodeSettings.Network)
		if remoteNodeId == 0 {
			newNode := nodes.Node{
				PublicKey: channel.RemotePubkey,
				Chain:     nodeSettings.Chain,
				Network:   nodeSettings.Network,
			}
			remoteNodeId, err = nodes.AddNodeWhenNew(db, newNode)
			if err != nil {
				return errors.Wrapf(err, "Adding node (publicKey: %v shortChannelId: %v)", channel.RemotePubkey, shortChannelId)
			}
		}

		// check if we have seen this channel before and if not store in the channel table
		channelRecord := channels.Channel{
			ShortChannelID:         shortChannelId,
			FundingTransactionHash: fundingTransactionHash,
			FundingOutputIndex:     fundingOutputIndex,
			ClosingTransactionHash: &channel.ClosingTxHash,
			FirstNodeId:            nodeSettings.NodeId,
			SecondNodeId:           remoteNodeId,
			LNDShortChannelID:      channel.ChanId,
			Status:                 channels.GetClosureStatus(channel.CloseType),
		}
		channelId, err := channels.AddChannelOrUpdateChannelStatus(db, channelRecord)
		if err != nil {
			return errors.Wrapf(err, "Adding or updating channel status (channelId: %v shortChannelId: %v)", channelId, shortChannelId)
		}
		// skip if we already have channel close channel event for this channel
		for _, existingChannelId := range existingChannelIds {
			if channelId == existingChannelId {
				continue icoLoop
			}
		}

		jb, err := json.Marshal(channel)
		if err != nil {
			return errors.Wrap(err, "JSON Marshal")
		}

		err = insertChannelEvent(db, time.Now().UTC(), lnrpc.ChannelEventUpdate_CLOSED_CHANNEL, nodeSettings.NodeId, channelId, true, jb)
		if err != nil {
			return errors.Wrap(err, "Enrich and insert channel event")
		}
	}
	return nil
}

func insertChannelEvent(db *sqlx.DB, eventTime time.Time, eventType lnrpc.ChannelEventUpdate_UpdateType,
	nodeId, channelId int, imported bool, jb []byte) error {

	var sqlStm = `INSERT INTO channel_event (time, event_type, channel_id, imported, event, node_id)
		VALUES($1, $2, $3, $4, $5, $6);`

	_, err := db.Exec(sqlStm, eventTime, eventType, channelId, imported, jb, nodeId)
	if err != nil {
		return errors.Wrap(err, "DB Exec")
	}
	return nil
}
