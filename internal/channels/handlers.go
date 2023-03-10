package channels

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/internal/tags"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

type ChannelBody struct {
	NodeId                       int                  `json:"nodeId"`
	PeerNodeId                   int                  `json:"peerNodeId"`
	ChannelId                    int                  `json:"channelId"`
	ChannelPoint                 string               `json:"channelPoint"`
	NodeName                     string               `json:"nodeName"`
	Tags                         []tags.Tag           `json:"tags"`
	Active                       bool                 `json:"active"`
	RemoteActive                 bool                 `json:"remoteActive"`
	CurrentBlockHeight           uint32               `json:"currentBlockHeight"`
	Gauge                        float64              `json:"gauge"`
	RemotePubkey                 string               `json:"remotePubkey"`
	FundingTransactionHash       string               `json:"fundingTransactionHash"`
	FundingOutputIndex           int                  `json:"fundingOutputIndex"`
	FundingBlockHeight           *uint32              `json:"fundingBlockHeight"`
	FundingBlockHeightDelta      *uint32              `json:"fundingBlockHeightDelta"`
	FundedOn                     *time.Time           `json:"fundedOn"`
	FundedOnSecondsDelta         *uint64              `json:"fundedOnSecondsDelta"`
	ClosingBlockHeight           *uint32              `json:"closingBlockHeight"`
	ClosingBlockHeightDelta      *uint32              `json:"closingBlockHeightDelta"`
	ClosedOn                     *time.Time           `json:"closedOn"`
	ClosedOnSecondsDelta         *uint64              `json:"closedOnSecondsDelta"`
	LNDShortChannelId            string               `json:"lndShortChannelId"`
	ShortChannelId               string               `json:"shortChannelId"`
	Capacity                     int64                `json:"capacity"`
	PeerChannelCapacity          int64                `json:"peerChannelCapacity"`
	PeerChannelCount             int                  `json:"peerChannelCount"`
	PeerLocalBalance             int64                `json:"peerLocalBalance"`
	PeerGauge                    float64              `json:"peerGauge"`
	LocalBalance                 int64                `json:"localBalance"`
	RemoteBalance                int64                `json:"remoteBalance"`
	UnsettledBalance             int64                `json:"unsettledBalance"`
	CommitFee                    int64                `json:"commitFee"`
	CommitWeight                 int64                `json:"commitWeight"`
	FeePerKw                     int64                `json:"feePerKw"`
	FeeBase                      int64                `json:"feeBase"`
	MinHtlc                      uint64               `json:"minHtlc"`
	MaxHtlc                      uint64               `json:"maxHtlc"`
	TimeLockDelta                uint32               `json:"timeLockDelta"`
	FeeRateMilliMsat             int64                `json:"feeRateMilliMsat"`
	RemoteFeeBase                int64                `json:"remoteFeeBase"`
	RemoteMinHtlc                uint64               `json:"remoteMinHtlc"`
	RemoteMaxHtlc                uint64               `json:"remoteMaxHtlc"`
	RemoteTimeLockDelta          uint32               `json:"remoteTimeLockDelta"`
	RemoteFeeRateMilliMsat       int64                `json:"remoteFeeRateMilliMsat"`
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
	FeeRateMillMsat int64  `json:"feeRateMillMsat" db:"fee_rate_mill_msat"`
	ShortChannelId  string `json:"shortChannelId" db:"short_channel_id"`
	FeeBaseMsat     int64  `json:"feeBaseMsat" db:"fee_base_msat"`
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
	Alias          *string `json:"alias"`
	Type           string  `json:"type" db:"type"`
}

type NodeForTag struct {
	NodeId int    `json:"nodeId" db:"node_id"`
	Alias  string `json:"alias"`
	Type   string `json:"type" db:"type"`
}

type ClosedChannel struct {
	ChannelID               int        `json:"channelId"`
	Tags                    []tags.Tag `json:"tags"`
	ShortChannelID          *string    `json:"shortChannelId"`
	FundingTransactionHash  string     `json:"fundingTransactionHash"`
	ClosingTransactionHash  *string    `json:"closingTransactionHash"`
	LNDShortChannelID       string     `json:"lndShortChannelId"`
	Capacity                int64      `json:"capacity"`
	FirstNodeId             int        `json:"nodeId"`
	SecondNodeId            int        `json:"peerNodeId"`
	InitiatingNodeId        *int       `json:"initiatingNodeId"`
	AcceptingNodeId         *int       `json:"acceptingNodeId"`
	ClosingNodeId           *int       `json:"closingNodeId"`
	Status                  string     `json:"status"`
	ClosingBlockHeight      *uint32    `json:"closingBlockHeight"`
	ClosedOn                *time.Time `json:"closedOn"`
	FundedOn                *time.Time `json:"fundedOn"`
	NodeName                string     `json:"nodeName"`
	PublicKey               string     `json:"pubKey"`
	PeerAlias               string     `json:"peerAlias"`
	ClosingNodeName         string     `json:"closingNodeName"`
	FundedOnSecondsDelta    *uint64    `json:"fundedOnSecondsDelta"`
	FundingBlockHeightDelta *uint32    `json:"fundingBlockHeightDelta"`
	ClosingBlockHeightDelta *uint32    `json:"closingBlockHeightDelta"`
	ClosedOnSecondsDelta    *uint64    `json:"closedOnSecondsDelta"`
}

