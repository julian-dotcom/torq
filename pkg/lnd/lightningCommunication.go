package lnd

import (
	"context"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/graph_events"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

func LightningCommunicationService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer) {

	defer log.Info().Msgf("LightningCommunicationService terminated for nodeId: %v", nodeId)

	client := lnrpc.NewLightningClient(conn)
	router := routerrpc.NewRouterClient(conn)

	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	wg := sync.WaitGroup{}
	listener := broadcaster.SubscribeLightningRequest()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ctx.Done() {
			broadcaster.CancelSubscriptionWebSocketResponse(listener)
			return
		}
	}()
	go func() {
		for lightningRequest := range listener {
			if request, ok := lightningRequest.(commons.ChannelStatusUpdateRequest); ok {
				if request.NodeId != nodeSettings.NodeId {
					continue
				}
				response := processChannelStatusUpdateRequest(ctx, db, request, router)
				if request.ResponseChannel != nil {
					request.ResponseChannel <- response
				}
			}
			if request, ok := lightningRequest.(commons.RoutingPolicyUpdateRequest); ok {
				if request.NodeId != nodeSettings.NodeId {
					continue
				}
				response := processRoutingPolicyUpdateRequest(ctx, db, request, client)
				if request.ResponseChannel != nil {
					request.ResponseChannel <- response
				}
			}
			if request, ok := lightningRequest.(commons.SignatureVerificationRequest); ok {
				if request.NodeId != nodeSettings.NodeId {
					continue
				}
				response := processSignatureVerificationRequest(ctx, request, client)
				if request.ResponseChannel != nil {
					request.ResponseChannel <- response
				}
			}
			if request, ok := lightningRequest.(commons.SignMessageRequest); ok {
				if request.NodeId != nodeSettings.NodeId {
					continue
				}
				response := processSignMessageRequest(ctx, request, client)
				if request.ResponseChannel != nil {
					request.ResponseChannel <- response
				}
			}
		}
	}()
	wg.Wait()
}

func processSignMessageRequest(ctx context.Context, request commons.SignMessageRequest,
	client lnrpc.LightningClient) commons.SignMessageResponse {

	response := commons.SignMessageResponse{
		Request: request,
		CommunicationResponse: commons.CommunicationResponse{
			Status: commons.Inactive,
		},
	}

	signMsgReq := lnrpc.SignMessageRequest{
		Msg: []byte(request.Message),
	}
	if request.SingleHash != nil {
		signMsgReq.SingleHash = *request.SingleHash
	}

	signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Status = commons.Active
	response.Signature = signMsgResp.Signature
	return response
}

func processSignatureVerificationRequest(ctx context.Context, request commons.SignatureVerificationRequest,
	client lnrpc.LightningClient) commons.SignatureVerificationResponse {

	response := commons.SignatureVerificationResponse{
		Request: request,
		CommunicationResponse: commons.CommunicationResponse{
			Status: commons.Inactive,
		},
	}

	verifyMsgReq := lnrpc.VerifyMessageRequest{
		Msg:       []byte(request.Message),
		Signature: request.Signature,
	}
	verifyMsgResp, err := client.VerifyMessage(ctx, &verifyMsgReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	if !verifyMsgResp.Valid {
		response.Message = "Signature is not valid"
		return response
	}

	response.Status = commons.Active
	response.PublicKey = verifyMsgResp.Pubkey
	response.Valid = verifyMsgResp.GetValid()
	return response
}

func processChannelStatusUpdateRequest(ctx context.Context, db *sqlx.DB, request commons.ChannelStatusUpdateRequest,
	router routerrpc.RouterClient) commons.ChannelStatusUpdateResponse {
	response := validateChannelStatusUpdateRequest(request)
	if response != nil {
		return *response
	}

	if !channelStatusUpdateRequestContainsUpdates(request) {
		return commons.ChannelStatusUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status:  commons.Active,
				Message: "Nothing changed so update is ignored",
			},
		}
	}

	response = channelStatusUpdateRequestIsRepeated(db, request)
	if response != nil {
		return *response
	}

	_, err := router.UpdateChanStatus(ctx, constructUpdateChanStatusRequest(request))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to update channel status for channelId: %v on nodeId: %v", request.ChannelId, request.NodeId)
		return commons.ChannelStatusUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  err.Error(),
			},
		}
	}
	return commons.ChannelStatusUpdateResponse{
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

func channelStatusUpdateRequestIsRepeated(db *sqlx.DB, request commons.ChannelStatusUpdateRequest) *commons.ChannelStatusUpdateResponse {
	secondsAgo := commons.ROUTING_POLICY_UPDATE_LIMITER_SECONDS
	channelEventsFromGraph, err := graph_events.GetChannelEventFromGraph(db, request.ChannelId, &secondsAgo)
	if err != nil {
		return &commons.ChannelStatusUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  err.Error(),
			},
		}
	}

	if len(channelEventsFromGraph) > 1 {
		disabled := channelEventsFromGraph[0].Disabled
		disabledCounter := 0
		for i := 0; i < len(channelEventsFromGraph); i++ {
			if disabled != channelEventsFromGraph[i].Disabled {
				disabledCounter++
				disabled = channelEventsFromGraph[i].Disabled
			}
		}
		if disabledCounter > 2 {
			return &commons.ChannelStatusUpdateResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Inactive,
					Error:  err.Error(),
				},
			}
		}
	}
	return nil
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

