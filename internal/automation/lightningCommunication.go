package automation

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

type RoutingPolicyHistory struct {
	ChannelId     int
	ExecutionTime time.Time
	RoutingPolicy lnrpc.RoutingPolicy
}

func LightningCommunicationService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer, eventChannel chan interface{}) {

	client := lnrpc.NewLightningClient(conn)
	router := routerrpc.NewRouterClient(conn)

	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	// TODO FIXME IMPLEMENT SAME CHANNEL POLICY UPDATE RATE LIMITER

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
			var response interface{}
			var responseChannel chan interface{}
			if request, ok := event.(commons.ChannelStatusUpdateRequest); ok {
				if request.NodeId != nodeSettings.NodeId {
					return
				}
				if request.ResponseChannel != nil {
					responseChannel = request.ResponseChannel
				}
				response = processChannelStatusUpdateRequest(ctx, request, router)
			}
			if request, ok := event.(commons.RoutingPolicyUpdateRequest); ok {
				if request.NodeId != nodeSettings.NodeId {
					return
				}
				if request.ResponseChannel != nil {
					responseChannel = request.ResponseChannel
				}
				response = processRoutingPolicyUpdateRequest(ctx, request, client)
			}
			if response != nil {
				if responseChannel != nil {
					responseChannel <- response
				}
				if eventChannel != nil {
					eventChannel <- response
				}
			}
		}
	}
}

func processChannelStatusUpdateRequest(ctx context.Context, request commons.ChannelStatusUpdateRequest, router routerrpc.RouterClient) *commons.ChannelStatusUpdateResponse {
	response := validateChannelStatusUpdateRequest(request)
	if response != nil {
		return response
	}
	if !channelStatusUpdateRequestContainsUpdates(request) {
		return &commons.ChannelStatusUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status:  commons.Active,
				Message: "Nothing changed so update is ignored",
			},
		}
	}
	_, err := router.UpdateChanStatus(ctx, constructUpdateChanStatusRequest(request))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v", request.ChannelId, request.NodeId)
		return &commons.ChannelStatusUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  err.Error(),
			},
		}
	}
	return &commons.ChannelStatusUpdateResponse{
		Request: request,
		CommunicationResponse: commons.CommunicationResponse{
			Status: commons.Active,
		},
	}
}

func constructUpdateChanStatusRequest(request commons.ChannelStatusUpdateRequest) *routerrpc.UpdateChanStatusRequest {
	action := routerrpc.ChanStatusAction_DISABLE
	if request.ChannelStatus == commons.Active {
		action = routerrpc.ChanStatusAction_ENABLE
	}
	channelSettings := commons.GetChannelSettingByChannelId(request.ChannelId)
	return &routerrpc.UpdateChanStatusRequest{
		ChanPoint: &lnrpc.ChannelPoint{
			FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: channelSettings.FundingTransactionHash},
			OutputIndex: uint32(channelSettings.FundingOutputIndex)},
		Action: action,
	}
}

func channelStatusUpdateRequestContainsUpdates(request commons.ChannelStatusUpdateRequest) bool {
	channelState := commons.GetChannelState(request.NodeId, request.ChannelId, true)
	if request.ChannelStatus == commons.Active && channelState.LocalDisabled {
		return true
	}
	if request.ChannelStatus == commons.Inactive && !channelState.LocalDisabled {
		return true
	}
	return false
}

func validateChannelStatusUpdateRequest(request commons.ChannelStatusUpdateRequest) *commons.ChannelStatusUpdateResponse {
	if request.ChannelId == 0 {
		return &commons.ChannelStatusUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  "ChannelId is 0",
			},
		}
	}
	if request.ChannelStatus != commons.Active &&
		request.ChannelStatus != commons.Inactive {
		return &commons.ChannelStatusUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  "ChannelStatus is not Active nor Inactive",
			},
		}
	}
	return nil
}

