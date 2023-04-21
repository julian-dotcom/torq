package cln

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/graph_events"
	"github.com/lncapital/torq/internal/lightning_helpers"
	"github.com/lncapital/torq/pkg/cln_connect"
	"github.com/lncapital/torq/proto/cln"
)

const routingPolicyUpdateLimiterSeconds = 5 * 60

var (
	connectionWrapperOnce sync.Once           //nolint:gochecknoglobals
	connectionWrapper     *connectionsWrapper //nolint:gochecknoglobals
)

type connectionsWrapper struct {
	mu               sync.Mutex
	connections      map[int]*grpc.ClientConn
	grpcAddresses    map[int]string
	certificateBytes map[int][]byte
	keyBytes         map[int][]byte
}

func getConnection(nodeId int) (*grpc.ClientConn, error) {
	connectionWrapperOnce.Do(func() {
		log.Debug().Msg("Loading Connection Wrapper.")
		connectionWrapper = &connectionsWrapper{
			mu:               sync.Mutex{},
			connections:      make(map[int]*grpc.ClientConn),
			grpcAddresses:    make(map[int]string),
			certificateBytes: make(map[int][]byte),
			keyBytes:         make(map[int][]byte),
		}
	})

	connectionWrapper.mu.Lock()
	defer connectionWrapper.mu.Unlock()

	ncd := cache.GetNodeConnectionDetails(nodeId)

	existingConnection, exists := connectionWrapper.connections[nodeId]
	if !exists ||
		connectionWrapper.grpcAddresses[nodeId] != ncd.GRPCAddress ||
		!bytes.Equal(connectionWrapper.certificateBytes[nodeId], ncd.CertificateFileBytes) ||
		!bytes.Equal(connectionWrapper.keyBytes[nodeId], ncd.KeyFileBytes) {

		conn, err := cln_connect.Connect(ncd.GRPCAddress, ncd.CertificateFileBytes, ncd.KeyFileBytes)
		if err != nil {
			log.Error().Err(err).Msgf("GRPC connection Failed for node id: %v", nodeId)
			return nil, errors.Wrapf(err, "Connecting to GRPC.")
		}
		connectionWrapper.connections[nodeId] = conn
		connectionWrapper.grpcAddresses[nodeId] = ncd.GRPCAddress
		connectionWrapper.certificateBytes[nodeId] = ncd.CertificateFileBytes
		connectionWrapper.keyBytes[nodeId] = ncd.KeyFileBytes
		if exists && existingConnection != nil {
			err = existingConnection.Close()
			if err != nil {
				log.Error().Err(err).Msgf("GRPC close connection failed for node id: %v", nodeId)
			}
		}
	}
	return connectionWrapper.connections[nodeId], nil
}

type lightningService struct {
	limit chan struct{}
}

func Information(
	request lightning_helpers.InformationRequest) lightning_helpers.InformationResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.InformationResponse); ok {
		return res
	}
	return lightning_helpers.InformationResponse{}
}

func SignMessage(
	request lightning_helpers.SignMessageRequest) lightning_helpers.SignMessageResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.SignMessageResponse); ok {
		return res
	}
	return lightning_helpers.SignMessageResponse{}
}

func SignatureVerification(
	request lightning_helpers.SignatureVerificationRequest) lightning_helpers.SignatureVerificationResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.SignatureVerificationResponse); ok {
		return res
	}
	return lightning_helpers.SignatureVerificationResponse{}
}

func RoutingPolicyUpdate(
	request lightning_helpers.RoutingPolicyUpdateRequest) lightning_helpers.RoutingPolicyUpdateResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.RoutingPolicyUpdateResponse); ok {
		return res
	}
	return lightning_helpers.RoutingPolicyUpdateResponse{}
}

func ConnectPeer(
	request lightning_helpers.ConnectPeerRequest) lightning_helpers.ConnectPeerResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.ConnectPeerResponse); ok {
		return res
	}
	return lightning_helpers.ConnectPeerResponse{}
}

