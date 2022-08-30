package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"github.com/rs/zerolog/log"
	"net/http"
)

type UpdateResponse struct {
	Status        string
	FailedUpdates []failedUpdate
	ChanPoint     string
}

func UpdateChannelHandler(c *gin.Context, db *sqlx.DB) {
	requestBody := updateChanRequestBody{}

	if err := c.BindJSON(&requestBody); err != nil {
		log.Error().Msgf("JSON binding the request body")
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}

	connectionDetails, err := settings.GetConnectionDetails(db)
	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Connecting to LND")
	}

	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)

	resp, err := UpdateChannel(client, requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	reqResp := UpdateResponse{
		Status:        resp.Status,
		FailedUpdates: resp.FailedUpdates,
	}
	if requestBody.ChannelPoint != nil {
		reqResp.ChanPoint = *requestBody.ChannelPoint
	}

	c.JSON(http.StatusOK, reqResp)
}
