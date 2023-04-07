package automation

import (
	"context"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterAutomationRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("rebalance", func(c *gin.Context) { rebalanceHandler(c, db) })
}

func rebalanceHandler(c *gin.Context, db *sqlx.DB) {
	var rr core.RebalanceRequests
	if err := c.BindJSON(&rr); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	if rr.NodeId == 0 {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	if !cache.IsChannelBalanceCacheStreamActive(rr.NodeId) {
		server_errors.SendBadRequest(c, "The node for the provided nodeId is not active (yet?).")
		return
	}

	response := workflows.RebalanceRequests(context.Background(), db, rr, rr.NodeId)
	if len(response) > 0 && response[0].Error != "" {
		server_errors.SendBadRequest(c, response[0].Error)
		return
	}
	c.JSON(http.StatusOK, response)
}
