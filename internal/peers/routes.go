package peers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterPeerRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getAllPeersHandler(c, db) })
	r.POST("/connect", func(c *gin.Context) { connectNewPeerHandler(c, db) })
	r.PATCH("/disconnect", func(c *gin.Context) { disconnectPeerHandler(c, db) })
	r.PATCH("/reconnect", func(c *gin.Context) { reconnectPeerHandler(c, db) })
	r.PATCH("/update", func(c *gin.Context) { updatePeer(c, db) })
}