type PendingChannel struct {
	ChannelID               int        `json:"channelId"`
	Tags                    []tags.Tag `json:"tags"`
	ShortChannelID          *string    `json:"shortChannelId"`
	FundingTransactionHash  string     `json:"fundingTransactionHash"`
	ClosingTransactionHash  *string    `json:"closingTransactionHash"`
	LNDShortChannelID       string     `json:"lndShortChannelId"`
	Capacity                int64      `json:"capacity"`
	FirstNodeId             int        `json:"nodeId"`
	SecondNodeId            int        `json:"peerNodeId"`
	InitiatingNodeId        *int       `json:"initiatingNodeId"`
	AcceptingNodeId         *int       `json:"acceptingNodeId"`
	ClosingNodeId           *int       `json:"closingNodeId"`
	Status                  string     `json:"status"`
	ClosingBlockHeight      *uint32    `json:"closingBlockHeight"`
	ClosedOn                *time.Time `json:"closedOn"`
	FundedOn                *time.Time `json:"fundedOn"`
	NodeName                string     `json:"nodeName"`
	PublicKey               string     `json:"pubKey"`
	PeerAlias               string     `json:"peerAlias"`
	ClosingNodeName         string     `json:"closingNodeName"`
	FundedOnSecondsDelta    *uint64    `json:"fundedOnSecondsDelta"`
	FundingBlockHeightDelta *uint32    `json:"fundingBlockHeightDelta"`
	ClosingBlockHeightDelta *uint32    `json:"closingBlockHeightDelta"`
	ClosedOnSecondsDelta    *uint64    `json:"closedOnSecondsDelta"`
}

