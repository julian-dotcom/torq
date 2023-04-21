package lightning

import (
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/cln"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/lightning_helpers"
	"github.com/lncapital/torq/internal/lnd"
)

var ServiceInactiveError = errors.New("service is not active")         //nolint:gochecknoglobals
var UnsupportedOperationError = errors.New("request is not supported") //nolint:gochecknoglobals

func GetInformation(nodeId int) (lightning_helpers.InformationResponse, error) {
	request := lightning_helpers.InformationRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
	}

	response := lightning_helpers.InformationResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return lightning_helpers.InformationResponse{}, ServiceInactiveError
		}
		response = lnd.Information(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return lightning_helpers.InformationResponse{}, ServiceInactiveError
		}
		response = cln.Information(request)
	}
	if response.Error != "" {
		return lightning_helpers.InformationResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func SignMessage(nodeId int, message string, singleHash *bool) (string, error) {
	request := lightning_helpers.SignMessageRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		Message:    message,
		SingleHash: singleHash,
	}

	response := lightning_helpers.SignMessageResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return "", ServiceInactiveError
		}
		response = lnd.SignMessage(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return "", ServiceInactiveError
		}
		response = cln.SignMessage(request)
	}
	if response.Error != "" {
		return "", errors.New(response.Error)
	}
	return response.Signature, nil
}

func SignatureVerification(nodeId int, message string, signature string) (string, bool, error) {
	request := lightning_helpers.SignatureVerificationRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		Message:   message,
		Signature: signature,
	}

	response := lightning_helpers.SignatureVerificationResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return "", false, ServiceInactiveError
		}
		response = lnd.SignatureVerification(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return "", false, ServiceInactiveError
		}
		response = cln.SignatureVerification(request)
	}
	if response.Error != "" {
		return "", false, errors.New(response.Error)
	}
	return response.PublicKey, response.Valid, nil
}

func SetRoutingPolicy(db *sqlx.DB,
	nodeId int,
	rateLimitSeconds int,
	rateLimitCount int,
	channelId int,
	feeRateMilliMsat *int64,
	feeBaseMsat *int64,
	maxHtlcMsat *uint64,
	minHtlcMsat *uint64,
	timeLockDelta *uint32) (core.Status, string, error) {

	request := lightning_helpers.RoutingPolicyUpdateRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		Db:               db,
		RateLimitSeconds: rateLimitSeconds,
		RateLimitCount:   rateLimitCount,
		ChannelId:        channelId,
		FeeRateMilliMsat: feeRateMilliMsat,
		FeeBaseMsat:      feeBaseMsat,
		MaxHtlcMsat:      maxHtlcMsat,
		MinHtlcMsat:      minHtlcMsat,
		TimeLockDelta:    timeLockDelta,
	}

	response := lightning_helpers.RoutingPolicyUpdateResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return core.Inactive, "", ServiceInactiveError
		}
		response = lnd.RoutingPolicyUpdate(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return core.Inactive, "", ServiceInactiveError
		}
		response = cln.RoutingPolicyUpdate(request)
	}
	if response.Error != "" {
		return core.Status(response.Status), response.Message, errors.New(response.Error)
	}
	return core.Status(response.Status), response.Message, nil
}

func ConnectPeer(nodeId int, publicKey string, host string) (bool, error) {
	request := lightning_helpers.ConnectPeerRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		PublicKey: publicKey,
		Host:      host,
	}

	response := lightning_helpers.ConnectPeerResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return false, ServiceInactiveError
		}
		response = lnd.ConnectPeer(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return false, ServiceInactiveError
		}
		response = cln.ConnectPeer(request)
	}
	if response.Error != "" {
		return false, errors.New(response.Error)
	}
	return response.RequestFailCurrentlyConnected, nil
}

func DisconnectPeer(nodeId int, peerNodeId int) (bool, error) {
	request := lightning_helpers.DisconnectPeerRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		PeerNodeId: peerNodeId,
	}

	response := lightning_helpers.DisconnectPeerResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return false, ServiceInactiveError
		}
		response = lnd.DisconnectPeer(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return false, ServiceInactiveError
		}
		response = cln.DisconnectPeer(request)
	}
	if response.Error != "" {
		return false, errors.New(response.Error)
	}
	return response.RequestFailedCurrentlyDisconnected, nil
}

func GetWalletBalance(nodeId int) (totalBalance int64, confirmedBalance int64, unconfirmedBalance int64,
	lockedBalance int64, reservedBalanceAnchorChan int64, err error) {

	request := lightning_helpers.WalletBalanceRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
	}

	response := lightning_helpers.WalletBalanceResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return 0, 0, 0, 0, 0, ServiceInactiveError
		}
		response = lnd.WalletBalance(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return 0, 0, 0, 0, 0, ServiceInactiveError
		}
		response = cln.WalletBalance(request)
	}
	if response.Error != "" {
		return 0, 0, 0, 0, 0, errors.New(response.Error)
	}
	return response.TotalBalance, response.ConfirmedBalance, response.UnconfirmedBalance, response.LockedBalance,
		response.ReservedBalanceAnchorChan, nil
}

func ListPeers(nodeId int, latestError bool) (map[string]lightning_helpers.Peer, error) {
	request := lightning_helpers.ListPeersRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		LatestError: latestError,
	}

	response := lightning_helpers.ListPeersResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return nil, ServiceInactiveError
		}
		response = lnd.ListPeers(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return nil, ServiceInactiveError
		}
		response = cln.ListPeers(request)
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}
	return response.Peers, nil
}

func NewAddress(request lightning_helpers.NewAddressRequest) (string, error) {
	response := lightning_helpers.NewAddressResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return "", ServiceInactiveError
		}
		response = lnd.NewAddress(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return "", ServiceInactiveError
		}
		response = cln.NewAddress(request)
	}
	if response.Error != "" {
		return "", errors.New(response.Error)
	}
	return response.Address, nil
}

func OpenChannel(request lightning_helpers.OpenChannelRequest) (lightning_helpers.OpenChannelResponse, error) {
	response := lightning_helpers.OpenChannelResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.OpenChannelResponse{}, ServiceInactiveError
		}
		response = lnd.OpenChannel(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.OpenChannelResponse{}, ServiceInactiveError
		}
		response = cln.OpenChannel(request)
	}
	if response.Error != "" {
		return lightning_helpers.OpenChannelResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func BatchOpenChannel(request lightning_helpers.BatchOpenChannelRequest) (lightning_helpers.BatchOpenChannelResponse, error) {
	response := lightning_helpers.BatchOpenChannelResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.BatchOpenChannelResponse{}, ServiceInactiveError
		}
		response = lnd.BatchOpenChannel(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.BatchOpenChannelResponse{}, ServiceInactiveError
		}
		return lightning_helpers.BatchOpenChannelResponse{}, UnsupportedOperationError
	}
	if response.Error != "" {
		return lightning_helpers.BatchOpenChannelResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func CloseChannel(request lightning_helpers.CloseChannelRequest) (lightning_helpers.CloseChannelResponse, error) {
	response := lightning_helpers.CloseChannelResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.CloseChannelResponse{}, ServiceInactiveError
		}
		response = lnd.CloseChannel(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.CloseChannelResponse{}, ServiceInactiveError
		}
		response = cln.CloseChannel(request)
	}
	if response.Error != "" {
		return lightning_helpers.CloseChannelResponse{}, errors.New(response.Error)
	}
	return response, nil
}
