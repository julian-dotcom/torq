package views

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterTableViewRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getTableViewsHandler(c, db) })
	r.POST("", func(c *gin.Context) { insertTableViewsHandler(c, db) })
	r.PUT("", func(c *gin.Context) { updateTableViewsHandler(c, db) })
	r.DELETE(":viewId", func(c *gin.Context) { deleteTableViewsHandler(c, db) })
}
