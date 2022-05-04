package lnd

import (
	"context"
	"database/sql"
	"io"
	"log"

	// "github.com/cockroachdb/errors"
	// "github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	// "github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/testutil"
	// "golang.org/x/sync/errgroup"

	// "google.golang.org/grpc"
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
		expected := channelEventData{Chan_id: 1338, Chan_point: "closed point break", Pub_key: "closed remote pub key",
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

	db.Close()
	err = srv.Cleanup()
	if err != nil {
		t.Fatal(err)
	}
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
		log.Fatalf("channel events %v", channelEvents)
		t.Fatal("Expected to get a single open channel event record")
	}

	if channelEvents[0].Chan_id != expected.Chan_id ||
		channelEvents[0].Pub_key != expected.Pub_key {
		t.Fatal("Data not stored correctly")
	}
}
