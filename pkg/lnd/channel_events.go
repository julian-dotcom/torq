package lnd

import (
	"context"
	"database/sql"
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
	"gopkg.in/guregu/null.v4"
)

// websocket channel event
type wsChannelEvent struct {
	Type             string `json:"type"`
	ChannelEventType string `json:"channelEventType"`
	ShortChannelId   string `json:"shortChannelId,omitempty"`
	LNDChannelPoint  string `json:"lndChannelPoint"`
	PubKey           string `json:"pubKey,omitempty"`
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
func storeChannelEvent(db *sqlx.DB, ce *lnrpc.ChannelEventUpdate, nodeId int, wsChan chan interface{}) error {

	timestampMs := time.Now().UTC()

	var chanID uint64
	var channelPoint string
	var publicKey string
	var wsChanEvent wsChannelEvent

	wsChanEvent.Type = "channelEvent"
	wsChanEvent.ChannelEventType = ce.GetType().String()

	switch ce.Type {
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		c := ce.GetOpenChannel()
		chanID = c.ChanId
		channelPoint = c.ChannelPoint
		publicKey = c.RemotePubkey
		shortChannelId := channels.ConvertLNDShortChannelID(chanID)

		node := nodes.Node{
			PublicKey: c.RemotePubkey,
		}
		remoteNodeId, err := nodes.AddNodeWhenNew(db, node)
		if err != nil {
			return errors.Wrap(err, "Adding node")
		}

		// This allows torq to listen to the graph for node updates
		commons.ManagedNodeChannel <- commons.ManagedNode{
			NodeId:    remoteNodeId,
			PublicKey: c.RemotePubkey,
			Type:      commons.WRITE_CHANNEL_NODE,
		}

		channel := channels.Channel{
			ShortChannelID:    shortChannelId,
			LNDChannelPoint:   null.StringFrom(channelPoint),
			FirstNodeId:       nodeId,
			SecondNodeId:      remoteNodeId,
			LNDShortChannelID: chanID,
			Status:            channels.Open,
		}
		channelId, err := channels.AddChannelOrUpdateChannelStatus(db, channel)
		if err != nil {
			return errors.Wrap(err, "Adding channel")
		}

		// This allows torq to listen to the graph for channel updates
		commons.ManagedChannelChannel <- commons.ManagedChannel{
			ChannelId:       channelId,
			ShortChannelId:  channel.ShortChannelID,
			LndChannelPoint: channelPoint,
			StatusId:        int(channels.Open),
			Type:            commons.WRITE_CHANNEL,
		}

		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, channelId, false, chanID, channelPoint, publicKey, jb)
		if err != nil {
			return errors.Wrap(err, "Insert Open Channel Event")
		}

		wsChanEvent.ShortChannelId = shortChannelId
		wsChanEvent.LNDChannelPoint = channelPoint
		wsChanEvent.PubKey = publicKey
		wsChan <- wsChanEvent

		return nil

	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
		c := ce.GetClosedChannel()
		chanID = c.ChanId
		channelPoint = c.ChannelPoint
		publicKey = c.RemotePubkey
		shortChannelId := channels.ConvertLNDShortChannelID(c.ChanId)
		channelId := commons.GetChannelIdFromShortChannelId(shortChannelId)

		// This stops the graph from listening to channel updates
		managedChannel := commons.ManagedChannel{
			ShortChannelId:  shortChannelId,
			LndChannelPoint: channelPoint,
			ChannelId:       channelId,
			Type:            commons.DELETE_CHANNEL,
		}
		commons.ManagedChannelChannel <- managedChannel

		var err error
		remoteNodeId := commons.GetNodeIdFromPublicKey(publicKey)
		if remoteNodeId == 0 {
			newNode := nodes.Node{
				PublicKey: publicKey,
			}
			remoteNodeId, err = nodes.AddNodeWhenNew(db, newNode)
			if err != nil {
				return errors.Wrapf(err, "Adding node (publicKey: %v shortChannelId: %v)", publicKey, shortChannelId)
			}
		}

		channel := channels.Channel{
			ShortChannelID:    shortChannelId,
			LNDChannelPoint:   null.StringFrom(channelPoint),
			FirstNodeId:       nodeId,
			SecondNodeId:      remoteNodeId,
			LNDShortChannelID: chanID,
			Status:            channels.GetClosureStatus(c.CloseType),
		}
		channelId, err = channels.AddChannelOrUpdateChannelStatus(db, channel)
		if err != nil {
			return errors.Wrapf(err, "Adding or updating channel (channelId: %v, shortChannelId: %v)", channelId, shortChannelId)
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, channelId, false, chanID, channelPoint, publicKey, jb)
		if err != nil {
			return errors.Wrap(err, "Insert Closed Channel Event")
		}

		// This stops the graph from listening to node updates
		chans, err := channels.GetChannelsForNodeId(db, remoteNodeId)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to verify if remote node still has open channels: %v", remoteNodeId)
		}
		if len(chans) == 0 {
			managedNode := commons.ManagedNode{
				PublicKey: publicKey,
				Type:      commons.DELETE_CHANNEL_NODE,
			}
			commons.ManagedNodeChannel <- managedNode
		}

		wsChanEvent.ShortChannelId = shortChannelId
		wsChanEvent.LNDChannelPoint = channelPoint
		wsChanEvent.PubKey = publicKey
		wsChan <- wsChanEvent

		return nil

	case lnrpc.ChannelEventUpdate_FULLY_RESOLVED_CHANNEL:
		c := ce.GetFullyResolvedChannel()
		channelPoint, err := chanPointFromByte(c.GetFundingTxidBytes(), c.GetOutputIndex())
		channelId := commons.GetChannelIdFromChannelPoint(channelPoint)
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, channelId, false, chanID, channelPoint, "", jb)
		if err != nil {
			return errors.Wrap(err, "Insert Fully Resolved Channel Event")
		}

		wsChanEvent.LNDChannelPoint = channelPoint
		wsChan <- wsChanEvent

	case lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL:
		c := ce.GetActiveChannel()
		channelPoint, err := chanPointFromByte(c.GetFundingTxidBytes(), c.GetOutputIndex())
		channelId := commons.GetChannelIdFromChannelPoint(channelPoint)
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, channelId, false, 0, channelPoint, "", jb)
		if err != nil {
			return errors.Wrap(err, "Insert Active Channel Event")
		}

		wsChanEvent.LNDChannelPoint = channelPoint
		wsChan <- wsChanEvent

		return nil
	case lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL:
		c := ce.GetInactiveChannel()
		channelPoint, err := chanPointFromByte(c.GetFundingTxidBytes(), c.GetOutputIndex())
		channelId := commons.GetChannelIdFromChannelPoint(channelPoint)
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, channelId, false, 0, channelPoint, "", jb)
		if err != nil {
			return errors.Wrap(err, "Insert Inactive Channel Event")
		}

		wsChanEvent.LNDChannelPoint = channelPoint
		wsChan <- wsChanEvent

		return nil
	case lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL:
		c := ce.GetPendingOpenChannel()
		channelPoint, err := chanPointFromByte(c.GetTxid(), c.GetOutputIndex())
		channelId := commons.GetChannelIdFromChannelPoint(channelPoint)
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall")
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, channelId, false, 0, channelPoint, "", jb)
		if err != nil {
			return errors.Wrap(err, "Insert Pending Open Channel Event")
		}

		wsChanEvent.LNDChannelPoint = channelPoint
		wsChan <- wsChanEvent

		return nil
	default:
	}

	return nil
}

