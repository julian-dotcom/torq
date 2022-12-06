package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterChannelRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.PUT("update", func(c *gin.Context) { updateChannelsHandler(c, db) })
	r.POST("openbatch", func(c *gin.Context) { batchOpenHandler(c, db) })
	r.GET("", func(c *gin.Context) { getChannelListHandler(c, db) })
}
