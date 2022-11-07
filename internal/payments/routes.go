package payments

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)


func RegisterPaymentsRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getPaymentsHandler(c, db) })
	r.GET(":identifier", func(c *gin.Context) { getPaymentHandler(c, db) })
}
