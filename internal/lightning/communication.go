package lightning

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/lnd"
)

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

	if !cache.IsLndServiceActive(nodeId) {
		return core.Inactive, "",
			errors.New(fmt.Sprintf("LND service is not active for nodeId: %v", nodeId))
	}
	request := lnd.RoutingPolicyUpdateRequest{
		CommunicationRequest: lnd.CommunicationRequest{
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
	response := lnd.RoutingPolicyUpdate(request)
	if response.Error != "" {
		return core.Status(response.Status), response.Message, errors.New(response.Error)
	}
	return core.Status(response.Status), response.Message, nil
}

func SignMessage(nodeId int, message string, singleHash *bool) (string, error) {
	if !cache.IsLndServiceActive(nodeId) {
		return "", errors.New(fmt.Sprintf("LND service is not active for nodeId: %v", nodeId))
	}

	request := lnd.SignMessageRequest{
		CommunicationRequest: lnd.CommunicationRequest{
			NodeId: nodeId,
		},
		Message:    message,
		SingleHash: singleHash,
	}
	response := lnd.SignMessage(request)
	if response.Error != "" {
		return "", errors.New(response.Error)
	}
	return response.Signature, nil
}

func SignatureVerification(nodeId int, message string, signature string) (string, bool, error) {
	request := lnd.SignatureVerificationRequest{
		CommunicationRequest: lnd.CommunicationRequest{
			NodeId: nodeId,
		},
		Message:   message,
		Signature: signature,
	}
	response := lnd.SignatureVerification(request)
	if response.Error != "" {
		return "", false, errors.New(response.Error)
	}
	return response.PublicKey, response.Valid, nil
}

func GetWalletBalance(nodeId int) (totalBalance int64, confirmedBalance int64, unconfirmedBalance int64,
	lockedBalance int64, reservedBalanceAnchorChan int64, err error) {

	request := lnd.WalletBalanceRequest{
		CommunicationRequest: lnd.CommunicationRequest{
			NodeId: nodeId,
		},
	}
	response := lnd.WalletBalance(request)
	if response.Error != "" {
		return 0, 0, 0, 0, 0, errors.New(response.Error)
	}
	return response.TotalBalance, response.ConfirmedBalance, response.UnconfirmedBalance, response.LockedBalance,
		response.ReservedBalanceAnchorChan, nil
}

func GetInformationRequest(nodeId int) (core.InformationResponse, error) {
	request := lnd.InformationRequest{
		CommunicationRequest: lnd.CommunicationRequest{
			NodeId: nodeId,
		},
	}
	response := lnd.Information(request)
	if response.Error != "" {
		return core.InformationResponse{}, errors.New(response.Error)
	}
	return core.InformationResponse{
		NodeId:                  nodeId,
		Implementation:          core.LND,
		Version:                 response.Version,
		PublicKey:               response.PublicKey,
		Alias:                   response.Alias,
		Color:                   response.Color,
		PendingChannelCount:     response.PendingChannelCount,
		ActiveChannelCount:      response.ActiveChannelCount,
		InactiveChannelCount:    response.InactiveChannelCount,
		PeerCount:               response.PeerCount,
		BlockHeight:             response.BlockHeight,
		BlockHash:               response.BlockHash,
		BestHeaderTimestamp:     response.BestHeaderTimestamp,
		ChainSynced:             response.ChainSynced,
		GraphSynced:             response.GraphSynced,
		Addresses:               response.Addresses,
		HtlcInterceptorRequired: response.HtlcInterceptorRequired,
	}, nil
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

func ConnectPeer(nodeId int, publicKey string, host string) (bool, error) {
	request := lnd.ConnectPeerRequest{
		CommunicationRequest: lnd.CommunicationRequest{
			NodeId: nodeId,
		},
		PublicKey: publicKey,
		Host:      host,
	}
	response := lnd.ConnectPeer(request)
	if response.Error != nil {
		return false, response.Error
	}
	return response.RequestFailCurrentlyConnected, nil
}

func DisconnectPeer(nodeId int, peerNodeId int) (bool, error) {
	request := lnd.DisconnectPeerRequest{
		CommunicationRequest: lnd.CommunicationRequest{
			NodeId: nodeId,
		},
		PeerNodeId: peerNodeId,
	}
	response := lnd.DisconnectPeer(request)
	if response.Error != nil {
		return false, response.Error
	}
	return response.RequestFailedCurrentlyDisconnected, nil
}

func ListPeers(nodeId int, latestError bool) (map[string]core.Peer, error) {
	request := lnd.ListPeersRequest{
		CommunicationRequest: lnd.CommunicationRequest{
			NodeId: nodeId,
		},
		NodeId:      nodeId,
		LatestError: latestError,
	}
	response := lnd.ListPeers(request)
	if response.Error != nil {
		return nil, response.Error
	}
	return response.Peers, nil
}
