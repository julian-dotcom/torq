package lnd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/lightning_requests"
	"github.com/lncapital/torq/proto/lnrpc/routerrpc"

	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/graph_events"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

const routingPolicyUpdateLimiterSeconds = 5 * 60

// 70 because a reconnection is attempted every 60 seconds
const avoidChannelAndPolicyImportRerunTimeSeconds = 70

var (
	connectionWrapperOnce sync.Once           //nolint:gochecknoglobals
	connectionWrapper     *connectionsWrapper //nolint:gochecknoglobals
)

type connectionsWrapper struct {
	mu            sync.Mutex
	connections   map[int]*grpc.ClientConn
	grpcAddresses map[int]string
	tlsBytes      map[int][]byte
	macaroonBytes map[int][]byte
}

func getConnection(nodeId int) (*grpc.ClientConn, error) {
	connectionWrapperOnce.Do(func() {
		log.Debug().Msg("Loading Connection Wrapper.")
		connectionWrapper = &connectionsWrapper{
			mu:            sync.Mutex{},
			connections:   make(map[int]*grpc.ClientConn),
			grpcAddresses: make(map[int]string),
			tlsBytes:      make(map[int][]byte),
			macaroonBytes: make(map[int][]byte),
		}
	})

	connectionWrapper.mu.Lock()
	defer connectionWrapper.mu.Unlock()

	ncd := cache.GetNodeConnectionDetails(nodeId)

	existingConnection, exists := connectionWrapper.connections[nodeId]
	if !exists ||
		connectionWrapper.grpcAddresses[nodeId] != ncd.GRPCAddress ||
		!bytes.Equal(connectionWrapper.tlsBytes[nodeId], ncd.TLSFileBytes) ||
		!bytes.Equal(connectionWrapper.macaroonBytes[nodeId], ncd.MacaroonFileBytes) {

		conn, err := lnd_connect.Connect(ncd.GRPCAddress, ncd.TLSFileBytes, ncd.MacaroonFileBytes)
		if err != nil {
			log.Error().Err(err).Msgf("GRPC connection Failed for node id: %v", nodeId)
			return nil, errors.Wrapf(err, "Connecting to GRPC.")
		}
		connectionWrapper.connections[nodeId] = conn
		connectionWrapper.grpcAddresses[nodeId] = ncd.GRPCAddress
		connectionWrapper.tlsBytes[nodeId] = ncd.TLSFileBytes
		connectionWrapper.macaroonBytes[nodeId] = ncd.MacaroonFileBytes
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

func Information(request lightning_requests.InformationRequest) lightning_requests.InformationResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.InformationResponse); ok {
		return res
	}
	return lightning_requests.InformationResponse{}
}

func SignMessage(request lightning_requests.SignMessageRequest) lightning_requests.SignMessageResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.SignMessageResponse); ok {
		return res
	}
	return lightning_requests.SignMessageResponse{}
}

func SignatureVerification(request lightning_requests.SignatureVerificationRequest) lightning_requests.SignatureVerificationResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.SignatureVerificationResponse); ok {
		return res
	}
	return lightning_requests.SignatureVerificationResponse{}
}

func RoutingPolicyUpdate(request lightning_requests.RoutingPolicyUpdateRequest) lightning_requests.RoutingPolicyUpdateResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.RoutingPolicyUpdateResponse); ok {
		return res
	}
	return lightning_requests.RoutingPolicyUpdateResponse{}
}

func ConnectPeer(request lightning_requests.ConnectPeerRequest) lightning_requests.ConnectPeerResponse {
	responseChan := make(chan any)
	process(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_requests.ConnectPeerResponse); ok {
		return res
	}
	return lightning_requests.ConnectPeerResponse{}
}

