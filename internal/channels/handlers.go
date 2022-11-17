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
	LNDShortChannelId            uint64               `json:"lndShortChannelId"`
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
			channelFee, err := client.GetChanInfo(context.Background(), &lnrpc.ChanInfoRequest{ChanId: channel.ChanId})
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Channel info")
				return
			}
			shortChannelId := ConvertLNDShortChannelID(channel.ChanId)
			stringLNDShortChannelId := strconv.FormatUint(channel.ChanId, 10)

			pendingHTLCs := calculateHTLCs(channel.PendingHtlcs)

			gauge := (float64(channel.LocalBalance) / float64(channel.Capacity)) * 100
			fundingTransactionHash, fundingOutputIndex := ParseChannelPoint(channel.ChannelPoint)
			chanBody := channelBody{
				NodeId:                       node.NodeId,
				NodeName:                     node.Name,
				Active:                       channel.Active,
				ChannelPoint:                 channel.ChannelPoint,
				Gauge:                        gauge,
				RemotePubkey:                 channel.RemotePubkey,
				FundingTransactionHash:       fundingTransactionHash,
				FundingOutputIndex:           fundingOutputIndex,
				LNDShortChannelId:            channel.ChanId,
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
				BaseFeeMsat:                  channelFee.Node1Policy.FeeBaseMsat,
				MinHtlc:                      channelFee.Node1Policy.MinHtlc,
				MaxHtlcMsat:                  channelFee.Node1Policy.MaxHtlcMsat,
				TimeLockDelta:                channelFee.Node1Policy.TimeLockDelta,
				FeeRatePpm:                   channelFee.Node1Policy.FeeRateMilliMsat,
				NumUpdates:                   channel.NumUpdates,
				Initiator:                    channel.Initiator,
				ChanStatusFlags:              channel.ChanStatusFlags,
				CommitmentType:               channel.CommitmentType,
				Lifetime:                     channel.Lifetime,
				MempoolSpace:                 MEMPOOL + stringLNDShortChannelId,
				AmbossSpace:                  AMBOSS + shortChannelId,
				OneMl:                        ONEML + stringLNDShortChannelId,
			}

			peerInfo, err := client.GetNodeInfo(context.Background(),
				&lnrpc.NodeInfoRequest{IncludeChannels: true, PubKey: channel.RemotePubkey})
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Node info")
				return
			}
			chanBody.PeerAlias = peerInfo.Node.Alias

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
