package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"github.com/rs/zerolog/log"
	"net/http"
)

type failedUpdate struct {
	OutPoint struct {
		Txid    string
		OutIndx uint32
	}
	Reason      string
	UpdateError string
}

type updateResponse struct {
	Status        string         `json:"status"`
	FailedUpdates []failedUpdate `json:"failedUpdates"`
}

type updateChanRequestBody struct {
	ChannelPoint  *string `json:"channelPoint"`
	FeeRatePpm    *uint32 `json:"feeRatePpm"`
	BaseFeeMsat   *int64  `json:"baseFeeMsat"`
	MaxHtlcMsat   *uint64 `json:"maxHtlcMsat"`
	MinHtlcMsat   *uint64 `json:"minHtlcMsat"`
	TimeLockDelta uint32  `json:"timeLockDelta"`
}

func updateChannelsHandler(c *gin.Context, db *sqlx.DB) {
	requestBody := updateChanRequestBody{}

	if err := c.BindJSON(&requestBody); err != nil {
		log.Error().Msgf("JSON binding the request body")
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}
	//log.Debug().Msgf("Received request body: %v", requestBody)

	response, err := updateChannels(db, requestBody)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Update channel/s policy")
		return
	}

	c.JSON(http.StatusOK, response)
}

func RegisterChannelRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("update", func(c *gin.Context) { updateChannelsHandler(c, db) })
	//r.POST("openbatch", func(c *gin.Context) { batchOpenHandler(c, db) })
}
