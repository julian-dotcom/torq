package nodes

import (
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

type PeerNode struct {
	NodeId    int       `json:"nodeId"`
	PublicKey string    `json:"pubKey"`
	Chain     string    `json:"chain"`
	Network   string    `json:"network"`
	CreatedOn time.Time `json:"createdOn"`
}

type ConnectPeerRequest struct {
	NodeId           int             `json:"nodeId"`
	ConnectionString string          `json:"connectionString"`
	Network          commons.Network `json:"network"`
}

func RegisterNodeRoutes(r *gin.RouterGroup, db *sqlx.DB, lightningRequestChannel chan<- interface{}) {
	r.GET("/:network/peers", func(c *gin.Context) { getAllPeersHandler(c, db) })
	r.POST("/peers/connect", func(c *gin.Context) { connectPeerHandler(c, db, lightningRequestChannel) })
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

func getAllPeersHandler(c *gin.Context, db *sqlx.DB) {
	network, err := strconv.Atoi(c.Param("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
		return
	}
	nodes, err := GetPeerNodes(db, commons.Network(network))
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting all Peer nodes.")
		return
	}

	peers := make([]PeerNode, 0)
	for _, n := range nodes {
		peer := PeerNode{
			NodeId:    n.NodeId,
			PublicKey: n.PublicKey,
			Chain:     n.Chain.String(),
			Network:   n.Network.String(),
			CreatedOn: n.CreatedOn,
		}
		peers = append(peers, peer)
	}

	c.JSON(http.StatusOK, peers)
}

func connectPeerHandler(c *gin.Context, db *sqlx.DB, lightningRequestChannel chan<- interface{}) {

	var rb ConnectPeerRequest

	if err := c.BindJSON(&rb); err != nil {
		server_errors.SendBadRequest(c, "Can't process ConnectPeerRequest")
		return
	}

	s := strings.Split(rb.ConnectionString, "@")
	if len(s) != 2 || s[0] == "" || s[1] == "" {
		server_errors.SendBadRequest(c, "Invalid connectionString format.")
		return
	}

	r := commons.ConnectPeer(rb.NodeId, s[0], s[1], lightningRequestChannel)
	if r.CommunicationResponse.Error != "" {
		server_errors.WrapLogAndSendServerError(c, errors.New(r.CommunicationResponse.Error), "Error connecting to peer.")
		return
	}

	// save peer node
	_, err := AddNodeWhenNew(db, Node{NodeId: rb.NodeId, PublicKey: s[0], Network: rb.Network})
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully connected to peer.")})
}
