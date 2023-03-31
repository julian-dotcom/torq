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
	NodeId                      int             `json:"nodeId" db:"node_id"`
	PublicKey                   string          `json:"pubKey" db:"public_key"`
	Chain                       commons.Chain   `json:"chain" db:"chain"`
	Network                     commons.Network `json:"network" db:"network"`
	CreatedOn                   time.Time       `json:"createdOn" db:"created_on"`
	NodeConnectionDetailsNodeId *int            `json:"nodeConnectionDetailsNodeId" db:"node_connection_details_node_id" description:"The node that established a connection to this node"`
	ConnectionStatusId          commons.Status  `json:"connectionStatus" db:"connection_status_id"`
	Host                        string          `json:"host" db:"host"`
	UpdatedOn                   time.Time       `json:"updatedOn" db:"updated_on"`
	Alias                       string          `json:"peerAlias" db:"node_alias"`
}

type ConnectPeerRequest struct {
	NodeId           int             `json:"nodeId"`
	ConnectionString string          `json:"connectionString"`
	Network          commons.Network `json:"network"`
}

type DisconnectPeerRequest struct {
	NodeId                      int `json:"nodeId"`
	NodeConnectionDetailsNodeId int `json:"nodeConnectionDetailsNodeId"`
}

type ReconnectPeerRequest struct {
	NodeId                      int `json:"nodeId"`
	NodeConnectionDetailsNodeId int `json:"nodeConnectionDetailsNodeId"`
}

func RegisterNodeRoutes(r *gin.RouterGroup, db *sqlx.DB, lightningRequestChannel chan<- interface{}) {
	r.GET("/:network/peers", func(c *gin.Context) { getAllPeersHandler(c, db) })
	r.POST("/peers/connect", func(c *gin.Context) { connectNewPeerHandler(c, db, lightningRequestChannel) })
	r.GET("/:network/nodes", func(c *gin.Context) { getNodesByNetworkHandler(c, db) })
	r.GET("/:network/walletBalances", func(c *gin.Context) { getNodesWalletBalancesHandler(c, db, lightningRequestChannel) })
	r.DELETE(":nodeId", func(c *gin.Context) { removeNodeHandler(c, db) })
	r.PATCH("/peers/disconnect", func(c *gin.Context) { disconnectPeerHandler(c, db, lightningRequestChannel) })
	r.PATCH("/peers/reconnect", func(c *gin.Context) { reconnectPeerHandler(c, db, lightningRequestChannel) })
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
	peerNodes, err := GetPeerNodes(db, commons.Network(network))
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting all Peer nodes.")
		return
	}

	c.JSON(http.StatusOK, peerNodes)
}

func connectNewPeerHandler(c *gin.Context, db *sqlx.DB, lightningRequestChannel chan<- interface{}) {

	var req ConnectPeerRequest

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

	r := commons.ConnectPeer(req.NodeId, pubKey, host, lightningRequestChannel)
	if r.CommunicationResponse.Error != "" {
		server_errors.WrapLogAndSendServerError(c, errors.New(r.CommunicationResponse.Error), "Error connecting to peer.")
		return
	}

	// save peer node
	node := Node{NodeId: req.NodeId, PublicKey: pubKey, Network: req.Network, Host: host}
	nodeId, err := AddNodeWhenNew(db, node)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node.")
		return
	}

	node.NodeId = nodeId
	node.ConnectionStatusId = commons.Active
	err = updateNodeConnectionStatus(db, node)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Updating node connection status.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully connected to peer.")})
}

func disconnectPeerHandler(c *gin.Context, db *sqlx.DB, lightningRequestChannel chan<- interface{}) {
	var req DisconnectPeerRequest
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequest(c, "Cannot process DisconnectPeerRequest")
		return
	}

	node, err := GetNodeById(db, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node by id.")
		return
	}

	// saving of the host is necessary if user wants to reconnect to the peer
	if node.Host == "" {
		host, err := getHostFromPeer(*node.NodeConnectionDetailsNodeId, node.PublicKey, lightningRequestChannel)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Getting host from peer.")
			return
		}
		node.Host = host
		err = updateNodeHost(db, node)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Updating node host.")
			return
		}
	}

	disconnectResp := commons.DisconnectPeer(req.NodeConnectionDetailsNodeId, node.PublicKey, lightningRequestChannel)
	if disconnectResp.CommunicationResponse.Error != "" {
		if disconnectResp.RequestFailedCurrentlyDisconnected {
			//correct the status of the node
			node.ConnectionStatusId = commons.Inactive
			err := updateNodeConnectionStatus(db, node)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Updating node connection status.")
				return
			}
		}
		server_errors.WrapLogAndSendServerError(c, errors.New(disconnectResp.CommunicationResponse.Error), "Error disconnecting peer.")
		return
	}

	node.ConnectionStatusId = commons.Inactive
	err = updateNodeConnectionStatus(db, node)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully disconnected peer.")})
}

func reconnectPeerHandler(c *gin.Context, db *sqlx.DB, lightningRequestChannel chan<- interface{}) {
	var req ReconnectPeerRequest
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequest(c, "Cannot process DisconnectPeerRequest")
		return
	}

	node, err := GetNodeById(db, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node by id.")
		return
	}

	if node.Host == "" && node.NodeConnectionDetailsNodeId != nil {
		host, hostErr := getHostFromPeer(*node.NodeConnectionDetailsNodeId, node.PublicKey, lightningRequestChannel)
		if hostErr != nil {
			server_errors.WrapLogAndSendServerError(c, hostErr, "Getting host from peer.")
			return
		}
		node.Host = host
		err = updateNodeHost(db, node)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Updating node host.")
			return
		}
	}

	connectResp := commons.ConnectPeer(req.NodeConnectionDetailsNodeId, node.PublicKey, node.Host, lightningRequestChannel)
	if connectResp.CommunicationResponse.Error != "" {
		if connectResp.RequestFailCurrentlyConnected {
			//correct the status of the node
			node.ConnectionStatusId = commons.Active
			err := updateNodeConnectionStatus(db, node)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Updating node connection status.")
				return
			}
		}
		server_errors.WrapLogAndSendServerError(c, errors.New(connectResp.CommunicationResponse.Error), "Error reconnecting peer.")
		return
	}

	node.ConnectionStatusId = commons.Active

	err = updateNodeConnectionStatus(db, node)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully disconnected peer.")})
}

func getHostFromPeer(connectionDetailsNodeId int, pubKey string, lightningRequestChannel chan<- interface{}) (string, error) {
	peersRsp := commons.ListPeers(connectionDetailsNodeId, lightningRequestChannel)
	p, ok := peersRsp.Peers[pubKey]
	if ok {
		return p.Address, nil
	}

	return "", errors.New("Peer public key not found.")
}
