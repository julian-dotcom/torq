package channel_groups

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

func RegisterChannelGroupRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("", func(c *gin.Context) { addChannelGroupHandler(c, db) })
	r.DELETE("/channelGroup/:channelGroupId", func(c *gin.Context) { removeChannelGroupHandler(c, db) })
	r.DELETE("/category/:categoryId", func(c *gin.Context) { removeCategoryHandler(c, db) })
	r.DELETE("/tag/:tagId", func(c *gin.Context) { removeTagHandler(c, db) })
}

func addChannelGroupHandler(c *gin.Context, db *sqlx.DB) {
	var cg channelGroup
	if err := c.BindJSON(&cg); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	if cg.TagId == 0 && cg.CategoryId == 0 {
		server_errors.SendUnprocessableEntity(c, "Failed to find tagId or categoryId in the request.")
		return
	}
	if cg.NodeId == 0 && cg.ChannelId == 0 {
		server_errors.SendUnprocessableEntity(c, "Failed to find nodeId or channelId in the request.")
		return
	}
	if cg.ChannelId != 0 && cg.NodeId == 0 {
		server_errors.SendUnprocessableEntity(c, "Failed to find nodeId in the request.")
		return
	}
	corridor := corridors.Corridor{CorridorTypeId: corridors.Tag().CorridorTypeId, Flag: 1}
	corridor.ReferenceId = &cg.TagId
	if cg.NodeId != 0 {
		corridor.FromNodeId = &cg.NodeId
	}
	if cg.CategoryId != 0 {
		corridor.FromCategoryId = &cg.CategoryId
	}
	if cg.ChannelId != 0 {
		corridor.ChannelId = &cg.ChannelId
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
		err := GenerateChannelGroups(db)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate channel groups.")
		}
	}()
	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully added channel group configuration."})
}

func removeChannelGroupHandler(c *gin.Context, db *sqlx.DB) {
	channelGroupId, err := strconv.Atoi(c.Param("channelGroupId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse channelGroupId in the request.")
		return
	}
	ct, err := getChannelGroup(db, channelGroupId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Obtaining channelGroup for channelGroupId: %v", channelGroupId))
		return
	}
	corridorKey := corridors.CorridorKey{CorridorType: corridors.Tag()}
	corridorKey.ReferenceId = ct.TagId
	corridorKey.FromNodeId = ct.NodeId
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
		err := GenerateChannelGroups(db)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate channel groups.")
		}
	}()
	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully deleted channel group(s)."})
}

func removeCategoryHandler(c *gin.Context, db *sqlx.DB) {
	categoryId, err := strconv.Atoi(c.Param("categoryId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse categoryId in the request.")
		return
	}
	count, err := removeCategory(db, categoryId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Removing category for categoryId: %v", categoryId))
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully deleted %v category(s).", count)})
}

func removeTagHandler(c *gin.Context, db *sqlx.DB) {
	tagId, err := strconv.Atoi(c.Param("tagId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse tagId in the request.")
		return
	}
	count, err := removeTag(db, tagId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Removing tag for tagId: %v", tagId))
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully deleted %v tag(s).", count)})
}
