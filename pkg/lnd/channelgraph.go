package lnd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/pkg/commons"
)

type subscribeChannelGrpahClient interface {
	SubscribeChannelGraph(ctx context.Context, in *lnrpc.GraphTopologySubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelGraphClient, error)
}

// SubscribeAndStoreChannelGraph Subscribes to channel updates
func SubscribeAndStoreChannelGraph(ctx context.Context, client subscribeChannelGrpahClient, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings) error {

	req := lnrpc.GraphTopologySubscription{}
	stream, err := client.SubscribeChannelGraph(ctx, &req)
	if err != nil {
		return errors.Wrap(err, "LND Subscribe Channel Graph")
	}

	rl := ratelimit.New(1) // 1 per second maximum rate limit

	for {

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		gpu, err := stream.Recv()

		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				break
			}
			log.Error().Err(err).Msgf("Subscribe channel graph stream receive")
			// rate limited resubscribe
			log.Info().Msg("Attempting reconnect to channel graph")
			for {
				rl.Take()
				stream, err = client.SubscribeChannelGraph(ctx, &req)
				if err == nil {
					log.Info().Msg("Reconnected to channel graph")
					break
				}
				log.Debug().Err(err).Msg("Reconnecting to channel graph")
			}
			continue
		}

		err = processNodeUpdates(gpu.NodeUpdates, db, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "Process node updates")
		}

		err = processChannelUpdates(gpu.ChannelUpdates, db, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "Process channel updates")
		}

	}

	return nil
}

func processNodeUpdates(nus []*lnrpc.NodeUpdate, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings) error {
	for _, nu := range nus {
		eventNodeId := commons.GetNodeIdFromPublicKey(nu.IdentityKey, nodeSettings.Chain, nodeSettings.Network)
		if eventNodeId != 0 {
			err := insertNodeEvent(db, time.Now().UTC(), eventNodeId, nu.Alias, nu.Color,
				nu.NodeAddresses, nu.Features, nodeSettings.NodeId)
			if err != nil {
				return errors.Wrapf(err, "Insert node event")
			}
		}
	}
	return nil
}

func processChannelUpdates(cus []*lnrpc.ChannelEdgeUpdate, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings) error {
	for _, cu := range cus {
		channelPoint, err := chanPointFromByte(cu.ChanPoint.GetFundingTxidBytes(), cu.ChanPoint.GetOutputIndex())
		if err != nil {
			return errors.Wrap(err, "Creating channel point from byte")
		}

		channelId := commons.GetChannelIdFromChannelPoint(channelPoint)
		if channelId != 0 {
			// If one of our nodes is advertising the channel update
			// (meaning we have changed our the channel policy so outbound).
			err := insertRoutingPolicy(db, time.Now().UTC(),
				commons.GetActiveTorqNodeIdFromPublicKey(cu.AdvertisingNode, nodeSettings.Chain, nodeSettings.Network) != 0,
				channelId, nodeSettings, cu)
			if err != nil {
				return errors.Wrap(err, "Insert routing policy")
			}
		}
	}
	return nil
}

const rpQuery = `
INSERT INTO routing_policy (ts,
	outbound,
	disabled,
	time_lock_delta,
	min_htlc,
	max_htlc_msat,
	fee_base_msat,
	fee_rate_mill_msat,
    channel_id,
    announcing_node_id,
    connecting_node_id,
    node_id)
select $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
WHERE NOT EXISTS (
	select true
	from (select
			last(disabled,ts) disabled,
			last(time_lock_delta,ts) time_lock_delta,
			last(min_htlc,ts) min_htlc,
			last(max_htlc_msat,ts) max_htlc_msat,
			last(fee_base_msat,ts) fee_base_msat,
			last(fee_rate_mill_msat, ts) fee_rate_mill_msat
		from routing_policy
		group by channel_id) as a
	where a.disabled = $3 and
		  a.time_lock_delta = $4 and
		  a.min_htlc = $5 and
		  a.max_htlc_msat = $6 and
		  a.fee_base_msat = $7 and
		  a.fee_rate_mill_msat = $8
);`

func insertRoutingPolicy(
	db *sqlx.DB,
	eventTime time.Time,
	outbound bool,
	channelId int,
	nodeSettings commons.ManagedNodeSettings,
	cu *lnrpc.ChannelEdgeUpdate) error {

	var err error
	if cu == nil || cu.RoutingPolicy == nil {
		log.Warn().Msg("Routing policy nil, skipping")
		return nil
	}

	announcingNodeId := commons.GetNodeIdFromPublicKey(cu.AdvertisingNode, nodeSettings.Chain, nodeSettings.Network)
	if announcingNodeId == 0 {
		newNode := nodes.Node{
			PublicKey: cu.AdvertisingNode,
			Chain:     nodeSettings.Chain,
			Network:   nodeSettings.Network,
		}
		announcingNodeId, err = nodes.AddNodeWhenNew(db, newNode)
		if err != nil {
			return errors.Wrapf(err, "Adding node (publicKey: %v)", cu.AdvertisingNode)
		}
	}
	connectingNodeId := commons.GetNodeIdFromPublicKey(cu.ConnectingNode, nodeSettings.Chain, nodeSettings.Network)
	if connectingNodeId == 0 {
		newNode := nodes.Node{
			PublicKey: cu.ConnectingNode,
			Chain:     nodeSettings.Chain,
			Network:   nodeSettings.Network,
		}
		connectingNodeId, err = nodes.AddNodeWhenNew(db, newNode)
		if err != nil {
			return errors.Wrapf(err, "Adding node (publicKey: %v shortChannelId: %v)", cu.ConnectingNode)
		}
	}

	_, err = db.Exec(rpQuery, eventTime, outbound,
		cu.RoutingPolicy.Disabled, cu.RoutingPolicy.TimeLockDelta, cu.RoutingPolicy.MinHtlc,
		cu.RoutingPolicy.MaxHtlcMsat, cu.RoutingPolicy.FeeBaseMsat, cu.RoutingPolicy.FeeRateMilliMsat,
		channelId, announcingNodeId, connectingNodeId, nodeSettings.NodeId)

	if err != nil {
		return errors.Wrapf(err, "DB Exec")
	}

	return nil
}

const neQuery = `
INSERT INTO node_event (timestamp, event_node_id, alias, color, node_addresses, features, node_id)
SELECT $1,$2,$3,$4,$5,$6,$7
WHERE NOT EXISTS (
select true
from (select last(alias, timestamp) as  alias,
        last(color,timestamp) as color,
        last(node_addresses,timestamp) as node_addresses,
        last(features,timestamp) as features
    from node_event
    group by pub_key) as a
where a.alias = $4 and
      a.color = $5 and
      a.node_addresses = $6 and
      a.features = $7
);`

func insertNodeEvent(db *sqlx.DB, ts time.Time, eventNodeId int, alias string, color string,
	na []*lnrpc.NodeAddress, f map[uint32]*lnrpc.Feature, nodeId int) error {

	// Create json byte object from node address map
	najb, err := json.Marshal(na)
	if err != nil {
		return errors.Wrap(err, "JSON Marshall node address map")
	}

	// Create json byte object from features list
	fjb, err := json.Marshal(f)
	if err != nil {
		return errors.Wrap(err, "JSON Marshal feature list")
	}

	if _, err = db.Exec(neQuery, ts, eventNodeId, alias, color, najb, fjb, nodeId); err != nil {
		return errors.Wrap(err, "Executing SQL")
	}

	return nil
}
