package lnd

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/lightning_helpers"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/lnrpc/routerrpc"
	"github.com/lncapital/torq/proto/lnrpc/walletrpc"

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

func Information(request lightning_helpers.InformationRequest) lightning_helpers.InformationResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.InformationResponse); ok {
		return res
	}
	return lightning_helpers.InformationResponse{}
}

func SignMessage(request lightning_helpers.SignMessageRequest) lightning_helpers.SignMessageResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.SignMessageResponse); ok {
		return res
	}
	return lightning_helpers.SignMessageResponse{}
}

func SignatureVerification(request lightning_helpers.SignatureVerificationRequest) lightning_helpers.SignatureVerificationResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.SignatureVerificationResponse); ok {
		return res
	}
	return lightning_helpers.SignatureVerificationResponse{}
}

func RoutingPolicyUpdate(request lightning_helpers.RoutingPolicyUpdateRequest) lightning_helpers.RoutingPolicyUpdateResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.RoutingPolicyUpdateResponse); ok {
		return res
	}
	return lightning_helpers.RoutingPolicyUpdateResponse{}
}

func ConnectPeer(request lightning_helpers.ConnectPeerRequest) lightning_helpers.ConnectPeerResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.ConnectPeerResponse); ok {
		return res
	}
	return lightning_helpers.ConnectPeerResponse{}
}

func DisconnectPeer(request lightning_helpers.DisconnectPeerRequest) lightning_helpers.DisconnectPeerResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), disconnectPeerTimeoutInSeconds, request, responseChan)
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

func BatchOpenChannel(request lightning_helpers.BatchOpenChannelRequest) lightning_helpers.BatchOpenChannelResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 300, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.BatchOpenChannelResponse); ok {
		return res
	}
	return lightning_helpers.BatchOpenChannelResponse{}
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

func ChannelStatusUpdate(request ChannelStatusUpdateRequest) ChannelStatusUpdateResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ChannelStatusUpdateResponse); ok {
		return res
	}
	return ChannelStatusUpdateResponse{}
}

func ImportAllChannels(request ImportAllChannelsRequest) ImportAllChannelsResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportAllChannelsResponse); ok {
		return res
	}
	return ImportAllChannelsResponse{}
}

func ImportPendingChannels(request ImportPendingChannelsRequest) ImportPendingChannelsResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportPendingChannelsResponse); ok {
		return res
	}
	return ImportPendingChannelsResponse{}
}

func ImportChannelRoutingPolicies(request ImportChannelRoutingPoliciesRequest) ImportChannelRoutingPoliciesResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportChannelRoutingPoliciesResponse); ok {
		return res
	}
	return ImportChannelRoutingPoliciesResponse{}
}

func ImportNodeInformation(request ImportNodeInformationRequest) ImportNodeInformationResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportNodeInformationResponse); ok {
		return res
	}
	return ImportNodeInformationResponse{}
}