type lndClientSubscribeChannelEvent interface {
	SubscribeChannelEvents(ctx context.Context, in *lnrpc.ChannelEventSubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error)
}

// SubscribeAndStoreChannelEvents Subscribes to channel events from LND and stores them in the
// database as a time series
func SubscribeAndStoreChannelEvents(ctx context.Context, client lndClientSubscribeChannelEvent,
	db *sqlx.DB, nodeId int, wsChan chan interface{}) error {

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

		err = storeChannelEvent(db, chanEvent, nodeId, wsChan)
		if err != nil {
			log.Error().Err(err).Msg("Subscribe channel events store event error")
			// rate limit for caution but hopefully not needed
			rl.Take()
			continue
		}

	}

	return nil
}

func ImportChannelList(t lnrpc.ChannelEventUpdate_UpdateType, db *sqlx.DB, client lnrpc.LightningClient, nodeId int) error {

	ctx := context.Background()
	switch t {
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		req := lnrpc.ListChannelsRequest{}
		r, err := client.ListChannels(ctx, &req)
		if err != nil {
			return errors.Wrap(err, "LND: List channels")
		}

		err = storeImportedOpenChannels(db, r.Channels, nodeId)
		if err != nil {
			return errors.Wrap(err, "Store imported open channels")
		}

	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
		req := lnrpc.ClosedChannelsRequest{}
		r, err := client.ClosedChannels(ctx, &req)
		if err != nil {
			return errors.Wrap(err, "LND: Get closed channels")
		}

		err = storeImportedClosedChannels(db, r.Channels, nodeId)
		if err != nil {
			return errors.Wrap(err, "Store imported closed channels")
		}

	}

	return nil
}