func DisconnectPeer(request lightning_requests.DisconnectPeerRequest) lightning_requests.DisconnectPeerResponse {
	responseChan := make(chan any)
	process(context.Background(), disconnectPeerTimeoutInSeconds, request, responseChan)
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

func ChannelStatusUpdate(request ChannelStatusUpdateRequest) ChannelStatusUpdateResponse {
	responseChan := make(chan any)
	process(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ChannelStatusUpdateResponse); ok {
		return res
	}
	return ChannelStatusUpdateResponse{}
}

func ImportAllChannels(request ImportAllChannelsRequest) ImportAllChannelsResponse {
	responseChan := make(chan any)
	process(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportAllChannelsResponse); ok {
		return res
	}
	return ImportAllChannelsResponse{}
}

func ImportPendingChannels(request ImportPendingChannelsRequest) ImportPendingChannelsResponse {
	responseChan := make(chan any)
	process(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportPendingChannelsResponse); ok {
		return res
	}
	return ImportPendingChannelsResponse{}
}

func ImportChannelRoutingPolicies(request ImportChannelRoutingPoliciesRequest) ImportChannelRoutingPoliciesResponse {
	responseChan := make(chan any)
	process(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportChannelRoutingPoliciesResponse); ok {
		return res
	}
	return ImportChannelRoutingPoliciesResponse{}
}

func ImportNodeInformation(request ImportNodeInformationRequest) ImportNodeInformationResponse {
	responseChan := make(chan any)
	process(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportNodeInformationResponse); ok {
		return res
	}
	return ImportNodeInformationResponse{}
}

func ImportPeerStatus(request ImportPeerStatusRequest) ImportPeerStatusResponse {
	responseChan := make(chan any)
	process(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportPeerStatusResponse); ok {
		return res
	}
	return ImportPeerStatusResponse{}
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
	case ChannelStatusUpdateRequest:
		responseChan <- processChannelStatusUpdateRequest(ctx, r)
		return
	case ImportAllChannelsRequest:
		responseChan <- processImportAllChannelsRequest(ctx, r)
		return
	case ImportPendingChannelsRequest:
		responseChan <- processImportPendingChannelsRequest(ctx, r)
		return
	case ImportChannelRoutingPoliciesRequest:
		responseChan <- processImportChannelRoutingPoliciesRequest(ctx, r)
		return
	case ImportNodeInformationRequest:
		responseChan <- processImportNodeInformationRequest(ctx, r)
		return
	case ImportPeerStatusRequest:
		responseChan <- processImportPeerStatusRequest(ctx, r)
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

type ResponseStatus int

const (
	Inactive = ResponseStatus(core.Inactive)
	Active   = ResponseStatus(core.Active)
)

type CommunicationRequest struct {
	NodeId int
}

type CommunicationResponse struct {
	Status  ResponseStatus
	Message string
	Error   string
}

type ChannelStatusUpdateRequest struct {
	CommunicationRequest
	Db            *sqlx.DB
	ChannelId     int
	ChannelStatus core.Status
}

type ChannelStatusUpdateResponse struct {
	CommunicationResponse
	Request ChannelStatusUpdateRequest
}

type FailedRequest struct {
	Reason string
	Error  string
}

type ImportRequest struct {
	CommunicationRequest
	Db    *sqlx.DB
	Force bool
}

type ImportResponse struct {
	CommunicationResponse
	Error error
}

type ImportAllChannelsRequest struct {
	ImportRequest
}

type ImportAllChannelsResponse struct {
	Request ImportAllChannelsRequest
	ImportResponse
}

type ImportPendingChannelsRequest struct {
	ImportRequest
}

type ImportPendingChannelsResponse struct {
	Request ImportPendingChannelsRequest
	ImportResponse
}

type ImportChannelRoutingPoliciesRequest struct {
	ImportRequest
}

type ImportChannelRoutingPoliciesResponse struct {
	Request ImportChannelRoutingPoliciesRequest
	ImportResponse
}

type ImportNodeInformationRequest struct {
	ImportRequest
}

type ImportNodeInformationResponse struct {
	Request ImportNodeInformationRequest
	ImportResponse
}

type ImportPeerStatusRequest struct {
	ImportRequest
}

type ImportPeerStatusResponse struct {
	Request ImportPeerStatusRequest
	ImportResponse
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

	listPeersRequest := lnrpc.ListPeersRequest{LatestError: request.LatestError}
	client := lnrpc.NewLightningClient(connection)
	rsp, err := client.ListPeers(ctx, &listPeersRequest)

	if err != nil {
		response.Error = err.Error()
		return response
	}

	peers := make(map[string]lightning_requests.Peer)
	for _, peer := range rsp.Peers {
		peers[peer.PubKey] = lightning_requests.GetPeerLND(peer)
	}

	response.Status = lightning_requests.Active
	response.Peers = peers

	return response
}

const disconnectPeerTimeoutInSeconds = 10
const disconnectPeerAttemptDelayInSeconds = 1
const maximumAttempts = 5

func processDisconnectPeerRequest(ctx context.Context,
	request lightning_requests.DisconnectPeerRequest) lightning_requests.DisconnectPeerResponse {

	response := lightning_requests.DisconnectPeerResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
		},
		RequestFailedCurrentlyDisconnected: false,
	}

	publicKey := cache.GetNodeSettingsByNodeId(request.PeerNodeId).PublicKey
	disconnectPeerRequest := lnrpc.DisconnectPeerRequest{
		PubKey: publicKey,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	iterationCounter := 0
	client := lnrpc.NewLightningClient(connection)
	for {
		_, err = client.DisconnectPeer(ctx, &disconnectPeerRequest)
		if err != nil {
			if strings.Contains(err.Error(), "not connected") {
				if iterationCounter == 0 {
					response.RequestFailedCurrentlyDisconnected = true
					return response
				}
				response.Status = lightning_requests.Active
				return response
			}
			log.Debug().Err(err).Msgf(
				"LND peer disconnection request failed for unknown reason but we ignore this and try again.")
			if !core.Sleep(ctx, disconnectPeerAttemptDelayInSeconds*time.Second) {
				break
			}
			continue
		}
		// Increment when LND request was successful
		iterationCounter++

		peer, err := getPeerByPublicKeyDelayed(ctx, client, publicKey)
		if err != nil {
			response.Error = err.Error()
			return response
		}
		if peer == nil {
			response.Status = lightning_requests.Active
			return response
		}
		if iterationCounter == maximumAttempts {
			break
		}
	}
	response.Error = "Disconnection unsuccessful"
	return response
}

func getPeerByPublicKeyDelayed(ctx context.Context,
	client lnrpc.LightningClient,
	publicKey string) (*lnrpc.Peer, error) {

	//ticker is needed to give lnd time to disconnect from the peer
	ticker := time.NewTicker(disconnectPeerAttemptDelayInSeconds * time.Second)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return nil, errors.New("Context ended")
	case <-ticker.C:
		peer, err := getPeerByPublicKey(ctx, client, publicKey)
		if err != nil {
			return nil, err
		}
		return peer, nil
	}
}

