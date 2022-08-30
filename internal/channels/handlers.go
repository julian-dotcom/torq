package channels

import (
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
)

type updateChanRequestBody struct {
	ChannelPoint  string
	FeeRate       float64
	TimeLockDelta uint32
}

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

	if requestBody.ChannelPoint == "" {
		server_errors.LogAndSendServerError(c, errors.New("Invalid channel point"))
		return
	}

	splitChanPoint := strings.Split(requestBody.ChannelPoint, ":")
	if len(splitChanPoint) != 2 {
		server_errors.LogAndSendServerError(c, errors.New("Channel point missing a colon"))
		return
	}

	fundingTxid := splitChanPoint[0]

	oIndxUint, err := strconv.ParseUint(splitChanPoint[1], 10, 1)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Parsing channel point output index")
		return
	}
	outputIndex := uint32(oIndxUint)

	timeLock := requestBody.TimeLockDelta

	//Minimum supported value for TimeLockDelta supported is 18
	if timeLock < 18 {
		timeLock = 18
	}
	//log.Debug().Msgf("Funding: %v, index: %v", fundingTxid, outputIndex)

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

	resp, err := UpdateChannel(client, fundingTxid, outputIndex, requestBody.FeeRate, timeLock)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Closing channel")
		return
	}

	reqResp := UpdateResponse{
		Status:        resp.Status,
		FailedUpdates: resp.FailedUpdates,
		ChanPoint:     requestBody.ChannelPoint,
	}

	c.JSON(http.StatusOK, reqResp)
}
