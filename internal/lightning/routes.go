package lightning

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterLightningRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.PUT("updateRoutingPolicy", func(c *gin.Context) { updateRoutingPolicyHandler(c, db) })
	r.GET("/:network/walletBalances", func(c *gin.Context) { getNodesWalletBalancesHandler(c, db) })
	r.POST("/peers/connect", func(c *gin.Context) { connectNewPeerHandler(c, db) })
	r.PATCH("/peers/disconnect", func(c *gin.Context) { disconnectPeerHandler(c, db) })
	r.PATCH("/peers/reconnect", func(c *gin.Context) { reconnectPeerHandler(c, db) })
	r.PATCH("/peers/update", func(c *gin.Context) { updatePeer(c, db) })
}