func processConnectPeerRequest(ctx context.Context,
	request lightning_requests.ConnectPeerRequest) lightning_requests.ConnectPeerResponse {

	response := lightning_requests.ConnectPeerResponse{
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
		},
		RequestFailCurrentlyConnected: false,
	}

	connectPeerRequest := lnrpc.ConnectPeerRequest{
		Addr: &lnrpc.LightningAddress{Pubkey: request.PublicKey, Host: request.Host},
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	client := lnrpc.NewLightningClient(connection)
	_, err = client.ConnectPeer(ctx, &connectPeerRequest)

	if err != nil {
		if strings.Contains(err.Error(), "already connected to peer") {
			response.RequestFailCurrentlyConnected = true
			return response
		}
		response.Error = err.Error()
		return response
	}

	//ticker is needed to give lnd time to connect to the peer
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		response.Error = "Context ended"
		return response
	case <-ticker.C:
		// call lnd again to see if the peer is still connected
		peer, err := getPeerByPublicKey(ctx, client, request.PublicKey)
		if err != nil {
			response.Error = err.Error()
			return response
		}

		if peer == nil {
			response.Error = "Connection unsuccessful"
			return response
		}
		response.Status = lightning_requests.Active
		return response
	}
}

// It seems that the errors are not returned in the Connect/Disconnect response, so we need to check the peers list
func getPeerByPublicKey(ctx context.Context, client lnrpc.LightningClient, publicKey string) (*lnrpc.Peer, error) {
	listPeersRequest := lnrpc.ListPeersRequest{}
	peersResponse, err := client.ListPeers(ctx, &listPeersRequest)
	if err != nil {
		return nil, errors.Wrap(err, "Obtaining list of peers.")
	}

	var peer *lnrpc.Peer
	for _, p := range peersResponse.Peers {
		if p.PubKey == publicKey {
			peer = p
			break
		}
	}
	return peer, nil
}

