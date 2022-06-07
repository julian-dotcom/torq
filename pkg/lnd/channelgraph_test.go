package lnd

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/testutil"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"io"
	"testing"
	"time"
)

type stubLNDSubscribeChannelGraphRPC struct {
	grpc.ClientStream
	GraphTopologyUpdate []*lnrpc.GraphTopologyUpdate
}

func (s *stubLNDSubscribeChannelGraphRPC) Recv() (*lnrpc.GraphTopologyUpdate, error) {
	if len(s.GraphTopologyUpdate) == 0 {
		return nil, io.EOF
	}
	var gtu interface{}
	gtu, s.GraphTopologyUpdate = s.GraphTopologyUpdate[0], s.GraphTopologyUpdate[1:]
	if u, ok := gtu.(*lnrpc.GraphTopologyUpdate); ok {
		return u, nil
	}
	return nil, io.EOF
}

func (s *stubLNDSubscribeChannelGraphRPC) SubscribeChannelGraph(ctx context.Context, in *lnrpc.GraphTopologySubscription,
	opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelGraphClient, error) {

	return &stubLNDSubscribeChannelGraphRPC{GraphTopologyUpdate: s.GraphTopologyUpdate}, nil
}

func TestSubscribeChannelGraphUpdates(t *testing.T) {
	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, err := srv.NewTestDatabase(true)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Routing Policy", func(t *testing.T) {
		//expected := channelEventData{Chan_id: 1337, Chan_point: "point break", Pub_key: "remote pub key",
		//	Event_type: int(lnrpc.ChannelEventUpdate_OPEN_CHANNEL), Capacity: 100000000}

		chanPoint := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.
			ChannelPoint_FundingTxidBytes{
			FundingTxidBytes: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
			OutputIndex: 3}
		chanPointStr, err := getChanPoint([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, chanPoint.OutputIndex)
		if err != nil {
			t.Fatal(err)
		}

		ourNodePubKeys = append(ourNodePubKeys, "ourNodePubkey")
		pubKeyList = append(pubKeyList, "theirNodePubkey")
		// Add the channel to the list of relevant channels
		chanPointList = append(chanPointList, chanPointStr)

		updateEvent := lnrpc.GraphTopologyUpdate{
			NodeUpdates: nil,
			ChannelUpdates: []*lnrpc.ChannelEdgeUpdate{{
				ChanId:    1234567,
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
				AdvertisingNode: "theirNodePubkey",
				ConnectingNode:  "ourNodePubkey",
			}},
			ClosedChans: nil,
		}

		expected := routingPolicyData{
			Ts:               time.Time{},
			ChanPoint:        chanPointStr,
			ChanId:           fmt.Sprint(updateEvent.ChannelUpdates[0].ChanId),
			Outbound:         updateEvent.ChannelUpdates[0].AdvertisingNode == "ourNodePubkey",
			AnnouncingPubKey: updateEvent.ChannelUpdates[0].AdvertisingNode,
			FeeRateMillMsat:  updateEvent.ChannelUpdates[0].RoutingPolicy.FeeRateMilliMsat,
			FeeBaseMsat:      updateEvent.ChannelUpdates[0].RoutingPolicy.FeeBaseMsat,
			MaxHtlcMsat:      updateEvent.ChannelUpdates[0].RoutingPolicy.MaxHtlcMsat,
			MinHtlc:          updateEvent.ChannelUpdates[0].RoutingPolicy.MinHtlc,
			TimeLockDelta:    updateEvent.ChannelUpdates[0].RoutingPolicy.TimeLockDelta,
			Disabled:         updateEvent.ChannelUpdates[0].RoutingPolicy.Disabled,
		}

		runChannelGraphTest(t, db, updateEvent, expected)
	})
}

