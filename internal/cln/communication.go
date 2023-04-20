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
	"github.com/lncapital/torq/internal/lightning_requests"
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
	request lightning_requests.InformationRequest) lightning_requests.InformationResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.InformationResponse); ok {
		return res
	}
	return lightning_requests.InformationResponse{}
}

func SignMessage(
	request lightning_requests.SignMessageRequest) lightning_requests.SignMessageResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.SignMessageResponse); ok {
		return res
	}
	return lightning_requests.SignMessageResponse{}
}

func SignatureVerification(
	request lightning_requests.SignatureVerificationRequest) lightning_requests.SignatureVerificationResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.SignatureVerificationResponse); ok {
		return res
	}
	return lightning_requests.SignatureVerificationResponse{}
}

func RoutingPolicyUpdate(
	request lightning_requests.RoutingPolicyUpdateRequest) lightning_requests.RoutingPolicyUpdateResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.RoutingPolicyUpdateResponse); ok {
		return res
	}
	return lightning_requests.RoutingPolicyUpdateResponse{}
}

func ConnectPeer(
	request lightning_requests.ConnectPeerRequest) lightning_requests.ConnectPeerResponse {
	responseChan := make(chan any)
	process(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.ConnectPeerResponse); ok {
		return res
	}
	return lightning_requests.ConnectPeerResponse{}
}

func DisconnectPeer(
	request lightning_requests.DisconnectPeerRequest) lightning_requests.DisconnectPeerResponse {
	responseChan := make(chan any)
	process(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.DisconnectPeerResponse); ok {
		return res
	}
	return lightning_requests.DisconnectPeerResponse{}
}

func WalletBalance(request lightning_requests.WalletBalanceRequest) lightning_requests.WalletBalanceResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.WalletBalanceResponse); ok {
		return res
	}
	return lightning_requests.WalletBalanceResponse{}
}

func ListPeers(request lightning_requests.ListPeersRequest) lightning_requests.ListPeersResponse {
	responseChan := make(chan any)
	process(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.ListPeersResponse); ok {
		return res
	}
	return lightning_requests.ListPeersResponse{}
}

func NewAddress(request lightning_requests.NewAddressRequest) lightning_requests.NewAddressResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.NewAddressResponse); ok {
		return res
	}
	return lightning_requests.NewAddressResponse{}
}

const concurrentWorkLimit = 1

var service = lightningService{limit: make(chan struct{}, concurrentWorkLimit)} //nolint:gochecknoglobals

func processRequest(ctx context.Context, cancel context.CancelFunc, req any, responseChan chan<- any) {

	defer func() {
		cancel()
		<-service.limit
	}()

	select {
	case <-ctx.Done():
		responseChan <- nil
		return
	default:
	}

	switch r := req.(type) {
	case lightning_requests.InformationRequest:
		responseChan <- processGetInfoRequest(ctx, r)
		return
	case lightning_requests.SignMessageRequest:
		responseChan <- processSignMessageRequest(ctx, r)
		return
	case lightning_requests.SignatureVerificationRequest:
		responseChan <- processSignatureVerificationRequest(ctx, r)
		return
	case lightning_requests.RoutingPolicyUpdateRequest:
		responseChan <- processRoutingPolicyUpdateRequest(ctx, r)
		return
	case lightning_requests.ConnectPeerRequest:
		responseChan <- processConnectPeerRequest(ctx, r)
		return
	case lightning_requests.DisconnectPeerRequest:
		responseChan <- processDisconnectPeerRequest(ctx, r)
		return
	case lightning_requests.WalletBalanceRequest:
		responseChan <- processWalletBalanceRequest(ctx, r)
		return
	case lightning_requests.ListPeersRequest:
		responseChan <- processListPeersRequest(ctx, r)
		return
	case lightning_requests.NewAddressRequest:
		responseChan <- processNewAddressRequest(ctx, r)
		return
	}

	responseChan <- nil
}

