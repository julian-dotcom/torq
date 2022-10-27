package corridors

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
	"strconv"
)

func RegisterCorridorRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getCorridorsHandler(c, db) })
}

func getCorridorsHandler(c *gin.Context, db *sqlx.DB) {
	corridorTypeId, err := strconv.Atoi(c.Param("corridorTypeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse corridorTypeId in the request.")
		return
	}
	corridors, err := getCorridorsByCorridorTypeId(db, corridorTypeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting corridors for corridorTypeId: %v", corridorTypeId))
		return
	}
	c.JSON(http.StatusOK, corridors)
}
