package lnd

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/testutil"
)

type stubLNDSubscribeChannelEventRPC struct {
	grpc.ClientStream
	ChannelEvents []interface{}
	CancelFunc    func()
}

func (s *stubLNDSubscribeChannelEventRPC) Recv() (*lnrpc.ChannelEventUpdate, error) {
	if len(s.ChannelEvents) == 0 {
		s.CancelFunc()
		return nil, context.Canceled
	}
	var channelEvent interface{}
	channelEvent, s.ChannelEvents = s.ChannelEvents[0], nil
	if eventUpdate, ok := channelEvent.(*lnrpc.ChannelEventUpdate); ok {
		return eventUpdate, nil
	}
	s.CancelFunc()
	return nil, context.Canceled
}

type stubLNDSubscribeChannelEvent struct {
	ChannelEvents []interface{}
	CancelFunc    func()
}

func (c *stubLNDSubscribeChannelEvent) SubscribeChannelEvents(
	ctx context.Context, in *lnrpc.ChannelEventSubscription,
	opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error) {

	return &stubLNDSubscribeChannelEventRPC{ChannelEvents: c.ChannelEvents, CancelFunc: c.CancelFunc}, nil
}

func (c *stubLNDSubscribeChannelEvent) PendingChannels(
	ctx context.Context, in *lnrpc.PendingChannelsRequest,
	opts ...grpc.CallOption) (*lnrpc.PendingChannelsResponse, error) {

	channel1 := lnrpc.PendingChannelsResponse_PendingChannel{}
	channel1.ChannelPoint = testutil.TestChannelPoint1
	channel1.RemoteNodePub = testutil.TestPublicKey2
	pendingOpenChannel1 := lnrpc.PendingChannelsResponse_PendingOpenChannel{
		Channel: &channel1,
	}
	channel2 := lnrpc.PendingChannelsResponse_PendingChannel{}
	channel2.ChannelPoint = testutil.TestChannelPoint2
	channel2.RemoteNodePub = testutil.TestPublicKey2
	pendingOpenChannel2 := lnrpc.PendingChannelsResponse_PendingOpenChannel{
		Channel: &channel2,
	}
	channel3 := lnrpc.PendingChannelsResponse_PendingChannel{}
	channel3.ChannelPoint = testutil.TestChannelPoint3
	channel3.RemoteNodePub = testutil.TestPublicKey2
	pendingOpenChannel3 := lnrpc.PendingChannelsResponse_PendingOpenChannel{
		Channel: &channel3,
	}
	channel4 := lnrpc.PendingChannelsResponse_PendingChannel{}
	channel4.ChannelPoint = testutil.TestChannelPoint4
	channel4.RemoteNodePub = testutil.TestPublicKey2
	pendingOpenChannel4 := lnrpc.PendingChannelsResponse_PendingOpenChannel{
		Channel: &channel4,
	}
	channel5 := lnrpc.PendingChannelsResponse_PendingChannel{}
	channel5.ChannelPoint = testutil.TestChannelPoint5_NOTINDB
	channel5.RemoteNodePub = testutil.TestPublicKey2
	pendingOpenChannel5 := lnrpc.PendingChannelsResponse_PendingOpenChannel{
		Channel: &channel5,
	}
	var pendingOpenChannels []*lnrpc.PendingChannelsResponse_PendingOpenChannel
	pendingOpenChannels = append(pendingOpenChannels, &pendingOpenChannel1)
	pendingOpenChannels = append(pendingOpenChannels, &pendingOpenChannel2)
	pendingOpenChannels = append(pendingOpenChannels, &pendingOpenChannel3)
	pendingOpenChannels = append(pendingOpenChannels, &pendingOpenChannel4)
	pendingOpenChannels = append(pendingOpenChannels, &pendingOpenChannel5)
	response := lnrpc.PendingChannelsResponse{
		PendingOpenChannels: pendingOpenChannels,
	}
	return &response, nil
}

