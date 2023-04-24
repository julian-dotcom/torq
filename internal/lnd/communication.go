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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

func NewInvoice(request lightning_helpers.NewInvoiceRequest) lightning_helpers.NewInvoiceResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.NewInvoiceResponse); ok {
		return res
	}
	return lightning_helpers.NewInvoiceResponse{}
}

func OnChainPayment(request lightning_helpers.OnChainPaymentRequest) lightning_helpers.OnChainPaymentResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.OnChainPaymentResponse); ok {
		return res
	}
	return lightning_helpers.OnChainPaymentResponse{}
}

// NewPayment - send new payment
// A new payment can be made either by providing an invoice or by providing:
// dest - the identity pubkey of the payment recipient
// amt(number of satoshi) or amt_msat(number of millisatoshi)
// amt and amt_msat are mutually exclusive
// payments hash - the hash to use within the payment's HTLC
// timeout seconds is mandatory
func NewPayment(request lightning_helpers.NewPaymentRequest) lightning_helpers.NewPaymentResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 120, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.NewPaymentResponse); ok {
		return res
	}
	return lightning_helpers.NewPaymentResponse{}
}

func DecodeInvoice(request lightning_helpers.DecodeInvoiceRequest) lightning_helpers.DecodeInvoiceResponse {
	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.DecodeInvoiceResponse); ok {
		return res
	}
	return lightning_helpers.DecodeInvoiceResponse{}
}

func ChannelStatusUpdate(
	request lightning_helpers.ChannelStatusUpdateRequest) lightning_helpers.ChannelStatusUpdateResponse {

	responseChan := make(chan any)
	processSequential(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.ChannelStatusUpdateResponse); ok {
		return res
	}
	return lightning_helpers.ChannelStatusUpdateResponse{}
}

func ImportAllChannelsUnshared(request ImportAllChannelsRequest) ImportAllChannelsResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportAllChannelsResponse); ok {
		return res
	}
	return ImportAllChannelsResponse{}
}

func ImportPendingChannelsUnshared(request ImportPendingChannelsRequest) ImportPendingChannelsResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportPendingChannelsResponse); ok {
		return res
	}
	return ImportPendingChannelsResponse{}
}

func ImportChannelRoutingPoliciesUnshared(request ImportChannelRoutingPoliciesRequest) ImportChannelRoutingPoliciesResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportChannelRoutingPoliciesResponse); ok {
		return res
	}
	return ImportChannelRoutingPoliciesResponse{}
}

func ImportNodeInformationUnshared(request ImportNodeInformationRequest) ImportNodeInformationResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(ImportNodeInformationResponse); ok {
		return res
	}
	return ImportNodeInformationResponse{}
}

