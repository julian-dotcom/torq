package lnd

import (
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
)

type OpenChanRequestBody struct {
	LndAddress string
	Amount     int64
}

type closeChanRequestBody struct {
	ChannelPoint string
}

type Response struct {
	Response string
}

func OpenChannelHandler(c *gin.Context, client lnrpc.LightningClient) {
	var requestBody OpenChanRequestBody
	r := Response{}

	//Get request body
	if err := c.BindJSON(&requestBody); err != nil {
		log.Error().Msgf("Error getting request body")
		r.Response = "Error getting request body"
		c.JSON(http.StatusInternalServerError, r)
	}

	lndAddressStr := requestBody.LndAddress
	fundingAmt := requestBody.Amount

	//pubkey to hex
	pubKeyHex, err := hex.DecodeString(lndAddressStr)
	if err != nil {
		log.Error().Msgf("Unable to decode node public key: %v", err)
		r.Response = "Error hexing pubkey"
		c.JSON(http.StatusInternalServerError, r)
	}

	//resp, err := openChannel(pubKeyHex, fundingAmt, db)
	resp, err := openChannel(pubKeyHex, fundingAmt, client)
	if err != nil {
		log.Error().Msgf("Error opening channel: %v", err)
		r.Response = "Error opening channel"
		c.JSON(http.StatusInternalServerError, r)
	}

	c.JSON(http.StatusOK, resp)

}

func CloseChannelHandler(c *gin.Context, client lnrpc.LightningClient) {
	var requestBody closeChanRequestBody
	r := Response{}

	if err := c.BindJSON(&requestBody); err != nil {
		log.Error().Msgf("Error getting request body")
	}

	splitChanPoint := strings.Split(requestBody.ChannelPoint, ":")
	if len(splitChanPoint) != 2 {
		log.Error().Msgf("Wrong chanpoint format")
		r.Response = "Wrong chanpoint format"
		c.JSON(http.StatusInternalServerError, r)
		return
	}
	log.Debug().Msgf("split: %v - %v", splitChanPoint[0], splitChanPoint[1])

	fundingTxid := &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: splitChanPoint[0]}

	oIndxUint, err := strconv.ParseUint(splitChanPoint[1], 10, 1)
	if err != nil {
		log.Error().Msgf("Error parsing output index to uint: %v", err)
		r.Response = "Error parsing output index to uint"
		c.JSON(http.StatusInternalServerError, r)
		return
	}
	outputIndex := uint32(oIndxUint)

	log.Debug().Msgf("Funding: %v, index: %v", fundingTxid, outputIndex)

	resp, err := closeChannel(fundingTxid, outputIndex, client)

	if err != nil {
		log.Error().Msgf("Error closing channel: %v", err)
		r.Response = "Error closing channel"
		c.JSON(http.StatusInternalServerError, r)
	}

	c.JSON(http.StatusOK, resp)
}
