package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterChannelRoutes(r *gin.RouterGroup, db *sqlx.DB, lightningRequestChannel chan interface{}) {
	r.PUT("update", func(c *gin.Context) { updateChannelsHandler(c, lightningRequestChannel) })
	r.POST("openbatch", func(c *gin.Context) { batchOpenHandler(c, db) })
	r.GET("", func(c *gin.Context) { getChannelListHandler(c, db) })
	r.GET("closedchannels", func(c *gin.Context) { getClosedChannelsListHandler(c, db) })
	r.GET("nodes", func(c *gin.Context) { getChannelAndNodeListHandler(c, db) })
}
