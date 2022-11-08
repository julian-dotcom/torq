package nodes

import (
	"fmt"
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
	NodeId    int    `json:"nodeId"`
	PublicKey string `json:"publicKey"`
	Alias     string `json:"alias"`
	Color     string `json:"color"`
}

func RegisterNodeRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getNodesHandler(c, db) })
	r.DELETE(":nodeId", func(c *gin.Context) { removeNodeHandler(c, db) })
}

func getNodesHandler(c *gin.Context, db *sqlx.DB) {
	nds, err := getNodeInformationAll(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting nodes.")
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
