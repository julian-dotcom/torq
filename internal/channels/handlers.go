package channels

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"github.com/rs/zerolog/log"
)

type failedUpdate struct {
	OutPoint struct {
		Txid    string
		OutIndx uint32
	}
	Reason      string
	UpdateError string
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
	LocalNodeId           int                  `json:"localNodeId"`
	Active                bool                 `json:"active"`
	Gauge                 float64              `json:"gauge"`
	RemotePubkey          string               `json:"remotePubkey"`
	LNDChannelPoint       string               `json:"lndChannelPoint"`
	LNDShortChannelId     uint64               `json:"lndShortChannelId"`
	ShortChannelId        string               `json:"shortChannelId"`
	Capacity              int64                `json:"capacity"`
	LocalBalance          int64                `json:"localBalance"`
	RemoteBalance         int64                `json:"remoteBalance"`
	UnsettledBalance      int64                `json:"unsettledBalance"`
	CommitFee             int64                `json:"commitFee"`
	CommitWeight          int64                `json:"commitWeight"`
	FeePerKw              int64                `json:"feePerKw"`
	BaseFeeMsat           int64                `json:"baseFeeMsat"`
	MinHtlc               int64                `json:"minHtlc"`
	MaxHtlcMsat           uint64               `json:"maxHtlcMsat"`
	PendingHtlcs          int64                `json:"pendingHtlcs"`
	TotalSatoshisSent     int64                `json:"totalSatoshisSent"`
	NumUpdates            uint64               `json:"numUpdates"`
	Initiator             bool                 `json:"initiator"`
	ChanStatusFlags       string               `json:"chanStatusFlags"`
	LocalChanReserveSat   int64                `json:"localChanReserveSat"`
	RemoteChanReserveSat  int64                `json:"remoteChanReserveSat"`
	CommitmentType        lnrpc.CommitmentType `json:"commitmentType"`
	Lifetime              int64                `json:"lifetime"`
	TotalSatoshisReceived int64                `json:"totalSatoshisReceived"`
}

func updateChannelsHandler(c *gin.Context, db *sqlx.DB) {
	requestBody := updateChanRequestBody{}

	if err := c.BindJSON(&requestBody); err != nil {
		log.Error().Msgf("JSON binding the request body")
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}
	//log.Debug().Msgf("Received request body: %v", requestBody)

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
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
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
			errorMsg := fmt.Sprintf("Connect to node %d\n", node.LocalNodeId)
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

			gauge := (float64(channel.LocalBalance) / float64(channel.Capacity)) * 100
			chanBody := channelBody{
				LocalNodeId:           node.LocalNodeId,
				Active:                channel.Active,
				Gauge:                 gauge,
				RemotePubkey:          channel.RemotePubkey,
				LNDChannelPoint:       channel.ChannelPoint,
				LNDShortChannelId:     channel.ChanId,
				ShortChannelId:        ConvertLNDShortChannelID(channel.ChanId),
				Capacity:              channel.Capacity,
				LocalBalance:          channel.LocalBalance,
				RemoteBalance:         channel.RemoteBalance,
				UnsettledBalance:      channel.UnsettledBalance,
				TotalSatoshisSent:     channel.TotalSatoshisSent,
				TotalSatoshisReceived: channel.TotalSatoshisReceived,
				PendingHtlcs:          sumHTLCs(channel.PendingHtlcs),
				CommitFee:             channel.CommitFee,
				CommitWeight:          channel.CommitWeight,
				FeePerKw:              channel.FeePerKw,
				BaseFeeMsat:           channelFee.Node1Policy.FeeBaseMsat,
				MinHtlc:               channelFee.Node1Policy.MinHtlc,
				MaxHtlcMsat:           channelFee.Node1Policy.MaxHtlcMsat,
				NumUpdates:            channel.NumUpdates,
				Initiator:             channel.Initiator,
				ChanStatusFlags:       channel.ChanStatusFlags,
				CommitmentType:        channel.CommitmentType,
				Lifetime:              channel.Lifetime,
			}
			channelsBody = append(channelsBody, chanBody)
		}
	}
	c.JSON(http.StatusOK, channelsBody)
}

func sumHTLCs(htlcs []*lnrpc.HTLC) int64 {
	var pendingHtlc int64
	if len(htlcs) < 1 {
		return pendingHtlc
	} else {
		for _, htlc := range htlcs {
			if htlc.ForwardingHtlcIndex == 0 {
				continue
			}
			pendingHtlc += htlc.Amount
		}
	}

	return pendingHtlc
}