func TestSubscribeChannelEvents(t *testing.T) {
	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, err := srv.NewTestDatabase(true)
	// TODO FIXME WHY?
	defer time.Sleep(1 * time.Second)
	defer cancel()
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
		expected := channelEventData{LNDShortChannelId: 1337, LNDChannelPoint: "point_break",
			FirstNodePublicKey: testutil.TestPublicKey1, SecondNodePublicKey: "remote_pub_key",
			EventType: int(lnrpc.ChannelEventUpdate_OPEN_CHANNEL), Capacity: 100000000}
		channel := &lnrpc.Channel{ChanId: expected.LNDShortChannelId, ChannelPoint: expected.LNDChannelPoint,
			RemotePubkey: expected.SecondNodePublicKey, Capacity: expected.Capacity}
		channelEvent := lnrpc.ChannelEventUpdate_OpenChannel{OpenChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_OPEN_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Fully Resolved Channel Event", func(t *testing.T) {
		expected := channelEventData{LNDChannelPoint: testutil.TestChannelPoint5_NOTINDB,
			FirstNodePublicKey: testutil.TestPublicKey1, SecondNodePublicKey: testutil.TestPublicKey2, EventType: int(lnrpc.ChannelEventUpdate_FULLY_RESOLVED_CHANNEL)}
		fundingTxBytes := []byte{5, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		channel := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.ChannelPoint_FundingTxidBytes{FundingTxidBytes: fundingTxBytes}, OutputIndex: 3}
		channelEvent := lnrpc.ChannelEventUpdate_FullyResolvedChannel{FullyResolvedChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_FULLY_RESOLVED_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Active Channel Event", func(t *testing.T) {
		expected := channelEventData{LNDShortChannelId: 2222, LNDChannelPoint: testutil.TestChannelPoint2,
			FirstNodePublicKey: testutil.TestPublicKey1, SecondNodePublicKey: testutil.TestPublicKey2, EventType: int(lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL)}
		fundingTxBytes := []byte{2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		channel := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.ChannelPoint_FundingTxidBytes{FundingTxidBytes: fundingTxBytes}, OutputIndex: 3}
		channelEvent := lnrpc.ChannelEventUpdate_ActiveChannel{ActiveChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Inactive Channel Event", func(t *testing.T) {
		expected := channelEventData{LNDShortChannelId: 3333, LNDChannelPoint: testutil.TestChannelPoint3,
			FirstNodePublicKey: testutil.TestPublicKey1, SecondNodePublicKey: testutil.TestPublicKey2, EventType: int(lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL)}
		fundingTxBytes := []byte{3, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		channel := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.ChannelPoint_FundingTxidBytes{FundingTxidBytes: fundingTxBytes}, OutputIndex: 3}
		channelEvent := lnrpc.ChannelEventUpdate_InactiveChannel{InactiveChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Pending Open Channel Event", func(t *testing.T) {
		expected := channelEventData{LNDShortChannelId: 4444, LNDChannelPoint: testutil.TestChannelPoint4,
			FirstNodePublicKey: testutil.TestPublicKey1, SecondNodePublicKey: testutil.TestPublicKey2, EventType: int(lnrpc.ChannelEventUpdate_PENDING_OPEN_CHANNEL)}
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
			Short_channel_id  string
			LND_Channel_point string
		}
		var channels []channel
		err = db.Select(&channels, `
			SELECT short_channel_id, lnd_channel_point FROM channel WHERE short_channel_id=$1;`, "0x0x1337")
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

		if channels[0].Short_channel_id != "0x0x1337" ||
			channels[0].LND_Channel_point != "point_break" {

			t.Fatal("Channel data not stored correctly")
		}

	})

	t.Run("Closed Channel Event", func(t *testing.T) {
		expected := channelEventData{LNDShortChannelId: 1337, LNDChannelPoint: "point_break", SecondNodePublicKey: "remote_pub_key",
			EventType: int(lnrpc.ChannelEventUpdate_CLOSED_CHANNEL), Capacity: 100000000}
		channel := &lnrpc.ChannelCloseSummary{ChanId: expected.LNDShortChannelId, ChannelPoint: expected.LNDChannelPoint,
			RemotePubkey: expected.SecondNodePublicKey, Capacity: expected.Capacity}
		channelEvent := lnrpc.ChannelEventUpdate_ClosedChannel{ClosedChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_CLOSED_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

}

type channelEventData struct {
	LNDShortChannelId   uint64 `db:"lnd_short_channel_id"`
	LNDChannelPoint     string `db:"lnd_channel_point"`
	FirstNodePublicKey  string `db:"first_node_public_key"`
	SecondNodePublicKey string `db:"second_node_public_key"`
	EventType           int    `db:"event_type"`
	Capacity            int64  `db:"capacity"`
}

func runChannelEventTest(t *testing.T, db *sqlx.DB, channelEvent interface{}, expected channelEventData) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	errs, ctx := errgroup.WithContext(ctx)

	client := &stubLNDSubscribeChannelEvent{ChannelEvents: []interface{}{channelEvent}, CancelFunc: cancel}
	errs.Go(func() error {
		err := SubscribeAndStoreChannelEvents(ctx, client, db,
			commons.GetNodeSettingsByNodeId(
				commons.GetNodeIdFromPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet)), nil)
		if err != nil {
			t.Fatalf("Problem subscribing to channel events: %v", err)
		}
		return err
	})

	// Wait for subscriptions to complete
	err := errs.Wait()
	if err != nil {
		t.Fatal(err)
	}

	var channelEvents []channelEventData
	err = db.Select(&channelEvents, `
			SELECT c.lnd_short_channel_id,
			       c.lnd_channel_point,
			       fn.public_key AS first_node_public_key,
			       sn.public_key AS second_node_public_key,
			       event_type,
			       coalesce((event->'capacity')::numeric, 0) as capacity
			FROM channel_event ce
			JOIN channel c ON c.channel_id=ce.channel_id
			JOIN node fn ON fn.node_id=c.first_node_id
			JOIN node sn ON sn.node_id=c.second_node_id
			WHERE c.lnd_channel_point = $1 AND ce.event_type = $2;`,
		expected.LNDChannelPoint, expected.EventType)

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

	if channelEvents[0].LNDShortChannelId != expected.LNDShortChannelId ||
		channelEvents[0].SecondNodePublicKey != expected.SecondNodePublicKey {
		t.Fatal("Channel event data not stored correctly")
	}

	// Check that channel open events stores the capacity correctly
	if channelEvents[0].EventType == 0 && channelEvents[0].Capacity != expected.Capacity {
		t.Fatalf("Channel capacity is not stored correctly. Expected: %d, got: %d", expected.Capacity,
			channelEvents[0].Capacity)
	}

	// Check that channel close events stores the capacity correctly
	if channelEvents[0].EventType == 1 && channelEvents[0].Capacity != expected.Capacity {
		t.Fatalf("Channel capacity is not stored correctly. Expected: %d, got: %d", expected.Capacity,
			channelEvents[0].Capacity)
	}
}