func processImportAllChannelsRequest(ctx context.Context, request ImportAllChannelsRequest) ImportAllChannelsResponse {
	importType := core.ImportAllChannels
	nodeSettings := cache.GetNodeSettingsByNodeId(request.NodeId)

	response := ImportAllChannelsResponse{
		Request: request,
		ImportResponse: ImportResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Inactive,
			},
		},
	}

	successTimes, skip := getSuccessTimes(request.NodeId, request.Force, importType)
	if skip {
		return response
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err
		return response
	}

	//Import Pending channels
	err = ImportPendingChannelsFromLnd(ctx, request.Db, lnrpc.NewLightningClient(connection), nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import pending channels.")
		response.Error = err
		return response
	}

	//Import Open channels
	err = ImportOpenChannels(ctx, request.Db, lnrpc.NewLightningClient(connection), nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import open channels.")
		response.Error = err
		return response
	}

	// Import Closed channels
	err = ImportClosedChannels(ctx, request.Db, lnrpc.NewLightningClient(connection), nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import closed channels.")
		response.Error = err
		return response
	}

	err = settings.InitializeChannelsCache(request.Db)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to Initialize ChannelsCacheHandler.")
		response.Error = err
		return response
	}

	setSuccessTime(request.NodeId, successTimes, importType)
	response.Status = Active
	return response
}

func processImportPendingChannelsRequest(ctx context.Context,
	request ImportPendingChannelsRequest) ImportPendingChannelsResponse {

	importType := core.ImportPendingChannels
	nodeSettings := cache.GetNodeSettingsByNodeId(request.NodeId)

	response := ImportPendingChannelsResponse{
		Request: request,
		ImportResponse: ImportResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Inactive,
			},
		},
	}

	successTimes, skip := getSuccessTimes(request.NodeId, request.Force, importType)
	if skip {
		return response
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err
		return response
	}

	err = ImportPendingChannelsFromLnd(ctx, request.Db, lnrpc.NewLightningClient(connection), nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import pending channels.")
		response.Error = err
		return response
	}

	err = settings.InitializeChannelsCache(request.Db)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to Initialize ChannelsCacheHandler.")
		response.Error = err
		return response
	}

	setSuccessTime(request.NodeId, successTimes, importType)
	response.Status = Active
	return response
}

func processImportChannelRoutingPoliciesRequest(ctx context.Context,
	request ImportChannelRoutingPoliciesRequest) ImportChannelRoutingPoliciesResponse {

	importType := core.ImportChannelRoutingPolicies
	nodeSettings := cache.GetNodeSettingsByNodeId(request.NodeId)

	response := ImportChannelRoutingPoliciesResponse{
		Request: request,
		ImportResponse: ImportResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Inactive,
			},
		},
	}

	successTimes, skip := getSuccessTimes(request.NodeId, request.Force, importType)
	if skip {
		return response
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err
		return response
	}

	err = ImportRoutingPolicies(ctx, lnrpc.NewLightningClient(connection), request.Db, nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import routing policies.")
		response.Error = err
		return response
	}
	setSuccessTime(request.NodeId, successTimes, importType)
	response.Status = Active
	return response
}

func processImportNodeInformationRequest(ctx context.Context,
	request ImportNodeInformationRequest) ImportNodeInformationResponse {

	importType := core.ImportNodeInformation
	nodeSettings := cache.GetNodeSettingsByNodeId(request.NodeId)

	response := ImportNodeInformationResponse{
		Request: request,
		ImportResponse: ImportResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Inactive,
			},
		},
	}

	successTimes, skip := getSuccessTimes(request.NodeId, request.Force, importType)
	if skip {
		return response
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err
		return response
	}

	err = ImportNodeInfo(ctx, lnrpc.NewLightningClient(connection), request.Db, nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import node information.")
		response.Error = err
		return response
	}
	setSuccessTime(request.NodeId, successTimes, importType)
	response.Status = Active
	return response
}

func processImportPeerStatusRequest(ctx context.Context, request ImportPeerStatusRequest) ImportPeerStatusResponse {
	importType := core.ImportPeerStatus
	nodeSettings := cache.GetNodeSettingsByNodeId(request.NodeId)

	response := ImportPeerStatusResponse{
		Request: request,
		ImportResponse: ImportResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Inactive,
			},
		},
	}

	successTimes, skip := getSuccessTimes(request.NodeId, request.Force, importType)
	if skip {
		return response
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err
		return response
	}

	err = ImportPeerStatusFromLnd(ctx, lnrpc.NewLightningClient(connection), request.Db, nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import peer status.")
		response.Error = err
		return response
	}
	setSuccessTime(request.NodeId, successTimes, importType)
	response.Status = Active
	return response
}

