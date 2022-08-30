package channels

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterControlChannelRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("update", func(c *gin.Context) { UpdateChannelHandler(c, db) })
}
