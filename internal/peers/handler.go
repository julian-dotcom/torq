package peers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/lightning"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/pkg/server_errors"
)

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

type ConnectionStatus int

const (
	Disconnected = ConnectionStatus(iota)
	Connected
)

func (c ConnectionStatus) String() string {
	switch c {
	case Connected:
		return "Connected"
	case Disconnected:
		return "Disconnected"
	default:
		return "unknown"
	}
}

func getAllPeersHandler(c *gin.Context, db *sqlx.DB) {
	network, err := strconv.Atoi(c.Query("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
		return
	}
	peerNodes, err := GetPeerNodes(db, core.Network(network))
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting all Peer nodes.")
		return
	}
	for peerIndex, peer := range peerNodes {
		peerNodes[peerIndex].Tags = tags.GetTagsByTagIds(cache.GetTagIdsByNodeId(peer.NodeId))
	}

	c.JSON(http.StatusOK, peerNodes)
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

	_, err := lightning.ConnectPeer(req.NodeId, pubKey, host)
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
	connected := core.NodeConnectionStatusConnected
	err = settings.AddNodeConnectionHistory(db, req.NodeId, nodeId, &host, &req.Setting, &connected)
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

	address, setting, _, err := settings.GetNodeConnectionHistoryWithDetail(db, req.TorqNodeId, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node by id.")
		return
	}

	if address == nil || *address == "" {
		host, err := getHostFromPeer(req.TorqNodeId, req.NodeId)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Getting host from peer.")
			return
		}
		address = &host
	}

	disconnected := core.NodeConnectionStatusDisconnected
	requestFailedCurrentlyDisconnected, err := lightning.DisconnectPeer(req.TorqNodeId, req.NodeId)
	if err != nil {
		if requestFailedCurrentlyDisconnected {
			err = settings.AddNodeConnectionHistory(db, req.TorqNodeId, req.NodeId, address, setting, &disconnected)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Saving peer node connection history.")
				return
			}
		}
		server_errors.WrapLogAndSendServerError(c, err, "Error disconnecting peer.")
		return
	}

	err = settings.AddNodeConnectionHistory(db, req.TorqNodeId, req.NodeId, address, setting, &disconnected)
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

	address, setting, _, err := settings.GetNodeConnectionHistoryWithDetail(db, req.TorqNodeId, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node connection history by id.")
		return
	}

	if address == nil || *address == "" {
		host, err := getHostFromPeer(req.TorqNodeId, req.NodeId)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Getting host from peer.")
			return
		}
		address = &host
	}

	connected := core.NodeConnectionStatusConnected
	publicKey := cache.GetNodeSettingsByNodeId(req.NodeId).PublicKey
	requestFailCurrentlyConnected, err := lightning.ConnectPeer(req.TorqNodeId, publicKey, *address)
	if err != nil {
		if requestFailCurrentlyConnected {
			err = settings.AddNodeConnectionHistory(db, req.TorqNodeId, req.NodeId, address, setting, &connected)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Saving peer node connection history.")
				return
			}
		}
		server_errors.WrapLogAndSendServerError(c, err, "Updating Peer, already connected.")
		return
	}

	err = settings.AddNodeConnectionHistory(db, req.TorqNodeId, req.NodeId, address, setting, &connected)
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

	address, _, status, err := settings.GetNodeConnectionHistoryWithDetail(db, req.TorqNodeId, req.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node connection history by id.")
		return
	}

	if address == nil || *address == "" {
		host, err := getHostFromPeer(req.TorqNodeId, req.NodeId)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Getting host from peer.")
			return
		}
		address = &host
	}

	err = settings.AddNodeConnectionHistory(db, req.TorqNodeId, req.NodeId, address, req.Setting, status)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "updating peer node.")
		return
	}

}

func getHostFromPeer(connectionDetailsNodeId int, nodeId int) (string, error) {
	nodeSettings := cache.GetNodeSettingsByNodeId(nodeId)
	peers, err := lightning.ListPeers(connectionDetailsNodeId, true)
	if err != nil {
		return "", errors.Wrap(err, "Getting list of peers.")
	}
	p, ok := peers[nodeSettings.PublicKey]
	if ok {
		return p.Address, nil
	}
	return "", errors.New("Peer public key not found.")
}