func DisconnectPeer(
	request lightning_helpers.DisconnectPeerRequest) lightning_helpers.DisconnectPeerResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.DisconnectPeerResponse); ok {
		return res
	}
	return lightning_helpers.DisconnectPeerResponse{}
}

func WalletBalance(request lightning_helpers.WalletBalanceRequest) lightning_helpers.WalletBalanceResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.WalletBalanceResponse); ok {
		return res
	}
	return lightning_helpers.WalletBalanceResponse{}
}

func ListPeers(request lightning_helpers.ListPeersRequest) lightning_helpers.ListPeersResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.ListPeersResponse); ok {
		return res
	}
	return lightning_helpers.ListPeersResponse{}
}

func NewAddress(request lightning_helpers.NewAddressRequest) lightning_helpers.NewAddressResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.NewAddressResponse); ok {
		return res
	}
	return lightning_helpers.NewAddressResponse{}
}

func OpenChannel(request lightning_helpers.OpenChannelRequest) lightning_helpers.OpenChannelResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 300, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.OpenChannelResponse); ok {
		return res
	}
	return lightning_helpers.OpenChannelResponse{}
}

func CloseChannel(request lightning_helpers.CloseChannelRequest) lightning_helpers.CloseChannelResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 300, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.CloseChannelResponse); ok {
		return res
	}
	return lightning_helpers.CloseChannelResponse{}
}

const concurrentWorkLimit = 10

var serviceSequential = lightningService{limit: make(chan struct{}, 1)}                   //nolint:gochecknoglobals
var serviceConcurrent = lightningService{limit: make(chan struct{}, concurrentWorkLimit)} //nolint:gochecknoglobals

func processSequential(ctx context.Context, timeoutInSeconds int, req any, responseChan chan any) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)

	select {
	case <-ctx.Done():
		cancel()
		return
	case serviceSequential.limit <- struct{}{}:
	}

	go processRequestSequential(ctx, cancel, req, responseChan)
}

func processRequestSequential(ctx context.Context, cancel context.CancelFunc, req any, responseChan chan<- any) {
	defer func() {
		cancel()
		<-serviceSequential.limit
	}()

	select {
	case <-ctx.Done():
		responseChan <- nil
		return
	default:
	}

	processRequestByType(ctx, req, responseChan)
}

func processConcurrent(ctx context.Context, timeoutInSeconds int, req any, responseChan chan any) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)

	select {
	case <-ctx.Done():
		cancel()
		return
	case serviceConcurrent.limit <- struct{}{}:
	}

	go processRequestConcurrent(ctx, cancel, req, responseChan)
}

func processRequestConcurrent(ctx context.Context, cancel context.CancelFunc, req any, responseChan chan<- any) {
	defer func() {
		cancel()
		<-serviceConcurrent.limit
	}()

	select {
	case <-ctx.Done():
		responseChan <- nil
		return
	default:
	}

	processRequestByType(ctx, req, responseChan)
}

func processRequestByType(ctx context.Context, req any, responseChan chan<- any) {
	switch r := req.(type) {
	case lightning_helpers.InformationRequest:
		responseChan <- processGetInfoRequest(ctx, r)
		return
	case lightning_helpers.SignMessageRequest:
		responseChan <- processSignMessageRequest(ctx, r)
		return
	case lightning_helpers.SignatureVerificationRequest:
		responseChan <- processSignatureVerificationRequest(ctx, r)
		return
	case lightning_helpers.RoutingPolicyUpdateRequest:
		responseChan <- processRoutingPolicyUpdateRequest(ctx, r)
		return
	case lightning_helpers.ConnectPeerRequest:
		responseChan <- processConnectPeerRequest(ctx, r)
		return
	case lightning_helpers.DisconnectPeerRequest:
		responseChan <- processDisconnectPeerRequest(ctx, r)
		return
	case lightning_helpers.WalletBalanceRequest:
		responseChan <- processWalletBalanceRequest(ctx, r)
		return
	case lightning_helpers.ListPeersRequest:
		responseChan <- processListPeersRequest(ctx, r)
		return
	case lightning_helpers.NewAddressRequest:
		responseChan <- processNewAddressRequest(ctx, r)
		return
	case lightning_helpers.OpenChannelRequest:
		responseChan <- processOpenChannelRequest(ctx, r)
		return
	case lightning_helpers.CloseChannelRequest:
		responseChan <- processCloseChannelRequest(ctx, r)
		return
	}

	responseChan <- nil
}

