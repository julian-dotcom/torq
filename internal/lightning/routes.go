package lightning

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterLightningRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.PUT("updateRoutingPolicy", func(c *gin.Context) { updateRoutingPolicyHandler(c, db) })
	r.GET("/:network/walletBalances", func(c *gin.Context) { getNodesWalletBalancesHandler(c, db) })
}
