package lightning

import (
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/cln"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/lightning_requests"
	"github.com/lncapital/torq/internal/lnd"
)

var ServiceInactiveError = errors.New("service is not active") //nolint:gochecknoglobals

func GetInformation(nodeId int) (lightning_requests.InformationResponse, error) {
	request := lightning_requests.InformationRequest{
		CommunicationRequest: lightning_requests.CommunicationRequest{
			NodeId: nodeId,
		},
	}

	response := lightning_requests.InformationResponse{
		Request: request,
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return lightning_requests.InformationResponse{}, ServiceInactiveError
		}
		response = lnd.Information(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return lightning_requests.InformationResponse{}, ServiceInactiveError
		}
		response = cln.Information(request)
	}
	if response.Error != "" {
		return lightning_requests.InformationResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func SignMessage(nodeId int, message string, singleHash *bool) (string, error) {
	request := lightning_requests.SignMessageRequest{
		CommunicationRequest: lightning_requests.CommunicationRequest{
			NodeId: nodeId,
		},
		Message:    message,
		SingleHash: singleHash,
	}

	response := lightning_requests.SignMessageResponse{
		Request: request,
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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
	request := lightning_requests.SignatureVerificationRequest{
		CommunicationRequest: lightning_requests.CommunicationRequest{
			NodeId: nodeId,
		},
		Message:   message,
		Signature: signature,
	}

	response := lightning_requests.SignatureVerificationResponse{
		Request: request,
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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

	request := lightning_requests.RoutingPolicyUpdateRequest{
		CommunicationRequest: lightning_requests.CommunicationRequest{
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

	response := lightning_requests.RoutingPolicyUpdateResponse{
		Request: request,
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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
	request := lightning_requests.ConnectPeerRequest{
		CommunicationRequest: lightning_requests.CommunicationRequest{
			NodeId: nodeId,
		},
		PublicKey: publicKey,
		Host:      host,
	}

	response := lightning_requests.ConnectPeerResponse{
		Request: request,
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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
	request := lightning_requests.DisconnectPeerRequest{
		CommunicationRequest: lightning_requests.CommunicationRequest{
			NodeId: nodeId,
		},
		PeerNodeId: peerNodeId,
	}

	response := lightning_requests.DisconnectPeerResponse{
		Request: request,
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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

	request := lightning_requests.WalletBalanceRequest{
		CommunicationRequest: lightning_requests.CommunicationRequest{
			NodeId: nodeId,
		},
	}

	response := lightning_requests.WalletBalanceResponse{
		Request: request,
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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

func ListPeers(nodeId int, latestError bool) (map[string]lightning_requests.Peer, error) {
	request := lightning_requests.ListPeersRequest{
		CommunicationRequest: lightning_requests.CommunicationRequest{
			NodeId: nodeId,
		},
		LatestError: latestError,
	}

	response := lightning_requests.ListPeersResponse{
		Request: request,
		CommunicationResponse: lightning_requests.CommunicationResponse{
			Status: lightning_requests.Inactive,
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

func ImportAllChannels(db *sqlx.DB,
	force bool,
	nodeId int) error {

	request := lnd.ImportAllChannelsRequest{
		ImportRequest: lnd.ImportRequest{
			CommunicationRequest: lnd.CommunicationRequest{
				NodeId: nodeId,
			},
			Db:    db,
			Force: force,
		},
	}
	response := lnd.ImportAllChannels(request)
	if response.Error != nil {
		return response.Error
	}
	return nil
}

func ImportPendingChannels(db *sqlx.DB,
	force bool,
	nodeId int) error {

	request := lnd.ImportPendingChannelsRequest{
		ImportRequest: lnd.ImportRequest{
			CommunicationRequest: lnd.CommunicationRequest{
				NodeId: nodeId,
			},
			Db:    db,
			Force: force,
		},
	}
	response := lnd.ImportPendingChannels(request)
	if response.Error != nil {
		return response.Error
	}
	return nil
}

func ImportChannelRoutingPolicies(db *sqlx.DB,
	force bool,
	nodeId int) error {

	request := lnd.ImportChannelRoutingPoliciesRequest{
		ImportRequest: lnd.ImportRequest{
			CommunicationRequest: lnd.CommunicationRequest{
				NodeId: nodeId,
			},
			Db:    db,
			Force: force,
		},
	}
	response := lnd.ImportChannelRoutingPolicies(request)
	if response.Error != nil {
		return response.Error
	}
	return nil
}

func ImportNodeInformation(db *sqlx.DB,
	force bool,
	nodeId int) error {

	request := lnd.ImportNodeInformationRequest{
		ImportRequest: lnd.ImportRequest{
			CommunicationRequest: lnd.CommunicationRequest{
				NodeId: nodeId,
			},
			Db:    db,
			Force: force,
		},
	}
	response := lnd.ImportNodeInformation(request)
	if response.Error != nil {
		return response.Error
	}
	return nil
}

func ImportPeerStatus(db *sqlx.DB,
	force bool,
	nodeId int) error {

	request := lnd.ImportPeerStatusRequest{
		ImportRequest: lnd.ImportRequest{
			CommunicationRequest: lnd.CommunicationRequest{
				NodeId: nodeId,
			},
			Db:    db,
			Force: force,
		},
	}
	response := lnd.ImportPeerStatus(request)
	if response.Error != nil {
		return response.Error
	}
	return nil
}
