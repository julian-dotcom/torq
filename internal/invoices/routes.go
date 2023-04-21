package invoices

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterInvoicesRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getInvoicesHandler(c, db) })
	r.GET(":identifier", func(c *gin.Context) { getInvoiceHandler(c, db) })
}
