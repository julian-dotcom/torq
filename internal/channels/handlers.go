package channels

import (
	"encoding/hex"
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

type OpenChanRequestBody struct {
	LndAddress  string
	Amount      int64
	SatPerVbyte *uint64
}

type closeChanRequestBody struct {
	ChannelPoint string
	SatPerVbyte  *uint64
}

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

func CloseChannelHandler(c *gin.Context, db *sqlx.DB) {
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
	var requestBody closeChanRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}

	splitChanPoint := strings.Split(requestBody.ChannelPoint, ":")
	if len(splitChanPoint) != 2 {
		server_errors.LogAndSendServerError(c, errors.New("Channel point missing a colon"))
		return
	}

	fundingTxid := &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: splitChanPoint[0]}

	oIndxUint, err := strconv.ParseUint(splitChanPoint[1], 10, 1)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Parsing channel point output index")
		return
	}
	outputIndex := uint32(oIndxUint)

	log.Debug().Msgf("Funding: %v, index: %v", fundingTxid, outputIndex)

	resp, err := CloseChannel(client, fundingTxid, outputIndex, requestBody.SatPerVbyte)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Closing channel")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func UpdateChannelHandler(c *gin.Context, db *sqlx.DB) {
	requestBody := updateChanRequestBody{}
	if err := c.BindJSON(&requestBody); err != nil {
		log.Error().Msgf("JSON binding the request body")
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}

	if requestBody.ChannelPoint == "" {
		//log.Debug().Msgf("Invalid channel point")
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
	//log.Debug().Msgf("handler Response: %v", resp)
	c.JSON(http.StatusOK, reqResp)
}