func getExistingChannelEvents(t lnrpc.ChannelEventUpdate_UpdateType, db *sqlx.DB, cp []string) ([]string, error) {
	// Prepare the query with an array of channel points
	q := "select lnd_channel_point from channel_event where (lnd_channel_point in (?)) and (event_type = ?);"
	qs, args, err := sqlx.In(q, cp, t)
	if err != nil {
		return []string{}, errors.Wrap(err, "SQLX In")
	}

	// Query and create the list of existing channel points (ecp)
	var ecp []string
	qsr := db.Rebind(qs)
	rows, err := db.Query(qsr, args...)
	if err != nil {
		return []string{}, errors.Wrap(err, "DB Query")
	}
	for rows.Next() {
		var cp sql.NullString
		err = rows.Scan(&cp)
		if err != nil {
			return nil, err
		}
		if cp.Valid {
			ecp = append(ecp, cp.String)
		}
	}

	return ecp, nil
}

func enrichAndInsertChannelEvent(db *sqlx.DB, eventType lnrpc.ChannelEventUpdate_UpdateType,
	channelId int, imported bool, chanId uint64, chanPoint string, pubKey string, jb []byte) error {

	// Use current time for imported channel events (open/close).
	// The time used to open/close events is the timestamp of the opening transaction.
	timestampMs := time.Now().UTC()

	err := insertChannelEvent(db, timestampMs, eventType, channelId, imported, chanId, chanPoint, pubKey, jb)
	if err != nil {
		return errors.Wrap(err, "Insert channel event")
	}
	return nil
}

func storeImportedOpenChannels(db *sqlx.DB, c []*lnrpc.Channel, nodeId int) error {

	if len(c) == 0 {
		return nil
	}

	// Creates a list of channel points in the request result.
	var cp []string
	for _, channel := range c {
		cp = append(cp, channel.ChannelPoint)
	}

	ecp, err := getExistingChannelEvents(lnrpc.ChannelEventUpdate_OPEN_CHANNEL, db, cp)
	if err != nil {
		return err
	}

icoLoop:
	for _, channel := range c {
		shortChannelId := channels.ConvertLNDShortChannelID(channel.ChanId)
		remoteNodeId := commons.GetNodeIdFromPublicKey(channel.RemotePubkey)
		if remoteNodeId == 0 {
			newNode := nodes.Node{
				PublicKey: channel.RemotePubkey,
			}
			remoteNodeId, err = nodes.AddNodeWhenNew(db, newNode)
			if err != nil {
				return errors.Wrapf(err, "Adding node (publicKey: %v shortChannelId: %v)", channel.RemotePubkey, shortChannelId)
			}
		}

		// check if we have seen this channel before and if not store in the channel table
		channelRecord := channels.Channel{
			ShortChannelID:    shortChannelId,
			LNDChannelPoint:   null.StringFrom(channel.ChannelPoint),
			FirstNodeId:       nodeId,
			SecondNodeId:      remoteNodeId,
			LNDShortChannelID: channel.ChanId,
			Status:            channels.Open,
		}
		channelId, err := channels.AddChannelOrUpdateChannelStatus(db, channelRecord)
		if err != nil {
			return errors.Wrapf(err, "Adding or updating channel status (channelId: %v shortChannelId: %v)", channelId, shortChannelId)
		}
		// skip if we have an existing channel open channel event
		for _, e := range ecp {
			if channel.ChannelPoint == e {
				continue icoLoop
			}
		}

		jb, err := json.Marshal(channel)
		if err != nil {
			return errors.Wrap(err, "JSON Marshal")
		}

		err = enrichAndInsertChannelEvent(db, lnrpc.ChannelEventUpdate_OPEN_CHANNEL, channelId,
			true, channel.ChanId, channel.ChannelPoint, channel.RemotePubkey, jb)
		if err != nil {
			return errors.Wrap(err, "Enrich and insert channel event")
		}
	}
	return nil
}

