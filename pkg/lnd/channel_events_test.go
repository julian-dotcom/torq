package lnd

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/settings"
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

func (c *stubLNDSubscribeChannelEvent) ListChannels(
	ctx context.Context, in *lnrpc.ListChannelsRequest, opts ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error) {
	var channels []*lnrpc.Channel
	return &lnrpc.ListChannelsResponse{
		Channels: channels,
	}, nil
}

func (c *stubLNDSubscribeChannelEvent) GetChanInfo(
	ctx context.Context, in *lnrpc.ChanInfoRequest, opts ...grpc.CallOption) (*lnrpc.ChannelEdge, error) {
	return nil, nil
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

	db, cancel, _, err := srv.NewTestDatabase(true)
	defer cancel()
	if err != nil {
		t.Fatal(err)
	}

	err = settings.InitializeManagedSettingsCache(db, commons.VectorUrl)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing ManagedSettings cache: %v", err)
	}

	err = settings.InitializeManagedNodeCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing ManagedNode cache: %v", err)
	}

	err = channels.InitializeManagedChannelCache(db)
	if err != nil {
		cancel()
		log.Fatal().Err(err).Msgf("Problem initializing ManagedChannel cache: %v", err)
	}

	t.Run("Open Channel Event", func(t *testing.T) {
		lndShortChannelId := uint64(1337)
		expected := channelEventData{LNDShortChannelId: &lndShortChannelId, FundingTransactionHash: "point_break", FundingOutputIndex: 3,
			FirstNodePublicKey: testutil.TestPublicKey1, SecondNodePublicKey: "remote_pub_key",
			EventType: int(lnrpc.ChannelEventUpdate_OPEN_CHANNEL), Capacity: 100000000}
		channel := &lnrpc.Channel{ChanId: *expected.LNDShortChannelId, ChannelPoint: "point_break:3",
			RemotePubkey: expected.SecondNodePublicKey, Capacity: expected.Capacity}
		channelEvent := lnrpc.ChannelEventUpdate_OpenChannel{OpenChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_OPEN_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

	t.Run("Fully Resolved Channel Event", func(t *testing.T) {
		expected := channelEventData{FundingTransactionHash: testutil.TestFundingTransactionHash5_NOTINDB, FundingOutputIndex: 3,
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
		lndShortChannelId := uint64(2222)
		expected := channelEventData{LNDShortChannelId: &lndShortChannelId, FundingTransactionHash: testutil.TestFundingTransactionHash2, FundingOutputIndex: 3,
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
		lndShortChannelId := uint64(3333)
		expected := channelEventData{LNDShortChannelId: &lndShortChannelId, FundingTransactionHash: testutil.TestFundingTransactionHash3, FundingOutputIndex: 3,
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
		lndShortChannelId := uint64(4444)
		expected := channelEventData{LNDShortChannelId: &lndShortChannelId, FundingTransactionHash: testutil.TestFundingTransactionHash4, FundingOutputIndex: 3,
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
			Short_channel_id         *string
			Funding_transaction_hash string
		}
		var channels []channel
		err = db.Select(&channels, `
			SELECT short_channel_id, funding_transaction_hash FROM channel WHERE short_channel_id=$1;`, "0x0x1337")
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

		if (channels[0].Short_channel_id != nil &&
			*channels[0].Short_channel_id != "0x0x1337") ||
			channels[0].Funding_transaction_hash != "point_break" {

			t.Fatal("Channel data not stored correctly")
		}

	})

	t.Run("Closed Channel Event", func(t *testing.T) {
		closingTransactionHash := "closing_point_break"
		lndShortChannelId := uint64(4444)
		expected := channelEventData{LNDShortChannelId: &lndShortChannelId, FundingTransactionHash: testutil.TestFundingTransactionHash4, FundingOutputIndex: 3,
			FirstNodePublicKey: testutil.TestPublicKey1, SecondNodePublicKey: testutil.TestPublicKey2, EventType: int(lnrpc.ChannelEventUpdate_CLOSED_CHANNEL),
			Capacity: 100000000, ClosingTransactionHash: &closingTransactionHash}
		channel := &lnrpc.ChannelCloseSummary{ChanId: *expected.LNDShortChannelId, ChannelPoint: testutil.TestChannelPoint4,
			RemotePubkey: expected.SecondNodePublicKey, Capacity: expected.Capacity, ClosingTxHash: closingTransactionHash, CloseInitiator: lnrpc.Initiator_INITIATOR_LOCAL}
		channelEvent := lnrpc.ChannelEventUpdate_ClosedChannel{ClosedChannel: channel}
		channelEventUpdate := &lnrpc.ChannelEventUpdate{
			Type:    lnrpc.ChannelEventUpdate_CLOSED_CHANNEL,
			Channel: &channelEvent}
		runChannelEventTest(t, db, channelEventUpdate, expected)
	})

}

type channelEventData struct {
	LNDShortChannelId      *uint64 `db:"lnd_short_channel_id"`
	FundingTransactionHash string  `db:"funding_transaction_hash"`
	FundingOutputIndex     int     `db:"funding_output_index"`
	ClosingTransactionHash *string `db:"closing_transaction_hash"`
	FirstNodePublicKey     string  `db:"first_node_public_key"`
	SecondNodePublicKey    string  `db:"second_node_public_key"`
	EventType              int     `db:"event_type"`
	Capacity               int64   `db:"capacity"`
}

func runChannelEventTest(t *testing.T, db *sqlx.DB, channelEvent interface{}, expected channelEventData) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	client := &stubLNDSubscribeChannelEvent{ChannelEvents: []interface{}{channelEvent}, CancelFunc: cancel}
	go func() {
		defer wg.Done()
		SubscribeAndStoreChannelEvents(ctx, client, db,
			commons.GetNodeSettingsByNodeId(
				commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet)), nil, nil, nil)
	}()
	wg.Wait()

	var channelEvents []channelEventData
	err := db.Select(&channelEvents, `
			SELECT c.lnd_short_channel_id,
			       c.funding_transaction_hash,
			       c.funding_output_index,
			       c.closing_transaction_hash,
			       fn.public_key AS first_node_public_key,
			       sn.public_key AS second_node_public_key,
			       event_type,
			       coalesce((event->'capacity')::numeric, 0) as capacity
			FROM channel_event ce
			JOIN channel c ON c.channel_id=ce.channel_id
			JOIN node fn ON fn.node_id=c.first_node_id
			JOIN node sn ON sn.node_id=c.second_node_id
			WHERE c.funding_transaction_hash = $1 AND c.funding_output_index = $2 AND ce.event_type = $3;`,
		expected.FundingTransactionHash, expected.FundingOutputIndex, expected.EventType)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// t.Fatalf("Error was %v", err)
			t.Fatal("There were no channel events but I did expect there to be some")
		}
		t.Fatalf("Problem executing sql: %v", err)
	}

	if len(channelEvents) != 1 {
		t.Fatalf("Expected to find a single channel event record stored in the database for this channel event update but found %d", len(channelEvents))
	}

	if channelEvents[0].LNDShortChannelId != nil && *channelEvents[0].LNDShortChannelId != *expected.LNDShortChannelId ||
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
	if channelEvents[0].EventType == 1 && expected.ClosingTransactionHash != nil && *expected.ClosingTransactionHash != "" && channelEvents[0].ClosingTransactionHash == nil {
		t.Fatalf("Channel ClosingTransactionHash is not stored correctly. Expected: %v, got: %v", *expected.ClosingTransactionHash,
			channelEvents[0].ClosingTransactionHash)
	}
	if channelEvents[0].EventType == 1 && expected.ClosingTransactionHash != nil && *expected.ClosingTransactionHash != "" && *channelEvents[0].ClosingTransactionHash != *expected.ClosingTransactionHash {
		t.Fatalf("Channel ClosingTransactionHash is not stored correctly. Expected: %v, got: %v", *expected.ClosingTransactionHash,
			channelEvents[0].ClosingTransactionHash)
	}
}
