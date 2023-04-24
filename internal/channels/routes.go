package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterChannelRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getChannelListHandler(c) })
	r.GET("closed", func(c *gin.Context) { getClosedChannelsListHandler(c, db) })
	r.GET("pending", func(c *gin.Context) { getChannelsPendingListHandler(c, db) })
	r.GET("nodes", func(c *gin.Context) { getChannelAndNodeListHandler(c, db) })
}
