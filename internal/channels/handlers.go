package channels

import (
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
)

type OpenChanRequestBody struct {
	LndAddress  string
	Amount      int64
	SatPerVbyte *uint64
}

type Response struct {
	Response string
}

func OpenChannelHandler(c *gin.Context, db *sqlx.DB) {
	connectionDetails, err := settings.GetConnectionDetails(db)
	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Connecting to LND")
		return
	}

	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)
	var requestBody OpenChanRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}

	pubKeyHex, err := hex.DecodeString(requestBody.LndAddress)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Decoding public key hex")
		return
	}

	resp, err := OpenChannel(client, pubKeyHex, requestBody.Amount, requestBody.SatPerVbyte)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Opening channel")
		return
	}

	c.JSON(http.StatusOK, resp)

}
