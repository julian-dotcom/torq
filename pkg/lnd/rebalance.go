package lnd

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

type Rebalancer struct {
	OutgoingChannelIds []int
	IncomingChannelId  int
	InitiationTime     time.Time
	MaximumConcurrency int
	MaximumCost        int
	GlobalCtx          context.Context
	Ctx                context.Context
	Cancel             context.CancelFunc
	Failures           int
	Runners            []RebalanceRunner
}

type RebalanceRunner struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

type RebalanceSuccess struct {
	OutgoingChannelId int
	IncomingChannelId int
	Time              time.Time
	Cost              int
}

func RebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer, eventChannel chan interface{}) {

	//client := lnrpc.NewLightningClient(conn)
	router := routerrpc.NewRouterClient(conn)

	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	// rebalancers = map[originType]map[originReference]map[IncomingChannelId]Rebalancer
	rebalancers := make(map[commons.RebalanceRequestOrigin]map[int]map[int]Rebalancer)
	rebalancers[commons.RebalanceRequestWorkflow] = make(map[int]map[int]Rebalancer)
	rebalancers[commons.RebalanceRequestManual] = make(map[int]map[int]Rebalancer)
	rebalanceSuccessCache := make(map[int]RebalanceSuccess)

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
			if request, ok := event.(commons.RebalanceRequest); ok {
				if request.NodeId != nodeSettings.NodeId {
					continue
				}
				processRebalanceRequest(ctx, request, router, rebalancers, rebalanceSuccessCache, eventChannel)
			}
		}
	}
}

func processRebalanceRequest(ctx context.Context, request commons.RebalanceRequest, router routerrpc.RouterClient,
	rebalancers map[commons.RebalanceRequestOrigin]map[int]map[int]Rebalancer,
	rebalanceSuccessCache map[int]RebalanceSuccess, eventChannel chan interface{}) {

	response := validateRebalanceRequest(request, rebalancers)
	if response != nil {
		sendResponse(request, response, eventChannel)
		return
	}

	rebalancersByOrigin, exists := rebalancers[request.OriginType][request.OriginReference]
	if !exists {
		rebalancersByOrigin = make(map[int]Rebalancer)
	}
	rebalanceCtx, rebalanceCancel := context.WithCancel(context.Background())
	rebalancersByOrigin[request.IncomingChannelId] = Rebalancer{
		GlobalCtx:          ctx,
		OutgoingChannelIds: request.OutgoingChannelIds,
		IncomingChannelId:  request.IncomingChannelId,
		InitiationTime:     time.Now().UTC(),
		MaximumCost:        request.MaximumCost,
		Ctx:                rebalanceCtx,
		Cancel:             rebalanceCancel,
	}
	rebalancers[request.OriginType][request.OriginReference] = rebalancersByOrigin

	previousSuccess, exists := rebalanceSuccessCache[request.IncomingChannelId]
	if exists && time.Since(previousSuccess.Time).Seconds() > commons.REBALLANCE_SUCCESS_TIMEOUT_SECONDS {
		previousSuccess = RebalanceSuccess{}
	}
	// TODO start rebalancing in go routine (when something is in previousSuccess then do that first)
}

func sendResponse(request commons.RebalanceRequest, response *commons.RebalanceResponse, eventChannel chan interface{}) {
	if request.ResponseChannel != nil {
		request.ResponseChannel <- response
	}
	if eventChannel != nil {
		eventChannel <- response
	}
}

func validateRebalanceRequest(request commons.RebalanceRequest,
	rebalancers map[commons.RebalanceRequestOrigin]map[int]map[int]Rebalancer) *commons.RebalanceResponse {

	if request.IncomingChannelId == 0 {
		return &commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Active,
				Error:  "IncomingChannelId is 0",
			},
		}
	}
	if len(request.OutgoingChannelIds) == 0 {
		return &commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Active,
				Error:  "OutgoingChannelIds are not specified",
			},
		}
	}
	rebalancersByOrigin, exists := rebalancers[request.OriginType][request.OriginReference]
	if exists {
		_, exists = rebalancersByOrigin[request.IncomingChannelId]
		if exists {
			return &commons.RebalanceResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Active,
					Error: fmt.Sprintf(
						"IncomingChannelId: %v already has a running rebalancer for origin: %v with reference number: %v",
						request.IncomingChannelId, request.OriginType, request.OriginReference),
				},
			}
		}
	}
	return nil
}