func updateChannelsHandler(c *gin.Context, lightningRequestChannel chan<- interface{}) {
	var requestBody commons.RoutingPolicyUpdateRequest
	if err := c.BindJSON(&requestBody); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	// DISABLE the rate limiter
	requestBody.RateLimitSeconds = 1
	requestBody.RateLimitCount = 10

	response := SetRoutingPolicy(requestBody, lightningRequestChannel)
	if response.Status != commons.Active {
		err := errors.New(response.Error)
		c.JSON(http.StatusInternalServerError, server_errors.SingleServerError(err.Error()))
		err = errors.Wrap(err, "Problem when setting routing policy")
		log.Error().Err(err).Send()
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

func GetChannelsByNetwork(network commons.Network) ([]ChannelBody, error) {
	var channelsBody []ChannelBody
	chain := commons.Bitcoin
	nodeIds := commons.GetAllTorqNodeIdsByNetwork(chain, network)
	for _, nodeId := range nodeIds {
		// Force Response because we don't care about balance accuracy
		channelIds := commons.GetChannelStateChannelIds(nodeId, true)
		channelsBodyByNode, err := GetChannelsByIds(nodeId, channelIds)
		if err != nil {
			return nil, errors.Wrapf(err, "Obtain channels for nodeId: %v", nodeId)
		}
		channelsBody = append(channelsBody, channelsBodyByNode...)
	}
	return channelsBody, nil
}

func GetChannelsByIds(nodeId int, channelIds []int) ([]ChannelBody, error) {
	var channelsBody []ChannelBody
	for _, channelId := range channelIds {
		// Force Response because we don't care about balance accuracy
		channel := commons.GetChannelState(nodeId, channelId, true)
		channelSettings := commons.GetChannelSettingByChannelId(channel.ChannelId)
		lndShortChannelIdString := strconv.FormatUint(channelSettings.LndShortChannelId, 10)

		pendingHTLCs := calculateHTLCs(channel.PendingHtlcs)

		chanBody := ChannelBody{
			NodeId:                       nodeId,
			PeerNodeId:                   channel.RemoteNodeId,
			Tags:                         tags.GetTagsByTagIds(commons.GetTagIdsByChannelId(channel.RemoteNodeId, channelSettings.ChannelId)),
			ChannelId:                    channelSettings.ChannelId,
			NodeName:                     *commons.GetNodeSettingsByNodeId(nodeId).Name,
			Active:                       !channel.LocalDisabled,
			RemoteActive:                 !channel.RemoteDisabled,
			ChannelPoint:                 commons.CreateChannelPoint(channelSettings.FundingTransactionHash, channelSettings.FundingOutputIndex),
			Gauge:                        (float64(channel.LocalBalance) / float64(channelSettings.Capacity)) * 100,
			RemotePubkey:                 commons.GetNodeSettingsByNodeId(channel.RemoteNodeId).PublicKey,
			PeerAlias:                    commons.GetNodeAlias(channel.RemoteNodeId),
			FundingTransactionHash:       channelSettings.FundingTransactionHash,
			FundingOutputIndex:           channelSettings.FundingOutputIndex,
			CurrentBlockHeight:           commons.GetBlockHeight(),
			FundingBlockHeight:           channelSettings.FundingBlockHeight,
			FundedOn:                     channelSettings.FundedOn,
			ClosingBlockHeight:           channelSettings.ClosingBlockHeight,
			ClosedOn:                     channelSettings.ClosedOn,
			LNDShortChannelId:            lndShortChannelIdString,
			ShortChannelId:               channelSettings.ShortChannelId,
			Capacity:                     channelSettings.Capacity,
			PeerChannelCapacity:          channel.PeerChannelCapacity,
			PeerChannelCount:             channel.PeerChannelCount,
			PeerLocalBalance:             channel.PeerLocalBalance,
			PeerGauge:                    (float64(channel.PeerLocalBalance) / float64(channel.PeerChannelCapacity)) * 100,
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
			FeeBase:                      channel.LocalFeeBaseMsat / 1000,
			MinHtlc:                      channel.LocalMinHtlcMsat / 1000,
			MaxHtlc:                      channel.LocalMaxHtlcMsat / 1000,
			TimeLockDelta:                channel.LocalTimeLockDelta,
			FeeRateMilliMsat:             channel.LocalFeeRateMilliMsat,
			RemoteFeeBase:                channel.RemoteFeeBaseMsat / 1000,
			RemoteMinHtlc:                channel.RemoteMinHtlcMsat / 1000,
			RemoteMaxHtlc:                channel.RemoteMaxHtlcMsat / 1000,
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

		if channelSettings.FundingBlockHeight != nil {
			delta := commons.GetBlockHeight() - *channelSettings.FundingBlockHeight
			chanBody.FundingBlockHeightDelta = &delta
		}
		if channelSettings.FundedOn != nil {
			deltaSeconds := uint64(time.Since(*channelSettings.FundedOn).Seconds())
			chanBody.FundedOnSecondsDelta = &deltaSeconds
		}
		if channelSettings.ClosingBlockHeight != nil {
			delta := commons.GetBlockHeight() - *channelSettings.ClosingBlockHeight
			chanBody.ClosingBlockHeightDelta = &delta
		}
		if channelSettings.ClosedOn != nil {
			deltaSeconds := uint64(time.Since(*channelSettings.ClosedOn).Seconds())
			chanBody.ClosedOnSecondsDelta = &deltaSeconds
		}
		channelsBody = append(channelsBody, chanBody)
	}
	return channelsBody, nil
}

func getChannelListHandler(c *gin.Context) {
	network, err := strconv.Atoi(c.Query("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
		return
	}

	channelsBody, err := GetChannelsByNetwork(commons.Network(network))
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get channel tags for channel")
		return
	}
	c.JSON(http.StatusOK, channelsBody)
}

func getClosedChannelsListHandler(c *gin.Context, db *sqlx.DB) {
	network, err := strconv.Atoi(c.Query("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
		return
	}

	channels, err := getClosedChannels(db, network)

	if err != nil {
		c.JSON(http.StatusInternalServerError, server_errors.SingleServerError(err.Error()))
		err = errors.Wrap(err, "Problem getting closed channels from db")
		log.Error().Err(err).Send()
		return
	}

	closedChannels := make([]ClosedChannel, len(channels))
	torqNodeIds := commons.GetAllTorqNodeIds()

	for i, channel := range channels {

		torqNodeId := channel.FirstNodeId
		peerNodeId := channel.SecondNodeId

		if !slices.Contains(torqNodeIds, channel.FirstNodeId) {
			torqNodeId = channel.SecondNodeId
			peerNodeId = channel.FirstNodeId
		}

		closedChannels[i] = ClosedChannel{
			ChannelID:              channel.ChannelID,
			Tags:                   channel.Tags,
			ShortChannelID:         channel.ShortChannelID,
			FundingTransactionHash: channel.FundingTransactionHash,
			ClosingTransactionHash: channel.ClosingTransactionHash,
			Capacity:               channel.Capacity,
			FirstNodeId:            channel.FirstNodeId,
			SecondNodeId:           channel.SecondNodeId,
			InitiatingNodeId:       channel.InitiatingNodeId,
			AcceptingNodeId:        channel.AcceptingNodeId,
			ClosingNodeId:          channel.ClosingNodeId,
			Status:                 channel.Status.String(),
			ClosingBlockHeight:     channel.ClosingBlockHeight,
			ClosedOn:               channel.ClosedOn,
			FundedOn:               channel.FundedOn,
			NodeName:               commons.GetNodeAlias(torqNodeId),
			PublicKey:              commons.GetNodeSettingsByNodeId(torqNodeId).PublicKey,
			PeerAlias:              commons.GetNodeAlias(peerNodeId),
		}

		if channel.ClosingNodeId != nil {
			closedChannels[i].ClosingNodeName = commons.GetNodeAlias(*channel.ClosingNodeId)
		}
		if channel.FundingBlockHeight != nil {
			delta := commons.GetBlockHeight() - *channel.FundingBlockHeight
			closedChannels[i].FundingBlockHeightDelta = &delta
		}
		if channel.FundedOn != nil {
			deltaSeconds := uint64(time.Since(*channel.FundedOn).Seconds())
			closedChannels[i].FundedOnSecondsDelta = &deltaSeconds
		}
		if channel.ClosingBlockHeight != nil {
			delta := commons.GetBlockHeight() - *channel.ClosingBlockHeight
			closedChannels[i].ClosingBlockHeightDelta = &delta
		}
		if channel.ClosedOn != nil {
			deltaSeconds := uint64(time.Since(*channel.ClosedOn).Seconds())
			closedChannels[i].ClosedOnSecondsDelta = &deltaSeconds
		}

	}

	c.JSON(http.StatusOK, closedChannels)
}

func getChannelsPendingListHandler(c *gin.Context, db *sqlx.DB) {
	network, err := strconv.Atoi(c.Query("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
		return
	}

	channels, err := getPendingChannels(db, network)

	if err != nil {
		c.JSON(http.StatusInternalServerError, server_errors.SingleServerError(err.Error()))
		err = errors.Wrap(err, "Problem getting pending channels from db")
		log.Error().Err(err).Send()
		return
	}

	closedChannels := make([]PendingChannel, len(channels))
	torqNodeIds := commons.GetAllTorqNodeIds()

	for i, channel := range channels {

		torqNodeId := channel.FirstNodeId
		peerNodeId := channel.SecondNodeId

		if !slices.Contains(torqNodeIds, channel.FirstNodeId) {
			torqNodeId = channel.SecondNodeId
			peerNodeId = channel.FirstNodeId
		}

		closedChannels[i] = PendingChannel{
			ChannelID:              channel.ChannelID,
			Tags:                   channel.Tags,
			ShortChannelID:         channel.ShortChannelID,
			FundingTransactionHash: channel.FundingTransactionHash,
			ClosingTransactionHash: channel.ClosingTransactionHash,
			Capacity:               channel.Capacity,
			FirstNodeId:            channel.FirstNodeId,
			SecondNodeId:           channel.SecondNodeId,
			InitiatingNodeId:       channel.InitiatingNodeId,
			AcceptingNodeId:        channel.AcceptingNodeId,
			ClosingNodeId:          channel.ClosingNodeId,
			Status:                 channel.Status.String(),
			ClosingBlockHeight:     channel.ClosingBlockHeight,
			ClosedOn:               channel.ClosedOn,
			FundedOn:               channel.FundedOn,
			NodeName:               commons.GetNodeAlias(torqNodeId),
			PublicKey:              commons.GetNodeSettingsByNodeId(torqNodeId).PublicKey,
			PeerAlias:              commons.GetNodeAlias(peerNodeId),
		}

		if channel.ClosingNodeId != nil {
			closedChannels[i].ClosingNodeName = commons.GetNodeAlias(*channel.ClosingNodeId)
		}
		if channel.FundingBlockHeight != nil {
			delta := commons.GetBlockHeight() - *channel.FundingBlockHeight
			closedChannels[i].FundingBlockHeightDelta = &delta
		}
		if channel.FundedOn != nil {
			deltaSeconds := uint64(time.Since(*channel.FundedOn).Seconds())
			closedChannels[i].FundedOnSecondsDelta = &deltaSeconds
		}
		if channel.ClosingBlockHeight != nil {
			delta := commons.GetBlockHeight() - *channel.ClosingBlockHeight
			closedChannels[i].ClosingBlockHeightDelta = &delta
		}
		if channel.ClosedOn != nil {
			deltaSeconds := uint64(time.Since(*channel.ClosedOn).Seconds())
			closedChannels[i].ClosedOnSecondsDelta = &deltaSeconds
		}

	}

	c.JSON(http.StatusOK, closedChannels)
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
