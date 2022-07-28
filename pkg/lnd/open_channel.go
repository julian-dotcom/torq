package lnd

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type OpenChanRequestBody struct {
	LndAddress string
	Amount     int64
}

func openChannel(c *gin.Context, db *sqlx.DB) {
	var requestBody OpenChanRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		log.Error().Msgf("Error getting request body")
	}

	lndAddressStr := requestBody.LndAddress
	fundingAmt := requestBody.Amount
	fmt.Println(lndAddressStr)

	pubKeyHex, err := hex.DecodeString(lndAddressStr)
	if err != nil {
		log.Error().Msgf("Unable to decode node public key: %v", err)
	}

	openChanReq := lnrpc.OpenChannelRequest{
		NodePubkey:         pubKeyHex,
		LocalFundingAmount: fundingAmt,
	}
	connectionDetails, err := settings.GetConnectionDetails(db)
	ctx := context.Background()
	//errs, ctx := errgroup.WithContext(ctx)
	conn, err := Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	client := lnrpc.NewLightningClient(conn)
	openChanRes, err := client.OpenChannel(ctx, &openChanReq)
	if err != nil {
		log.Error().Msgf("Err opening channel: %v", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Error().Msgf("Context error: %v", ctx.Err())
				break
			default:
			}

			resp, err := openChanRes.Recv()
			if err == io.EOF {

				log.Debug().Msgf("Open channel EOF")
				break
			}
			if err != nil {
				//m["error"] = fmt.Sprintf("%v", err)
				//c.JSON(http.StatusNotImplemented, m)
				log.Error().Msgf("Err receive %v", err.Error())
				break
			}
			log.Debug().Msgf("Channel openning status: %v", resp.String())

		}
		//close(waitc)
	}()
	r := make(map[string]string)
	r["response"] = "Channel openning"
	c.JSON(http.StatusOK, r)
}