func ImportPeerStatus(request ImportPeerStatusRequest) ImportPeerStatusResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportPeerStatusResponse); ok {
		return res
	}
	return ImportPeerStatusResponse{}
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
	case lightning_helpers.BatchOpenChannelRequest:
		responseChan <- processBatchOpenChannelRequest(ctx, r)
		return
	case lightning_helpers.CloseChannelRequest:
		responseChan <- processCloseChannelRequest(ctx, r)
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
	request lightning_helpers.ListPeersRequest) lightning_helpers.ListPeersResponse {

	response := lightning_helpers.ListPeersResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
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

	peers := make(map[string]lightning_helpers.Peer)
	for _, peer := range rsp.Peers {
		peers[peer.PubKey] = lightning_helpers.GetPeerLND(peer)
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
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	lndAddressRequest := &walletrpc.AddrRequest{}
	if request.Account != "" {
		lndAddressRequest.Account = request.Account
	}
	switch request.Type {
	case lightning_helpers.P2WPKH:
		lndAddressRequest.Type = walletrpc.AddressType_WITNESS_PUBKEY_HASH
	case lightning_helpers.P2WKH:
		lndAddressRequest.Type = walletrpc.AddressType_NESTED_WITNESS_PUBKEY_HASH
	case lightning_helpers.NP2WKH:
		lndAddressRequest.Type = walletrpc.AddressType_HYBRID_NESTED_WITNESS_PUBKEY_HASH
	case lightning_helpers.P2TR:
		lndAddressRequest.Type = walletrpc.AddressType_TAPROOT_PUBKEY
	default:
		response.Error = "unknown address type"
		return response
	}

	rsp, err := walletrpc.NewWalletKitClient(connection).NextAddr(ctx, lndAddressRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	response.Status = lightning_helpers.Active
	response.Address = rsp.Addr

	return response
}

const openChannelTimeoutInSeconds = 60

func processOpenChannelRequest(ctx context.Context,
	request lightning_helpers.OpenChannelRequest) lightning_helpers.OpenChannelResponse {

	response := lightning_helpers.OpenChannelResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
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

	client := lnrpc.NewLightningClient(connection)

	//If host provided - check if node is connected to peer and if not, connect peer
	if request.NodePubKey != "" && request.Host != nil {
		if err := checkConnectPeer(request.NodeId, request.NodePubKey, *request.Host); err != nil {
			response.Error = "could not connect to peer"
			return response
		}
	}

	// Send open channel request
	response, err = openChannelProcess(ctx, client, openChanReq, request)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	response.Status = lightning_helpers.Active
	return response
}

func prepareOpenRequest(request lightning_helpers.OpenChannelRequest) (r *lnrpc.OpenChannelRequest, err error) {
	if request.NodeId == 0 {
		return &lnrpc.OpenChannelRequest{}, errors.New("Node id is missing")
	}

	if request.SatPerVbyte != nil && request.TargetConf != nil {
		return &lnrpc.OpenChannelRequest{}, errors.New("Cannot set both SatPerVbyte and TargetConf")
	}

	pubKeyHex, err := hex.DecodeString(request.NodePubKey)
	if err != nil {
		return &lnrpc.OpenChannelRequest{}, errors.New("error decoding public key hex")
	}

	//open channel request
	openChanReq := &lnrpc.OpenChannelRequest{
		NodePubkey: pubKeyHex,

		// This is the amount we are putting into the channel (channel size)
		LocalFundingAmount: request.LocalFundingAmount,
	}

	// The amount to give the other node in the opening process.
	// NB: This means you will give the other node this amount of sats
	if request.PushSat != nil {
		openChanReq.PushSat = *request.PushSat
	}

	if request.SatPerVbyte != nil {
		openChanReq.SatPerVbyte = *request.SatPerVbyte
	}

	if request.TargetConf != nil {
		openChanReq.TargetConf = *request.TargetConf
	}

	if request.Private != nil {
		openChanReq.Private = *request.Private
	}

	if request.MinHtlcMsat != nil {
		openChanReq.MinHtlcMsat = int64(*request.MinHtlcMsat)
	}

	if request.RemoteCsvDelay != nil {
		openChanReq.RemoteCsvDelay = *request.RemoteCsvDelay
	}

	if request.MinConfs != nil {
		openChanReq.MinConfs = *request.MinConfs
	}

	if request.SpendUnconfirmed != nil {
		openChanReq.SpendUnconfirmed = *request.SpendUnconfirmed
	}

	if request.CloseAddress != nil {
		openChanReq.CloseAddress = *request.CloseAddress
	}
	return openChanReq, nil
}

func checkConnectPeer(nodeId int, remotePublicKey string, host string) error {
	peerList := ListPeers(lightning_helpers.ListPeersRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: nodeId},
		LatestError:          false,
	})
	if peerList.Error != "" {
		return errors.Wrap(errors.New(peerList.Error), "List peers")
	}

	for _, peer := range peerList.Peers {
		if peer.PubKey == remotePublicKey {
			// peer found
			//log.Debug().Msgf("Peer is connected")
			return nil
		}
	}

	req := lightning_helpers.ConnectPeerRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: nodeId},
		PublicKey:            remotePublicKey,
		Host:                 host,
	}

	res := ConnectPeer(req)
	if res.Error != "" {
		return errors.Wrap(errors.New(res.Error), "Connect peer")
	}

	return nil
}

