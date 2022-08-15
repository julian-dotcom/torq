package forwards

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterForwardsRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getForwardsTableHandler(c, db) })
}