func processGetInfoRequest(ctx context.Context,
	request lightning_helpers.InformationRequest) lightning_helpers.InformationResponse {

	response := lightning_helpers.InformationResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	infoRequest := cln.GetinfoRequest{}
	info, err := cln.NewNodeClient(connection).Getinfo(ctx, &infoRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Implementation = core.CLN
	response.Status = lightning_helpers.Active
	response.PublicKey = hex.EncodeToString(info.Id)
	response.Version = info.Version
	response.Alias = info.Alias
	response.Color = hex.EncodeToString(info.Color)
	response.PendingChannelCount = int(info.NumPendingChannels)
	response.ActiveChannelCount = int(info.NumActiveChannels)
	response.InactiveChannelCount = int(info.NumInactiveChannels)
	response.PeerCount = int(info.NumPeers)
	response.BlockHeight = info.Blockheight
	response.ChainSynced = info.WarningBitcoindSync == nil || *info.WarningBitcoindSync == ""
	response.GraphSynced = info.WarningLightningdSync == nil || *info.WarningLightningdSync == ""
	//response.Addresses = info.Address
	//response.Network = info.Network
	return response
}

func processSignMessageRequest(ctx context.Context,
	request lightning_helpers.SignMessageRequest) lightning_helpers.SignMessageResponse {

	response := lightning_helpers.SignMessageResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	signMsgReq := cln.SignmessageRequest{
		Message: request.Message,
	}
	signMsgResp, err := cln.NewNodeClient(connection).SignMessage(ctx, &signMsgReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Status = lightning_helpers.Active
	response.Signature = signMsgResp.Zbase
	return response
}

func processSignatureVerificationRequest(ctx context.Context,
	request lightning_helpers.SignatureVerificationRequest) lightning_helpers.SignatureVerificationResponse {

	response := lightning_helpers.SignatureVerificationResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	verifyMsgReq := cln.CheckmessageRequest{
		Message: request.Message,
		Zbase:   request.Signature,
	}
	verifyMsgResp, err := cln.NewNodeClient(connection).CheckMessage(ctx, &verifyMsgReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	if !verifyMsgResp.Verified {
		response.Message = "Signature is not valid"
		return response
	}

	response.Status = lightning_helpers.Active
	response.PublicKey = hex.EncodeToString(verifyMsgResp.Pubkey)
	response.Valid = verifyMsgResp.GetVerified()
	return response
}

func processRoutingPolicyUpdateRequest(ctx context.Context,
	request lightning_helpers.RoutingPolicyUpdateRequest) lightning_helpers.RoutingPolicyUpdateResponse {

	response := validateRoutingPolicyUpdateRequest(request)
	if response != nil {
		return *response
	}

	channelState := cache.GetChannelState(request.NodeId, request.ChannelId, true)
	if channelState == nil {
		return lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
			},
			Request: request,
		}
	}
	if !routingPolicyUpdateRequestContainsUpdates(request, channelState) {
		return lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Active,
			},
			Request: request,
		}
	}

	response = routingPolicyUpdateRequestIsRepeated(request)
	if response != nil {
		return *response
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return *response
	}

	resp, err := cln.NewNodeClient(connection).SetChannel(ctx, constructPolicyUpdateRequest(request, channelState))

	// TODO FIXME TIMELOCK CANNOT BE SET VIA SetChannel

	return processRoutingPolicyUpdateResponse(request, resp, err)
}