func openChannelProcess(ctx context.Context,
	client lnrpc.LightningClient,
	openChannelReq *lnrpc.OpenChannelRequest,
	request lightning_helpers.OpenChannelRequest) (lightning_helpers.OpenChannelResponse, error) {

	// Create a context with a timeout.
	timeoutCtx, cancel := context.WithTimeout(ctx, openChannelTimeoutInSeconds*time.Second)
	defer cancel()

	// Call OpenChannel with the timeout context.
	openReq, err := client.OpenChannel(timeoutCtx, openChannelReq)
	if err != nil {
		return lightning_helpers.OpenChannelResponse{}, errors.Wrap(err, "Close channel request")
	}

	// Loop until we receive an open channel response or the context times out.
	for {
		select {
		case <-timeoutCtx.Done():
			return lightning_helpers.OpenChannelResponse{}, errors.New("Close channel request timeout")
		default:
		}

		// Receive the next close channel response message.
		resp, err := openReq.Recv()
		if err != nil {
			if err == io.EOF {
				// TODO FIXME this doesn't look ok?
				// No more messages to receive, the channel is open.
				return lightning_helpers.OpenChannelResponse{}, nil
			}
			return lightning_helpers.OpenChannelResponse{}, errors.Wrap(err, "LND Open channel")
		}

		r := lightning_helpers.OpenChannelResponse{
			Request: request,
		}
		if resp.Update == nil {
			continue
		}

		switch resp.GetUpdate().(type) {
		case *lnrpc.OpenStatusUpdate_ChanPending:
			r.ChannelStatus = core.Opening
			ch, err := chainhash.NewHash(resp.GetChanPending().Txid)
			if err != nil {
				return lightning_helpers.OpenChannelResponse{}, errors.Wrap(err, "Getting closing transaction hash")
			}
			r.FundingTransactionHash = ch.String()
			r.FundingOutputIndex = resp.GetChanPending().OutputIndex
			r.ChannelPoint = fmt.Sprintf("%s:%d", ch.String(), resp.GetChanPending().OutputIndex)
			return r, nil
		}
	}
}

