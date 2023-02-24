package automation

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterAutomationRoutes(r *gin.RouterGroup, db *sqlx.DB, rebalanceRequestChannel chan commons.RebalanceRequest) {
	r.POST("rebalance", func(c *gin.Context) { rebalanceHandler(c, db, rebalanceRequestChannel) })
}

func rebalanceHandler(c *gin.Context, db *sqlx.DB, rebalanceRequestChannel chan commons.RebalanceRequest) {
	var rr commons.RebalanceRequest
	if err := c.BindJSON(&rr); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	if rr.NodeId == 0 {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(rr.NodeId) != commons.Active {
		server_errors.SendBadRequest(c, "The node for the provided nodeId is not active (yet?).")
		return
	}

	responseChannel := make(chan commons.RebalanceResponse)
	rr.ResponseChannel = responseChannel
	rebalanceRequestChannel <- rr
	response := <-responseChannel
	if response.Error != "" {
		server_errors.SendBadRequest(c, response.Error)
		return
	}
	c.JSON(http.StatusOK, response)
}
