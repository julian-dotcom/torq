package lightning

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/lightning_helpers"
	"github.com/lncapital/torq/pkg/server_errors"
)

type routingPolicyUpdateRequest struct {
	NodeId           int     `json:"nodeId"`
	RateLimitSeconds int     `json:"rateLimitSeconds"`
	RateLimitCount   int     `json:"rateLimitCount"`
	ChannelId        int     `json:"channelId"`
	FeeRateMilliMsat *int64  `json:"feeRateMilliMsat"`
	FeeBaseMsat      *int64  `json:"feeBaseMsat"`
	MaxHtlcMsat      *uint64 `json:"maxHtlcMsat"`
	MinHtlcMsat      *uint64 `json:"minHtlcMsat"`
	TimeLockDelta    *uint32 `json:"timeLockDelta"`
}

type nodeWalletBalance struct {
	NodeId                    int   `json:"nodeId"`
	TotalBalance              int64 `json:"totalBalance"`
	ConfirmedBalance          int64 `json:"confirmedBalance"`
	UnconfirmedBalance        int64 `json:"unconfirmedBalance"`
	LockedBalance             int64 `json:"lockedBalance"`
	ReservedBalanceAnchorChan int64 `json:"reservedBalanceAnchorChan"`
}

func batchOpenHandler(c *gin.Context) {
	var batchOpnReq lightning_helpers.BatchOpenChannelRequest
	if err := c.BindJSON(&batchOpnReq); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	response, err := BatchOpenChannel(batchOpnReq)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Batch open channels")
		return
	}

	c.JSON(http.StatusOK, response)
}

