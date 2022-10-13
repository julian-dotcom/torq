package peers

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"google.golang.org/grpc"
	"net/http"
	"strconv"
)

type LndAddress struct {
	PubKey string `json:"pubKey"`
	Host   string `json:"host"`
}

type ConnectPeerRequest struct {
	NodeId     int        `json:"nodeId"`
	LndAddress LndAddress `json:"lndAddress"`
	Perm       *bool      `json:"perm"`
	TimeOut    *uint64    `json:"timeOut"`
}

func connectPeerHandler(c *gin.Context, db *sqlx.DB) {
	var requestBody ConnectPeerRequest
	//var errAlreadyConnected = errors.New("Already connected")

	if err := c.BindJSON(&requestBody); err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}

	conn, err := connectLND(db, requestBody.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "can't connect to LND")
		return
	}
	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)
	ctx := context.Background()

	resp, err := ConnectPeer(client, ctx, requestBody)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Connect peer")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func listPeersHandler(c *gin.Context, db *sqlx.DB) {

	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	latestErr := c.Param("le")

	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting node id")
		return
	}

	conn, err := connectLND(db, nodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Connecting to LND")
		return
	}
	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)
	ctx := context.Background()

	resp, err := ListPeers(client, ctx, latestErr)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Listing peers")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func connectLND(db *sqlx.DB, nodeId int) (conn *grpc.ClientConn, err error) {
	connectionDetails, err := settings.GetNodeConnectionDetailsById(db, nodeId)

	if err != nil {
		return nil, errors.Wrap(err, "Getting node connection details from the db")
	}

	conn, err = lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Connecting to LND")
	}

	return conn, nil

}

func RegisterPeersRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("add", func(c *gin.Context) { connectPeerHandler(c, db) })
	r.GET("list/:nodeId/*le", func(c *gin.Context) { listPeersHandler(c, db) })
}
