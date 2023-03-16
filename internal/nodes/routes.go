package nodes

import (
	"fmt"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/pkg/server_errors"
)

type LndAddress struct {
	PubKey string `json:"pubKey"`
	Host   string `json:"host"`
}

type ConnectNodeRequest struct {
	NodeId     int        `json:"nodeId"`
	LndAddress LndAddress `json:"lndAddress"`
	Perm       *bool      `json:"perm"`
	TimeOut    *uint64    `json:"timeOut"`
}

type NodeInformation struct {
	NodeId    int            `json:"nodeId"`
	PublicKey string         `json:"publicKey"`
	Alias     string         `json:"alias"`
	TorqAlias string         `json:"torqAlias"`
	Color     string         `json:"color"`
	Addresses *[]NodeAddress `json:"addresses"`
	Status    commons.Status `json:"status"`
}

type NodeWalletBalance struct {
	NodeId                    int   `json:"nodeId"`
	TotalBalance              int64 `json:"totalBalance"`
	ConfirmedBalance          int64 `json:"confirmedBalance"`
	UnconfirmedBalance        int64 `json:"unconfirmedBalance"`
	LockedBalance             int64 `json:"lockedBalance"`
	ReservedBalanceAnchorChan int64 `json:"reservedBalanceAnchorChan"`
}

func RegisterNodeRoutes(r *gin.RouterGroup, db *sqlx.DB, lightningRequestChannel chan<- interface{}) {
	r.GET("/:network/nodes", func(c *gin.Context) { getNodesByNetworkHandler(c, db) })
	r.GET("/:network/walletBalances", func(c *gin.Context) { getNodesWalletBalancesHandler(c, db, lightningRequestChannel) })
	r.DELETE(":nodeId", func(c *gin.Context) { removeNodeHandler(c, db) })
}

func getNodesByNetworkHandler(c *gin.Context, db *sqlx.DB) {
	network, err := strconv.Atoi(c.Param("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
	}
	nds, err := getAllNodeInformationByNetwork(db, commons.Network(network))
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

func getNodesWalletBalancesHandler(c *gin.Context, db *sqlx.DB, lightningRequestChannel chan<- interface{}) {
	network, err := strconv.Atoi(c.Param("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
	}

	nodes, err := getNodesByNetwork(db, false, commons.Network(network))

	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Unable to get nodes")
		return
	}
	walletBalances := make([]NodeWalletBalance, 0)
	for _, v := range nodes {
		r := commons.GetNodeWalletBalance(v.NodeId, lightningRequestChannel)
		if r.Error != "" {
			errorMsg := fmt.Sprintf("Error retrieving wallet balance for nodeId: %v", v.NodeId)
			server_errors.WrapLogAndSendServerError(c, err, errorMsg)
			log.Error().Msg(errorMsg)
			return
		}
		walletBalances = append(walletBalances, NodeWalletBalance{
			NodeId:                    v.NodeId,
			TotalBalance:              r.TotalBalance,
			ConfirmedBalance:          r.ConfirmedBalance,
			UnconfirmedBalance:        r.UnconfirmedBalance,
			LockedBalance:             r.LockedBalance,
			ReservedBalanceAnchorChan: r.ReservedBalanceAnchorChan,
		})
	}

	c.JSON(http.StatusOK, walletBalances)

}
