package automation

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

type RebalanceAttempt struct {
	OutgoingChannelIds []int
	IncomingChannelId  int
	InitiationTime     time.Time
	MaximumCost        int
}

func RebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer, eventChannel chan interface{}) {

	client := lnrpc.NewLightningClient(conn)
	router := routerrpc.NewRouterClient(conn)

	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	rebalanceAttempts := make(map[int]RebalanceAttempt)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		listener := broadcaster.Subscribe()
		for event := range listener {
			select {
			case <-ctx.Done():
				broadcaster.CancelSubscription(listener)
				return
			default:
			}
			if rebalanceEvent, ok := event.(commons.RebalanceRequest); ok {
				if rebalanceEvent.NodeId != nodeSettings.NodeId {
					return
				}
				if rebalanceEvent.IncomingChannelId == 0 {
					return
				}
				if len(rebalanceEvent.OutgoingChannelIds) == 0 {
					return
				}
				// TODO check if rebalancer isn't already running
				rebalanceAttempts[rebalanceEvent.IncomingChannelId] = RebalanceAttempt{
					OutgoingChannelIds: rebalanceEvent.OutgoingChannelIds,
					IncomingChannelId:  rebalanceEvent.IncomingChannelId,
					InitiationTime:     time.Now().UTC(),
					MaximumCost:        rebalanceEvent.MaximumCost,
				}
				// TODO start rebalancing in go routine
			}
		}
	}
}
