package channel_history

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterChannelHistoryRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET(":chanIds", func(c *gin.Context) { getChannelHistoryHandler(c, db) })
}

func RegisterRebalancingTotals(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET(":chanIds/rebalancing", func(c *gin.Context) { getChannelRebalancingTotalHandler(c, db) })
}
