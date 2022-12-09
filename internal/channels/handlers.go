package channels

import (
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

type failedUpdate struct {
	OutPoint    OutPoint `json:"outPoint"`
	Reason      string   `json:"reason"`
	UpdateError string   `json:"updateError"`
}

type OutPoint struct {
	Txid        string `json:"txId"`
	OutputIndex uint32 `json:"outputIndex"`
}

type updateResponse struct {
	Status        string         `json:"status"`
	FailedUpdates []failedUpdate `json:"failedUpdates"`
}

type updateChanRequestBody struct {
	NodeId                 int     `json:"nodeId"`
	FundingTransactionHash *string `json:"fundingTransactionHash"`
	FundingOutputIndex     *int    `json:"fundingOutputIndex"`
	FeeRatePpm             *uint32 `json:"feeRatePpm"`
	BaseFeeMsat            *int64  `json:"baseFeeMsat"`
	MaxHtlcMsat            *uint64 `json:"maxHtlcMsat"`
	MinHtlcMsat            *uint64 `json:"minHtlcMsat"`
	TimeLockDelta          uint32  `json:"timeLockDelta"`
}
type pendingChannel struct {
	PendingChannelPoint string `json:"pendingChannelPoint"`
}

type channelBody struct {
	NodeId                       int                  `json:"nodeId"`
	ChannelPoint                 string               `json:"channelPoint"`
	NodeName                     string               `json:"nodeName"`
	Active                       bool                 `json:"active"`
	Gauge                        float64              `json:"gauge"`
	RemotePubkey                 string               `json:"remotePubkey"`
	FundingTransactionHash       string               `json:"fundingTransactionHash"`
	FundingOutputIndex           int                  `json:"fundingOutputIndex"`
	LNDShortChannelId            uint64               `json:"lndShortChannelId"`
	ShortChannelId               string               `json:"shortChannelId"`
	Capacity                     int64                `json:"capacity"`
	LocalBalance                 int64                `json:"localBalance"`
	RemoteBalance                int64                `json:"remoteBalance"`
	UnsettledBalance             int64                `json:"unsettledBalance"`
	CommitFee                    int64                `json:"commitFee"`
	CommitWeight                 int64                `json:"commitWeight"`
	FeePerKw                     int64                `json:"feePerKw"`
	BaseFeeMsat                  uint64               `json:"baseFeeMsat"`
	MinHtlc                      int64                `json:"minHtlc"`
	MaxHtlcMsat                  uint64               `json:"maxHtlcMsat"`
	TimeLockDelta                uint32               `json:"timeLockDelta"`
	FeeRatePpm                   uint64               `json:"feeRatePpm"`
	RemoteBaseFeeMsat            uint64               `json:"remoteBaseFeeMsat"`
	RemoteMinHtlc                int64                `json:"remoteMinHtlc"`
	RemoteMaxHtlcMsat            uint64               `json:"remoteMaxHtlcMsat"`
	RemoteTimeLockDelta          uint32               `json:"remoteTimeLockDelta"`
	RemoteFeeRatePpm             uint64               `json:"remoteFeeRatePpm"`
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
	MinHtlc         int64  `json:"minHtlc" db:"min_htlc"`
	MaxHtlcMsat     uint64 `json:"maxHtlcMsat" db:"max_htlc_msat"`
	FeeRateMillMsat uint64 `json:"feeRateMillMsat" db:"fee_rate_mill_msat"`
	ShortChannelId  string `json:"shortChannelId" db:"short_channel_id"`
	FeeBaseMsat     uint64 `json:"feeBaseMsat" db:"fee_base_msat"`
	NodeId          int    `json:"nodeId" db:"node_id"`
	RemoteNodeId    int    `json:"RemoteodeId" db:"remote_node_id"`
}

const (
	MEMPOOL string = "https://mempool.space/lightning/channel/"
	AMBOSS  string = "https://amboss.space/edge/"
	ONEML   string = "https://1ml.com/channel/"
)

func updateChannelsHandler(c *gin.Context, db *sqlx.DB) {
	requestBody := updateChanRequestBody{}

	if err := c.BindJSON(&requestBody); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	response, err := updateChannels(db, requestBody)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Update channel/s policy")
		return
	}

	c.JSON(http.StatusOK, response)
}

