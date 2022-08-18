package channels

import (
	"encoding/hex"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"net/http"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/server_errors"
	"github.com/rs/zerolog/log"
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
