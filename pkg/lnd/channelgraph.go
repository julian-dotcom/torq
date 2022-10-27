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

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/pkg/commons"
)

type subscribeChannelGrpahClient interface {
	SubscribeChannelGraph(ctx context.Context, in *lnrpc.GraphTopologySubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelGraphClient, error)
}

// SubscribeAndStoreChannelGraph Subscribes to channel updates
func SubscribeAndStoreChannelGraph(ctx context.Context, client subscribeChannelGrpahClient, db *sqlx.DB) error {

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

		err = processNodeUpdates(gpu.NodeUpdates, db)
		if err != nil {
			return errors.Wrap(err, "Process node updates")
		}

		err = processChannelUpdates(gpu.ChannelUpdates, db)
		if err != nil {
			return errors.Wrap(err, "Process channel updates")
		}

	}

	return nil
}

func processNodeUpdates(nus []*lnrpc.NodeUpdate, db *sqlx.DB) error {
	for _, nu := range nus {
		nodeId := commons.GetNodeIdFromPublicKey(nu.IdentityKey)
		if nodeId != 0 {
			err := insertNodeEvent(db, time.Now().UTC(), nodeId, nu.IdentityKey, nu.Alias, nu.Color, nu.NodeAddresses,
				nu.Features)
			if err != nil {
				return errors.Wrapf(err, "Insert node event")
			}
		}
	}
	return nil
}

func processChannelUpdates(cus []*lnrpc.ChannelEdgeUpdate, db *sqlx.DB) error {
	for _, cu := range cus {
		channelPoint, err := chanPointFromByte(cu.ChanPoint.GetFundingTxidBytes(), cu.ChanPoint.GetOutputIndex())
		if err != nil {
			return errors.Wrap(err, "Creating channel point from byte")
		}

		channelId := commons.GetChannelIdFromChannelPoint(channelPoint)
		if channelId != 0 {
			// If one of our nodes is advertising the channel update
			// (meaning we have changed our the channel policy so outbound).
			err := insertRoutingPolicy(db, time.Now().UTC(), commons.GetActiveTorqNodeIdFromPublicKey(cu.AdvertisingNode) != 0, channelId, cu)
			if err != nil {
				return errors.Wrap(err, "Insert routing policy")
			}
		}
	}
	return nil
}

const rpQuery = `
INSERT INTO routing_policy (ts,
	lnd_short_channel_id,
	short_channel_id,
	announcing_pub_key,
	lnd_channel_point,
	outbound,
	disabled,
	time_lock_delta,
	min_htlc,
	max_htlc_msat,
	fee_base_msat,
	fee_rate_mill_msat,
    channel_id,
    announcing_node_id,
    connecting_node_id)
select $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
WHERE NOT EXISTS (
	select true
	from (select
            last(lnd_short_channel_id,ts) lnd_short_channel_id,
            last(short_channel_id,ts) short_channel_id,
			last(announcing_pub_key, ts) as announcing_pub_key,
			last(disabled,ts) disabled,
			last(time_lock_delta,ts) time_lock_delta,
			last(min_htlc,ts) min_htlc,
			last(max_htlc_msat,ts) max_htlc_msat,
			last(fee_base_msat,ts) fee_base_msat,
			last(fee_rate_mill_msat, ts) fee_rate_mill_msat
		from routing_policy
		group by lnd_short_channel_id, announcing_pub_key) as a
	where a.lnd_short_channel_id = $2 and
		  a.short_channel_id = $3 and
		  a.announcing_pub_key = $4 and
		  a.disabled = $7 and
		  a.time_lock_delta = $8 and
		  a.min_htlc = $9 and
		  a.max_htlc_msat = $10 and
		  a.fee_base_msat = $11 and
		  a.fee_rate_mill_msat = $12
);`

func insertRoutingPolicy(
	db *sqlx.DB,
	eventTime time.Time,
	outbound bool,
	channelId int,
	cu *lnrpc.ChannelEdgeUpdate) error {

	if cu == nil || cu.RoutingPolicy == nil {
		log.Warn().Msg("Routing policy nil, skipping")
		return nil
	}

	cp, err := chanPointFromByte(cu.ChanPoint.GetFundingTxidBytes(), cu.ChanPoint.GetOutputIndex())
	if err != nil {
		return errors.Wrap(err, "Creating channel point from byte")
	}

	shortChannelId := channels.ConvertLNDShortChannelID(cu.ChanId)
	announcingNodeId := commons.GetNodeIdFromPublicKey(cu.AdvertisingNode)
	if announcingNodeId == 0 {
		newNode := nodes.Node{
			PublicKey: cu.AdvertisingNode,
		}
		announcingNodeId, err = nodes.AddNodeWhenNew(db, newNode)
		if err != nil {
			return errors.Wrapf(err, "Adding node (publicKey: %v shortChannelId: %v)", cu.AdvertisingNode, shortChannelId)
		}
	}
	connectingNodeId := commons.GetNodeIdFromPublicKey(cu.ConnectingNode)
	if connectingNodeId == 0 {
		newNode := nodes.Node{
			PublicKey: cu.ConnectingNode,
		}
		connectingNodeId, err = nodes.AddNodeWhenNew(db, newNode)
		if err != nil {
			return errors.Wrapf(err, "Adding node (publicKey: %v shortChannelId: %v)", cu.ConnectingNode, shortChannelId)
		}
	}

	_, err = db.Exec(rpQuery, eventTime, cu.ChanId, shortChannelId, cu.AdvertisingNode, cp, outbound,
		cu.RoutingPolicy.Disabled, cu.RoutingPolicy.TimeLockDelta, cu.RoutingPolicy.MinHtlc,
		cu.RoutingPolicy.MaxHtlcMsat, cu.RoutingPolicy.FeeBaseMsat, cu.RoutingPolicy.FeeRateMilliMsat,
		channelId, announcingNodeId, connectingNodeId)

	if err != nil {
		return errors.Wrapf(err, "DB Exec")
	}

	return nil
}

const neQuery = `INSERT INTO node_event (timestamp, node_id, pub_key, alias, color, node_addresses, features)
SELECT $1,$2,$3,$4,$5,$6,$7
WHERE NOT EXISTS (
select true
from (select pub_key,
        last(alias, timestamp) as  alias,
        last(color,timestamp) as color,
        last(node_addresses,timestamp) as node_addresses,
        last(features,timestamp) as features
    from node_event
    group by pub_key) as a
where a.pub_key = $3 and
      a.alias = $4 and
      a.color = $5 and
      a.node_addresses = $6 and
      a.features = $7
);`

func insertNodeEvent(db *sqlx.DB, ts time.Time, nodeId int, pubKey string, alias string, color string,
	na []*lnrpc.NodeAddress, f map[uint32]*lnrpc.Feature) error {

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

	if _, err = db.Exec(neQuery, ts, nodeId, pubKey, alias, color, najb, fjb); err != nil {
		return errors.Wrap(err, "Executing SQL")
	}

	return nil
}