func ImportPeerStatusUnshared(request ImportPeerStatusRequest) ImportPeerStatusResponse {
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
	case lightning_helpers.NewInvoiceRequest:
		responseChan <- processNewInvoiceRequest(ctx, r)
		return
	case lightning_helpers.OnChainPaymentRequest:
		responseChan <- processOnChainPaymentRequest(ctx, r)
		return
	case lightning_helpers.NewPaymentRequest:
		responseChan <- processNewPaymentRequest(ctx, r)
		return
	case lightning_helpers.DecodeInvoiceRequest:
		responseChan <- processDecodeInvoiceRequest(ctx, r)
		return
	case lightning_helpers.ChannelStatusUpdateRequest:
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

type ImportRequest struct {
	lightning_helpers.CommunicationRequest
	Db    *sqlx.DB
	Force bool
}

type ImportResponse struct {
	lightning_helpers.CommunicationResponse
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
		Request: request,
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
		Request: request,
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
		Request: request,
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

func processNewInvoiceRequest(ctx context.Context,
	request lightning_helpers.NewInvoiceRequest) lightning_helpers.NewInvoiceResponse {

	response := lightning_helpers.NewInvoiceResponse{
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

	newInvoiceReq, err := processInvoiceReq(request)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	resp, err := lnrpc.NewLightningClient(connection).AddInvoice(ctx, newInvoiceReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.PaymentRequest = resp.GetPaymentRequest()
	response.AddIndex = resp.GetAddIndex()
	response.PaymentAddress = hex.EncodeToString(resp.GetPaymentAddr())
	response.Status = lightning_helpers.Active
	return response
}

func processInvoiceReq(request lightning_helpers.NewInvoiceRequest) (*lnrpc.Invoice, error) {
	inv := &lnrpc.Invoice{}

	if request.NodeId == 0 {
		return &lnrpc.Invoice{}, errors.New("Node id is missing")
	}

	if request.Memo != nil {
		inv.Memo = *request.Memo
	}

	if request.RPreImage != nil {
		rPreImage, err := hex.DecodeString(*request.RPreImage)
		if err != nil {
			return &lnrpc.Invoice{}, errors.New("error decoding preimage")
		}
		inv.RPreimage = rPreImage
	}

	if request.ValueMsat != nil {
		inv.ValueMsat = *request.ValueMsat
	}

	if request.Expiry != nil {
		inv.Expiry = *request.Expiry
	}

	if request.FallBackAddress != nil {
		inv.FallbackAddr = *request.FallBackAddress
	}

	if request.Private != nil {
		inv.Private = *request.Private
	}

	if request.IsAmp != nil {
		inv.IsAmp = *request.IsAmp
	}

	return inv, nil
}

func processOnChainPaymentRequest(ctx context.Context,
	request lightning_helpers.OnChainPaymentRequest) lightning_helpers.OnChainPaymentResponse {

	response := lightning_helpers.OnChainPaymentResponse{
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

	sendCoinsReq, err := processSendRequest(request)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	resp, err := lnrpc.NewLightningClient(connection).SendCoins(ctx, sendCoinsReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.TxId = resp.Txid
	response.Status = lightning_helpers.Active
	return response

}

func processSendRequest(req lightning_helpers.OnChainPaymentRequest) (*lnrpc.SendCoinsRequest, error) {
	r := &lnrpc.SendCoinsRequest{}

	if req.NodeId == 0 {
		return &lnrpc.SendCoinsRequest{}, errors.New("Node id is missing")
	}

	if req.Address == "" {
		log.Error().Msgf("Address must be provided")
		return &lnrpc.SendCoinsRequest{}, errors.New("Address must be provided")
	}

	if req.AmountSat <= 0 {
		log.Error().Msgf("Invalid amount")
		return &lnrpc.SendCoinsRequest{}, errors.New("Invalid amount")
	}

	if req.TargetConf != nil && req.SatPerVbyte != nil {
		log.Error().Msgf("Either targetConf or satPerVbyte accepted")
		return &lnrpc.SendCoinsRequest{}, errors.New("Either targetConf or satPerVbyte accepted")
	}

	r.Addr = req.Address
	r.Amount = req.AmountSat

	if req.TargetConf != nil {
		r.TargetConf = *req.TargetConf
	}

	if req.SatPerVbyte != nil {
		r.SatPerVbyte = *req.SatPerVbyte
	}

	if req.SendAll != nil {
		r.SendAll = *req.SendAll
	}

	if req.Label != nil {
		r.Label = *req.Label
	}

	if req.MinConfs != nil {
		r.MinConfs = *req.MinConfs
	}

	if req.SpendUnconfirmed != nil {
		r.SpendUnconfirmed = *req.SpendUnconfirmed
	}

	return r, nil
}

func processNewPaymentRequest(ctx context.Context,
	request lightning_helpers.NewPaymentRequest) lightning_helpers.NewPaymentResponse {

	response := lightning_helpers.NewPaymentResponse{
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

	return sendPayment(ctx, routerrpc.NewRouterClient(connection), request, response)
}

func newSendPaymentRequest(npReq lightning_helpers.NewPaymentRequest) (*routerrpc.SendPaymentRequest, error) {
	newPayReq := &routerrpc.SendPaymentRequest{
		TimeoutSeconds: npReq.TimeOutSecs,
	}

	if npReq.Invoice != nil {
		newPayReq.PaymentRequest = *npReq.Invoice
	}

	if npReq.FeeLimitMsat != nil && *npReq.FeeLimitMsat != 0 {
		newPayReq.FeeLimitMsat = *npReq.FeeLimitMsat
	}

	if npReq.AmtMSat != nil {
		newPayReq.AmtMsat = *npReq.AmtMSat
	}

	if npReq.AllowSelfPayment != nil {
		newPayReq.AllowSelfPayment = *npReq.AllowSelfPayment
	}

	// TODO: Add support for Keysend, needs to solve issue related to payment hash generation
	//if npReq.Dest != nil {
	//	fmt.Println("It was a keysend")
	//	destHex, err := hex.DecodeString(*npReq.Dest)
	//	if err != nil {
	//		return r, errors.New("Could not decode destination pubkey (keysend)")
	//	}
	//	newPayReq.Dest = destHex
	// //	newPayReq.PaymentHash = make([]byte, 32)
	//}

	return newPayReq, nil
}

func sendPayment(ctx context.Context,
	client routerrpc.RouterClient,
	request lightning_helpers.NewPaymentRequest,
	response lightning_helpers.NewPaymentResponse) lightning_helpers.NewPaymentResponse {

	// Create and validate payment request details
	newPayReq, err := newSendPaymentRequest(request)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	req, err := client.SendPaymentV2(ctx, newPayReq)
	if err != nil {
		response.Error = "sending payment"
		return response
	}

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				response.Error = "payment timeout"
				return response
			}
			response.Error = "context ended"
			return response
		default:
		}

		resp, err := req.Recv()
		switch err {
		case nil:
			break
		case io.EOF:
		case status.Error(codes.AlreadyExists, ""):
			response.Error = "ALREADY_PAID"
			return response
		case status.Error(codes.NotFound, "lnrpc.Lightning_PayInvoice.UnknownPaymentHash"):
			response.Error = "INVALID_HASH"
			return response
		case status.Error(codes.InvalidArgument, "lnrpc.Lightning_PayReq.InvalidPaymentRequest"):
			response.Error = "INVALID_PAYMENT_REQUEST"
			return response
		case status.Error(codes.InvalidArgument, "lnrpc.Lightning_SendPaymentRequest.CheckPaymentRequest"):
			response.Error = "CHECKSUM_FAILED"
			return response
		case status.Error(codes.InvalidArgument, "amount must be specified when paying a zero amount invoice"):
			response.Error = "AMOUNT_REQUIRED"
			return response
		case status.Error(codes.InvalidArgument, "amount must not be specified when paying a non-zero amount invoice"):
			response.Error = "AMOUNT_NOT_ALLOWED"
			return response
		default:
			log.Error().Msgf("Unknown payment error %v", err)
			response.Error = "UNKNOWN_ERROR"
			return response
		}

		if request.ProgressReportChannel != nil {
			// Do a non-blocking write as the pub sub + websockets current dead locks itself
			// TODO: Make it so it can't deadlock itself
			select {
			case request.ProgressReportChannel <- processResponse(resp, request):
			default:
			}
		}

		// TODO: If LND fails to update us that the payment succeeded or failed for whatever reason
		// this for loop will run forever (a memory leak).
		// We should probably have some kind of timeout which exits regarless after a certain amount of time
		switch resp.GetStatus() {
		case lnrpc.Payment_SUCCEEDED:
			response.Status = lightning_helpers.Active
			return response
		case lnrpc.Payment_FAILED:
			return response
		}
	}
}

func processResponse(p *lnrpc.Payment, req lightning_helpers.NewPaymentRequest) lightning_helpers.NewPaymentResponse {
	r := lightning_helpers.NewPaymentResponse{
		Request:       req,
		PaymentStatus: p.Status.String(),
		Hash:          p.PaymentHash,
		Preimage:      p.PaymentPreimage,
		AmountMsat:    p.ValueMsat,
		CreationDate:  time.Unix(0, p.CreationTimeNs),
		FailureReason: p.FailureReason.String(),
		FeePaidMsat:   p.FeeMsat,
	}
	for _, attempt := range p.GetHtlcs() {
		r.Attempt.AttemptId = attempt.AttemptId
		r.Attempt.Status = attempt.Status.String()
		r.Attempt.AttemptTimeNs = time.Unix(0, attempt.AttemptTimeNs)
		r.Attempt.ResolveTimeNs = time.Unix(0, attempt.ResolveTimeNs)
		r.Attempt.Preimage = hex.EncodeToString(attempt.Preimage)

		if attempt.Failure != nil {
			r.Attempt.Failure.Reason = attempt.Failure.Code.String()
			r.Attempt.Failure.FailureSourceIndex = attempt.Failure.FailureSourceIndex
			r.Attempt.Failure.Height = attempt.Failure.Height
		}

		for _, hop := range attempt.Route.Hops {
			h := lightning_helpers.Hops{
				ChanId:           core.ConvertLNDShortChannelID(hop.ChanId),
				AmtToForwardMsat: hop.AmtToForwardMsat,
				Expiry:           hop.Expiry,
				PubKey:           hop.PubKey,
			}
			if hop.MppRecord != nil {
				h.MppRecord.TotalAmtMsat = hop.MppRecord.TotalAmtMsat
				h.MppRecord.PaymentAddr = hex.EncodeToString(hop.MppRecord.PaymentAddr)
			}
			r.Attempt.Route.Hops = append(r.Attempt.Route.Hops, h)
		}

		r.Attempt.Route.TotalTimeLock = attempt.Route.TotalTimeLock
		r.Attempt.Route.TotalAmtMsat = attempt.Route.TotalAmtMsat
	}
	return r
}

func constructRouteHints(routeHints []*lnrpc.RouteHint) []lightning_helpers.RouteHint {
	var r []lightning_helpers.RouteHint

	for _, rh := range routeHints {
		var hopHints []lightning_helpers.HopHint
		for _, hh := range rh.HopHints {
			hopHints = append(hopHints, lightning_helpers.HopHint{
				LNDShortChannelId: hh.ChanId,
				ShortChannelId:    core.ConvertLNDShortChannelID(hh.ChanId),
				NodeId:            hh.NodeId,
				FeeBase:           hh.FeeBaseMsat,
				CltvExpiryDelta:   hh.CltvExpiryDelta,
				FeeProportional:   hh.FeeProportionalMillionths,
			})
		}
		r = append(r, lightning_helpers.RouteHint{
			HopHints: hopHints,
		})
	}

	return r
}

func constructFeatureMap(features map[uint32]*lnrpc.Feature) lightning_helpers.FeatureMap {

	f := lightning_helpers.FeatureMap{}
	for n, v := range features {
		f[n] = lightning_helpers.Feature{
			Name:       v.Name,
			IsKnown:    v.IsKnown,
			IsRequired: v.IsRequired,
		}
	}

	return f
}

func constructDecodedInvoice(decodedInvoice *lnrpc.PayReq,
	response lightning_helpers.DecodeInvoiceResponse) lightning_helpers.DecodeInvoiceResponse {

	response.DestinationPubKey = decodedInvoice.Destination
	response.RHash = decodedInvoice.PaymentHash
	response.Memo = decodedInvoice.Description
	response.ValueMsat = decodedInvoice.NumMsat
	response.FallbackAddr = decodedInvoice.FallbackAddr
	response.CreatedAt = decodedInvoice.Timestamp
	response.Expiry = decodedInvoice.Expiry
	response.CltvExpiry = decodedInvoice.CltvExpiry
	response.RouteHints = constructRouteHints(decodedInvoice.RouteHints)
	response.Features = constructFeatureMap(decodedInvoice.Features)
	response.PaymentAddr = hex.EncodeToString(decodedInvoice.PaymentAddr)
	return response
}

func processDecodeInvoiceRequest(ctx context.Context,
	request lightning_helpers.DecodeInvoiceRequest) lightning_helpers.DecodeInvoiceResponse {

	response := lightning_helpers.DecodeInvoiceResponse{
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

	client := lnrpc.NewLightningClient(connection)
	// TODO: Handle different error types like incorrect checksum etc to explain why the decode failed.
	decodedInvoice, err := client.DecodePayReq(ctx, &lnrpc.PayReqString{
		PayReq: request.Invoice,
	})
	if err != nil {
		response.Error = err.Error()
		return response
	}

	nodeInfo, err := client.GetNodeInfo(ctx, &lnrpc.NodeInfoRequest{
		PubKey:          decodedInvoice.Destination,
		IncludeChannels: false,
	})
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response = constructDecodedInvoice(decodedInvoice, response)
	response.NodeAlias = nodeInfo.Node.Alias

	return response
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
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
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

	client := lnrpc.NewLightningClient(connection)

	//Import Pending channels
	err = ImportPendingChannelsFromLnd(ctx, request.Db, client, nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import pending channels.")
		response.Error = err
		return response
	}

	//Import Open channels
	err = ImportOpenChannelsFromLnd(ctx, request.Db, client, nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import open channels.")
		response.Error = err
		return response
	}

	// Import Closed channels
	err = ImportClosedChannelsFromLnd(ctx, request.Db, client, nodeSettings)
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
	response.Status = lightning_helpers.Active
	return response
}

func processImportPendingChannelsRequest(ctx context.Context,
	request ImportPendingChannelsRequest) ImportPendingChannelsResponse {

	importType := services_helpers.ImportPendingChannels
	nodeSettings := cache.GetNodeSettingsByNodeId(request.NodeId)

	response := ImportPendingChannelsResponse{
		Request: request,
		ImportResponse: ImportResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
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
	response.Status = lightning_helpers.Active
	return response
}

func processImportChannelRoutingPoliciesRequest(ctx context.Context,
	request ImportChannelRoutingPoliciesRequest) ImportChannelRoutingPoliciesResponse {

	importType := services_helpers.ImportChannelRoutingPolicies
	nodeSettings := cache.GetNodeSettingsByNodeId(request.NodeId)

	response := ImportChannelRoutingPoliciesResponse{
		Request: request,
		ImportResponse: ImportResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
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

	err = ImportRoutingPoliciesFromLnd(ctx, lnrpc.NewLightningClient(connection), request.Db, nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import routing policies.")
		response.Error = err
		return response
	}
	setSuccessTime(request.NodeId, successTimes, importType)
	response.Status = lightning_helpers.Active
	return response
}

func processImportNodeInformationRequest(ctx context.Context,
	request ImportNodeInformationRequest) ImportNodeInformationResponse {

	importType := services_helpers.ImportNodeInformation
	nodeSettings := cache.GetNodeSettingsByNodeId(request.NodeId)

	response := ImportNodeInformationResponse{
		Request: request,
		ImportResponse: ImportResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
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

	err = ImportNodeInfoFromLnd(ctx, lnrpc.NewLightningClient(connection), request.Db, nodeSettings)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to import node information.")
		response.Error = err
		return response
	}
	setSuccessTime(request.NodeId, successTimes, importType)
	response.Status = lightning_helpers.Active
	return response
}

func processImportPeerStatusRequest(ctx context.Context, request ImportPeerStatusRequest) ImportPeerStatusResponse {
	importType := services_helpers.ImportPeerStatus
	nodeSettings := cache.GetNodeSettingsByNodeId(request.NodeId)

	response := ImportPeerStatusResponse{
		Request: request,
		ImportResponse: ImportResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
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
	response.Status = lightning_helpers.Active
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
	request lightning_helpers.ChannelStatusUpdateRequest) lightning_helpers.ChannelStatusUpdateResponse {

	response := validateChannelStatusUpdateRequest(request)
	if response != nil {
		return *response
	}

	if !channelStatusUpdateRequestContainsUpdates(request) {
		return lightning_helpers.ChannelStatusUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Active,
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
		return lightning_helpers.ChannelStatusUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  err.Error(),
			},
			Request: request,
		}
	}
	return lightning_helpers.ChannelStatusUpdateResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Active,
		},
		Request: request,
	}
}

func constructUpdateChanStatusRequest(
	request lightning_helpers.ChannelStatusUpdateRequest) *routerrpc.UpdateChanStatusRequest {

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

func channelStatusUpdateRequestIsRepeated(
	request lightning_helpers.ChannelStatusUpdateRequest) *lightning_helpers.ChannelStatusUpdateResponse {

	secondsAgo := routingPolicyUpdateLimiterSeconds
	channelEventsFromGraph, err := graph_events.GetChannelEventFromGraph(request.Db, request.ChannelId, &secondsAgo)
	if err != nil {
		return &lightning_helpers.ChannelStatusUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
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
			return &lightning_helpers.ChannelStatusUpdateResponse{
				CommunicationResponse: lightning_helpers.CommunicationResponse{
					Status: lightning_helpers.Inactive,
				},
				Request: request,
			}
		}
	}
	return nil
}

func channelStatusUpdateRequestContainsUpdates(request lightning_helpers.ChannelStatusUpdateRequest) bool {
	channelState := cache.GetChannelState(request.NodeId, request.ChannelId, true)
	if request.ChannelStatus == core.Active && channelState.LocalDisabled {
		return true
	}
	if request.ChannelStatus == core.Inactive && !channelState.LocalDisabled {
		return true
	}
	return false
}

func validateChannelStatusUpdateRequest(
	request lightning_helpers.ChannelStatusUpdateRequest) *lightning_helpers.ChannelStatusUpdateResponse {

	if request.ChannelId == 0 {
		return &lightning_helpers.ChannelStatusUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  "ChannelId is 0",
			},
			Request: request,
		}
	}
	if request.ChannelStatus != core.Active &&
		request.ChannelStatus != core.Inactive {
		return &lightning_helpers.ChannelStatusUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  "ChannelStatus is not Active nor Inactive",
			},
			Request: request,
		}
	}
	channelSettings := cache.GetChannelSettingByChannelId(request.ChannelId)
	if channelSettings.FundingTransactionHash == nil || *channelSettings.FundingTransactionHash == "" ||
		channelSettings.FundingOutputIndex == nil {
		return &lightning_helpers.ChannelStatusUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
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
