package channels

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
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
	NodeId        int     `json:"nodeId"`
	ChannelPoint  *string `json:"channelPoint"`
	FeeRatePpm    *uint32 `json:"feeRatePpm"`
	BaseFeeMsat   *int64  `json:"baseFeeMsat"`
	MaxHtlcMsat   *uint64 `json:"maxHtlcMsat"`
	MinHtlcMsat   *uint64 `json:"minHtlcMsat"`
	TimeLockDelta uint32  `json:"timeLockDelta"`
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
	LNDShortChannelId            string               `json:"lndShortChannelId"`
	ShortChannelId               string               `json:"shortChannelId"`
	Capacity                     int64                `json:"capacity"`
	LocalBalance                 int64                `json:"localBalance"`
	RemoteBalance                int64                `json:"remoteBalance"`
	UnsettledBalance             int64                `json:"unsettledBalance"`
	CommitFee                    int64                `json:"commitFee"`
	CommitWeight                 int64                `json:"commitWeight"`
	FeePerKw                     int64                `json:"feePerKw"`
	BaseFeeMsat                  int64                `json:"baseFeeMsat"`
	MinHtlc                      int64                `json:"minHtlc"`
	MaxHtlcMsat                  uint64               `json:"maxHtlcMsat"`
	TimeLockDelta                uint32               `json:"timeLockDelta"`
	FeeRatePpm                   int64                `json:"feeRatePpm"`
	RemoteBaseFeeMsat            int64                `json:"remoteBaseFeeMsat"`
	RemoteMinHtlc                int64                `json:"remoteMinHtlc"`
	RemoteMaxHtlcMsat            uint64               `json:"remoteMaxHtlcMsat"`
	RemoteTimeLockDelta          uint32               `json:"remoteTimeLockDelta"`
	RemoteFeeRatePpm             int64                `json:"remoteFeeRatePpm"`
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
	OneMl                        string               `json:"1ml"`
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
	TimeLockDelta   uint32 `json:"timeLockDelta" db:"time_lock_delta"`
	MinHtlc         int64  `json:"minHtlc" db:"min_htlc"`
	MaxHtlcMsat     uint64 `json:"maxHtlcMsat" db:"max_htlc_msat"`
	FeeRateMillMsat int64  `json:"feeRateMillMsat" db:"fee_rate_mill_msat"`
	ShortChannelId  string `json:"shortChannelId" db:"short_channel_id"`
	FeeBaseMsat     int64  `json:"feeBaseMsat" db:"fee_base_msat"`
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

func getChannelListhandler(c *gin.Context, db *sqlx.DB) {
	var channelsBody []channelBody
	nodes, err := settings.GetActiveNodesConnectionDetails(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get active nodes")
		return
	}

	for _, node := range nodes {
		conn, err := lnd_connect.Connect(
			node.GRPCAddress,
			node.TLSFileBytes,
			node.MacaroonFileBytes)
		if err != nil {
			errorMsg := fmt.Sprintf("Connect to node %d\n", node.NodeId)
			server_errors.WrapLogAndSendServerError(c, err, errorMsg)
			return
		}

		defer conn.Close()

		client := lnrpc.NewLightningClient(conn)

		r, err := client.ListChannels(context.Background(), &lnrpc.ListChannelsRequest{})
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "List channels")
			return
		}

		for _, channel := range r.Channels {
			localPolicy, err := GetLocalRoutingPolicy(channel.ChanId, node.NodeId, db)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Local Channel policy")
				return
			}
			remotePolicy, err := GetRemoteRoutingPolicy(channel.ChanId, node.NodeId, db)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Remote Channel policy")
				return
			}

			stringLNDShortChannelId := strconv.FormatUint(channel.ChanId, 10)
			pendingHTLCs := calculateHTLCs(channel.PendingHtlcs)

			gauge := (float64(channel.LocalBalance) / float64(channel.Capacity)) * 100
			fundingTransactionHash, fundingOutputIndex := ParseChannelPoint(channel.ChannelPoint)

			shortChannelId := localPolicy.ShortChannelId
			localNodeId := localPolicy.NodeId
			remoteNodeId := localPolicy.RemoteNodeId
			if localPolicy.ShortChannelId == "" && remotePolicy.ShortChannelId != "" {
				shortChannelId = remotePolicy.ShortChannelId
				localNodeId = remotePolicy.RemoteNodeId
				remoteNodeId = remotePolicy.NodeId
			}
			chanBody := channelBody{
				NodeId:                       node.NodeId,
				NodeName:                     node.Name,
				Active:                       channel.Active,
				ChannelPoint:                 channel.ChannelPoint,
				Gauge:                        gauge,
				RemotePubkey:                 channel.RemotePubkey,
				FundingTransactionHash:       fundingTransactionHash,
				FundingOutputIndex:           fundingOutputIndex,
				LNDShortChannelId:            stringLNDShortChannelId,
				ShortChannelId:               shortChannelId,
				Capacity:                     channel.Capacity,
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
				BaseFeeMsat:                  localPolicy.FeeBaseMsat,
				MinHtlc:                      localPolicy.MinHtlc,
				MaxHtlcMsat:                  localPolicy.MaxHtlcMsat,
				TimeLockDelta:                localPolicy.TimeLockDelta,
				FeeRatePpm:                   localPolicy.FeeRateMillMsat,
				RemoteBaseFeeMsat:            remotePolicy.FeeBaseMsat,
				RemoteMinHtlc:                remotePolicy.MinHtlc,
				RemoteMaxHtlcMsat:            remotePolicy.MaxHtlcMsat,
				RemoteTimeLockDelta:          remotePolicy.TimeLockDelta,
				RemoteFeeRatePpm:             remotePolicy.FeeRateMillMsat,
				NumUpdates:                   channel.NumUpdates,
				Initiator:                    channel.Initiator,
				ChanStatusFlags:              channel.ChanStatusFlags,
				CommitmentType:               channel.CommitmentType,
				Lifetime:                     channel.Lifetime,
				MempoolSpace:                 MEMPOOL + stringLNDShortChannelId,
				AmbossSpace:                  AMBOSS + shortChannelId,
				OneMl:                        ONEML + stringLNDShortChannelId,
			}

			peerInfo, err := GetNodePeerAlias(localNodeId, remoteNodeId, db)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Node Alias")
				return
			}
			chanBody.PeerAlias = peerInfo
			channelsBody = append(channelsBody, chanBody)
		}
	}
	c.JSON(http.StatusOK, channelsBody)
}

func calculateHTLCs(htlcs []*lnrpc.HTLC) PendingHtlcs {
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
		pendingHTLCs.TotalAmount = pendingHTLCs.ForwardingAmount + pendingHTLCs.ForwardingAmount
		pendingHTLCs.TotalCount = pendingHTLCs.ForwardingCount + pendingHTLCs.LocalCount
	}

	return pendingHTLCs
}
