package lnd

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

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

type stubLNDSubscribeChannelGraphRPC struct {
	grpc.ClientStream
	GraphTopologyUpdate []*lnrpc.GraphTopologyUpdate
	CancelFunc          func()
}

func (s *stubLNDSubscribeChannelGraphRPC) ListChannels(
	ctx context.Context, in *lnrpc.ListChannelsRequest, opts ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error) {
	var channels []*lnrpc.Channel
	return &lnrpc.ListChannelsResponse{
		Channels: channels,
	}, nil
}

func (s *stubLNDSubscribeChannelGraphRPC) GetChanInfo(
	ctx context.Context, in *lnrpc.ChanInfoRequest, opts ...grpc.CallOption) (*lnrpc.ChannelEdge, error) {
	return nil, nil
}

func (s *stubLNDSubscribeChannelGraphRPC) GetNodeInfo(
	ctx context.Context, in *lnrpc.NodeInfoRequest, opts ...grpc.CallOption) (*lnrpc.NodeInfo, error) {
	return nil, nil
}

func (s *stubLNDSubscribeChannelGraphRPC) Recv() (*lnrpc.GraphTopologyUpdate, error) {
	if len(s.GraphTopologyUpdate) == 0 {
		s.CancelFunc()
		return nil, context.Canceled
	}
	var gtu interface{}
	gtu, s.GraphTopologyUpdate = s.GraphTopologyUpdate[0], s.GraphTopologyUpdate[1:]
	if u, ok := gtu.(*lnrpc.GraphTopologyUpdate); ok {
		return u, nil
	}
	s.CancelFunc()
	return nil, context.Canceled
}

func (s *stubLNDSubscribeChannelGraphRPC) SubscribeChannelGraph(ctx context.Context, in *lnrpc.GraphTopologySubscription,
	opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelGraphClient, error) {

	return s, nil
}

