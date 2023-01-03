package corridors

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterCorridorRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getCorridorsHandler(c, db) })
	r.GET(":tagId", func(c *gin.Context) { getCorridorsByTagHandler(c, db) })
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

func getCorridorsByTagHandler(c *gin.Context, db *sqlx.DB) {
	tagId, err := strconv.Atoi(c.Param("tagId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse getCorridorsByTagId in the request.")
		return
	}
	var corridorsWithTotal CorridorNodesChannels
	corridors, err := getCorridorsByTagId(db, tagId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting corridors for getCorridorsByTagId: %v", tagId))
		return
	}
	totalNodes, err := getCorridorsNodesByTagId(db, tagId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting corridors for getCorridorsNodesByTagId: %v", tagId))
		return
	}
	TotalChannels, err := getCorridorsChannelsByTagId(db, tagId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting corridors for getCorridorsChannelsByTagId: %v", tagId))
		return
	}

	corridorsWithTotal.Corridors = corridors
	corridorsWithTotal.TotalChannels = TotalChannels
	corridorsWithTotal.TotalNodes = totalNodes

	c.JSON(http.StatusOK, corridorsWithTotal)
}