func process(ctx context.Context, timeoutInSeconds int, req any, responseChan chan any) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)

	select {
	case <-ctx.Done():
		cancel()
		return
	case service.limit <- struct{}{}:
	}

	go processRequest(ctx, cancel, req, responseChan)
}

func processGetInfoRequest(ctx context.Context,
	request lightning_requests.InformationRequest) lightning_requests.InformationResponse {

	response := lightning_requests.InformationResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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
	response.Status = lightning_requests.Active
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
	request lightning_requests.SignMessageRequest) lightning_requests.SignMessageResponse {

	response := lightning_requests.SignMessageResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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

	response.Status = lightning_requests.Active
	response.Signature = signMsgResp.Zbase
	return response
}

func processSignatureVerificationRequest(ctx context.Context,
	request lightning_requests.SignatureVerificationRequest) lightning_requests.SignatureVerificationResponse {

	response := lightning_requests.SignatureVerificationResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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

	response.Status = lightning_requests.Active
	response.PublicKey = hex.EncodeToString(verifyMsgResp.Pubkey)
	response.Valid = verifyMsgResp.GetVerified()
	return response
}

func processRoutingPolicyUpdateRequest(ctx context.Context,
	request lightning_requests.RoutingPolicyUpdateRequest) lightning_requests.RoutingPolicyUpdateResponse {

	response := validateRoutingPolicyUpdateRequest(request)
	if response != nil {
		return *response
	}

	channelState := cache.GetChannelState(request.NodeId, request.ChannelId, true)
	if channelState == nil {
		return lightning_requests.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_requests.CommunicationResponse{
				Status: lightning_requests.Inactive,
			},
			Request: request,
		}
	}
	if !routingPolicyUpdateRequestContainsUpdates(request, channelState) {
		return lightning_requests.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_requests.CommunicationResponse{
				Status: lightning_requests.Active,
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

func processRoutingPolicyUpdateResponse(request lightning_requests.RoutingPolicyUpdateRequest,
	resp *cln.SetchannelResponse,
	err error) lightning_requests.RoutingPolicyUpdateResponse {

	if err != nil && resp == nil {
		log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v",
			request.ChannelId, request.NodeId)
		return lightning_requests.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_requests.CommunicationResponse{
				Status: lightning_requests.Inactive,
			},
			Request: request,
		}
	}
	var failedUpdateArray []lightning_requests.FailedRequest
	for _, failedUpdate := range resp.Channels {
		if failedUpdate.WarningHtlcmaxTooHigh != nil {
			log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (cln-grpc error: %v)",
				request.ChannelId, request.NodeId, *failedUpdate.WarningHtlcmaxTooHigh)
			failedUpdateArray = append(failedUpdateArray, lightning_requests.FailedRequest{
				Reason: *failedUpdate.WarningHtlcmaxTooHigh,
				Error:  *failedUpdate.WarningHtlcmaxTooHigh,
			})
		}
		if failedUpdate.WarningHtlcminTooLow != nil {
			log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (cln-grpc error: %v)",
				request.ChannelId, request.NodeId, *failedUpdate.WarningHtlcminTooLow)
			failedUpdateArray = append(failedUpdateArray, lightning_requests.FailedRequest{
				Reason: *failedUpdate.WarningHtlcminTooLow,
				Error:  *failedUpdate.WarningHtlcminTooLow,
			})
		}
	}
	if err != nil || len(failedUpdateArray) != 0 {
		log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v",
			request.ChannelId, request.NodeId)
		return lightning_requests.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_requests.CommunicationResponse{
				Status: lightning_requests.Inactive,
			},
			Request:       request,
			FailedUpdates: failedUpdateArray,
		}
	}
	return lightning_requests.RoutingPolicyUpdateResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Active,
		},
		Request: request,
	}
}

