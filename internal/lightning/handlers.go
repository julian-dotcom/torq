package lightning

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/cache"
	"github.com/lncapital/torq/pkg/core"
	"github.com/lncapital/torq/pkg/lightning"
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

func updateRoutingPolicyHandler(c *gin.Context, db *sqlx.DB) {
	var requestBody routingPolicyUpdateRequest
	if err := c.BindJSON(&requestBody); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	// DISABLE the rate limiter
	requestBody.RateLimitSeconds = 1
	requestBody.RateLimitCount = 10

	_, responseMessage, err := lightning.SetRoutingPolicy(db, requestBody.NodeId,
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
			lightning.GetWalletBalance(activeTorqNode.NodeId)
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