func getSuccessTimes(nodeId int, force bool, importType core.ImportType) (map[core.ImportType]time.Time, bool) {
	// TODO FIXME For now there is no concurrency enabled for lightning communication
	// When concurrency is enabled this need to be revisited
	successTimes := cache.GetSuccessTimes(nodeId)
	if !force {
		successTime, exists := cache.GetSuccessTimes(nodeId)[importType]
		if exists && time.Since(successTime).Seconds() < avoidChannelAndPolicyImportRerunTimeSeconds {
			log.Info().Msgf("%v were imported very recently for nodeId: %v.", importType.String(), nodeId)
			return nil, true
		}
	}
	if force {
		log.Info().Msgf("Forced import of %v for nodeId: %v.", importType.String(), nodeId)
	}
	return successTimes, false
}

func setSuccessTime(nodeId int, successTimes map[core.ImportType]time.Time, importType core.ImportType) {
	log.Info().Msgf("%v was imported successfully for nodeId: %v.", importType.String(), nodeId)
	successTimes[importType] = time.Now()
	cache.SetSuccessTimes(nodeId, successTimes)
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

	signMsgReq := lnrpc.SignMessageRequest{
		Msg: []byte(request.Message),
	}
	if request.SingleHash != nil {
		signMsgReq.SingleHash = *request.SingleHash
	}
	signMsgResp, err := lnrpc.NewLightningClient(connection).SignMessage(ctx, &signMsgReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Status = lightning_requests.Active
	response.Signature = signMsgResp.Signature
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

	verifyMsgReq := lnrpc.VerifyMessageRequest{
		Msg:       []byte(request.Message),
		Signature: request.Signature,
	}
	verifyMsgResp, err := lnrpc.NewLightningClient(connection).VerifyMessage(ctx, &verifyMsgReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	if !verifyMsgResp.Valid {
		response.Message = "Signature is not valid"
		return response
	}

	response.Status = lightning_requests.Active
	response.PublicKey = verifyMsgResp.Pubkey
	response.Valid = verifyMsgResp.GetValid()
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

	lndWalletBalanceRequest := lnrpc.WalletBalanceRequest{}
	wb, err := lnrpc.NewLightningClient(connection).WalletBalance(ctx, &lndWalletBalanceRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Status = lightning_requests.Active
	response.ReservedBalanceAnchorChan = wb.ReservedBalanceAnchorChan
	response.UnconfirmedBalance = wb.UnconfirmedBalance
	response.ConfirmedBalance = wb.ConfirmedBalance
	response.TotalBalance = wb.TotalBalance
	response.LockedBalance = wb.LockedBalance

	return response
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

	infoRequest := lnrpc.GetInfoRequest{}
	info, err := lnrpc.NewLightningClient(connection).GetInfo(ctx, &infoRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Implementation = core.LND
	response.Status = lightning_requests.Active
	response.Version = info.Version
	response.PublicKey = info.IdentityPubkey
	response.Alias = info.Alias
	response.Color = info.Color
	response.PendingChannelCount = int(info.NumPendingChannels)
	response.ActiveChannelCount = int(info.NumActiveChannels)
	response.InactiveChannelCount = int(info.NumInactiveChannels)
	response.PeerCount = int(info.NumPeers)
	response.BlockHeight = info.BlockHeight
	response.BlockHash = info.BlockHash
	response.BestHeaderTimestamp = time.Unix(info.BestHeaderTimestamp, 0)
	response.ChainSynced = info.SyncedToChain
	response.GraphSynced = info.SyncedToGraph
	response.Addresses = info.Uris
	response.HtlcInterceptorRequired = info.RequireHtlcInterceptor
	//for _, chain := range info.Chains {
	//for number, feature := range info.Features {
	return response
}

func processChannelStatusUpdateRequest(ctx context.Context,
	request ChannelStatusUpdateRequest) ChannelStatusUpdateResponse {

	response := validateChannelStatusUpdateRequest(request)
	if response != nil {
		return *response
	}

	if !channelStatusUpdateRequestContainsUpdates(request) {
		return ChannelStatusUpdateResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Active,
			},
			Request: request,
		}
	}

	response = channelStatusUpdateRequestIsRepeated(request)
	if response != nil {
		return *response
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return *response
	}

	_, err = routerrpc.NewRouterClient(connection).UpdateChanStatus(ctx, constructUpdateChanStatusRequest(request))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to update channel status for channelId: %v on nodeId: %v",
			request.ChannelId, request.NodeId)
		return ChannelStatusUpdateResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Inactive,
				Error:  err.Error(),
			},
			Request: request,
		}
	}
	return ChannelStatusUpdateResponse{
		CommunicationResponse: CommunicationResponse{
			Status: Active,
		},
		Request: request,
	}
}

