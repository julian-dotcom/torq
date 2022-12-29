package automation

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterAutomationRoutes(r *gin.RouterGroup, db *sqlx.DB, eventChannel chan interface{}) {
	r.POST("rebalance", func(c *gin.Context) { rebalanceHandler(c, db, eventChannel) })
}

func rebalanceHandler(c *gin.Context, db *sqlx.DB, eventChannel chan interface{}) {
	var rr commons.RebalanceRequest
	if err := c.BindJSON(&rr); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	responseChannel := make(chan commons.RebalanceResponse)
	rr.ResponseChannel = responseChannel
	eventChannel <- rr
	response := <-responseChannel
	if response.Error != "" {
		server_errors.SendBadRequest(c, response.Error)
		return
	}
	c.JSON(http.StatusOK, response)
}