// openChannelHandler opens a channel to a peer
func openChannelHandler(c *gin.Context) {
	var openChannelRequest lightning_helpers.OpenChannelRequest
	err := c.BindJSON(&openChannelRequest)
	if err != nil {
		server_errors.SendBadRequest(c, "Can't parse request")
		return
	}

	response, err := OpenChannel(openChannelRequest)
	switch {
	case err != nil && strings.Contains(err.Error(), "connecting to "):
		serr := server_errors.ServerError{}
		// TODO: Replace with error codes
		serr.AddServerError("Torq could not connect to your node.")
		server_errors.SendBadRequestFieldError(c, &serr)
		return
	case err != nil && strings.Contains(err.Error(), "could not connect to peer."):
		serr := server_errors.ServerError{}
		// TODO: Replace with error codes
		serr.AddServerError("Could not connect to peer node. This could be because the node is offline or the node is " +
			"not reachable from your node. Please check the node and try again.")
		server_errors.SendBadRequestFieldError(c, &serr)
		return
	case err != nil && strings.Contains(err.Error(), "Cannot set both SatPerVbyte and TargetConf"):
		serr := server_errors.ServerError{}
		// TODO: Replace with error codes
		serr.AddServerError("Cannot set both Sat per vbyte and Target confirmations. Choose one and try again.")
		server_errors.SendBadRequestFieldError(c, &serr)
		return
	case err != nil && strings.Contains(err.Error(), "error decoding public key hex"):
		serr := server_errors.ServerError{}
		serr.AddServerError("Invalid public key. Please check the public key and try again.")
		server_errors.SendBadRequestFieldError(c, &serr)
		return
	case err != nil && strings.Contains(err.Error(), "channels cannot be created before the wallet is fully synced"):
		serr := server_errors.ServerError{}
		serr.AddServerError("Channels cannot be created before the wallet is fully synced. Please wait for the wallet to sync and try again.")
		server_errors.SendBadRequestFieldError(c, &serr)
		return
	case err != nil && strings.Contains(err.Error(), "unknown peer"):
		serr := server_errors.ServerError{}
		serr.AddServerError("Unknown peer. Please check the public key and url.")
		serr.AddServerError("The peer node may be offline or unreachable.")
		server_errors.SendBadRequestFieldError(c, &serr)
		return
	case err != nil && strings.Contains(err.Error(), "channel funding aborted"):
		serr := server_errors.ServerError{}
		serr.AddServerError("Channel funding aborted.")
		server_errors.SendBadRequestFieldError(c, &serr)
		return
	case status.Code(err) == codes.InvalidArgument:
		serr := server_errors.ServerError{}
		serr.AddServerError("Invalid argument. Please check the values and try again or reach out to the Torq team for help using the \"Help\" button in the navigation bar.")
		server_errors.SendBadRequestFieldError(c, &serr)
		return
	case status.Code(err) == codes.FailedPrecondition:
		serr := server_errors.ServerError{}
		serr.AddServerError("Failed precondition. Please check the values and try again or reach out to the Torq team for help using the \"Help\" button in the navigation bar.")
		server_errors.SendBadRequestFieldError(c, &serr)
		return
	case err != nil:
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// CloseChannel
func closeChannelHandler(c *gin.Context, db *sqlx.DB) {
	var closeChannelRequest lightning_helpers.CloseChannelRequest
	err := c.BindJSON(&closeChannelRequest)
	if err != nil {
		server_errors.SendBadRequest(c, "Can't parse request")
		return
	}

	closeChannelRequest.Db = db
	response, err := CloseChannel(closeChannelRequest)
	if err != nil {
		// Check if the error was because the node could not connect to the peer
		if strings.Contains(err.Error(), "could not connect to peer.") {
			serr := server_errors.ServerError{}
			serr.AddServerError("Could not connect to peer node.")
			server_errors.SendBadRequestFieldError(c, &serr)
		}
		server_errors.SendBadRequestFromError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func updateRoutingPolicyHandler(c *gin.Context, db *sqlx.DB) {
	var requestBody routingPolicyUpdateRequest
	if err := c.BindJSON(&requestBody); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	// DISABLE the rate limiter
	requestBody.RateLimitSeconds = 1
	requestBody.RateLimitCount = 10

	_, responseMessage, err := SetRoutingPolicy(db, requestBody.NodeId,
		requestBody.RateLimitSeconds, requestBody.RateLimitCount,
		requestBody.ChannelId,
		requestBody.FeeRateMilliMsat, requestBody.FeeBaseMsat,
		requestBody.MaxHtlcMsat, requestBody.MinHtlcMsat,
		requestBody.TimeLockDelta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, server_errors.SingleServerError(err.Error()))
		err = errors.Wrap(err, "Problem when setting routing policy")
		log.Error().Err(err).Send()
		return
	}

	c.JSON(http.StatusOK, responseMessage)
}

func getNodesWalletBalancesHandler(c *gin.Context, db *sqlx.DB) {
	network, err := strconv.Atoi(c.Param("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
	}

	activeTorqNodes := cache.GetActiveTorqNodeSettings()

	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Unable to get nodes")
		return
	}
	walletBalances := make([]nodeWalletBalance, 0)
	for _, activeTorqNode := range activeTorqNodes {
		if activeTorqNode.Network != core.Network(network) {
			continue
		}
		totalBalance, confirmedBalance, unconfirmedBalance, lockedBalance, reservedBalanceAnchorChan, err :=
			GetWalletBalance(activeTorqNode.NodeId)
		if err != nil {
			errorMsg := fmt.Sprintf("Error retrieving wallet balance for nodeId: %v", activeTorqNode.NodeId)
			server_errors.WrapLogAndSendServerError(c, err, errorMsg)
			log.Error().Msg(errorMsg)
			return
		}
		walletBalances = append(walletBalances, nodeWalletBalance{
			NodeId:                    activeTorqNode.NodeId,
			TotalBalance:              totalBalance,
			ConfirmedBalance:          confirmedBalance,
			UnconfirmedBalance:        unconfirmedBalance,
			LockedBalance:             lockedBalance,
			ReservedBalanceAnchorChan: reservedBalanceAnchorChan,
		})
	}

	c.JSON(http.StatusOK, walletBalances)

}