func processBatchOpenChannelRequest(ctx context.Context,
	request lightning_helpers.BatchOpenChannelRequest) lightning_helpers.BatchOpenChannelResponse {

	response := lightning_helpers.BatchOpenChannelResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	bOpenChanReq, err := checkPrepareReq(request)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	client := lnrpc.NewLightningClient(connection)
	bocResponse, err := client.BatchOpenChannel(ctx, bOpenChanReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	r, err := processBatchOpenChannelResponse(bocResponse)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	return r

}

func checkPrepareReq(bocReq lightning_helpers.BatchOpenChannelRequest) (req *lnrpc.BatchOpenChannelRequest, err error) {

	if bocReq.NodeId == 0 {
		return req, errors.New("Node id is missing")
	}

	if len(bocReq.Channels) == 0 {
		log.Debug().Msgf("channel array empty")
		return req, errors.New("Channels array is empty")
	}

	if bocReq.TargetConf != nil && bocReq.SatPerVbyte != nil {
		log.Error().Msgf("Only one fee model accepted")
		return req, errors.New("Either targetConf or satPerVbyte accepted")
	}

	var boChannels []*lnrpc.BatchOpenChannel

	for i, channel := range bocReq.Channels {
		var boChannel lnrpc.BatchOpenChannel
		pubKeyHex, err := hex.DecodeString(channel.NodePublicKey)
		if err != nil {
			log.Error().Msgf("Err decoding string: %v, %v", i, err)
			return req, errors.Wrap(err, "Hex decode string")
		}
		boChannel.NodePubkey = pubKeyHex

		if channel.LocalFundingAmount == 0 {
			log.Debug().Msgf("local funding amt 0")
			return req, errors.New("Local funding amount 0")
		}
		boChannel.LocalFundingAmount = channel.LocalFundingAmount

		if channel.Private != nil {
			boChannel.Private = *channel.Private
		}

		if channel.PushSat != nil {
			boChannel.PushSat = *channel.PushSat
		}
		boChannels = append(boChannels, &boChannel)
	}

	batchOpnReq := &lnrpc.BatchOpenChannelRequest{
		Channels: boChannels,
	}

	if bocReq.SatPerVbyte != nil {
		batchOpnReq.SatPerVbyte = *bocReq.SatPerVbyte
	}

	if bocReq.TargetConf != nil {
		batchOpnReq.TargetConf = *bocReq.TargetConf
	}

	return batchOpnReq, nil
}

func processBatchOpenChannelResponse(
	resp *lnrpc.BatchOpenChannelResponse) (lightning_helpers.BatchOpenChannelResponse, error) {

	response := lightning_helpers.BatchOpenChannelResponse{}
	for _, pc := range resp.GetPendingChannels() {
		chanPoint, err := chanPointFromByte(pc.Txid, pc.OutputIndex)
		if err != nil {
			log.Error().Msgf("Translate channel point err: %v", err)
			return lightning_helpers.BatchOpenChannelResponse{}, err
		}
		response.PendingChannelPoints = append(response.PendingChannelPoints, chanPoint)
	}
	return response, nil
}

const closeChannelTimeoutInSeconds = 60

func processCloseChannelRequest(ctx context.Context,
	request lightning_helpers.CloseChannelRequest) lightning_helpers.CloseChannelResponse {

	response := lightning_helpers.CloseChannelResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
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

	client := lnrpc.NewLightningClient(connection)

	// Send Close channel request
	response, err = closeChannelProcess(ctx, client, closeChanReq, request)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	response.Status = lightning_helpers.Active
	return response
}

func prepareCloseRequest(ccReq lightning_helpers.CloseChannelRequest) (r *lnrpc.CloseChannelRequest, err error) {

	if ccReq.NodeId == 0 {
		return &lnrpc.CloseChannelRequest{}, errors.New("Node id is missing")
	}

	if ccReq.SatPerVbyte != nil && ccReq.TargetConf != nil {
		return &lnrpc.CloseChannelRequest{}, errors.New("cannot set both SatPerVbyte and TargetConf")
	}

	channelSettings := cache.GetChannelSettingByChannelId(ccReq.ChannelId)

	if channelSettings.FundingTransactionHash == nil || channelSettings.FundingOutputIndex == nil {
		return &lnrpc.CloseChannelRequest{}, errors.New("cannot find FundingTransactionHash or OutputIndex")
	}

	//Make the close channel request
	closeChanReq := &lnrpc.CloseChannelRequest{
		ChannelPoint: &lnrpc.ChannelPoint{
			FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
				FundingTxidStr: *channelSettings.FundingTransactionHash,
			},
			OutputIndex: uint32(*channelSettings.FundingOutputIndex),
		},
	}

	if ccReq.Force != nil {
		closeChanReq.Force = *ccReq.Force
	}

	if ccReq.TargetConf != nil {
		closeChanReq.TargetConf = *ccReq.TargetConf
	}

	if ccReq.SatPerVbyte != nil {
		closeChanReq.SatPerVbyte = *ccReq.SatPerVbyte
	}

	if ccReq.DeliveryAddress != nil {
		closeChanReq.DeliveryAddress = *ccReq.DeliveryAddress
	}

	return closeChanReq, nil
}

func closeChannelProcess(ctx context.Context,
	client lnrpc.LightningClient,
	lndRequest *lnrpc.CloseChannelRequest,
	request lightning_helpers.CloseChannelRequest) (lightning_helpers.CloseChannelResponse, error) {

	// Create a context with a timeout.
	timeoutCtx, cancel := context.WithTimeout(ctx, closeChannelTimeoutInSeconds*time.Second)
	defer cancel()

	// Call CloseChannel with the timeout context.
	closeChanRes, err := client.CloseChannel(timeoutCtx, lndRequest)
	if err != nil {
		err = errors.Wrap(err, "problem sending closing channel request to LND")
		log.Error().Err(err).Send()
		return lightning_helpers.CloseChannelResponse{}, err
	}

	// Loop until we receive a close channel response or the context times out.
	for {
		select {
		case <-timeoutCtx.Done():
			return lightning_helpers.CloseChannelResponse{}, errors.New("close channel request timeout")
		default:
		}

		// Receive the next close channel response message.
		resp, err := closeChanRes.Recv()
		if err != nil {
			if err == io.EOF {
				// TODO FIXME this doesn't look ok?
				// No more messages to receive, the channel is closed.
				return lightning_helpers.CloseChannelResponse{}, nil
			}
			log.Error().Err(errors.Wrap(err, "LND close channel")).Send()
			return lightning_helpers.CloseChannelResponse{}, errors.Wrap(err, "LND Close channel")
		}

		// Process the close channel response and see if the channel is pending closure.
		r := lightning_helpers.CloseChannelResponse{
			Request: request,
		}
		if resp.Update == nil {
			continue
		}

		switch resp.GetUpdate().(type) {
		case *lnrpc.CloseStatusUpdate_ClosePending:
			r.ChannelStatus = core.Closing
			ch, err := chainhash.NewHash(resp.GetClosePending().Txid)
			if err != nil {
				return lightning_helpers.CloseChannelResponse{}, errors.Wrap(err, "getting closing transaction hash")
			}
			r.ClosingTransactionHash = ch.String()

			err = channels.UpdateChannelToClosingByChannelId(request.Db, request.ChannelId, ch.String())
			if err != nil {
				return lightning_helpers.CloseChannelResponse{}, errors.Wrap(err, "updating channel to closing status in the db")
			}
			return r, nil
		}
	}
}

