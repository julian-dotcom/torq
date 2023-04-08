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

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/settings"
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

type connectPeerRequest struct {
	NodeId           int                        `json:"nodeId"`
	ConnectionString string                     `json:"connectionString"`
	Network          core.Network               `json:"network"`
	Setting          core.NodeConnectionSetting `json:"setting"`
}

type disconnectPeerRequest struct {
	NodeId     int `json:"nodeId"`
	TorqNodeId int `json:"torqNodeId"`
}

type reconnectPeerRequest struct {
	NodeId     int `json:"nodeId"`
	TorqNodeId int `json:"torqNodeId"`
}

type updatePeerRequest struct {
	NodeId     int                         `json:"nodeId"`
	TorqNodeId int                         `json:"torqNodeId"`
	Setting    *core.NodeConnectionSetting `json:"setting"`
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

func connectNewPeerHandler(c *gin.Context, db *sqlx.DB) {

	var req connectPeerRequest

	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequest(c, "Can't process ConnectPeerRequest")
		return
	}

	s := strings.Split(req.ConnectionString, "@")
	if len(s) != 2 || s[0] == "" || s[1] == "" {
		server_errors.SendBadRequest(c, "Invalid connectionString format.")
		return
	}

	pubKey := s[0]
	host := s[1]

	_, err := ConnectPeer(req.NodeId, pubKey, host)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Error connecting to peer.")
		return
	}

	// save peer node
	nodeId, err := settings.AddNodeWhenNew(db, pubKey, core.Bitcoin, req.Network)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node.")
		return
	}
	err = settings.AddNodeConnectionHistory(db, req.NodeId, nodeId, host, req.Setting, core.NodeConnectionStatusConnected)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node connection history.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully connected to peer."})
}

func disconnectPeerHandler(c *gin.Context, db *sqlx.DB) {
	var req disconnectPeerRequest
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequest(c, "Cannot process DisconnectPeerRequest")
		return
	}

	torqNodeId, address, setting, _, err := settings.GetNodeConnectionHistoryWithDetail(db, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node by id.")
		return
	}

	if address == "" {
		host, err := getHostFromPeer(torqNodeId, req.NodeId)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Getting host from peer.")
			return
		}
		address = host
	}

	requestFailedCurrentlyDisconnected, err := DisconnectPeer(req.TorqNodeId, req.NodeId)
	if err != nil {
		if requestFailedCurrentlyDisconnected {
			err = settings.AddNodeConnectionHistory(db, torqNodeId, req.NodeId, address, setting, core.NodeConnectionStatusDisconnected)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Saving peer node connection history.")
				return
			}
		}
		server_errors.WrapLogAndSendServerError(c, err, "Error disconnecting peer.")
		return
	}

	err = settings.AddNodeConnectionHistory(db, torqNodeId, req.NodeId, address, setting, core.NodeConnectionStatusDisconnected)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node connection history.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully disconnected peer."})
}

func reconnectPeerHandler(c *gin.Context, db *sqlx.DB) {
	var req reconnectPeerRequest
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequest(c, "Cannot process DisconnectPeerRequest")
		return
	}

	torqNodeId, address, setting, _, err := settings.GetNodeConnectionHistoryWithDetail(db, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node connection history by id.")
		return
	}

	if address == "" {
		host, err := getHostFromPeer(torqNodeId, req.NodeId)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Getting host from peer.")
			return
		}
		address = host
	}

	publicKey := cache.GetNodeSettingsByNodeId(req.NodeId).PublicKey
	requestFailCurrentlyConnected, err := ConnectPeer(req.TorqNodeId, publicKey, address)
	if err != nil {
		if requestFailCurrentlyConnected {
			err = settings.AddNodeConnectionHistory(db, torqNodeId, req.NodeId, address, setting, core.NodeConnectionStatusConnected)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Saving peer node connection history.")
				return
			}
		}
		server_errors.WrapLogAndSendServerError(c, err, "Updating Peer, already connected.")
		return
	}

	err = settings.AddNodeConnectionHistory(db, torqNodeId, req.NodeId, address, setting, core.NodeConnectionStatusConnected)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully disconnected peer."})
}

func updatePeer(c *gin.Context, db *sqlx.DB) {
	var req updatePeerRequest
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequest(c, "Cannot process DisconnectPeerRequest")
		return
	}

	if req.NodeId == 0 {
		server_errors.SendBadRequest(c, "Node id is required.")
		return
	}

	if req.TorqNodeId == 0 {
		server_errors.SendBadRequest(c, "Torq Node id is required.")
		return
	}

	if req.Setting == nil {
		server_errors.SendBadRequest(c, "Setting is required.")
		return
	}

	torqNodeId, address, _, status, err := settings.GetNodeConnectionHistoryWithDetail(db, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node connection history by id.")
		return
	}

	if address == "" {
		host, err := getHostFromPeer(torqNodeId, req.NodeId)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Getting host from peer.")
			return
		}
		address = host
	}

	err = settings.AddNodeConnectionHistory(db, torqNodeId, req.NodeId, address, *req.Setting, status)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "updating peer node.")
		return
	}

}

func getHostFromPeer(connectionDetailsNodeId int, nodeId int) (string, error) {
	nodeSettings := cache.GetNodeSettingsByNodeId(nodeId)
	peers, err := ListPeers(connectionDetailsNodeId)
	if err != nil {
		return "", errors.Wrap(err, "Getting list of peers.")
	}
	p, ok := peers[nodeSettings.PublicKey]
	if ok {
		return p.Address, nil
	}
	return "", errors.New("Peer public key not found.")
}
