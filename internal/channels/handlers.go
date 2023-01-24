package channels

import (
	"net/http"
	"strconv"
	"time"

	"github.com/lncapital/torq/internal/tags"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

type channelBody struct {
	NodeId                       int                  `json:"nodeId"`
	PeerNodeId                   int                  `json:"peerNodeId"`
	ChannelId                    int                  `json:"channelId"`
	ChannelPoint                 string               `json:"channelPoint"`
	NodeName                     string               `json:"nodeName"`
	Tags                         []tags.Tag           `json:"tags"`
	Active                       bool                 `json:"active"`
	RemoteActive                 bool                 `json:"remoteActive"`
	Gauge                        float64              `json:"gauge"`
	RemotePubkey                 string               `json:"remotePubkey"`
	FundingTransactionHash       string               `json:"fundingTransactionHash"`
	FundingOutputIndex           int                  `json:"fundingOutputIndex"`
	FundingBlockHeight           *int64               `json:"fundingBlockHeight"`
	FundedOn                     *time.Time           `json:"fundedOn"`
	ClosingBlockHeight           *int64               `json:"closingBlockHeight"`
	ClosedOn                     *time.Time           `json:"closedOn"`
	LNDShortChannelId            string               `json:"lndShortChannelId"`
	ShortChannelId               string               `json:"shortChannelId"`
	Capacity                     int64                `json:"capacity"`
	LocalBalance                 int64                `json:"localBalance"`
	RemoteBalance                int64                `json:"remoteBalance"`
	UnsettledBalance             int64                `json:"unsettledBalance"`
	CommitFee                    int64                `json:"commitFee"`
	CommitWeight                 int64                `json:"commitWeight"`
	FeePerKw                     int64                `json:"feePerKw"`
	FeeBaseMsat                  uint64               `json:"feeBaseMsat"`
	MinHtlcMsat                  uint64               `json:"minHtlcMsat"`
	MaxHtlcMsat                  uint64               `json:"maxHtlcMsat"`
	TimeLockDelta                uint32               `json:"timeLockDelta"`
	FeeRateMilliMsat             uint64               `json:"feeRateMilliMsat"`
	RemoteFeeBaseMsat            uint64               `json:"remoteFeeBaseMsat"`
	RemoteMinHtlcMsat            uint64               `json:"remoteMinHtlcMsat"`
	RemoteMaxHtlcMsat            uint64               `json:"remoteMaxHtlcMsat"`
	RemoteTimeLockDelta          uint32               `json:"remoteTimeLockDelta"`
	RemoteFeeRateMilliMsat       uint64               `json:"remoteFeeRateMilliMsat"`
	PendingForwardingHTLCsCount  int                  `json:"pendingForwardingHTLCsCount"`
	PendingForwardingHTLCsAmount int64                `json:"pendingForwardingHTLCsAmount"`
	PendingLocalHTLCsCount       int                  `json:"pendingLocalHTLCsCount"`
	PendingLocalHTLCsAmount      int64                `json:"pendingLocalHTLCsAmount"`
	PendingTotalHTLCsCount       int                  `json:"pendingTotalHTLCsCount"`
	PendingTotalHTLCsAmount      int64                `json:"pendingTotalHTLCsAmount"`
	TotalSatoshisSent            int64                `json:"totalSatoshisSent"`
	NumUpdates                   uint64               `json:"numUpdates"`
	Initiator                    bool                 `json:"initiator"`
	ChanStatusFlags              string               `json:"chanStatusFlags"`
	LocalChanReserveSat          int64                `json:"localChanReserveSat"`
	RemoteChanReserveSat         int64                `json:"remoteChanReserveSat"`
	CommitmentType               lnrpc.CommitmentType `json:"commitmentType"`
	Lifetime                     int64                `json:"lifetime"`
	TotalSatoshisReceived        int64                `json:"totalSatoshisReceived"`
	MempoolSpace                 string               `json:"mempoolSpace"`
	AmbossSpace                  string               `json:"ambossSpace"`
	OneMl                        string               `json:"oneMl"`
	PeerAlias                    string               `json:"peerAlias"`
}

type PendingHtlcs struct {
	ForwardingCount  int   `json:"forwardingCount"`
	ForwardingAmount int64 `json:"forwardingAmount"`
	LocalCount       int   `json:"localCount"`
	LocalAmount      int64 `json:"localAmount"`
	TotalCount       int   `json:"toalCount"`
	TotalAmount      int64 `json:"totalAmount"`
}

type ChannelPolicy struct {
	Disabled        bool   `json:"disabled" db:"disabled"`
	TimeLockDelta   uint32 `json:"timeLockDelta" db:"time_lock_delta"`
	MinHtlcMsat     uint64 `json:"minHtlcMsat" db:"min_htlc"`
	MaxHtlcMsat     uint64 `json:"maxHtlcMsat" db:"max_htlc_msat"`
	FeeRateMillMsat uint64 `json:"feeRateMillMsat" db:"fee_rate_mill_msat"`
	ShortChannelId  string `json:"shortChannelId" db:"short_channel_id"`
	FeeBaseMsat     uint64 `json:"feeBaseMsat" db:"fee_base_msat"`
	NodeId          int    `json:"nodeId" db:"node_id"`
	RemoteNodeId    int    `json:"RemoteodeId" db:"remote_node_id"`
}

type ChannelsNodes struct {
	Channels []ChannelForTag `json:"channels"`
	Nodes    []NodeForTag    `json:"nodes"`
}

type ChannelForTag struct {
	ShortChannelId *string `json:"shortChannelId" db:"short_channel_id"`
	ChannelId      int     `json:"channelId" db:"channel_id"`
	NodeId         int     `json:"nodeId" db:"node_id"`
	Alias          *string `json:"alias" db:"alias"`
	Type           string  `json:"type" db:"type"`
}

type NodeForTag struct {
	NodeId int    `json:"nodeId" db:"node_id"`
	Alias  string `json:"alias" db:"alias"`
	Type   string `json:"type" db:"type"`
}

func updateChannelsHandler(c *gin.Context, lightningRequestChannel chan interface{}) {
	var requestBody commons.RoutingPolicyUpdateRequest
	if err := c.BindJSON(&requestBody); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	response, err := routingPolicyUpdate(requestBody, lightningRequestChannel)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Update channel/s policy")
		return
	}

	c.JSON(http.StatusOK, response)
}

