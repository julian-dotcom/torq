package nodes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/lightning"
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

type PeerNode struct {
	NodeId           int                             `json:"nodeId" db:"node_id"`
	Alias            string                          `json:"peerAlias" db:"alias"`
	PublicKey        string                          `json:"pubKey" db:"public_key"`
	TorqNodeId       *int                            `json:"torqNodeId" db:"torq_node_id"`
	TorqNodeAlias    *string                         `json:"torqNodeAlias" db:"torq_node_alias"`
	Setting          *cache.LndNodeConnectionDetails `json:"setting" db:"setting"`
	ConnectionStatus *core.Status                    `json:"connectionStatus" db:"connection_status"`
	Address          *string                         `json:"address" db:"address"`
}

type ConnectPeerRequest struct {
	NodeId           int                            `json:"nodeId"`
	ConnectionString string                         `json:"connectionString"`
	Network          core.Network                   `json:"network"`
	Setting          cache.LndNodeConnectionDetails `json:"setting"`
}

type DisconnectPeerRequest struct {
	NodeId     int `json:"nodeId"`
	TorqNodeId int `json:"torqNodeId"`
}

type ReconnectPeerRequest struct {
	NodeId     int `json:"nodeId"`
	TorqNodeId int `json:"torqNodeId"`
}

type UpdatePeerRequest struct {
	NodeId     int                             `json:"nodeId"`
	TorqNodeId int                             `json:"torqNodeId"`
	Setting    *cache.LndNodeConnectionDetails `json:"setting"`
}

func RegisterNodeRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("/:network/peers", func(c *gin.Context) { getAllPeersHandler(c, db) })
	r.POST("/peers/connect", func(c *gin.Context) { connectNewPeerHandler(c, db) })
	r.GET("/:network/nodes", func(c *gin.Context) { getNodesByNetworkHandler(c, db) })
	r.DELETE(":nodeId", func(c *gin.Context) { removeNodeHandler(c, db) })
	r.PATCH("/peers/disconnect", func(c *gin.Context) { disconnectPeerHandler(c, db) })
	r.PATCH("/peers/reconnect", func(c *gin.Context) { reconnectPeerHandler(c, db) })
	r.PATCH("/peers/update", func(c *gin.Context) { updatePeer(c, db) })
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

func getAllPeersHandler(c *gin.Context, db *sqlx.DB) {
	network, err := strconv.Atoi(c.Param("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
		return
	}
	peerNodes, err := GetPeerNodes(db, core.Network(network))
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting all Peer nodes.")
		return
	}

	c.JSON(http.StatusOK, peerNodes)
}

func connectNewPeerHandler(c *gin.Context, db *sqlx.DB) {

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

	_, err := lightning.ConnectPeer(req.NodeId, pubKey, host)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Error connecting to peer.")
		return
	}

	// save peer node
	node := Node{NodeId: req.NodeId, PublicKey: pubKey, Network: req.Network}
	peerConnectionHistory := &NodeConnectionHistory{
		NodeId:           node.NodeId,
		TorqNodeId:       req.NodeId,
		Address:          host,
		ConnectionStatus: core.NodeConnectionStatusConnected,
		Setting:          req.Setting,
	}
	_, err = AddNodeWhenNew(db, node, peerConnectionHistory)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node.")
		return
	}

	//save node connection detail
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Updating node connection status.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully connected to peer.")})
}

func disconnectPeerHandler(c *gin.Context, db *sqlx.DB) {
	var req DisconnectPeerRequest
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequest(c, "Cannot process DisconnectPeerRequest")
		return
	}

	nch, err := getNodeConnectionHistoryWithDetail(db, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node by id.")
		return
	}

	if nch.Address == "" {
		host, err := getHostFromPeer(nch.TorqNodeId, nch.PubKey)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Getting host from peer.")
			return
		}
		nch.Address = host
	}

	requestFailedCurrentlyDisconnected, err := lightning.DisconnectPeer(req.TorqNodeId, nch.PubKey)
	if err != nil {
		if requestFailedCurrentlyDisconnected {
			//correct the status of the node
			err := addNodeConnectionHistory(db, &NodeConnectionHistory{
				NodeId:           nch.NodeId,
				TorqNodeId:       nch.TorqNodeId,
				Address:          nch.Address,
				Setting:          nch.Setting,
				ConnectionStatus: core.NodeConnectionStatusDisconnected,
			})
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Updating node connection status.")
				return
			}
		}
		server_errors.WrapLogAndSendServerError(c, err, "Error disconnecting peer.")
		return
	}

	err = addNodeConnectionHistory(db, &NodeConnectionHistory{
		NodeId:           nch.NodeId,
		TorqNodeId:       nch.TorqNodeId,
		Address:          nch.Address,
		Setting:          nch.Setting,
		ConnectionStatus: core.NodeConnectionStatusDisconnected,
	})
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully disconnected peer.")})
}

func reconnectPeerHandler(c *gin.Context, db *sqlx.DB) {
	var req ReconnectPeerRequest
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequest(c, "Cannot process DisconnectPeerRequest")
		return
	}

	nch, err := getNodeConnectionHistoryWithDetail(db, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node connection history by id.")
		return
	}

	if nch.Address == "" {
		host, err := getHostFromPeer(nch.TorqNodeId, nch.PubKey)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Getting host from peer.")
			return
		}
		nch.Address = host
	}

	requestFailCurrentlyConnected, err := lightning.ConnectPeer(req.TorqNodeId, nch.PubKey, nch.Address)
	if err != nil {
		if requestFailCurrentlyConnected {
			//correct the status of the node
			err = addNodeConnectionHistory(db, &NodeConnectionHistory{
				NodeId:           nch.NodeId,
				TorqNodeId:       nch.TorqNodeId,
				Address:          nch.Address,
				Setting:          nch.Setting,
				ConnectionStatus: core.NodeConnectionStatusConnected,
			})

			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Updating node connection status.")
				return
			}
		}
		server_errors.WrapLogAndSendServerError(c, err, "Updating Peer, already connected.")
		return
	}

	err = addNodeConnectionHistory(db, &NodeConnectionHistory{
		NodeId:           nch.NodeId,
		TorqNodeId:       nch.TorqNodeId,
		Address:          nch.Address,
		Setting:          nch.Setting,
		ConnectionStatus: core.NodeConnectionStatusConnected,
	})

	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Saving peer node.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully disconnected peer.")})
}

func updatePeer(c *gin.Context, db *sqlx.DB) {
	var req UpdatePeerRequest
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

	nch, err := getNodeConnectionHistoryWithDetail(db, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node connection history by id.")
		return
	}

	if nch.Address == "" {
		host, err := getHostFromPeer(nch.TorqNodeId, nch.PubKey)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Getting host from peer.")
			return
		}
		nch.Address = host
	}

	err = addNodeConnectionHistory(db, &NodeConnectionHistory{
		NodeId:           nch.NodeId,
		TorqNodeId:       nch.TorqNodeId,
		Address:          nch.Address,
		Setting:          *req.Setting,
		ConnectionStatus: nch.ConnectionStatus,
	})

	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "updating peer node.")
		return
	}

}

func getHostFromPeer(connectionDetailsNodeId int, pubKey string) (string, error) {
	peers, err := lightning.ListPeers(connectionDetailsNodeId)
	if err != nil {
		return "", errors.Wrap(err, "Getting list of peers.")
	}
	p, ok := peers[pubKey]
	if ok {
		return p.Address, nil
	}
	return "", errors.New("Peer public key not found.")
}
