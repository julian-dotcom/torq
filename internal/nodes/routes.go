package nodes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/pkg/server_errors"
)

type NodeInformation struct {
	NodeId    int            `json:"nodeId"`
	PublicKey string         `json:"publicKey"`
	Alias     string         `json:"alias"`
	TorqAlias string         `json:"torqAlias"`
	Color     string         `json:"color"`
	Addresses *[]NodeAddress `json:"addresses"`
	Status    core.Status    `json:"status"`
}

type NodeWalletBalance struct {
	NodeId                    int   `json:"nodeId"`
	TotalBalance              int64 `json:"totalBalance"`
	ConfirmedBalance          int64 `json:"confirmedBalance"`
	UnconfirmedBalance        int64 `json:"unconfirmedBalance"`
	LockedBalance             int64 `json:"lockedBalance"`
	ReservedBalanceAnchorChan int64 `json:"reservedBalanceAnchorChan"`
}

func RegisterNodeRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("/:network/nodes", func(c *gin.Context) { getNodesByNetworkHandler(c, db) })
	r.DELETE(":nodeId", func(c *gin.Context) { removeNodeHandler(c, db) })
}

func getNodesByNetworkHandler(c *gin.Context, db *sqlx.DB) {
	network, err := strconv.Atoi(c.Param("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
	}
	nds, err := getAllNodeInformationByNetwork(db, core.Network(network))
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting nodes by network.")
		return
	}
	c.JSON(http.StatusOK, nds)
}

func removeNodeHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	count, err := removeNode(db, nodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Removing node for nodeId: %v", nodeId))
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully deleted %v node(s).", count)})
}
