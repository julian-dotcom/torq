package lnd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/graph_events"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

const routingPolicyUpdateLimiterSeconds = 5 * 60

// 70 because a reconnection is attempted every 60 seconds
const avoidChannelAndPolicyImportRerunTimeSeconds = 70

func LightningCommunicationService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer, serviceEventChannel chan<- commons.ServiceEvent) {

	defer log.Info().Msgf("LightningCommunicationService terminated for nodeId: %v", nodeId)

	client := lnrpc.NewLightningClient(conn)
	router := routerrpc.NewRouterClient(conn)

	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	successTimes := make(map[commons.ImportType]time.Time, 0)

	wg := sync.WaitGroup{}
	listener := broadcaster.SubscribeLightningRequest()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ctx.Done() {
			broadcaster.CancelSubscriptionLightningRequest(listener)
			return
		}
	}()
	go func() {
		for lightningRequest := range listener {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
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
			if request, ok := lightningRequest.(commons.ImportRequest); ok {
				if request.NodeId != nodeSettings.NodeId {
					continue
				}
				response := processImportRequest(ctx, db, request, successTimes, client)
				if request.ResponseChannel != nil {
					request.ResponseChannel <- response
				}
			}
		}
	}()

	commons.SendServiceEvent(nodeId, serviceEventChannel, commons.ServicePending, commons.ServiceActive, commons.LightningCommunicationService, nil)

	wg.Wait()
}

func processImportRequest(ctx context.Context, db *sqlx.DB, request commons.ImportRequest,
	successTimes map[commons.ImportType]time.Time,
	client lnrpc.LightningClient) commons.ImportResponse {

	nodeSettings := commons.GetNodeSettingsByNodeId(request.NodeId)

	response := commons.ImportResponse{
		Request: request,
		CommunicationResponse: commons.CommunicationResponse{
			Status: commons.Inactive,
		},
	}

	if !request.Force {
		successTime, exists := successTimes[request.ImportType]
		if exists && time.Since(successTime).Seconds() < avoidChannelAndPolicyImportRerunTimeSeconds {
			switch request.ImportType {
			case commons.ImportAllChannels:
				log.Info().Msgf("All Channels were imported very recently for nodeId: %v.", request.NodeId)
			case commons.ImportPendingChannelsOnly:
				log.Info().Msgf("Pending Channels were imported very recently for nodeId: %v.", request.NodeId)
			case commons.ImportChannelRoutingPolicies:
				log.Info().Msgf("ChannelRoutingPolicies were imported very recently for nodeId: %v.", request.NodeId)
			case commons.ImportNodeInformation:
				log.Info().Msgf("NodeInformation were imported very recently for nodeId: %v.", request.NodeId)
			}
			return response
		}
	}
	if request.Force {
		switch request.ImportType {
		case commons.ImportAllChannels:
			log.Info().Msgf("Forced import of All Channels for nodeId: %v.", request.NodeId)
		case commons.ImportPendingChannelsOnly:
			log.Info().Msgf("Forced import of Pending Channels for nodeId: %v.", request.NodeId)
		case commons.ImportChannelRoutingPolicies:
			log.Info().Msgf("Forced import of ChannelRoutingPolicies for nodeId: %v.", request.NodeId)
		case commons.ImportNodeInformation:
			log.Info().Msgf("Forced import of NodeInformation for nodeId: %v.", request.NodeId)
		}
	}
	switch request.ImportType {
	case commons.ImportAllChannels:
		var err error
		//Import Pending channels
		err = ImportPendingChannels(ctx, db, client, nodeSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to import pending channels.")
			response.Error = err
			return response
		}

		//Import Open channels
		err = ImportOpenChannels(ctx, db, client, nodeSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to import open channels.")
			response.Error = err
			return response
		}

		// Import Closed channels
		err = ImportClosedChannels(ctx, db, client, nodeSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to import closed channels.")
			response.Error = err
			return response
		}

		// TODO FIXME channels with short_channel_id = null and status IN (1,2,100,101,102,103) should be fixed somehow???
		//  Open                   = 1
		//  Closing                = 2
		//	CooperativeClosed      = 100
		//	LocalForceClosed       = 101
		//	RemoteForceClosed      = 102
		//	BreachClosed           = 103

		err = channels.InitializeManagedChannelCache(db)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to Initialize ManagedChannelCache.")
			response.Error = err
			return response
		}
		log.Info().Msgf("All Channels were imported successfully for nodeId: %v.", nodeSettings.NodeId)
	case commons.ImportPendingChannelsOnly:
		err := ImportPendingChannels(ctx, db, client, nodeSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to import pending channels.")
			response.Error = err
			return response
		}

		err = channels.InitializeManagedChannelCache(db)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to Initialize ManagedChannelCache.")
			response.Error = err
			return response
		}
		log.Info().Msgf("Pending Channels were imported successfully for nodeId: %v.", nodeSettings.NodeId)
	case commons.ImportChannelRoutingPolicies:
		err := ImportRoutingPolicies(ctx, client, db, nodeSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to import routing policies.")
			response.Error = err
			return response
		}
		log.Info().Msgf("ChannelRoutingPolicies were imported successfully for nodeId: %v.", nodeSettings.NodeId)
	case commons.ImportNodeInformation:
		err := ImportNodeInfo(ctx, client, db, nodeSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to import node information.")
			response.Error = err
			return response
		}
		log.Info().Msgf("NodeInformation was imported successfully for nodeId: %v.", nodeSettings.NodeId)
	}
	successTimes[request.ImportType] = time.Now()
	response.Status = commons.Active
	return response
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
	secondsAgo := routingPolicyUpdateLimiterSeconds
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
	rateLimitSeconds := routingPolicyUpdateLimiterSeconds
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
					Error:  fmt.Sprintf("Routing policy update ignored due to rate limiter for channelId: %v", request.ChannelId),
				},
			}
		}
	}
	return nil
}