func TestSubscribeChannelGraphUpdates(t *testing.T) {
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

	chanPoint := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.
	ChannelPoint_FundingTxidBytes{
		FundingTxidBytes: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
		OutputIndex: 3}

	chanPointStr, err := chanPointFromByte([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, chanPoint.OutputIndex)
	if err != nil {
		t.Fatal(err)
	}
	fundingTransactionHash, fundingOutputIndex := commons.ParseChannelPoint(chanPointStr)

	t.Run("Irrelevant routing policy updates are ignored", func(t *testing.T) {

		irrelevatChannelPoint := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.
		ChannelPoint_FundingTxidBytes{
			FundingTxidBytes: []byte{2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1}},
			OutputIndex: 2}

		irrelecantUpdateEvent := lnrpc.GraphTopologyUpdate{
			NodeUpdates: nil,
			ChannelUpdates: []*lnrpc.ChannelEdgeUpdate{{
				ChanId:    12345678,
				ChanPoint: irrelevatChannelPoint,
				Capacity:  2000000,
				RoutingPolicy: &lnrpc.RoutingPolicy{
					TimeLockDelta:    0,
					MinHtlc:          0,
					FeeBaseMsat:      2,
					FeeRateMilliMsat: 200,
					Disabled:         true,
					MaxHtlcMsat:      1000,
					LastUpdate:       0,
				},
				AdvertisingNode: testutil.TestPublicKey2,
				ConnectingNode:  testutil.TestPublicKey1,
			}},
			ClosedChans: nil,
		}

		result := simulateChannelGraphUpdate(t, db, &stubLNDSubscribeChannelGraphRPC{GraphTopologyUpdate: []*lnrpc.
		GraphTopologyUpdate{&irrelecantUpdateEvent}}, fundingTransactionHash, fundingOutputIndex)

		if len(result) != 0 {
			testutil.Fatalf(t, "Expected to find no routing policy record stored in the database. Found %d",
				len(result))
		}
	})

	t.Run("Relevant routing policies are correctly stored", func(t *testing.T) {

		updateEvent := lnrpc.GraphTopologyUpdate{
			NodeUpdates: nil,
			ChannelUpdates: []*lnrpc.ChannelEdgeUpdate{{
				ChanId:    1111,
				ChanPoint: chanPoint,
				Capacity:  1000000,
				RoutingPolicy: &lnrpc.RoutingPolicy{
					TimeLockDelta:    0,
					MinHtlc:          0,
					FeeBaseMsat:      2,
					FeeRateMilliMsat: 100,
					Disabled:         true,
					MaxHtlcMsat:      1000,
					LastUpdate:       0,
				},
				AdvertisingNode: testutil.TestPublicKey2,
				ConnectingNode:  testutil.TestPublicKey1,
			}},
			ClosedChans: nil,
		}

		expected := routingPolicyData{
			Ts:                     time.Now(),
			FundingTransactionHash: fundingTransactionHash,
			FundingOutputIndex:     fundingOutputIndex,
			LNDShortChannelId:      &updateEvent.ChannelUpdates[0].ChanId,
			AnnouncingPubKey:       updateEvent.ChannelUpdates[0].AdvertisingNode,
			FeeRateMillMsat:        updateEvent.ChannelUpdates[0].RoutingPolicy.FeeRateMilliMsat,
			FeeBaseMsat:            updateEvent.ChannelUpdates[0].RoutingPolicy.FeeBaseMsat,
			MaxHtlcMsat:            updateEvent.ChannelUpdates[0].RoutingPolicy.MaxHtlcMsat,
			MinHtlcMsat:            uint64(updateEvent.ChannelUpdates[0].RoutingPolicy.MinHtlc),
			TimeLockDelta:          updateEvent.ChannelUpdates[0].RoutingPolicy.TimeLockDelta,
			Disabled:               updateEvent.ChannelUpdates[0].RoutingPolicy.Disabled,
		}

		result := simulateChannelGraphUpdate(t, db,
			&stubLNDSubscribeChannelGraphRPC{GraphTopologyUpdate: []*lnrpc.GraphTopologyUpdate{&updateEvent}},
			fundingTransactionHash, fundingOutputIndex)

		if len(result) != 1 {
			testutil.Fatalf(t, "Expected to find a single routing policy record stored in the database. Found %d",
				len(result))
		}

		if result[0].AnnouncingPubKey != expected.AnnouncingPubKey {
			testutil.Errorf(t, "Incorrect announcing pub key. Expected: %v, got %v", expected.AnnouncingPubKey,
				result[0].AnnouncingPubKey)
		}

		if result[0].FundingTransactionHash != expected.FundingTransactionHash {
			testutil.Errorf(t, "Incorrect Funding Transaction. Expected: %v, got %v", expected.FundingTransactionHash, result[0].FundingTransactionHash)
		}

		if result[0].FundingOutputIndex != expected.FundingOutputIndex {
			testutil.Errorf(t, "Incorrect Funding Output Index. Expected: %v, got %v", expected.FundingOutputIndex, result[0].FundingOutputIndex)
		}

		if *result[0].LNDShortChannelId != *expected.LNDShortChannelId {
			testutil.Errorf(t, "Incorrect channel id. Expected: %v, got %v", *expected.LNDShortChannelId, *result[0].LNDShortChannelId)
		}

		if result[0].Disabled != expected.Disabled {
			testutil.Errorf(t, "Incorrect channel disabled state. Expected: %v, got %v", expected.Disabled,
				result[0].Disabled)
		}

		if result[0].FeeRateMillMsat != expected.FeeRateMillMsat {
			testutil.Errorf(t, "Incorrect fee rate. Expected: %v, got %v", expected.FeeRateMillMsat,
				result[0].FeeRateMillMsat)
		}

		if result[0].FeeBaseMsat != expected.FeeBaseMsat {
			testutil.Errorf(t, "Incorrect base fee state. Expected: %v, got %v", expected.FeeBaseMsat,
				result[0].FeeBaseMsat)
		}

		if result[0].MinHtlcMsat != expected.MinHtlcMsat {
			testutil.Errorf(t, "Incorrect min htlc. Expected: %v, got %v", expected.MinHtlcMsat, result[0].MinHtlcMsat)
		}

		if result[0].MaxHtlcMsat != expected.MaxHtlcMsat {
			testutil.Errorf(t, "Incorrect max htlc. Expected: %v, got %v", expected.MaxHtlcMsat, result[0].MaxHtlcMsat)
		}

		if result[0].TimeLockDelta != expected.TimeLockDelta {
			testutil.Errorf(t, "Incorrect timelock delta. Expected: %v, got %v", expected.TimeLockDelta,
				result[0].TimeLockDelta)
		}

		r2 := simulateChannelGraphUpdate(t, db, &stubLNDSubscribeChannelGraphRPC{GraphTopologyUpdate: []*lnrpc.
		GraphTopologyUpdate{&updateEvent}}, fundingTransactionHash, fundingOutputIndex)

		if len(r2) != 1 {
			testutil.Fatalf(t, "Expected to find a single routing policy record stored in the database. Found %d",
				len(r2))
		}

		if t.Failed() {
			t.FailNow()
		}

		secondUpdateEvent := lnrpc.GraphTopologyUpdate{
			NodeUpdates: nil,
			ChannelUpdates: []*lnrpc.ChannelEdgeUpdate{{
				ChanId:    1111,
				ChanPoint: chanPoint,
				Capacity:  2000000,
				RoutingPolicy: &lnrpc.RoutingPolicy{
					TimeLockDelta:    0,
					MinHtlc:          0,
					FeeBaseMsat:      2,
					FeeRateMilliMsat: 200,
					Disabled:         true,
					MaxHtlcMsat:      1000,
					LastUpdate:       0,
				},
				AdvertisingNode: testutil.TestPublicKey1,
				ConnectingNode:  testutil.TestPublicKey2,
			}},
			ClosedChans: nil,
		}

		e3 := routingPolicyData{
			Ts:                     time.Now(),
			FundingTransactionHash: fundingTransactionHash,
			FundingOutputIndex:     fundingOutputIndex,
			LNDShortChannelId:      &secondUpdateEvent.ChannelUpdates[0].ChanId,
			AnnouncingPubKey:       testutil.TestPublicKey1,
			FeeRateMillMsat:        secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.FeeRateMilliMsat,
			FeeBaseMsat:            secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.FeeBaseMsat,
			MaxHtlcMsat:            secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.MaxHtlcMsat,
			MinHtlcMsat:            uint64(secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.MinHtlc),
			TimeLockDelta:          secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.TimeLockDelta,
			Disabled:               secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.Disabled,
		}

		r3 := simulateChannelGraphUpdate(t, db, &stubLNDSubscribeChannelGraphRPC{GraphTopologyUpdate: []*lnrpc.
		GraphTopologyUpdate{&secondUpdateEvent}}, fundingTransactionHash, fundingOutputIndex)

		if r3[1].AnnouncingPubKey != e3.AnnouncingPubKey {
			testutil.Errorf(t, "Incorrect announcing pub key. Expected: %v, got %v", e3.AnnouncingPubKey,
				r3[1].AnnouncingPubKey)
		}
	})

}