func processRoutingPolicyUpdateRequest(ctx context.Context, db *sqlx.DB, request commons.RoutingPolicyUpdateRequest,
	client lnrpc.LightningClient) commons.RoutingPolicyUpdateResponse {

	response := validateRoutingPolicyUpdateRequest(request)
	if response != nil {
		return *response
	}

	channelState := commons.GetChannelState(request.NodeId, request.ChannelId, true)
	if channelState == nil {
		return commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status:  commons.Inactive,
				Message: "channelState was nil",
			},
		}
	}
	if !routingPolicyUpdateRequestContainsUpdates(request, channelState) {
		return commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status:  commons.Active,
				Message: "Nothing changed so update is ignored",
			},
		}
	}

	response = routingPolicyUpdateRequestIsRepeated(db, request)
	if response != nil {
		return *response
	}

	resp, err := client.UpdateChannelPolicy(ctx, constructPolicyUpdateRequest(request, channelState))
	return processRoutingPolicyUpdateResponse(request, resp, err)
}

func processRoutingPolicyUpdateResponse(request commons.RoutingPolicyUpdateRequest, resp *lnrpc.PolicyUpdateResponse,
	err error) commons.RoutingPolicyUpdateResponse {

	if err != nil && resp == nil {
		log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v", request.ChannelId, request.NodeId)
		return commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
			},
		}
	}
	var failedUpdateArray []commons.FailedRequest
	for _, failedUpdate := range resp.GetFailedUpdates() {
		log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (lnd-grpc error: %v)",
			request.ChannelId, request.NodeId, failedUpdate.Reason)
		failedRequest := commons.FailedRequest{
			Reason: failedUpdate.UpdateError,
			Error:  failedUpdate.UpdateError,
		}
		failedUpdateArray = append(failedUpdateArray, failedRequest)
	}
	if err != nil || len(failedUpdateArray) != 0 {
		log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v", request.ChannelId, request.NodeId)
		return commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
			},
			FailedUpdates: failedUpdateArray,
		}
	}
	return commons.RoutingPolicyUpdateResponse{
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
		policyUpdateRequest.BaseFeeMsat = channelState.LocalFeeBaseMsat
	} else {
		policyUpdateRequest.BaseFeeMsat = *request.FeeBaseMsat
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

func routingPolicyUpdateRequestIsRepeated(db *sqlx.DB, request commons.RoutingPolicyUpdateRequest) *commons.RoutingPolicyUpdateResponse {
	rateLimitSeconds := commons.ROUTING_POLICY_UPDATE_LIMITER_SECONDS
	if request.RateLimitSeconds > 0 {
		rateLimitSeconds = request.RateLimitSeconds
	}
	channelEventsFromGraph, err := graph_events.GetChannelEventFromGraph(db, request.ChannelId, &rateLimitSeconds)
	if err != nil {
		return &commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  err.Error(),
			},
		}
	}

	if len(channelEventsFromGraph) > 1 {
		timeLockDelta := channelEventsFromGraph[0].TimeLockDelta
		timeLockDeltaCounter := 1
		minHtlcMsat := channelEventsFromGraph[0].MinHtlcMsat
		minHtlcMsatCounter := 1
		maxHtlcMsat := channelEventsFromGraph[0].MaxHtlcMsat
		maxHtlcMsatCounter := 1
		feeBaseMsat := channelEventsFromGraph[0].FeeBaseMsat
		feeBaseMsatCounter := 1
		feeRateMilliMsat := channelEventsFromGraph[0].FeeRateMilliMsat
		feeRateMilliMsatCounter := 1
		for i := 0; i < len(channelEventsFromGraph); i++ {
			if timeLockDelta != channelEventsFromGraph[i].TimeLockDelta {
				timeLockDeltaCounter++
				timeLockDelta = channelEventsFromGraph[i].TimeLockDelta
			}
			if minHtlcMsat != channelEventsFromGraph[i].MinHtlcMsat {
				minHtlcMsatCounter++
				minHtlcMsat = channelEventsFromGraph[i].MinHtlcMsat
			}
			if maxHtlcMsat != channelEventsFromGraph[i].MaxHtlcMsat {
				maxHtlcMsatCounter++
				maxHtlcMsat = channelEventsFromGraph[i].MaxHtlcMsat
			}
			if feeBaseMsat != channelEventsFromGraph[i].FeeBaseMsat {
				feeBaseMsatCounter++
				feeBaseMsat = channelEventsFromGraph[i].FeeBaseMsat
			}
			if feeRateMilliMsat != channelEventsFromGraph[i].FeeRateMilliMsat {
				feeRateMilliMsatCounter++
				feeRateMilliMsat = channelEventsFromGraph[i].FeeRateMilliMsat
			}
		}
		rateLimitCount := 2
		if request.RateLimitCount > 0 {
			rateLimitCount = request.RateLimitCount
		}
		if timeLockDeltaCounter >= rateLimitCount ||
			minHtlcMsatCounter >= rateLimitCount || maxHtlcMsatCounter >= rateLimitCount ||
			feeBaseMsatCounter >= rateLimitCount || feeRateMilliMsatCounter >= rateLimitCount {

			return &commons.RoutingPolicyUpdateResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Inactive,
					Error:  "Routing policy update ignored due to rate limiter",
				},
			}
		}
	}
	return nil
}