func constructPolicyUpdateRequest(request lightning_requests.RoutingPolicyUpdateRequest,
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
	request lightning_requests.RoutingPolicyUpdateRequest) *lightning_requests.RoutingPolicyUpdateResponse {

	if request.FeeRateMilliMsat == nil &&
		request.FeeBaseMsat == nil &&
		request.MaxHtlcMsat == nil &&
		request.MinHtlcMsat == nil &&
		request.TimeLockDelta == nil {
		return &lightning_requests.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_requests.CommunicationResponse{
				Status:  lightning_requests.Active,
				Message: "Nothing changed so update is ignored",
			},
			Request: request,
		}
	}
	if request.ChannelId == 0 {
		return &lightning_requests.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_requests.CommunicationResponse{
				Status: lightning_requests.Inactive,
				Error:  "ChannelId is 0",
			},
			Request: request,
		}
	}
	if request.TimeLockDelta != nil && *request.TimeLockDelta < 18 {
		return &lightning_requests.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_requests.CommunicationResponse{
				Status: lightning_requests.Inactive,
				Error:  "TimeLockDelta is < 18",
			},
			Request: request,
		}
	}
	return nil
}

func routingPolicyUpdateRequestContainsUpdates(request lightning_requests.RoutingPolicyUpdateRequest,
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
	request lightning_requests.RoutingPolicyUpdateRequest) *lightning_requests.RoutingPolicyUpdateResponse {

	rateLimitSeconds := routingPolicyUpdateLimiterSeconds
	if request.RateLimitSeconds > 0 {
		rateLimitSeconds = request.RateLimitSeconds
	}
	channelEventsFromGraph, err := graph_events.GetChannelEventFromGraph(request.Db, request.ChannelId, &rateLimitSeconds)
	if err != nil {
		return &lightning_requests.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_requests.CommunicationResponse{
				Status: lightning_requests.Inactive,
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

			return &lightning_requests.RoutingPolicyUpdateResponse{
				CommunicationResponse: lightning_requests.CommunicationResponse{
					Status: lightning_requests.Inactive,
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
	request lightning_requests.ConnectPeerRequest) lightning_requests.ConnectPeerResponse {

	response := lightning_requests.ConnectPeerResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
		},
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

	response.Status = lightning_requests.Active
	return response
}

func processDisconnectPeerRequest(ctx context.Context,
	request lightning_requests.DisconnectPeerRequest) lightning_requests.DisconnectPeerResponse {

	response := lightning_requests.DisconnectPeerResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
		},
		RequestFailedCurrentlyDisconnected: false,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	publicKey, err := hex.DecodeString(cache.GetNodeSettingsByNodeId(request.PeerNodeId).PublicKey)
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

	response.Status = lightning_requests.Active
	return response
}

func processWalletBalanceRequest(ctx context.Context,
	request lightning_requests.WalletBalanceRequest) lightning_requests.WalletBalanceResponse {

	response := lightning_requests.WalletBalanceResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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

	response.Status = lightning_requests.Active
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
	request lightning_requests.ListPeersRequest) lightning_requests.ListPeersResponse {

	response := lightning_requests.ListPeersResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
		},
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

	peers := make(map[string]lightning_requests.Peer)
	for _, peer := range rsp.Peers {
		if peer != nil && peer.Connected {
			peers[hex.EncodeToString(peer.Id)] = lightning_requests.GetPeerCLN(peer)
		}
	}

	response.Status = lightning_requests.Active
	response.Peers = peers

	return response
}

func processNewAddressRequest(ctx context.Context,
	request lightning_requests.NewAddressRequest) lightning_requests.NewAddressResponse {

	response := lightning_requests.NewAddressResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
		},
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
	case lightning_requests.P2WPKH:
		clnAddressRequest.Addresstype = &segwit
	case lightning_requests.P2WKH:
		clnAddressRequest.Addresstype = &segwit
	case lightning_requests.NP2WKH:
		clnAddressRequest.Addresstype = &segwit
	case lightning_requests.P2TR:
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
	response.Status = lightning_requests.Active
	if rsp.P2ShSegwit != nil {
		response.Address = *rsp.P2ShSegwit
	}
	if response.Address == "" && rsp.Bech32 != nil {
		response.Address = *rsp.Bech32
	}
	return response
}