func processRoutingPolicyUpdateResponse(request lightning_helpers.RoutingPolicyUpdateRequest,
	resp *cln.SetchannelResponse,
	err error) lightning_helpers.RoutingPolicyUpdateResponse {

	if err != nil && resp == nil {
		log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v",
			request.ChannelId, request.NodeId)
		return lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
			},
			Request: request,
		}
	}
	var failedUpdateArray []lightning_helpers.FailedRequest
	for _, failedUpdate := range resp.Channels {
		if failedUpdate.WarningHtlcmaxTooHigh != nil {
			log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (cln-grpc error: %v)",
				request.ChannelId, request.NodeId, *failedUpdate.WarningHtlcmaxTooHigh)
			failedUpdateArray = append(failedUpdateArray, lightning_helpers.FailedRequest{
				Reason: *failedUpdate.WarningHtlcmaxTooHigh,
				Error:  *failedUpdate.WarningHtlcmaxTooHigh,
			})
		}
		if failedUpdate.WarningHtlcminTooLow != nil {
			log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (cln-grpc error: %v)",
				request.ChannelId, request.NodeId, *failedUpdate.WarningHtlcminTooLow)
			failedUpdateArray = append(failedUpdateArray, lightning_helpers.FailedRequest{
				Reason: *failedUpdate.WarningHtlcminTooLow,
				Error:  *failedUpdate.WarningHtlcminTooLow,
			})
		}
	}
	if err != nil || len(failedUpdateArray) != 0 {
		log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v",
			request.ChannelId, request.NodeId)
		return lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
			},
			Request:       request,
			FailedUpdates: failedUpdateArray,
		}
	}
	return lightning_helpers.RoutingPolicyUpdateResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Active,
		},
		Request: request,
	}
}

func constructPolicyUpdateRequest(request lightning_helpers.RoutingPolicyUpdateRequest,
	channelState *cache.ChannelStateSettingsCache) *cln.SetchannelRequest {

	policyUpdateRequest := &cln.SetchannelRequest{}

	var feePpm uint32
	if request.FeeRateMilliMsat == nil {
		feePpm = uint32(channelState.LocalFeeRateMilliMsat)
	} else {
		feePpm = uint32(*request.FeeRateMilliMsat)
	}
	policyUpdateRequest.Feeppm = &feePpm

	var feeBase uint64
	if request.FeeBaseMsat == nil {
		feeBase = uint64(channelState.LocalFeeBaseMsat)
	} else {
		feeBase = uint64(*request.FeeBaseMsat)
	}
	policyUpdateRequest.Feebase = &cln.Amount{Msat: feeBase}

	var minHtlcMsat uint64
	if request.MinHtlcMsat == nil {
		minHtlcMsat = channelState.LocalMinHtlcMsat
	} else {
		minHtlcMsat = *request.MinHtlcMsat
	}
	policyUpdateRequest.Htlcmin = &cln.Amount{Msat: minHtlcMsat}

	var maxHtlcMsat uint64
	if request.MaxHtlcMsat == nil {
		maxHtlcMsat = channelState.LocalMaxHtlcMsat
	} else {
		maxHtlcMsat = *request.MaxHtlcMsat
	}
	policyUpdateRequest.Htlcmax = &cln.Amount{Msat: maxHtlcMsat}

	channelSettings := cache.GetChannelSettingByChannelId(request.ChannelId)
	policyUpdateRequest.Id = *channelSettings.ShortChannelId
	return policyUpdateRequest
}

func validateRoutingPolicyUpdateRequest(
	request lightning_helpers.RoutingPolicyUpdateRequest) *lightning_helpers.RoutingPolicyUpdateResponse {

	if request.FeeRateMilliMsat == nil &&
		request.FeeBaseMsat == nil &&
		request.MaxHtlcMsat == nil &&
		request.MinHtlcMsat == nil &&
		request.TimeLockDelta == nil {
		return &lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status:  lightning_helpers.Active,
				Message: "Nothing changed so update is ignored",
			},
			Request: request,
		}
	}
	if request.ChannelId == 0 {
		return &lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  "ChannelId is 0",
			},
			Request: request,
		}
	}
	if request.TimeLockDelta != nil && *request.TimeLockDelta < 18 {
		return &lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  "TimeLockDelta is < 18",
			},
			Request: request,
		}
	}
	return nil
}

