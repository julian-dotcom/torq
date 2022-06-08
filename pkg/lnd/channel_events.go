package lnd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/channels"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
	"gopkg.in/guregu/null.v4"
	"time"
)

func getChanPoint(cb []byte, oi uint32) (string, error) {

	ch, err := chainhash.NewHash(cb)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", ch.String(), oi), nil
}

// storeChannelEvent extracts the timestamp, channel ID and PubKey from the
// ChannelEvent and converts the original struct to json.
// Then it's stored in the database in the channel_event table.
func storeChannelEvent(db *sqlx.DB, ce *lnrpc.ChannelEventUpdate,
	pubKeyChan chan string, chanPointChan chan string) error {

	timestampMs := time.Now().UTC()

	var ChanID uint64
	var ChannelPoint string
	var PubKey string

	switch ce.Type {
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		c := ce.GetOpenChannel()
		ChanID = c.ChanId
		ChannelPoint = c.ChannelPoint
		PubKey = c.RemotePubkey

		// Add the remote public key to the list to listen to for graph updates.
		pubKeyChan <- c.RemotePubkey

		// Add the channel point to the chanPointList, this allows the
		// channel graph to listen for routing policy updates
		chanPointChan <- c.ChannelPoint

		channel := channels.Channel{
			ShortChannelID:    channels.ConvertLNDShortChannelID(ChanID),
			ChannelPoint:      null.StringFrom(ChannelPoint),
			DestinationPubKey: null.StringFrom(PubKey),
		}

		err := channels.AddChannelRecordIfDoesntExist(db, channel)
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("storeChannelEvent -> json.Marshal(%v): %v", c, err)
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		if err != nil {
			return errors.Wrapf(err, `storeChannelEvent -> insertChannelEventExec(%v, %s, %s, %t, %d, %s, %s, %v)`,
				db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		}

		return nil

	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
		c := ce.GetClosedChannel()
		ChanID = c.ChanId
		ChannelPoint = c.ChannelPoint
		PubKey = c.RemotePubkey

		// Updates the channel point list by removing the channel point from the chanPointList.
		chanPointChan <- c.ChannelPoint

		channel := channels.Channel{
			ShortChannelID:    channels.ConvertLNDShortChannelID(ChanID),
			ChannelPoint:      null.StringFrom(ChannelPoint),
			DestinationPubKey: null.StringFrom(PubKey),
		}
		err := channels.AddChannelRecordIfDoesntExist(db, channel)
		if err != nil {
			return err
		}

		jb, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("storeChannelEvent -> json.Marshal(%v): %v", c, err)
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		if err != nil {
			return errors.Wrapf(err, `storeChannelEvent -> insertChannelEventExec(%v, %s, %s, %t, %d, %s, %s, %v)`,
				db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		}

		return nil

	case lnrpc.ChannelEventUpdate_FULLY_RESOLVED_CHANNEL:
		c := ce.GetFullyResolvedChannel()
		ChannelPoint, err := getChanPoint(c.GetFundingTxidBytes(), c.GetOutputIndex())
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("storeChannelEvent -> json.Marshal(%v): %v", c, err)
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		if err != nil {
			return errors.Wrapf(err, `storeChannelEvent -> insertChannelEventExec(%v, %s, %s, %t, %d, %s, %s, %v)`,
				db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		}
	case lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL:
		c := ce.GetActiveChannel()
		ChannelPoint, err := getChanPoint(c.GetFundingTxidBytes(), c.GetOutputIndex())
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("storeChannelEvent -> json.Marshal(%v): %v", c, err)
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		if err != nil {
			return errors.Wrapf(err, `storeChannelEvent -> insertChannelEventExec(%v, %s, %s, %t, %d, %s, %s, %v)`,
				db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		}
		return nil
	case lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL:
		c := ce.GetInactiveChannel()
		ChannelPoint, err := getChanPoint(c.GetFundingTxidBytes(), c.GetOutputIndex())
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("storeChannelEvent -> json.Marshal(%v): %v", c, err)
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		if err != nil {
			return errors.Wrapf(err, `storeChannelEvent -> insertChannelEventExec(%v, %s, %s, %t, %d, %s, %s, %v)`,
				db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		}
		return nil
	case lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL:
		c := ce.GetPendingOpenChannel()
		ChannelPoint, err := getChanPoint(c.GetTxid(), c.GetOutputIndex())
		if err != nil {
			return err
		}
		jb, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("storeChannelEvent -> json.Marshal(%v): %v", c, err)
		}
		err = insertChannelEvent(db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		if err != nil {
			return errors.Wrapf(err, `storeChannelEvent -> insertChannelEventExec(%v, %s, %s, %t, %d, %s, %s, %v)`,
				db, timestampMs, ce.Type, false, ChanID, ChannelPoint, PubKey, jb)
		}
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
	db *sqlx.DB, pubKeyChan chan string, chanPoinChan chan string) error {

	cesr := lnrpc.ChannelEventSubscription{}
	stream, err := client.SubscribeChannelEvents(ctx, &cesr)
	if err != nil {
		return errors.Wrapf(err, "SubscribeAndStoreChannelEvents -> client.SubscribeChannelEvents(%v, %v)",
			ctx, cesr)
	}

	rl := ratelimit.New(1) // 1 per second maximum rate limit
	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		chanEvent, err := stream.Recv()

		if err != nil {
			fmt.Printf("Subscribe channel events stream receive error: %v", err)
			// rate limited resubscribe
			rl.Take()
			stream, err = client.SubscribeChannelEvents(ctx, &cesr)
			continue
		}

		err = storeChannelEvent(db, chanEvent, pubKeyChan, chanPoinChan)
		if err != nil {
			fmt.Printf("Subscribe channel events store event error: %v", err)
			// rate limit for caution but hopefully not needed
			rl.Take()
			continue
		}

	}

	return nil
}

func ImportChannelList(t lnrpc.ChannelEventUpdate_UpdateType, db *sqlx.DB, client lnrpc.LightningClient) error {

	ctx := context.Background()
	switch t {
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		req := lnrpc.ListChannelsRequest{}
		r, err := client.ListChannels(ctx, &req)
		if err != nil {
			return errors.Wrapf(err, "ImportChannelList -> client.ListChannels(%v, %v)", ctx, req)
		}

		err = storeImportedOpenChannels(db, r.Channels)
		if err != nil {
			return errors.Wrapf(err, "ImportChannelList -> storeImportedOpenChannels(%v, %v)", db, r.Channels)
		}

	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
		req := lnrpc.ClosedChannelsRequest{}
		r, err := client.ClosedChannels(ctx, &req)
		if err != nil {
			return errors.Wrapf(err, "ImportChannelList -> client.ClosedChannels(%v, %v)", ctx, req)
		}

		err = storeImportedClosedChannels(db, r.Channels)
		if err != nil {
			return errors.Wrapf(err, "ImportChannelList -> storeImportedClosedChannels(%v, %v)", db, r.Channels)
		}

	}

	return nil
}

func getExistingChannelEvents(t lnrpc.ChannelEventUpdate_UpdateType, db *sqlx.DB, cp []string) ([]string, error) {
	// Prepare the query with an array of channel points
	q := "select chan_point from channel_event where (chan_point in (?)) and (event_type = ?);"
	qs, args, err := sqlx.In(q, cp, t)
	if err != nil {
		return []string{}, errors.Wrapf(err, "sqlx.In(%s, %v, %d)", q, cp, t)
	}

	// Query and create the list of existing channel points (ecp)
	var ecp []string
	qsr := db.Rebind(qs)
	rows, err := db.Query(qsr, args...)
	if err != nil {
		return []string{}, errors.Wrapf(err, "getExistingChannelEvents -> db.Query(qsr, args...)")
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

func enrichAndInsertChannelEvent(db *sqlx.DB, eventType lnrpc.ChannelEventUpdate_UpdateType, imported bool, chanId uint64, chanPoint string, pubKey string, jb []byte) error {

	// Use current time for imported channel events (open/close).
	// The time used to open/close events is the timestamp of the opening transaction.
	timestampMs := time.Now().UTC()

	err := insertChannelEvent(db, timestampMs, eventType, imported, chanId, chanPoint, pubKey, jb)
	if err != nil {
		return errors.Wrapf(err, "storeChannelOpenList -> "+
			"insertChannelEventExec(%v, %s, %s, %t, %d, %s, %s, %v)",
			db, timestampMs, eventType, imported, chanId, chanPoint, pubKey, jb)
	}
	return nil
}

func storeImportedOpenChannels(db *sqlx.DB, c []*lnrpc.Channel) error {

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

		for _, e := range ecp {
			if channel.ChannelPoint == e {
				continue icoLoop
			}
		}

		jb, err := json.Marshal(channel)
		if err != nil {
			return errors.Wrapf(err, "storeChannelList -> json.Marshal(%v)", channel)
		}

		// check if we have seen this channel before and if not store in the channel table
		channelRecord := channels.Channel{
			ShortChannelID:    channels.ConvertLNDShortChannelID(channel.ChanId),
			ChannelPoint:      null.StringFrom(channel.ChannelPoint),
			DestinationPubKey: null.StringFrom(channel.RemotePubkey),
		}
		err = channels.AddChannelRecordIfDoesntExist(db, channelRecord)
		if err != nil {
			return err
		}

		err = enrichAndInsertChannelEvent(db, lnrpc.ChannelEventUpdate_OPEN_CHANNEL,
			true, channel.ChanId, channel.ChannelPoint, channel.RemotePubkey, jb)
		if err != nil {
			return errors.Wrapf(err, "storeChannelOpenList -> "+
				"enrichAndInsertChannelEvent(%v, %d, %t, %d, %s, %s, %v)", db,
				lnrpc.ChannelEventUpdate_OPEN_CHANNEL, true, channel.ChanId, channel.ChannelPoint,
				channel.RemotePubkey, jb)
		}
	}
	return nil
}

func storeImportedClosedChannels(db *sqlx.DB, c []*lnrpc.ChannelCloseSummary) error {

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

		for _, e := range ecp {
			if channel.ChannelPoint == e {
				continue icoLoop
			}
		}

		jb, err := json.Marshal(channel)
		if err != nil {
			return errors.Wrapf(err, "storeChannelList -> json.Marshal(%v)", channel)
		}

		// check if we have seen this channel before and if not store in the channel table
		channelRecord := channels.Channel{
			ShortChannelID:    channels.ConvertLNDShortChannelID(channel.ChanId),
			ChannelPoint:      null.StringFrom(channel.ChannelPoint),
			DestinationPubKey: null.StringFrom(channel.RemotePubkey),
		}
		err = channels.AddChannelRecordIfDoesntExist(db, channelRecord)
		if err != nil {
			return err
		}

		err = enrichAndInsertChannelEvent(db, lnrpc.ChannelEventUpdate_CLOSED_CHANNEL,
			true, channel.ChanId, channel.ChannelPoint, channel.RemotePubkey, jb)
		if err != nil {
			return errors.Wrapf(err, "storeImportedClosedChannels -> "+
				"enrichAndInsertChannelEvent(%v, %s, %t, %d, %s, %s, %v)", db,
				lnrpc.ChannelEventUpdate_CLOSED_CHANNEL, true, channel.ChanId, channel.ChannelPoint,
				channel.RemotePubkey, jb)
		}
	}
	return nil
}

var sqlStm = `INSERT INTO channel_event (time, event_type, imported, chan_id, chan_point, pub_key,
	event) VALUES($1, $2, $3, $4, $5, $6, $7);`

func insertChannelEvent(db *sqlx.DB, ts time.Time, eventType lnrpc.ChannelEventUpdate_UpdateType,
	imported bool, chanId uint64, chanPoint string, pubKey string, jb []byte) error {
	_, err := db.Exec(sqlStm, ts, eventType, imported, chanId, chanPoint, pubKey, jb)
	if err != nil {
		return errors.Wrapf(err, `insertChannelEvent -> db.Exec(%s, %s, %d, %t, %d, %s, %s, %v)`,
			sqlStm, ts, eventType, imported, chanId, chanPoint, pubKey, jb)
	}
	return nil
}