func processRoutingPolicyUpdateRequest(ctx context.Context, request commons.RoutingPolicyUpdateRequest, client lnrpc.LightningClient) *commons.RoutingPolicyUpdateResponse {
	response := validateRoutingPolicyUpdateRequest(request)
	if response != nil {
		return response
	}
	channelState := commons.GetChannelState(request.NodeId, request.ChannelId, true)
	if !routingPolicyUpdateRequestContainsUpdates(request, channelState) {
		return &commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status:  commons.Active,
				Message: "Nothing changed so update is ignored",
			},
		}
	}
	resp, err := client.UpdateChannelPolicy(ctx, constructPolicyUpdateRequest(request, channelState))
	var failedUpdateArray []commons.FailedRequest
	for _, failedUpdate := range resp.GetFailedUpdates() {
		log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (%v)",
			request.ChannelId, request.NodeId, failedUpdate.UpdateError)
		failedRequest := commons.FailedRequest{
			Reason: failedUpdate.UpdateError,
			Error:  failedUpdate.UpdateError,
		}
		failedUpdateArray = append(failedUpdateArray, failedRequest)
	}
	if err != nil || len(failedUpdateArray) > 0 {
		log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v", request.ChannelId, request.NodeId)
		return &commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
			},
			FailedUpdates: failedUpdateArray,
		}
	}
	return &commons.RoutingPolicyUpdateResponse{
		Request: request,
		CommunicationResponse: commons.CommunicationResponse{
			Status: commons.Active,
		},
	}
}

func constructPolicyUpdateRequest(request commons.RoutingPolicyUpdateRequest, channelState *commons.ManagedChannelStateSettings) *lnrpc.PolicyUpdateRequest {
	policyUpdateRequest := &lnrpc.PolicyUpdateRequest{}
	if request.TimeLockDelta == nil {
		policyUpdateRequest.TimeLockDelta = channelState.LocalTimeLockDelta
	} else {
		policyUpdateRequest.TimeLockDelta = *request.TimeLockDelta
	}
	if request.FeeRateMilliMsat == nil {
		policyUpdateRequest.FeeRatePpm = uint32(channelState.LocalFeeRateMilliMsat)
	} else {
		policyUpdateRequest.FeeRatePpm = uint32(*request.FeeRateMilliMsat)
	}
	if request.FeeBaseMsat == nil {
		policyUpdateRequest.BaseFeeMsat = int64(channelState.LocalFeeBaseMsat)
	} else {
		policyUpdateRequest.BaseFeeMsat = int64(*request.FeeBaseMsat)
	}
	if request.MinHtlcMsat == nil {
		policyUpdateRequest.MinHtlcMsat = channelState.LocalMinHtlcMsat
	} else {
		policyUpdateRequest.MinHtlcMsat = *request.MinHtlcMsat
	}
	policyUpdateRequest.MinHtlcMsatSpecified = true
	if request.MaxHtlcMsat == nil {
		policyUpdateRequest.MaxHtlcMsat = channelState.LocalMaxHtlcMsat
	} else {
		policyUpdateRequest.MaxHtlcMsat = *request.MaxHtlcMsat
	}
	channelSettings := commons.GetChannelSettingByChannelId(request.ChannelId)
	policyUpdateRequest.Scope = &lnrpc.PolicyUpdateRequest_ChanPoint{
		ChanPoint: &lnrpc.ChannelPoint{
			FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: channelSettings.FundingTransactionHash},
			OutputIndex: uint32(channelSettings.FundingOutputIndex)}}
	return policyUpdateRequest
}

func validateRoutingPolicyUpdateRequest(request commons.RoutingPolicyUpdateRequest) *commons.RoutingPolicyUpdateResponse {
	if request.FeeRateMilliMsat == nil &&
		request.FeeBaseMsat == nil &&
		request.MaxHtlcMsat == nil &&
		request.MinHtlcMsat == nil &&
		request.TimeLockDelta == nil {
		return &commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status:  commons.Active,
				Message: "Nothing changed so update is ignored",
			},
		}
	}
	if request.ChannelId == 0 {
		return &commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  "ChannelId is 0",
			},
		}
	}
	if request.TimeLockDelta != nil && *request.TimeLockDelta < 18 {
		return &commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  "TimeLockDelta is < 18",
			},
		}
	}
	return nil
}

func routingPolicyUpdateRequestContainsUpdates(request commons.RoutingPolicyUpdateRequest, channelState *commons.ManagedChannelStateSettings) bool {
	if request.TimeLockDelta != nil && *request.TimeLockDelta != channelState.LocalTimeLockDelta {
		return true
	}
	if request.FeeRateMilliMsat != nil && *request.FeeRateMilliMsat != channelState.LocalFeeRateMilliMsat {
		return true
	}
	if request.FeeBaseMsat != nil && *request.FeeBaseMsat != channelState.LocalFeeBaseMsat {
		return true
	}
	if request.MinHtlcMsat != nil && *request.MinHtlcMsat != channelState.LocalMinHtlcMsat {
		return true
	}
	if request.MaxHtlcMsat != nil && *request.MaxHtlcMsat != channelState.LocalMaxHtlcMsat {
		return true
	}
	return false
}