func routingPolicyUpdateRequestContainsUpdates(request lightning_helpers.RoutingPolicyUpdateRequest,
	channelState *cache.ChannelStateSettingsCache) bool {

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

func routingPolicyUpdateRequestIsRepeated(
	request lightning_helpers.RoutingPolicyUpdateRequest) *lightning_helpers.RoutingPolicyUpdateResponse {

	rateLimitSeconds := routingPolicyUpdateLimiterSeconds
	if request.RateLimitSeconds > 0 {
		rateLimitSeconds = request.RateLimitSeconds
	}
	channelEventsFromGraph, err := graph_events.GetChannelEventFromGraph(request.Db, request.ChannelId, &rateLimitSeconds)
	if err != nil {
		return &lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  err.Error(),
			},
			Request: request,
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

			return &lightning_helpers.RoutingPolicyUpdateResponse{
				CommunicationResponse: lightning_helpers.CommunicationResponse{
					Status: lightning_helpers.Inactive,
					Error: fmt.Sprintf("Routing policy update ignored due to rate limiter for channelId: %v",
						request.ChannelId),
				},
				Request: request,
			}
		}
	}
	return nil
}

func processConnectPeerRequest(ctx context.Context,
	request lightning_helpers.ConnectPeerRequest) lightning_helpers.ConnectPeerResponse {

	response := lightning_helpers.ConnectPeerResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request:                       request,
		RequestFailCurrentlyConnected: false,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	host := request.Host
	var port *uint32
	if strings.Contains(request.Host, ":") {
		host = request.Host[:strings.Index(request.Host, ":")]
		portInt, err := strconv.ParseUint(request.Host[strings.Index(request.Host, ":")+1:], 10, 64)
		if err == nil {
			p := uint32(portInt)
			port = &p
		}
	}
	connectPeerRequest := cln.ConnectRequest{
		Id:   request.PublicKey,
		Host: &host,
		Port: port,
	}

	_, err = cln.NewNodeClient(connection).ConnectPeer(ctx, &connectPeerRequest)

	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Status = lightning_helpers.Active
	return response
}

