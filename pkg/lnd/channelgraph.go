package lnd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lncapital/torq/internal/graph_events"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/pkg/commons"
)

type subscribeChannelGraphClient interface {
	lndClientChannelEvent
	SubscribeChannelGraph(ctx context.Context, in *lnrpc.GraphTopologySubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelGraphClient, error)
	GetNodeInfo(ctx context.Context, in *lnrpc.NodeInfoRequest,
		opts ...grpc.CallOption) (*lnrpc.NodeInfo, error)
}

// SubscribeAndStoreChannelGraph Subscribes to channel updates
func SubscribeAndStoreChannelGraph(ctx context.Context, client subscribeChannelGraphClient, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings,
	nodeGraphEventChannel chan<- commons.NodeGraphEvent,
	channelGraphEventChannel chan<- commons.ChannelGraphEvent,
	lightningRequestChannel chan<- interface{}) {

	defer log.Info().Msgf("SubscribeAndStoreChannelGraph terminated for nodeId: %v", nodeSettings.NodeId)

	var stream lnrpc.Lightning_SubscribeChannelGraphClient
	var err error
	var gpu *lnrpc.GraphTopologyUpdate
	serviceStatus := commons.ServiceInactive
	subscriptionStream := commons.GraphEventStream
	var delay bool

	for {
		if delay {
			ticker := clock.New().Tick(streamErrorSleepSeconds * time.Second)
			select {
			case <-ctx.Done():
				return
			case <-ticker:
			}
		} else {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}

		if stream == nil {
			serviceStatus = SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServicePending)
			stream, err = client.SubscribeChannelGraph(ctx, &lnrpc.GraphTopologySubscription{})
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				log.Error().Err(err).Msgf("Obtaining stream (SubscribeChannelGraph) from LND failed, will retry in %v seconds", streamErrorSleepSeconds)
				stream = nil
				delay = true
				continue
			}
			// HACK to know if the context is a testcase.
			if lightningRequestChannel != nil {

				now := time.Now()
				responseChannel := make(chan commons.ImportResponse)
				lightningRequestChannel <- commons.ImportRequest{
					CommunicationRequest: commons.CommunicationRequest{
						RequestId:   fmt.Sprintf("%v", now.Unix()),
						RequestTime: &now,
						NodeId:      nodeSettings.NodeId,
					},
					ImportType:      commons.ImportAllChannels,
					ResponseChannel: responseChannel,
				}
				response := <-responseChannel
				if response.Error != nil {
					log.Error().Err(response.Error).Msgf("Obtaining All Channels (SubscribeChannelGraph) from LND failed, will retry in %v seconds", streamErrorSleepSeconds)
					stream = nil
					delay = true
					continue
				}

				responseChannel = make(chan commons.ImportResponse)
				lightningRequestChannel <- commons.ImportRequest{
					CommunicationRequest: commons.CommunicationRequest{
						RequestId:   fmt.Sprintf("%v", now.Unix()),
						RequestTime: &now,
						NodeId:      nodeSettings.NodeId,
					},
					ImportType:      commons.ImportChannelRoutingPolicies,
					ResponseChannel: responseChannel,
				}
				response = <-responseChannel
				if response.Error != nil {
					log.Error().Err(response.Error).Msgf("Obtaining RoutingPolicies (SubscribeChannelGraph) from LND failed, will retry in %v seconds", streamErrorSleepSeconds)
					stream = nil
					delay = true
					continue
				}

				responseChannel = make(chan commons.ImportResponse)
				lightningRequestChannel <- commons.ImportRequest{
					CommunicationRequest: commons.CommunicationRequest{
						RequestId:   fmt.Sprintf("%v", now.Unix()),
						RequestTime: &now,
						NodeId:      nodeSettings.NodeId,
					},
					ImportType:      commons.ImportNodeInformation,
					ResponseChannel: responseChannel,
				}
				response = <-responseChannel
				if response.Error != nil {
					log.Error().Err(response.Error).Msgf("Obtaining Node Information (SubscribeChannelGraph) from LND failed, will retry in %v seconds", streamErrorSleepSeconds)
					stream = nil
					delay = true
					continue
				}
			}
			serviceStatus = SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServiceActive)
		}

		gpu, err = stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServicePending)
			log.Error().Err(err).Msgf("Receiving channel graph events from the stream failed, will retry in %v seconds", streamErrorSleepSeconds)
			stream = nil
			delay = true
			continue
		}

		err = processNodeUpdates(gpu.NodeUpdates, db, nodeSettings, nodeGraphEventChannel)
		if err != nil {
			// TODO FIXME STORE THIS SOMEWHERE??? NODE UPDATES ARE NOW IGNORED???
			log.Error().Err(err).Msgf("Failed to store node update events")
		}

		err = processChannelUpdates(gpu.ChannelUpdates, db, nodeSettings, channelGraphEventChannel)
		if err != nil {
			// TODO FIXME STORE THIS SOMEWHERE??? CHANNEL UPDATES ARE NOW IGNORED???
			log.Error().Err(err).Msgf("Failed to store channel update events")
		}
		delay = false
	}
}