func storeImportedClosedChannels(db *sqlx.DB, c []*lnrpc.ChannelCloseSummary, nodeId int) error {

	if len(c) == 0 {
		return nil
	}
	// Creates a list of channel points in the request result.
	var cp []string
	for _, channel := range c {
		cp = append(cp, channel.ChannelPoint)
	}

	ecp, err := getExistingChannelEvents(lnrpc.ChannelEventUpdate_CLOSED_CHANNEL, db, cp)
	if err != nil {
		return err
	}

icoLoop:
	for _, channel := range c {
		shortChannelId := channels.ConvertLNDShortChannelID(channel.ChanId)
		remoteNodeId := commons.GetNodeIdFromPublicKey(channel.RemotePubkey)
		if remoteNodeId == 0 {
			newNode := nodes.Node{
				PublicKey: channel.RemotePubkey,
			}
			remoteNodeId, err = nodes.AddNodeWhenNew(db, newNode)
			if err != nil {
				return errors.Wrapf(err, "Adding node (publicKey: %v shortChannelId: %v)", channel.RemotePubkey, shortChannelId)
			}
		}

		// check if we have seen this channel before and if not store in the channel table
		channelRecord := channels.Channel{
			ShortChannelID:    shortChannelId,
			LNDChannelPoint:   null.StringFrom(channel.ChannelPoint),
			FirstNodeId:       nodeId,
			SecondNodeId:      remoteNodeId,
			LNDShortChannelID: channel.ChanId,
			Status:            channels.GetClosureStatus(channel.CloseType),
		}
		channelId, err := channels.AddChannelOrUpdateChannelStatus(db, channelRecord)
		if err != nil {
			return errors.Wrapf(err, "Adding or updating channel status (channelId: %v shortChannelId: %v)", channelId, shortChannelId)
		}
		// skip if we already have channel close channel event for this channel
		for _, e := range ecp {
			if channel.ChannelPoint == e {
				continue icoLoop
			}
		}

		jb, err := json.Marshal(channel)
		if err != nil {
			return errors.Wrap(err, "JSON Marshal")
		}

		err = enrichAndInsertChannelEvent(db, lnrpc.ChannelEventUpdate_CLOSED_CHANNEL, channelId,
			true, channel.ChanId, channel.ChannelPoint, channel.RemotePubkey, jb)
		if err != nil {
			return errors.Wrap(err, "Enrich and insert channel event")
		}
	}
	return nil
}

func insertChannelEvent(db *sqlx.DB, eventTime time.Time, eventType lnrpc.ChannelEventUpdate_UpdateType,
	channelId int, imported bool, lndShortChannelId uint64, lndChannelPoint string, pubKey string, jb []byte) error {

	shortChannelId := channels.ConvertLNDShortChannelID(lndShortChannelId)

	var sqlStm = `INSERT INTO channel_event
		(time, event_type, channel_id, imported, short_channel_id, lnd_short_channel_id, lnd_channel_point, pub_key, event)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`

	_, err := db.Exec(sqlStm, eventTime, eventType, channelId, imported, shortChannelId, lndShortChannelId, lndChannelPoint, pubKey, jb)
	if err != nil {
		return errors.Wrap(err, "DB Exec")
	}
	return nil
}