func constructUpdateChanStatusRequest(request ChannelStatusUpdateRequest) *routerrpc.UpdateChanStatusRequest {
	action := routerrpc.ChanStatusAction_DISABLE
	if request.ChannelStatus == core.Active {
		action = routerrpc.ChanStatusAction_ENABLE
	}
	channelSettings := cache.GetChannelSettingByChannelId(request.ChannelId)
	return &routerrpc.UpdateChanStatusRequest{
		ChanPoint: &lnrpc.ChannelPoint{
			FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: *channelSettings.FundingTransactionHash},
			OutputIndex: uint32(*channelSettings.FundingOutputIndex)},
		Action: action,
	}
}

func channelStatusUpdateRequestIsRepeated(request ChannelStatusUpdateRequest) *ChannelStatusUpdateResponse {
	secondsAgo := routingPolicyUpdateLimiterSeconds
	channelEventsFromGraph, err := graph_events.GetChannelEventFromGraph(request.Db, request.ChannelId, &secondsAgo)
	if err != nil {
		return &ChannelStatusUpdateResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Inactive,
				Error:  err.Error(),
			},
			Request: request,
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
			return &ChannelStatusUpdateResponse{
				CommunicationResponse: CommunicationResponse{
					Status: Inactive,
				},
				Request: request,
			}
		}
	}
	return nil
}

func channelStatusUpdateRequestContainsUpdates(request ChannelStatusUpdateRequest) bool {
	channelState := cache.GetChannelState(request.NodeId, request.ChannelId, true)
	if request.ChannelStatus == core.Active && channelState.LocalDisabled {
		return true
	}
	if request.ChannelStatus == core.Inactive && !channelState.LocalDisabled {
		return true
	}
	return false
}

func validateChannelStatusUpdateRequest(request ChannelStatusUpdateRequest) *ChannelStatusUpdateResponse {
	if request.ChannelId == 0 {
		return &ChannelStatusUpdateResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Inactive,
				Error:  "ChannelId is 0",
			},
			Request: request,
		}
	}
	if request.ChannelStatus != core.Active &&
		request.ChannelStatus != core.Inactive {
		return &ChannelStatusUpdateResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Inactive,
				Error:  "ChannelStatus is not Active nor Inactive",
			},
			Request: request,
		}
	}
	channelSettings := cache.GetChannelSettingByChannelId(request.ChannelId)
	if channelSettings.FundingTransactionHash == nil || *channelSettings.FundingTransactionHash == "" ||
		channelSettings.FundingOutputIndex == nil {
		return &ChannelStatusUpdateResponse{
			CommunicationResponse: CommunicationResponse{
				Status: Inactive,
				Error:  "FundingTransaction information is not known",
			},
			Request: request,
		}
	}
	return nil
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

	resp, err := lnrpc.NewLightningClient(connection).
		UpdateChannelPolicy(ctx, constructPolicyUpdateRequest(request, channelState))
	return processRoutingPolicyUpdateResponse(request, resp, err)
}

func processRoutingPolicyUpdateResponse(request lightning_requests.RoutingPolicyUpdateRequest,
	resp *lnrpc.PolicyUpdateResponse,
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
	for _, failedUpdate := range resp.GetFailedUpdates() {
		log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (lnd-grpc error: %v)",
			request.ChannelId, request.NodeId, failedUpdate.Reason)
		failedRequest := lightning_requests.FailedRequest{
			Reason: failedUpdate.UpdateError,
			Error:  failedUpdate.UpdateError,
		}
		failedUpdateArray = append(failedUpdateArray, failedRequest)
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
	channelState *cache.ChannelStateSettingsCache) *lnrpc.PolicyUpdateRequest {

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
	channelSettings := cache.GetChannelSettingByChannelId(request.ChannelId)
	policyUpdateRequest.Scope = &lnrpc.PolicyUpdateRequest_ChanPoint{
		ChanPoint: &lnrpc.ChannelPoint{
			FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: *channelSettings.FundingTransactionHash},
			OutputIndex: uint32(*channelSettings.FundingOutputIndex)}}
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
	channelSettings := cache.GetChannelSettingByChannelId(request.ChannelId)
	if channelSettings.FundingTransactionHash == nil || *channelSettings.FundingTransactionHash == "" ||
		channelSettings.FundingOutputIndex == nil {
		return &lightning_requests.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_requests.CommunicationResponse{
				Status: lightning_requests.Inactive,
				Error:  "FundingTransaction information is not known",
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
