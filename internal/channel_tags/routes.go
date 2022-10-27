package channel_tags

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterChannelTagRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("", func(c *gin.Context) { addChannelTagHandler(c, db) })
	r.DELETE(":channelTagId", func(c *gin.Context) { removeChannelTagHandler(c, db) })
}

func addChannelTagHandler(c *gin.Context, db *sqlx.DB) {
	var ct channelTag
	if err := c.BindJSON(&ct); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	if ct.TagId == 0 {
		server_errors.SendUnprocessableEntity(c, "Failed to find tagId in the request.")
		return
	}
	if ct.FromNodeId == 0 && ct.ToNodeId == 0 && ct.ChannelId == 0 {
		server_errors.SendUnprocessableEntity(c, "Failed to find fromNodeId, toNodeId or channelId in the request.")
		return
	}
	if ct.ToNodeId != 0 && ct.FromNodeId == 0 {
		server_errors.SendUnprocessableEntity(c, "Failed to find fromNodeId in the request.")
		return
	}
	if ct.ChannelId != 0 {
		if ct.FromNodeId == 0 {
			server_errors.SendUnprocessableEntity(c, "Failed to find fromNodeId in the request.")
			return
		}
		if ct.ToNodeId == 0 {
			server_errors.SendUnprocessableEntity(c, "Failed to find toNodeId in the request.")
			return
		}
	}
	corridor := corridors.Corridor{CorridorTypeId: corridors.Tag().CorridorTypeId, Flag: 1}
	corridor.ReferenceId = &ct.TagId
	if ct.FromNodeId != 0 {
		corridor.FromNodeId = &ct.FromNodeId
	}
	if ct.ToNodeId != 0 {
		corridor.ToNodeId = &ct.ToNodeId
	}
	if ct.ChannelId != 0 {
		corridor.ChannelId = &ct.ChannelId
	}
	_, err := corridors.AddCorridor(db, corridor)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding corridor.")
		return
	}
	err = corridors.RefreshCorridorCacheByTypeId(db, corridor.CorridorTypeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Refresh Corridor Cache By Type.")
		return
	}
	go func() {
		err := GenerateChannelTag(db)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate channel tags.")
		}
	}()
	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully added channel tag configuration."})
}

func removeChannelTagHandler(c *gin.Context, db *sqlx.DB) {
	channelTagId, err := strconv.Atoi(c.Param("channelTagId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse channelTagId in the request.")
		return
	}
	ct, err := getChannelTag(db, channelTagId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Obtaining channelTag for channelTagId: %v", channelTagId))
		return
	}
	corridorKey := corridors.CorridorKey{CorridorType: corridors.Tag()}
	corridorKey.ReferenceId = ct.TagId
	corridorKey.FromNodeId = ct.FromNodeId
	corridorKey.ToNodeId = ct.ToNodeId
	corridorKey.ChannelId = ct.ChannelId
	corridor := corridors.GetBestCorridor(corridorKey)
	_, err = corridors.RemoveCorridor(db, corridor.CorridorId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Removing corridor with corridorId: %v", corridor.CorridorId))
		return
	}
	err = corridors.RefreshCorridorCacheByType(db, corridors.Tag())
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Refresh Corridor Cache By Type.")
		return
	}
	go func() {
		err := GenerateChannelTag(db)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate channel tags.")
		}
	}()
	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully deleted tag(s)."})
}
