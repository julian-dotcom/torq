package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterControlChannelRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("open", func(c *gin.Context) { OpenChannelHandler(c, db) })
	r.POST("close", func(c *gin.Context) { CloseChannelHandler(c, db) })
}
