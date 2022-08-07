package lnd

import (
	"github.com/gin-gonic/gin"
	"github.com/lightningnetwork/lnd/lnrpc"
)

func RegisterControlChannelRoutes(r *gin.RouterGroup, client lnrpc.LightningClient) {
	r.POST("open", func(c *gin.Context) { OpenChannelHandler(c, client) })
	r.POST("close", func(c *gin.Context) { CloseChannelHandler(c, client) })
}
