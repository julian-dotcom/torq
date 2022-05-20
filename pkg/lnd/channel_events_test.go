package lnd

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/testutil"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"io"
	"testing"
)

type stubLNDSubscribeChannelEventRPC struct {
	grpc.ClientStream
	ChannelEvents []interface{}
}

func (s *stubLNDSubscribeChannelEventRPC) Recv() (*lnrpc.ChannelEventUpdate, error) {
	if len(s.ChannelEvents) == 0 {
		return nil, io.EOF
	}
	var channelEvent interface{}
	channelEvent, s.ChannelEvents = s.ChannelEvents[0], nil
	if eventUpdate, ok := channelEvent.(*lnrpc.ChannelEventUpdate); ok {
		return eventUpdate, nil
	}
	return nil, io.EOF
}

type stubLNDSubscribeChannelEvent struct {
	ChannelEvents []interface{}
}

func (c *stubLNDSubscribeChannelEvent) SubscribeChannelEvents(
	ctx context.Context, in *lnrpc.ChannelEventSubscription,
	opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error) {

	return &stubLNDSubscribeChannelEventRPC{ChannelEvents: c.ChannelEvents}, nil
}

func TestSubscribeChannelEvents(t *testing.T) {
	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, err := srv.NewTestDatabase(true)
	if err != nil {
		t.Fatal(err)
	}

	// testDBCleanup := func() {
	// 	db.Close()
	// 	err = srv.Cleanup()
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// }
	// defer testDBCleanup()

	t.Run("Open Channel Event", func(t *testing.T) {
		expected := channelEventData{Chan_id: 1337, Chan_point: "point break", Pub_key: "remote pub key",
			Event_type: int(lnrpc.ChannelEventUpdate_OPEN_CHANNEL)}
		channel := &lnrpc.Channel{ChanId: expected.Chan_id, ChannelPoint: expected.Chan_point, RemotePubkey: expected.Pub_key}
		channelEvent := lnrpc.ChannelEventUpdate_OpenChannel{OpenChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_OPEN_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Closed Channel Event", func(t *testing.T) {
		expected := channelEventData{Chan_id: 1337, Chan_point: "closed point break", Pub_key: "closed remote pub key",
			Event_type: int(lnrpc.ChannelEventUpdate_CLOSED_CHANNEL)}
		channel := &lnrpc.ChannelCloseSummary{ChanId: expected.Chan_id, ChannelPoint: expected.Chan_point, RemotePubkey: expected.Pub_key}
		channelEvent := lnrpc.ChannelEventUpdate_ClosedChannel{ClosedChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_CLOSED_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Fully Resolved Channel Event", func(t *testing.T) {
		expected := channelEventData{Chan_id: 0, Chan_point: "0101010101010101010101010101010101010101010101010101010101010101:3", Pub_key: "",
			Event_type: int(lnrpc.ChannelEventUpdate_FULLY_RESOLVED_CHANNEL)}
		fundingTxBytes := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		channel := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.ChannelPoint_FundingTxidBytes{FundingTxidBytes: fundingTxBytes}, OutputIndex: 3}
		channelEvent := lnrpc.ChannelEventUpdate_FullyResolvedChannel{FullyResolvedChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_FULLY_RESOLVED_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Active Channel Event", func(t *testing.T) {
		expected := channelEventData{Chan_id: 0, Chan_point: "0101010101010101010101010101010101010101010101010101010101010102:3", Pub_key: "",
			Event_type: int(lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL)}
		fundingTxBytes := []byte{2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		channel := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.ChannelPoint_FundingTxidBytes{FundingTxidBytes: fundingTxBytes}, OutputIndex: 3}
		channelEvent := lnrpc.ChannelEventUpdate_ActiveChannel{ActiveChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Inactive Channel Event", func(t *testing.T) {
		expected := channelEventData{Chan_id: 0, Chan_point: "0101010101010101010101010101010101010101010101010101010101010103:3", Pub_key: "",
			Event_type: int(lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL)}
		fundingTxBytes := []byte{3, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		channel := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.ChannelPoint_FundingTxidBytes{FundingTxidBytes: fundingTxBytes}, OutputIndex: 3}
		channelEvent := lnrpc.ChannelEventUpdate_InactiveChannel{InactiveChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Pending Open Channel Event", func(t *testing.T) {
		expected := channelEventData{Chan_id: 0, Chan_point: "0101010101010101010101010101010101010101010101010101010101010104:3", Pub_key: "",
			Event_type: int(lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL)}
		TxBytes := []byte{4, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		channel := &lnrpc.PendingUpdate{Txid: TxBytes, OutputIndex: 3}
		channelEvent := lnrpc.ChannelEventUpdate_PendingOpenChannel{PendingOpenChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Newly discovered channel is stored", func(t *testing.T) {
		// The open channel and closed channel tests above used the same chan id of 1337
		// so a single new channel record should have been created

		type channel struct {
			Short_channel_id    string
			Channel_point       string
			Destination_pub_key string
		}
		var channels []channel
		err = db.Select(&channels, `SELECT short_channel_id, channel_point, destination_pub_key FROM channel;`)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// t.Fatalf("Error was %v", err)
				t.Fatal("There were no channel records but I did expect there to be one")
			}
			t.Fatalf("Problem executing sql: %v", err)
		}

		if len(channels) != 1 {
			t.Fatal("Expected to find a single channel record stored in the database for these channel event updates")
		}

		if channels[0].Short_channel_id != "0:0:1337" ||
			channels[0].Channel_point != "point break" ||
			channels[0].Destination_pub_key != "remote pub key" {

			t.Fatal("Channel data not stored correctly")
		}

	})
}

type channelEventData struct {
	Chan_id    uint64
	Chan_point string
	Pub_key    string
	Event_type int
}

func runChannelEventTest(t *testing.T, db *sqlx.DB, channelEvent interface{}, expected channelEventData) {
	ctx := context.Background()
	errs, ctx := errgroup.WithContext(ctx)

	pubKeyChan := make(chan string, 1)
	chanPointChan := make(chan string, 1)

	client := &stubLNDSubscribeChannelEvent{ChannelEvents: []interface{}{channelEvent}}

	errs.Go(func() error {
		err := SubscribeAndStoreChannelEvents(ctx, client, db, pubKeyChan, chanPointChan)
		if err != nil {
			t.Fatalf("Problem subscribing to channel events: %v", err)
		}
		return err
	})

	// wait for subscriptions to complete
	err := errs.Wait()
	if err != nil {
		t.Fatal(err)
	}

	var channelEvents []channelEventData
	err = db.Select(&channelEvents, `
SELECT chan_id, pub_key
FROM channel_event
WHERE chan_point = $1 AND event_type = $2;`, expected.Chan_point, expected.Event_type)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// t.Fatalf("Error was %v", err)
			t.Fatal("There were no channel events but I did expect there to be some")
		}
		t.Fatalf("Problem executing sql: %v", err)
	}

	if len(channelEvents) != 1 {
		t.Fatal("Expected to find a single channel event record stored in the database for this channel event update")
	}

	if channelEvents[0].Chan_id != expected.Chan_id ||
		channelEvents[0].Pub_key != expected.Pub_key {
		t.Fatal("Channel event data not stored correctly")
	}
}