func ImportNodeInfo(ctx context.Context, client subscribeChannelGraphClient, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings) error {
	// Get all node public keys with channels
	publicKeys := commons.GetAllChannelPublicKeys(nodeSettings.Chain, nodeSettings.Network)

	for _, publicKey := range publicKeys {
		ni, err := client.GetNodeInfo(ctx, &lnrpc.NodeInfoRequest{PubKey: publicKey, IncludeChannels: false})
		if err != nil {
			if e, ok := status.FromError(err); ok {
				switch e.Code() {
				case codes.NotFound:
					log.Debug().Err(err).Msgf("Node info not found error when importing node info for public key: %v", publicKey)
					continue
				default:
					return errors.Wrap(err, "Get node info")
				}
			}
		}
		err = insertNodeEvent(db, time.Now().UTC(),
			commons.GetNodeIdByPublicKey(publicKey, nodeSettings.Chain, nodeSettings.Network),
			ni.Node.Alias, ni.Node.Color, ni.Node.Addresses, ni.Node.Features, nodeSettings.NodeId, nil)
		if err != nil {
			return errors.Wrap(err, "Insert node event")
		}
	}
	return nil
}

func processNodeUpdates(nus []*lnrpc.NodeUpdate, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	nodeGraphEventChannel chan<- commons.NodeGraphEvent) error {
	for _, nu := range nus {
		eventNodeId := commons.GetActiveNodeIdByPublicKey(nu.IdentityKey, nodeSettings.Chain, nodeSettings.Network)
		if eventNodeId != 0 {
			err := insertNodeEvent(db, time.Now().UTC(), eventNodeId, nu.Alias, nu.Color,
				nu.NodeAddresses, nu.Features, nodeSettings.NodeId, nodeGraphEventChannel)
			if err != nil {
				return errors.Wrapf(err, "Insert node event")
			}
		}
	}
	return nil
}

func processChannelUpdates(cus []*lnrpc.ChannelEdgeUpdate, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, channelGraphEventChannel chan<- commons.ChannelGraphEvent) error {
	for _, cu := range cus {
		channelPoint, err := chanPointFromByte(cu.ChanPoint.GetFundingTxidBytes(), cu.ChanPoint.GetOutputIndex())
		if err != nil {
			return errors.Wrap(err, "Creating channel point from byte")
		}
		fundingTransactionHash, fundingOutputIndex := commons.ParseChannelPoint(channelPoint)

		channelId := commons.GetActiveChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
		if channelId != 0 {
			err := insertRoutingPolicy(db, time.Now().UTC(), channelId, nodeSettings, cu, channelGraphEventChannel)
			if err != nil {
				return errors.Wrap(err, "Insert routing policy")
			}
		}
	}
	return nil
}

