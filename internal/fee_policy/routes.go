package fee_policy

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterFeePolicyRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("", func(c *gin.Context) { addFeePolicyHandler(c, db) })
	r.GET("", func(c *gin.Context) { getFeePolicyListHandler(c, db) })
}

func addFeePolicyHandler(c *gin.Context, db *sqlx.DB) {
	return
}

func getFeePolicyListHandler(c *gin.Context, db *sqlx.DB) {
	return
}
