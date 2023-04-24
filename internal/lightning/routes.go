package lightning

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterLightningRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("open", func(c *gin.Context) { openChannelHandler(c) })
	r.POST("openbatch", func(c *gin.Context) { batchOpenHandler(c) })
	r.POST("close", func(c *gin.Context) { closeChannelHandler(c, db) })
	r.PUT("updateRoutingPolicy", func(c *gin.Context) { updateRoutingPolicyHandler(c, db) })
	r.GET("/:network/walletBalances", func(c *gin.Context) { getNodesWalletBalancesHandler(c) })
	r.POST("newinvoice", func(c *gin.Context) { newInvoiceHandler(c) })
	r.GET("decode", func(c *gin.Context) { decodeInvoiceHandler(c) })
	r.POST("sendcoins", func(c *gin.Context) { sendCoinsHandler(c) })
	r.POST("new-address", func(c *gin.Context) { newAddressHandler(c) })
}