func batchOpenHandler(c *gin.Context, db *sqlx.DB) {
	var batchOpnReq commons.BatchOpenRequest
	if err := c.BindJSON(&batchOpnReq); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	response, err := batchOpenChannels(db, batchOpnReq)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Batch open channels")
		return
	}

	c.JSON(http.StatusOK, response)
}

func getChannelListHandler(c *gin.Context, db *sqlx.DB) {
	var channelsBody []channelBody

	network, err := strconv.Atoi(c.Query("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
		return
	}

	chain := commons.Bitcoin

	nodeIds := commons.GetAllTorqNodeIds(chain, commons.Network(network))

	if len(nodeIds) != 0 {
		for _, nodeId := range nodeIds {
			// Force Response because we don't care about balance accuracy
			channelBalanceStates := commons.GetChannelStates(nodeId, true)
			nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
			for _, channel := range channelBalanceStates {
				channelSettings := commons.GetChannelSettingByChannelId(channel.ChannelId)
				lndShortChannelIdString := strconv.FormatUint(channelSettings.LndShortChannelId, 10)

				pendingHTLCs := calculateHTLCs(channel.PendingHtlcs)

				gauge := (float64(channel.LocalBalance) / float64(channelSettings.Capacity)) * 100

				channelTags, err := tags.GetChannelTags(db, tags.ChannelTagsRequest{
					ChannelId: channelSettings.ChannelId,
					NodeId:    &channel.RemoteNodeId,
				})
				if err != nil {
					server_errors.WrapLogAndSendServerError(c, err, "Get channel tags for channel")
					return
				}

				remoteNode := commons.GetNodeSettingsByNodeId(channel.RemoteNodeId)
				chanBody := channelBody{
					NodeId:                       nodeId,
					PeerNodeId:                   channel.RemoteNodeId,
					Tags:                         channelTags,
					ChannelId:                    channelSettings.ChannelId,
					NodeName:                     *nodeSettings.Name,
					Active:                       !channel.LocalDisabled,
					RemoteActive:                 !channel.RemoteDisabled,
					ChannelPoint:                 commons.CreateChannelPoint(channelSettings.FundingTransactionHash, channelSettings.FundingOutputIndex),
					Gauge:                        gauge,
					RemotePubkey:                 remoteNode.PublicKey,
					FundingTransactionHash:       channelSettings.FundingTransactionHash,
					FundingOutputIndex:           channelSettings.FundingOutputIndex,
					FundingBlockHeight:           channelSettings.FundingBlockHeight,
					FundedOn:                     channelSettings.FundedOn,
					ClosingBlockHeight:           channelSettings.ClosingBlockHeight,
					ClosedOn:                     channelSettings.ClosedOn,
					LNDShortChannelId:            lndShortChannelIdString,
					ShortChannelId:               channelSettings.ShortChannelId,
					Capacity:                     channelSettings.Capacity,
					LocalBalance:                 channel.LocalBalance,
					RemoteBalance:                channel.RemoteBalance,
					UnsettledBalance:             channel.UnsettledBalance,
					TotalSatoshisSent:            channel.TotalSatoshisSent,
					TotalSatoshisReceived:        channel.TotalSatoshisReceived,
					PendingForwardingHTLCsCount:  pendingHTLCs.ForwardingCount,
					PendingForwardingHTLCsAmount: pendingHTLCs.ForwardingAmount,
					PendingLocalHTLCsCount:       pendingHTLCs.LocalCount,
					PendingLocalHTLCsAmount:      pendingHTLCs.LocalAmount,
					PendingTotalHTLCsCount:       pendingHTLCs.TotalCount,
					PendingTotalHTLCsAmount:      pendingHTLCs.TotalAmount,
					CommitFee:                    channel.CommitFee,
					CommitWeight:                 channel.CommitWeight,
					FeePerKw:                     channel.FeePerKw,
					FeeBaseMsat:                  channel.LocalFeeBaseMsat,
					MinHtlcMsat:                  channel.LocalMinHtlcMsat,
					MaxHtlcMsat:                  channel.LocalMaxHtlcMsat,
					TimeLockDelta:                channel.LocalTimeLockDelta,
					FeeRateMilliMsat:             channel.LocalFeeRateMilliMsat,
					RemoteFeeBaseMsat:            channel.RemoteFeeBaseMsat,
					RemoteMinHtlcMsat:            channel.RemoteMinHtlcMsat,
					RemoteMaxHtlcMsat:            channel.RemoteMaxHtlcMsat,
					RemoteTimeLockDelta:          channel.RemoteTimeLockDelta,
					RemoteFeeRateMilliMsat:       channel.RemoteFeeRateMilliMsat,
					NumUpdates:                   channel.NumUpdates,
					Initiator:                    channelSettings.InitiatingNodeId != nil && *channelSettings.InitiatingNodeId == nodeId,
					ChanStatusFlags:              channel.ChanStatusFlags,
					CommitmentType:               channel.CommitmentType,
					Lifetime:                     channel.Lifetime,
					MempoolSpace:                 commons.MEMPOOL + lndShortChannelIdString,
					AmbossSpace:                  commons.AMBOSS + channelSettings.ShortChannelId,
					OneMl:                        commons.ONEML + lndShortChannelIdString,
				}

				peerInfo, err := GetNodePeerAlias(nodeId, channel.RemoteNodeId, db)
				if err == nil {
					chanBody.PeerAlias = peerInfo
				} else {
					log.Error().Err(err).Msgf("Could not obtain the alias of the peer with nodeId: %v (for Torq nodeId: %v)", channel.RemoteNodeId, nodeId)
				}
				channelsBody = append(channelsBody, chanBody)
			}
		}
	}
	c.JSON(http.StatusOK, channelsBody)
}

func calculateHTLCs(htlcs []commons.Htlc) PendingHtlcs {
	var pendingHTLCs PendingHtlcs
	if len(htlcs) < 1 {
		return pendingHTLCs
	}
	for _, htlc := range htlcs {
		if htlc.ForwardingHtlcIndex == 0 {
			pendingHTLCs.LocalCount++
			pendingHTLCs.LocalAmount += htlc.Amount
		} else {
			pendingHTLCs.ForwardingCount++
			pendingHTLCs.ForwardingAmount += htlc.Amount
		}
	}
	pendingHTLCs.TotalAmount = pendingHTLCs.ForwardingAmount + pendingHTLCs.LocalAmount
	pendingHTLCs.TotalCount = pendingHTLCs.ForwardingCount + pendingHTLCs.LocalCount

	return pendingHTLCs
}

func getChannelAndNodeListHandler(c *gin.Context, db *sqlx.DB) {
	channels, err := GetChannelsForTag(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "List channels")
		return
	}

	nodes, err := GetNodesForTag(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "List nodes")
		return
	}

	nodesChannels := ChannelsNodes{
		Channels: channels,
		Nodes:    nodes,
	}

	c.JSON(http.StatusOK, nodesChannels)
}