func processDisconnectPeerRequest(ctx context.Context,
	request lightning_helpers.DisconnectPeerRequest) lightning_helpers.DisconnectPeerResponse {

	response := lightning_helpers.DisconnectPeerResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request:                            request,
		RequestFailedCurrentlyDisconnected: false,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	publicKey, err := hex.DecodeString(cache.GetNodeSettingsByNodeId(request.PeerNodeId).PublicKey)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	force := true
	disconnectPeerRequest := cln.DisconnectRequest{
		Id:    publicKey,
		Force: &force,
	}

	_, err = cln.NewNodeClient(connection).Disconnect(ctx, &disconnectPeerRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Status = lightning_helpers.Active
	return response
}

func processWalletBalanceRequest(ctx context.Context,
	request lightning_helpers.WalletBalanceRequest) lightning_helpers.WalletBalanceResponse {

	response := lightning_helpers.WalletBalanceResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	funds, err := cln.NewNodeClient(connection).ListFunds(ctx, &cln.ListfundsRequest{})
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Status = lightning_helpers.Active
	var unconfirmedBalance int64
	var confirmedBalance int64
	var lockedBalance int64
	for _, utxo := range funds.Outputs {
		if utxo.AmountMsat != nil {
			switch utxo.Status {
			case cln.ListfundsOutputs_UNCONFIRMED, cln.ListfundsOutputs_IMMATURE:
				unconfirmedBalance += int64(utxo.AmountMsat.Msat) / 1_000
			case cln.ListfundsOutputs_CONFIRMED:
				confirmedBalance += int64(utxo.AmountMsat.Msat) / 1_000
			case cln.ListfundsOutputs_SPENT:
			}
			if utxo.Reserved {
				lockedBalance += int64(utxo.AmountMsat.Msat) / 1_000
			}
		}
	}
	response.ReservedBalanceAnchorChan = 0
	response.UnconfirmedBalance = unconfirmedBalance
	response.ConfirmedBalance = confirmedBalance
	response.TotalBalance = confirmedBalance + unconfirmedBalance
	response.LockedBalance = lockedBalance

	return response
}

func processListPeersRequest(ctx context.Context,
	request lightning_helpers.ListPeersRequest) lightning_helpers.ListPeersResponse {

	response := lightning_helpers.ListPeersResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	rsp, err := cln.NewNodeClient(connection).ListPeers(ctx, &cln.ListpeersRequest{})

	if err != nil {
		response.Error = err.Error()
		return response
	}

	peers := make(map[string]lightning_helpers.Peer)
	for _, peer := range rsp.Peers {
		if peer != nil && peer.Connected {
			peers[hex.EncodeToString(peer.Id)] = lightning_helpers.GetPeerCLN(peer)
		}
	}

	response.Status = lightning_helpers.Active
	response.Peers = peers

	return response
}

func processNewAddressRequest(ctx context.Context,
	request lightning_helpers.NewAddressRequest) lightning_helpers.NewAddressResponse {

	response := lightning_helpers.NewAddressResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	// TODO FIXME CLN implementation is temporary
	clnAddressRequest := &cln.NewaddrRequest{}
	segwit := cln.NewaddrRequest_P2SH_SEGWIT
	bech32 := cln.NewaddrRequest_BECH32
	switch request.Type {
	case lightning_helpers.P2WPKH:
		clnAddressRequest.Addresstype = &segwit
	case lightning_helpers.P2WKH:
		clnAddressRequest.Addresstype = &segwit
	case lightning_helpers.NP2WKH:
		clnAddressRequest.Addresstype = &segwit
	case lightning_helpers.P2TR:
		clnAddressRequest.Addresstype = &bech32
	default:
		response.Error = "unknown address type"
		return response
	}

	rsp, err := cln.NewNodeClient(connection).NewAddr(ctx, clnAddressRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	response.Status = lightning_helpers.Active
	if rsp.P2ShSegwit != nil {
		response.Address = *rsp.P2ShSegwit
	}
	if response.Address == "" && rsp.Bech32 != nil {
		response.Address = *rsp.Bech32
	}
	return response
}

func processOpenChannelRequest(ctx context.Context,
	request lightning_helpers.OpenChannelRequest) lightning_helpers.OpenChannelResponse {

	response := lightning_helpers.OpenChannelResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	openChanReq, err := prepareOpenRequest(request)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	channel, err := cln.NewNodeClient(connection).FundChannel(ctx, openChanReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.ChannelPoint = fmt.Sprintf("%v:%v", hex.EncodeToString(channel.Txid), channel.Outnum)
	response.ChannelStatus = core.Opening
	response.Request = request
	response.FundingTransactionHash = hex.EncodeToString(channel.Txid)
	response.FundingOutputIndex = channel.Outnum
	response.Status = lightning_helpers.Active
	return response
}

func prepareOpenRequest(request lightning_helpers.OpenChannelRequest) (*cln.FundchannelRequest, error) {
	if request.NodeId == 0 {
		return nil, errors.New("nodeId is missing")
	}

	if request.SatPerVbyte != nil && request.TargetConf != nil {
		return nil, errors.New("Cannot set both SatPerVbyte and TargetConf")
	}

	pubKeyHex, err := hex.DecodeString(request.NodePubKey)
	if err != nil {
		return nil, errors.New("error decoding public key hex")
	}

	//open channel request
	openChanReq := &cln.FundchannelRequest{
		Id: pubKeyHex,

		// This is the amount we are putting into the channel (channel size)
		Amount: &cln.AmountOrAll{Value: &cln.AmountOrAll_Amount{Amount: &cln.Amount{Msat: uint64(request.LocalFundingAmount * 1_000)},
		}},
	}

	// The amount to give the other node in the opening process.
	// NB: This means you will give the other node this amount of sats
	if request.PushSat != nil {
		openChanReq.PushMsat = &cln.Amount{Msat: uint64((*request.PushSat) * 1_000)}
	}

	if request.SatPerVbyte != nil {
		// TODO FIXME CLN verify
		openChanReq.Feerate = &cln.Feerate{Style: &cln.Feerate_Perkb{Perkb: uint32(*request.SatPerVbyte)}}
	}

	if request.TargetConf != nil {
		minDept := uint32(*request.TargetConf)
		openChanReq.Mindepth = &minDept
	}

	if request.Private != nil {
		announced := !*request.Private
		openChanReq.Announce = &announced
	}

	// TODO FIXME CLN verify
	//if request.MinHtlcMsat != nil {
	//	openChanReq.MinHtlcMsat = int64(*request.MinHtlcMsat)
	//}

	// TODO FIXME CLN verify
	//if request.RemoteCsvDelay != nil {
	//	openChanReq.RemoteCsvDelay = *request.RemoteCsvDelay
	//}

	if request.MinConfs != nil {
		minConf := uint32(*request.MinConfs)
		openChanReq.Minconf = &minConf
	}

	// TODO FIXME CLN verify
	//if request.SpendUnconfirmed != nil {
	//	openChanReq.SpendUnconfirmed = *request.SpendUnconfirmed
	//}

	if request.CloseAddress != nil {
		openChanReq.CloseTo = request.CloseAddress
	}
	return openChanReq, nil
}

func processCloseChannelRequest(ctx context.Context,
	request lightning_helpers.CloseChannelRequest) lightning_helpers.CloseChannelResponse {

	response := lightning_helpers.CloseChannelResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	closeChanReq, err := prepareCloseRequest(request)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	channel, err := cln.NewNodeClient(connection).Close(ctx, closeChanReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.ChannelStatus = core.Closing
	response.Request = request
	response.ClosingTransactionHash = hex.EncodeToString(channel.Txid)
	response.Status = lightning_helpers.Active
	return response
}

func prepareCloseRequest(request lightning_helpers.CloseChannelRequest) (*cln.CloseRequest, error) {
	if request.NodeId == 0 {
		return nil, errors.New("nodeId is missing")
	}

	if request.SatPerVbyte != nil && request.TargetConf != nil {
		return nil, errors.New("Cannot set both SatPerVbyte and TargetConf")
	}

	channel := cache.GetChannelSettingByChannelId(request.ChannelId)
	peerNodeId := channel.FirstNodeId
	if peerNodeId == request.NodeId {
		peerNodeId = channel.SecondNodeId
	}
	peerNode := cache.GetNodeSettingsByNodeId(peerNodeId)

	//open channel request
	closeChanReq := &cln.CloseRequest{
		Id: peerNode.PublicKey,
	}

	if request.Force != nil {
		closeChanReq.ForceLeaseClosed = request.Force
	}

	// TODO FIXME CLN verify
	//if request.TargetConf != nil {
	//	closeChanReq.TargetConf = *ccReq.TargetConf
	//}

	if request.SatPerVbyte != nil {
		closeChanReq.Feerange = append(closeChanReq.Feerange, &cln.Feerate{Style: &cln.Feerate_Perkb{Perkb: uint32(*request.SatPerVbyte)}})
		closeChanReq.Feerange = append(closeChanReq.Feerange, &cln.Feerate{Style: &cln.Feerate_Perkb{Perkb: uint32(*request.SatPerVbyte)}})
	}

	// TODO FIXME CLN verify
	//if request.DeliveryAddress != nil {
	//	closeChanReq.DeliveryAddress = *ccReq.DeliveryAddress
	//}

	return closeChanReq, nil
}