func insertRoutingPolicy(
	db *sqlx.DB,
	eventTime time.Time,
	channelId int,
	nodeSettings commons.ManagedNodeSettings,
	cu *lnrpc.ChannelEdgeUpdate,
	channelGraphEventChannel chan<- commons.ChannelGraphEvent) error {

	channelSettings := commons.GetChannelSettingByChannelId(channelId)

	var err error
	if cu == nil || cu.RoutingPolicy == nil {
		if !channelSettings.Private {
			log.Info().Msgf("Routing policy nil, skipping it for LND channel id: %v", cu.ChanId)
		}
		return nil
	}

	announcingNodeId := 0
	if cu.AdvertisingNode != "" {
		announcingNodeId = commons.GetNodeIdByPublicKey(cu.AdvertisingNode, nodeSettings.Chain, nodeSettings.Network)
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
	}
	connectingNodeId := 0
	if cu.ConnectingNode != "" {
		connectingNodeId = commons.GetNodeIdByPublicKey(cu.ConnectingNode, nodeSettings.Chain, nodeSettings.Network)
		if connectingNodeId == 0 {
			newNode := nodes.Node{
				PublicKey: cu.ConnectingNode,
				Chain:     nodeSettings.Chain,
				Network:   nodeSettings.Network,
			}
			connectingNodeId, err = nodes.AddNodeWhenNew(db, newNode)
			if err != nil {
				return errors.Wrapf(err, "Adding node (publicKey: %v)", cu.ConnectingNode)
			}
		}
	}
	if connectingNodeId == 0 && announcingNodeId == 0 {
		return errors.New(fmt.Sprintf("Cannot obtain announcingNodeId nor connectingNodeId (from AdvertisingNode: %v ConnectingNode: %v)", cu.AdvertisingNode, cu.ConnectingNode))
	}
	if connectingNodeId == 0 {
		if announcingNodeId == channelSettings.FirstNodeId {
			connectingNodeId = channelSettings.SecondNodeId
		} else {
			connectingNodeId = channelSettings.FirstNodeId
		}
	}
	if announcingNodeId == 0 {
		if connectingNodeId == channelSettings.FirstNodeId {
			announcingNodeId = channelSettings.SecondNodeId
		} else {
			announcingNodeId = channelSettings.FirstNodeId
		}
	}

	channelEvent := graph_events.ChannelEventFromGraph{}
	err = db.Get(&channelEvent, `
				SELECT *
				FROM routing_policy
				WHERE channel_id=$1 AND announcing_node_id=$2 AND connecting_node_id=$3
				ORDER BY ts DESC
				LIMIT 1;`, channelId, announcingNodeId, connectingNodeId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrapf(err, "insertNodeEvent -> getPreviousChannelEvent.")
		}
	}

	// If one of our active torq nodes is announcing_node_id then the channel update was by our node
	// TODO FIXME ignore if previous update was from the same node so if announcing_node_id=node_id on previous record
	// and the current parameters are announcing_node_id!=node_id
	if cu.RoutingPolicy.Disabled != channelEvent.Disabled ||
		cu.RoutingPolicy.FeeBaseMsat != channelEvent.FeeBaseMsat ||
		cu.RoutingPolicy.FeeRateMilliMsat != channelEvent.FeeRateMilliMsat ||
		cu.RoutingPolicy.MaxHtlcMsat != channelEvent.MaxHtlcMsat ||
		uint64(cu.RoutingPolicy.MinHtlc) != channelEvent.MinHtlcMsat ||
		cu.RoutingPolicy.TimeLockDelta != channelEvent.TimeLockDelta {

		_, err := db.Exec(`
		INSERT INTO routing_policy
			(ts,disabled,time_lock_delta,min_htlc,max_htlc_msat,fee_base_msat,fee_rate_mill_msat,
			 channel_id,announcing_node_id,connecting_node_id,node_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`, eventTime,
			cu.RoutingPolicy.Disabled, cu.RoutingPolicy.TimeLockDelta, cu.RoutingPolicy.MinHtlc,
			cu.RoutingPolicy.MaxHtlcMsat, cu.RoutingPolicy.FeeBaseMsat, cu.RoutingPolicy.FeeRateMilliMsat,
			channelId, announcingNodeId, connectingNodeId, nodeSettings.NodeId)
		if err != nil {
			return errors.Wrapf(err, "insertRoutingPolicy")
		}

		if channelGraphEventChannel != nil {
			channelGraphEvent := commons.ChannelGraphEvent{
				GraphEventData: commons.GraphEventData{
					EventData: commons.EventData{
						EventTime: eventTime,
						NodeId:    nodeSettings.NodeId,
					},
					AnnouncingNodeId: &announcingNodeId,
					ConnectingNodeId: &connectingNodeId,
					ChannelId:        &channelId,
				},
				ChannelGraphEventData: commons.ChannelGraphEventData{
					TimeLockDelta:    cu.RoutingPolicy.TimeLockDelta,
					FeeRateMilliMsat: int64(cu.RoutingPolicy.FeeRateMilliMsat),
					FeeBaseMsat:      int64(cu.RoutingPolicy.FeeBaseMsat),
					MaxHtlcMsat:      cu.RoutingPolicy.MaxHtlcMsat,
					Disabled:         cu.RoutingPolicy.Disabled,
					MinHtlcMsat:      uint64(cu.RoutingPolicy.MinHtlc),
				},
			}
			if channelEvent.ChannelId != 0 {
				channelGraphEvent.PreviousEventTime = &channelEvent.EventTime
				channelGraphEvent.PreviousEventData = &commons.ChannelGraphEventData{
					TimeLockDelta:    channelEvent.TimeLockDelta,
					FeeRateMilliMsat: int64(channelEvent.FeeRateMilliMsat),
					FeeBaseMsat:      int64(channelEvent.FeeBaseMsat),
					MaxHtlcMsat:      channelEvent.MaxHtlcMsat,
					Disabled:         channelEvent.Disabled,
					MinHtlcMsat:      channelEvent.MinHtlcMsat,
				}
			}
			channelGraphEventChannel <- channelGraphEvent
		}
	}
	return nil
}

