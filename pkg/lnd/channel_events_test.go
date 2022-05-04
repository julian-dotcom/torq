package lnd

import (
	"context"
	"database/sql"
	"io"
	// "github.com/cockroachdb/errors"
	// "github.com/jmoiron/sqlx"
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
	Channels []interface{}
}

func (s *stubLNDSubscribeChannelEventRPC) Recv() (*lnrpc.ChannelEventUpdate, error) {
	if len(s.Channels) == 0 {
		return nil, io.EOF
	}
	var channel interface{}
	channel, s.Channels = s.Channels[0], nil
	if openChannel, ok := channel.(lnrpc.ChannelEventUpdate_OpenChannel); ok {
		return &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_OPEN_CHANNEL,
			Channel: &openChannel}, nil
	}
	return nil, io.EOF
}

func (s *stubLNDSubscribeChannelEventRPC) AddChannelEvent(channel *lnrpc.ChannelEventUpdate_OpenChannel) error {
	s.Channels = append(s.Channels, channel)
	return nil
}

type stubLNDSubscribeChannelEvent struct {
	Channels []interface{}
}

func (c *stubLNDSubscribeChannelEvent) SubscribeChannelEvents(
	ctx context.Context, in *lnrpc.ChannelEventSubscription,
	opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error) {

	return &stubLNDSubscribeChannelEventRPC{Channels: c.Channels}, nil
}

func TestSubscribeChannelEvents(t *testing.T) {
	ctx := context.Background()
	errs, ctx := errgroup.WithContext(ctx)

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, err := srv.NewTestDatabase(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	pubKeyChan := make(chan string, 1)
	chanPointChan := make(chan string, 1)

	expectedEvent := &lnrpc.Channel{ChanId: 1337, ChannelPoint: "point break", RemotePubkey: "remote pub key"}
	client := &stubLNDSubscribeChannelEvent{Channels: []interface{}{
		lnrpc.ChannelEventUpdate_OpenChannel{OpenChannel: expectedEvent}}}

	errs.Go(func() error {
		err = SubscribeAndStoreChannelEvents(ctx, client, db, pubKeyChan, chanPointChan)
		if err != nil {
			t.Fatalf("Problem subscribing to channel events: %v", err)
		}
		return err
	})

	// wait for subscriptions to complete
	err = errs.Wait()
	if err != nil {
		t.Fatal(err)
	}

	type channelEvent struct {
		Chan_point string
		Pub_key    string
		Event_type int
	}
	var channelEvents []channelEvent
	err = db.Select(&channelEvents, "select chan_point, pub_key, event_type FROM channel_event WHERE chan_id = $1;",
		expectedEvent.ChanId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// t.Fatalf("Error was %v", err)
			t.Fatal("There were no channel events but I did expect there to be some")
		}
		t.Fatalf("Problem executing sql: %v", err)
	}

	if len(channelEvents) != 1 {
		t.Fatal("Expected to get a single open channel event record")
	}

	if channelEvents[0].Chan_point != expectedEvent.ChannelPoint ||
		channelEvents[0].Pub_key != expectedEvent.RemotePubkey ||
		channelEvents[0].Event_type != int(lnrpc.ChannelEventUpdate_OPEN_CHANNEL) {
		t.Fatal("Data not stored correctly")
	}

	db.Close()
	err = srv.Cleanup()
	if err != nil {
		t.Fatal(err)
	}
}
