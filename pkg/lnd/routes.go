package lnd

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterOpenChannelRoute(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("", func(c *gin.Context) { openChannel(c, db) })
}
