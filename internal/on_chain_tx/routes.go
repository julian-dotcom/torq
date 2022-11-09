package on_chain_tx

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)


func RegisterOnChainTxsRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getOnChainTxsHandler(c, db) })
	r.POST("sendcoins", func(c *gin.Context) { sendCoinsHandler(c, db) })
}