const disconnectPeerTimeoutInSeconds = 10
const disconnectPeerAttemptDelayInSeconds = 1
const maximumAttempts = 5

func processDisconnectPeerRequest(ctx context.Context,
	request lightning_helpers.DisconnectPeerRequest) lightning_helpers.DisconnectPeerResponse {

	response := lightning_helpers.DisconnectPeerResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
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
				response.Status = lightning_helpers.Active
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
			response.Status = lightning_helpers.Active
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
	request lightning_helpers.ConnectPeerRequest) lightning_helpers.ConnectPeerResponse {

	response := lightning_helpers.ConnectPeerResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
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
		response.Status = lightning_helpers.Active
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
	importType := services_helpers.ImportAllChannels
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

	importType := services_helpers.ImportPendingChannels
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

	importType := services_helpers.ImportChannelRoutingPolicies
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

	importType := services_helpers.ImportNodeInformation
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
	importType := services_helpers.ImportPeerStatus
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

func getSuccessTimes(nodeId int,
	force bool,
	importType services_helpers.ImportType) (map[services_helpers.ImportType]time.Time, bool) {

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

func setSuccessTime(nodeId int,
	successTimes map[services_helpers.ImportType]time.Time,
	importType services_helpers.ImportType) {

	log.Info().Msgf("%v was imported successfully for nodeId: %v.", importType.String(), nodeId)
	successTimes[importType] = time.Now()
	cache.SetSuccessTimes(nodeId, successTimes)
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

	response.Status = lightning_helpers.Active
	response.Signature = signMsgResp.Signature
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

	response.Status = lightning_helpers.Active
	response.PublicKey = verifyMsgResp.Pubkey
	response.Valid = verifyMsgResp.GetValid()
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

	lndWalletBalanceRequest := lnrpc.WalletBalanceRequest{}
	wb, err := lnrpc.NewLightningClient(connection).WalletBalance(ctx, &lndWalletBalanceRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Status = lightning_helpers.Active
	response.ReservedBalanceAnchorChan = wb.ReservedBalanceAnchorChan
	response.UnconfirmedBalance = wb.UnconfirmedBalance
	response.ConfirmedBalance = wb.ConfirmedBalance
	response.TotalBalance = wb.TotalBalance
	response.LockedBalance = wb.LockedBalance

	return response
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

	infoRequest := lnrpc.GetInfoRequest{}
	info, err := lnrpc.NewLightningClient(connection).GetInfo(ctx, &infoRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Implementation = core.LND
	response.Status = lightning_helpers.Active
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

	resp, err := lnrpc.NewLightningClient(connection).
		UpdateChannelPolicy(ctx, constructPolicyUpdateRequest(request, channelState))
	return processRoutingPolicyUpdateResponse(request, resp, err)
}

func processRoutingPolicyUpdateResponse(request lightning_helpers.RoutingPolicyUpdateRequest,
	resp *lnrpc.PolicyUpdateResponse,
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
	for _, failedUpdate := range resp.GetFailedUpdates() {
		log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (lnd-grpc error: %v)",
			request.ChannelId, request.NodeId, failedUpdate.Reason)
		failedRequest := lightning_helpers.FailedRequest{
			Reason: failedUpdate.UpdateError,
			Error:  failedUpdate.UpdateError,
		}
		failedUpdateArray = append(failedUpdateArray, failedRequest)
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
	channelSettings := cache.GetChannelSettingByChannelId(request.ChannelId)
	if channelSettings.FundingTransactionHash == nil || *channelSettings.FundingTransactionHash == "" ||
		channelSettings.FundingOutputIndex == nil {
		return &lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  "FundingTransaction information is not known",
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