type routingPolicyData struct {
	Ts                     time.Time
	FeeRateMillMsat        int64   `db:"fee_rate_mill_msat"`
	FeeBaseMsat            int64   `db:"fee_base_msat"`
	MaxHtlcMsat            uint64  `db:"max_htlc_msat"`
	MinHtlcMsat            uint64  `db:"min_htlc"`
	TimeLockDelta          uint32  `db:"time_lock_delta"`
	Disabled               bool    `db:"disabled"`
	ChannelId              int     `db:"channel_id"`
	AnnouncingNodeId       int     `db:"announcing_node_id"`
	AnnouncingPubKey       string  `db:"announcing_public_key"`
	ConnectingNodeId       int     `db:"connecting_node_id"`
	ConnectingPubKey       string  `db:"connecting_public_key"`
	NodeId                 int     `db:"node_id"`
	FundingTransactionHash string  `db:"funding_transaction_hash"`
	FundingOutputIndex     int     `db:"funding_output_index"`
	LNDShortChannelId      *uint64 `db:"lnd_short_channel_id"`
}

func simulateChannelGraphUpdate(t *testing.T, db *sqlx.DB, client *stubLNDSubscribeChannelGraphRPC,
	fundingTransactionHash string, fundingOutputIndex int) (r []routingPolicyData) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	client.CancelFunc = cancel

	lndShortChannelId := uint64(1111)
	shortChannelId := commons.ConvertLNDShortChannelID(lndShortChannelId)
	channel := channels.Channel{
		ShortChannelID:         &shortChannelId,
		FirstNodeId:            commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet),
		SecondNodeId:           commons.GetNodeIdByPublicKey(testutil.TestPublicKey2, commons.Bitcoin, commons.SigNet),
		Capacity:               1_000_000,
		LNDShortChannelID:      &lndShortChannelId,
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		Status:                 commons.Open,
	}
	channelId, err := channels.AddChannelOrUpdateChannelStatus(db, channel)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create channel %v", lndShortChannelId)
		t.Fatalf("Problem adding channel %v", channel)
	}
	t.Logf("channel added with channelId: %v", channelId)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		SubscribeAndStoreChannelGraph(ctx, client, db,
			commons.GetNodeSettingsByNodeId(
				commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet)), nil, nil, nil, nil)
	}()
	wg.Wait()

	var result []routingPolicyData
	err = db.Select(&result, `
			select rp.ts,
				   rp.fee_rate_mill_msat,
				   rp.fee_base_msat,
				   rp.max_htlc_msat,
				   rp.min_htlc,
				   rp.time_lock_delta,
				   rp.disabled,
				   c.channel_id,
				   rp.announcing_node_id,
				   rp.connecting_node_id,
				   rp.node_id,
				   an.public_key AS announcing_public_key,
				   cn.public_key AS connecting_public_key,
				   c.lnd_short_channel_id,
				   c.funding_transaction_hash,
				   c.funding_output_index
			from routing_policy rp
			JOIN channel c ON c.channel_id=rp.channel_id
			JOIN node an ON rp.announcing_node_id=an.node_id
			JOIN node cn ON rp.connecting_node_id=cn.node_id;`)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			t.Fatal("There were no routing policies but I did expect there to be some")
		}
		t.Fatalf("Problem executing sql: %v", err)
	}

	return result
}
