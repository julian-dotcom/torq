package channel_history

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterChannelHistoryRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET(":chanIds/history", func(c *gin.Context) { getChannelHistoryHandler(c, db) })
	r.GET(":chanIds/event", func(c *gin.Context) { getChannelEventHistoryHandler(c, db) })
	r.GET(":chanIds/balance", func(c *gin.Context) { getChannelBalanceHandler(c, db) })
	r.GET(":chanIds/rebalancing", func(c *gin.Context) { getChannelReBalancingHandler(c, db) })
	r.GET(":chanIds/onchaincost", func(c *gin.Context) { getTotalOnchainCostHandler(c, db) })
}