type routingPolicyData struct {
	Ts               time.Time
	ChanPoint        string `db:"chan_point"`
	ChanId           string `db:"chan_id"`
	Outbound         bool   `db:"outbound"`
	AnnouncingPubKey string `db:"announcing_pub_key"`
	FeeRateMillMsat  int64  `db:"fee_rate_mill_msat"`
	FeeBaseMsat      int64  `db:"fee_base_msat"`
	MaxHtlcMsat      uint64 `db:"max_htlc_msat"`
	MinHtlc          int64  `db:"min_htlc"`
	TimeLockDelta    uint32 `db:"time_lock_delta"`
	Disabled         bool   `db:"disabled"`
}

func runChannelGraphTest(t *testing.T, db *sqlx.DB, graphUpdate lnrpc.GraphTopologyUpdate, expected routingPolicyData) { //,
	// expected channelEventData
	ctx := context.Background()
	errs, ctx := errgroup.WithContext(ctx)

	client := &stubLNDSubscribeChannelGraphRPC{GraphTopologyUpdate: []*lnrpc.GraphTopologyUpdate{&graphUpdate}}

	errs.Go(func() error {
		err := SubscribeAndStoreChannelGraph(ctx, client, db)
		if err != nil {
			t.Fatalf("Problem subscribing to channel graph: %v", err)
		}
		return err
	})

	// Wait for subscriptions to complete
	err := errs.Wait()
	if err != nil {
		t.Fatal(err)
	}

	var result []routingPolicyData
	err = db.Select(&result, `
			select ts,
				   chan_point,
				   chan_id,
				   outbound,
				   announcing_pub_key,
				   fee_rate_mill_msat,
				   fee_base_msat,
				   max_htlc_msat,
				   min_htlc,
				   time_lock_delta,
				   disabled
			from routing_policy;`)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			t.Fatal("There were no routing policies but I did expect there to be some")
		}
		t.Fatalf("Problem executing sql: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected to find a single routing policy record stored in the database. Found %d",
			len(result))
	}

	if result[0].AnnouncingPubKey != expected.AnnouncingPubKey {
		t.Fatalf("Incorrect announcing pub key. Expected: %v, got %v", expected.AnnouncingPubKey, result[0].AnnouncingPubKey)
	}

	if result[0].ChanPoint != expected.ChanPoint {
		t.Fatalf("Incorrect channel point. Expected: %v, got %v", expected.ChanPoint, result[0].ChanPoint)
	}

	if result[0].ChanId != expected.ChanId {
		t.Fatalf("Incorrect channel id. Expected: %v, got %v", expected.ChanId, result[0].ChanId)
	}

	if result[0].Disabled != expected.Disabled {
		t.Fatalf("Incorrect channel disabled state. Expected: %v, got %v", expected.Disabled, result[0].Disabled)
	}

	if result[0].FeeRateMillMsat != expected.FeeRateMillMsat {
		t.Fatalf("Incorrect fee rate. Expected: %v, got %v", expected.FeeRateMillMsat, result[0].FeeRateMillMsat)
	}

	if result[0].FeeBaseMsat != expected.FeeBaseMsat {
		t.Fatalf("Incorrect base fee state. Expected: %v, got %v", expected.FeeBaseMsat, result[0].FeeBaseMsat)
	}

	if result[0].MinHtlc != expected.MinHtlc {
		t.Fatalf("Incorrect min htlc. Expected: %v, got %v", expected.MinHtlc, result[0].MinHtlc)
	}

	if result[0].MaxHtlcMsat != expected.MaxHtlcMsat {
		t.Fatalf("Incorrect max htlc. Expected: %v, got %v", expected.MaxHtlcMsat, result[0].MaxHtlcMsat)
	}

	if result[0].Outbound != expected.Outbound {
		t.Fatalf("Incorrect outbound state. Expected: %v, got %v", expected.Outbound, result[0].Outbound)
	}

	if result[0].TimeLockDelta != expected.TimeLockDelta {
		t.Fatalf("Incorrect timelock delta. Expected: %v, got %v", expected.TimeLockDelta, result[0].TimeLockDelta)
	}

}
