package cln

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/proto/cln"
)

const streamPeersTickerSeconds = 60

type client_ListPeers interface {
	ListPeers(ctx context.Context, in *cln.ListpeersRequest, opts ...grpc.CallOption) (*cln.ListpeersResponse, error)
}

func SubscribeAndStorePeers(ctx context.Context,
	client client_ListPeers,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.ClnServicePeersService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamPeersTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerPeers := ticker.C

	err := listAndProcessPeers(ctx, db, client, serviceType, nodeSettings, true)
	if err != nil {
		processError(ctx, serviceType, nodeSettings, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		case <-tickerPeers:
			err = listAndProcessPeers(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func processError(ctx context.Context, serviceType services_helpers.ServiceType, nodeSettings cache.NodeSettingsCache, err error) {
	if errors.Is(ctx.Err(), context.Canceled) {
		cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
		return
	}
	cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
	log.Error().Err(err).Msgf("%v failed to process for nodeId: %v", serviceType.String(), nodeSettings.NodeId)
}

func listAndProcessPeers(ctx context.Context, db *sqlx.DB, client client_ListPeers,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	peers, err := client.ListPeers(ctx, &cln.ListpeersRequest{})
	if err != nil {
		return errors.Wrapf(err, "listing peers for nodeId: %v", nodeSettings.NodeId)
	}

	err = storePeers(db, peers.Peers, nodeSettings)
	if err != nil {
		return errors.Wrapf(err, "storing peers for nodeId: %v", nodeSettings.NodeId)
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of peers is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func storePeers(db *sqlx.DB, peers []*cln.ListpeersPeers, nodeSettings cache.NodeSettingsCache) error {
	for _, peer := range peers {
		peerPublicKey := hex.EncodeToString(peer.Id)
		peerNodeId := cache.GetPeerNodeIdByPublicKey(peerPublicKey, nodeSettings.Chain, nodeSettings.Network)
		if peerNodeId == 0 {
			var err error
			peerNodeId, err = nodes.AddNodeWhenNew(db, nodes.Node{
				PublicKey: peerPublicKey,
				Chain:     nodeSettings.Chain,
				Network:   nodeSettings.Network,
			}, nil)
			if err != nil {
				return errors.Wrapf(err, "add new peer node for nodeId: %v", nodeSettings.NodeId)
			}
		}
		address, setting, connectionStatus, err := settings.GetNodeConnectionHistoryWithDetail(db, nodeSettings.NodeId, peerNodeId)
		if err != nil {
			return errors.Wrapf(err, "obtaining node connection history for peerNodeId: %v, nodeId: %v",
				peerNodeId, nodeSettings.NodeId)
		}
		if address == nil && peer.RemoteAddr != nil {
			remoteAddress := *peer.RemoteAddr
			address = &remoteAddress
		}
		//if address == nil && peer.Netaddr != nil {
		//}
		if peer.Connected && (connectionStatus == nil || *connectionStatus != core.NodeConnectionStatusConnected) {
			connected := core.NodeConnectionStatusConnected
			err = settings.AddNodeConnectionHistory(db, nodeSettings.NodeId, peerNodeId, address, setting, &connected)
			if err != nil {
				return errors.Wrapf(err, "add new node connection history for nodeId: %v", nodeSettings.NodeId)
			}
		}
		if !peer.Connected && (connectionStatus == nil || *connectionStatus != core.NodeConnectionStatusDisconnected) {
			disconnected := core.NodeConnectionStatusDisconnected
			err = settings.AddNodeConnectionHistory(db, nodeSettings.NodeId, peerNodeId, address, setting, &disconnected)
			if err != nil {
				return errors.Wrapf(err, "add new node disconnection history for nodeId: %v", nodeSettings.NodeId)
			}
		}

		// TODO FIXME THIS IS NOT WORKING ON CLN RIGHT NOW THERE IS A PENDING GITHUB ISSUE FOR IT.
		//for _, clnChannel := range peer.Channels {
		//	if clnChannel != nil {
		//		_, err = processPeerChannel(db, clnChannel, nodeSettings, peerNodeId)
		//		if err != nil {
		//			return errors.Wrapf(err, "add new node disconnection history for nodeId: %v", nodeSettings.NodeId)
		//		}
		//	}
		//}
	}
	return nil
}

//func processPeerChannel(db *sqlx.DB,
//	clnChannel *cln.ListpeersPeersChannels,
//	nodeSettings cache.NodeSettingsCache,
//	peerNodeId int) (int, error) {
//
//	var fundingOutputIndex *int
//	if clnChannel.FundingOutnum != nil {
//		foi := int(*clnChannel.FundingOutnum)
//		fundingOutputIndex = &foi
//	}
//	var fundingTransactionHash *string
//	if len(clnChannel.FundingTxid) != 0 {
//		fti := hex.EncodeToString(clnChannel.FundingTxid)
//		fundingTransactionHash = &fti
//	}
//	var channelId int
//	if clnChannel.ShortChannelId != nil {
//		channelId = cache.GetChannelIdByShortChannelId(clnChannel.ShortChannelId)
//	}
//	if channelId == 0 && len(clnChannel.FundingTxid) != 0 {
//		channelId = cache.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
//	}
//	var channel channels.Channel
//	if channelId == 0 {
//		channel = channels.Channel{
//			FundingTransactionHash: fundingTransactionHash,
//			FundingOutputIndex:     fundingOutputIndex,
//			FirstNodeId:            nodeSettings.NodeId,
//			SecondNodeId:           peerNodeId,
//		}
//	} else {
//		channelSettings := cache.GetChannelSettingByChannelId(channelId)
//		channel = channels.Channel{
//			ChannelID:              channelSettings.ChannelId,
//			ShortChannelID:         channelSettings.ShortChannelId,
//			FundingTransactionHash: fundingTransactionHash,
//			FundingOutputIndex:     fundingOutputIndex,
//			ClosingTransactionHash: channelSettings.ClosingTransactionHash,
//			Capacity:               channelSettings.Capacity,
//			Private:                channelSettings.Private,
//			FirstNodeId:            channelSettings.FirstNodeId,
//			SecondNodeId:           channelSettings.SecondNodeId,
//			InitiatingNodeId:       channelSettings.InitiatingNodeId,
//			AcceptingNodeId:        channelSettings.AcceptingNodeId,
//			ClosingNodeId:          channelSettings.ClosingNodeId,
//			Status:                 channelSettings.Status,
//			FundingBlockHeight:     channelSettings.FundingBlockHeight,
//			FundedOn:               channelSettings.FundedOn,
//			ClosingBlockHeight:     channelSettings.ClosingBlockHeight,
//			ClosedOn:               channelSettings.ClosedOn,
//			Flags:                  channelSettings.Flags,
//		}
//	}
//	if clnChannel.ShortChannelId != nil {
//		shortChannelId := *clnChannel.ShortChannelId
//		channel.ShortChannelID = &shortChannelId
//	}
//	if clnChannel.CloseToAddr != nil {
//		closeToAddr := *clnChannel.CloseToAddr
//		channel.ClosingTransactionHash = &closeToAddr
//	}
//	if clnChannel.TotalMsat != nil {
//		channel.Capacity = int64(clnChannel.TotalMsat.Msat / 1_000)
//	}
//
//	//"OPENINGD": The channel funding protocol with the peer is ongoing and both sides are negotiating parameters.
//	//"CHANNELD_AWAITING_LOCKIN": The peer and you have agreed on channel parameters and are just waiting for the channel funding transaction to be confirmed deeply. Both you and the peer must acknowledge the channel funding transaction to be confirmed deeply before entering the next state.
//	//"CHANNELD_NORMAL": The channel can be used for normal payments.
//	//"CHANNELD_SHUTTING_DOWN": A mutual close was requested (by you or peer) and both of you are waiting for HTLCs in-flight to be either failed or succeeded. The channel can no longer be used for normal payments and forwarding. Mutual close will proceed only once all HTLCs in the channel have either been fulfilled or failed.
//	//"CLOSINGD_SIGEXCHANGE": You and the peer are negotiating the mutual close onchain fee.
//	//"CLOSINGD_COMPLETE": You and the peer have agreed on the mutual close onchain fee and are awaiting the mutual close getting confirmed deeply.
//	//"AWAITING_UNILATERAL": You initiated a unilateral close, and are now waiting for the peer-selected unilateral close timeout to complete.
//	//"FUNDING_SPEND_SEEN": You saw the funding transaction getting spent (usually the peer initiated a unilateral close) and will now determine what exactly happened (i.e. if it was a theft attempt).
//	//"ONCHAIN": You saw the funding transaction getting spent and now know what happened (i.e. if it was a proper unilateral close by the peer, or a theft attempt).
//	//"CLOSED": The channel closure has been confirmed deeply. The channel will eventually be removed from this array.
//	switch clnChannel.State {
//	case cln.ListpeersPeersChannels_OPENINGD,
//		cln.ListpeersPeersChannels_CHANNELD_AWAITING_LOCKIN,
//		cln.ListpeersPeersChannels_DUALOPEND_OPEN_INIT,
//		cln.ListpeersPeersChannels_DUALOPEND_AWAITING_LOCKIN:
//		channel.Status = core.Opening
//	case cln.ListpeersPeersChannels_CHANNELD_NORMAL:
//		channel.Status = core.Open
//	case cln.ListpeersPeersChannels_CHANNELD_SHUTTING_DOWN,
//		cln.ListpeersPeersChannels_ONCHAIN:
//		channel.Status = core.Closing
//	case cln.ListpeersPeersChannels_CLOSINGD_SIGEXCHANGE:
//	case cln.ListpeersPeersChannels_CLOSINGD_COMPLETE:
//		channel.Status = core.CooperativeClosed
//	case cln.ListpeersPeersChannels_FUNDING_SPEND_SEEN:
//		channel.Status = core.FundingCancelledClosed
//	case cln.ListpeersPeersChannels_AWAITING_UNILATERAL:
//		channel.Status = core.LocalForceClosed
//		//RemoteForceClosed      = 102
//		//BreachClosed           = 103
//		//AbandonedClosed        = 105
//	}
//	if clnChannel.Private != nil {
//		channel.Private = *clnChannel.Private
//	}
//	if clnChannel.CloseTo != nil {
//		closeTo := hex.EncodeToString(clnChannel.CloseTo)
//		channel.ClosingTransactionHash = &closeTo
//	}
//	if clnChannel.Closer != nil {
//		switch *clnChannel.Closer {
//		case cln.ChannelSide_IN:
//			channel.ClosingNodeId = &peerNodeId
//		case cln.ChannelSide_OUT:
//			channel.ClosingNodeId = &nodeSettings.NodeId
//		}
//	}
//	switch clnChannel.Opener {
//	case cln.ChannelSide_IN:
//		channel.InitiatingNodeId = &peerNodeId
//		channel.AcceptingNodeId = &nodeSettings.NodeId
//	case cln.ChannelSide_OUT:
//		channel.InitiatingNodeId = &nodeSettings.NodeId
//		channel.AcceptingNodeId = &peerNodeId
//	}
//	channelId, err := channels.AddChannelOrUpdateChannelStatus(db, nodeSettings, channel)
//	if err != nil {
//		return 0, errors.Wrapf(err, "update channel data for channelId: %v, nodeId: %v",
//			channelId, nodeSettings.NodeId)
//	}
//	return channelId, nil
//}