func insertNodeEvent(db *sqlx.DB, eventTime time.Time, eventNodeId int, alias string, color string,
	nodeAddress []*lnrpc.NodeAddress, features map[uint32]*lnrpc.Feature, nodeId int, nodeGraphEventChannel chan<- commons.NodeGraphEvent) error {

	// Create json byte object from node address map
	najb, err := json.Marshal(nodeAddress)
	if err != nil {
		return errors.Wrap(err, "JSON Marshall node address map")
	}

	// Create json byte object from features list
	fjb, err := json.Marshal(features)
	if err != nil {
		return errors.Wrap(err, "JSON Marshal feature list")
	}

	nodeEvent := graph_events.NodeEventFromGraph{}
	err = db.Get(&nodeEvent, `
				SELECT *
				FROM node_event
				WHERE event_node_id=$1
				ORDER BY timestamp DESC
				LIMIT 1;`, eventNodeId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrapf(err, "insertNodeEvent -> getPreviousNodeEvent.")
		}
	}
	// TODO FIXME ignore if previous update was from the same node so if event_node_id=node_id on previous record
	// and the current parameters are event_node_id!=node_id
	if alias != nodeEvent.Alias ||
		color != nodeEvent.Color ||
		string(najb) != nodeEvent.NodeAddresses ||
		string(fjb) != nodeEvent.Features {

		_, err = db.Exec(`INSERT INTO node_event
    		(timestamp, event_node_id, alias, color, node_addresses, features, node_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7);`,
			eventTime, eventNodeId, alias, color, najb, fjb, nodeId)
		if err != nil {
			return errors.Wrap(err, "Executing SQL")
		}
		commons.SetNodeAlias(eventNodeId, alias)

		if nodeGraphEventChannel != nil {
			nodeGraphEvent := commons.NodeGraphEvent{
				GraphEventData: commons.GraphEventData{
					EventData: commons.EventData{
						EventTime: eventTime,
						NodeId:    nodeId,
					},
					EventNodeId: &eventNodeId,
				},
				NodeGraphEventData: commons.NodeGraphEventData{
					Alias:     alias,
					Color:     color,
					Addresses: string(najb),
					Features:  string(fjb),
				},
			}
			if nodeEvent.NodeId != 0 {
				nodeGraphEvent.PreviousEventTime = &nodeEvent.EventTime
				nodeGraphEvent.PreviousEventData = &commons.NodeGraphEventData{
					Alias:     nodeEvent.Alias,
					Color:     nodeEvent.Color,
					Addresses: nodeEvent.NodeAddresses,
					Features:  nodeEvent.Features,
				}
			}
			nodeGraphEventChannel <- nodeGraphEvent
		}
	}
	return nil
}