type batchOpenChannel struct {
	NodePubkey         string `json:"nodePubkey"`
	LocalFundingAmount int64  `json:"localFundingAmount"`
	PushSat            *int64 `json:"pushSat"`
	Private            *bool  `json:"private"`
	MinHtlcMsat        *int64 `json:"minHtlcMsat"`
}

type BatchOpenRequest struct {
	NodeId      int                `json:"nodeId"`
	Channels    []batchOpenChannel `json:"channels"`
	TargetConf  *int32             `json:"targetConf"`
	SatPerVbyte *int64             `json:"satPerVbyte"`
}

type BatchOpenResponse struct {
	PendingChannels []pendingChannel `json:"pendingChannels"`
}

func batchOpenHandler(c *gin.Context, db *sqlx.DB) {
	var batchOpnReq BatchOpenRequest
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
	activeNcds, err := settings.GetActiveNodesConnectionDetails(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "List channels")
		return
	}
	if len(activeNcds) != 0 {
		for _, ncd := range activeNcds {
			channelBalanceStates := commons.GetChannelStates(ncd.NodeId, true)
			nodeSettings := commons.GetNodeSettingsByNodeId(ncd.NodeId)
			for _, channel := range channelBalanceStates {
				channelSettings := commons.GetChannelSettingByChannelId(channel.ChannelId)
				lndShortChannelIdString := strconv.FormatUint(channelSettings.LndShortChannelId, 10)

				pendingHTLCs := calculateHTLCs(channel.PendingHtlcs)

				gauge := (float64(channel.LocalBalance) / float64(channelSettings.Capacity)) * 100

				remoteNode := commons.GetNodeSettingsByNodeId(channel.RemoteNodeId)
				chanBody := channelBody{
					NodeId:                       ncd.NodeId,
					NodeName:                     *nodeSettings.Name,
					Active:                       !channel.LocalDisabled,
					ChannelPoint:                 commons.CreateChannelPoint(channelSettings.FundingTransactionHash, channelSettings.FundingOutputIndex),
					Gauge:                        gauge,
					RemotePubkey:                 remoteNode.PublicKey,
					FundingTransactionHash:       channelSettings.FundingTransactionHash,
					FundingOutputIndex:           channelSettings.FundingOutputIndex,
					LNDShortChannelId:            channelSettings.LndShortChannelId,
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
					BaseFeeMsat:                  channel.LocalFeeBaseMsat,
					MinHtlc:                      channel.LocalMinHtlc,
					MaxHtlcMsat:                  channel.LocalMaxHtlcMsat,
					TimeLockDelta:                channel.LocalTimeLockDelta,
					FeeRatePpm:                   channel.LocalFeeRateMilliMsat,
					RemoteBaseFeeMsat:            channel.RemoteFeeBaseMsat,
					RemoteMinHtlc:                channel.RemoteMinHtlc,
					RemoteMaxHtlcMsat:            channel.RemoteMaxHtlcMsat,
					RemoteTimeLockDelta:          channel.RemoteTimeLockDelta,
					RemoteFeeRatePpm:             channel.RemoteFeeRateMilliMsat,
					NumUpdates:                   channel.NumUpdates,
					Initiator:                    channelSettings.InitiatingNodeId != nil && *channelSettings.InitiatingNodeId == ncd.NodeId,
					ChanStatusFlags:              channel.ChanStatusFlags,
					CommitmentType:               channel.CommitmentType,
					Lifetime:                     channel.Lifetime,
					MempoolSpace:                 MEMPOOL + lndShortChannelIdString,
					AmbossSpace:                  AMBOSS + channelSettings.ShortChannelId,
					OneMl:                        ONEML + lndShortChannelIdString,
				}

				peerInfo, err := GetNodePeerAlias(ncd.NodeId, channel.RemoteNodeId, db)
				if err == nil {
					chanBody.PeerAlias = peerInfo
				} else {
					log.Error().Err(err).Msgf("Could not obtain the alias of the peer with nodeId: %v (for nodeId: %v)", channel.RemoteNodeId, ncd.NodeId)
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
	} else {
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
	}

	return pendingHTLCs
}
