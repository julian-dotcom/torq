package automation

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterAutomationRoutes(r *gin.RouterGroup, rebalanceRequestChannel chan<- commons.RebalanceRequests) {
	r.POST("rebalance", func(c *gin.Context) { rebalanceHandler(c, rebalanceRequestChannel) })
}

func rebalanceHandler(c *gin.Context, rebalanceRequestChannel chan<- commons.RebalanceRequests) {
	var rr commons.RebalanceRequests
	if err := c.BindJSON(&rr); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	if rr.NodeId == 0 {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(rr.NodeId) != commons.ServiceActive {
		server_errors.SendBadRequest(c, "The node for the provided nodeId is not active (yet?).")
		return
	}

	responseChannel := make(chan []commons.RebalanceResponse)
	rr.ResponseChannel = responseChannel
	rebalanceRequestChannel <- rr
	response := <-responseChannel
	if len(response) > 0 && response[0].Error != "" {
		server_errors.SendBadRequest(c, response[0].Error)
		return
	}
	c.JSON(http.StatusOK, response)
}
