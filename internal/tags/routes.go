package tags

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterTagRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET(":tagId", func(c *gin.Context) { getTagHandler(c, db) })
	r.DELETE(":tagId", func(c *gin.Context) { deleteTagHandler(c, db) })
	r.GET("", func(c *gin.Context) { getTagsHandler(c, db) })
	r.GET("/category/:categoryId", func(c *gin.Context) { getTagsByCategoryHandler(c, db) })
	r.POST("", func(c *gin.Context) { createTagHandler(c, db) })
	r.PUT("", func(c *gin.Context) { updateTagHandler(c, db) })
	// Ads a tag to an entity (i.e. adds a tag to a channel or a node)
	r.POST("tag", func(c *gin.Context) { tagEntityHandler(c, db) })
	r.POST("untag", func(c *gin.Context) { untagEntityHandler(c, db) })

	// Get all tags for a channel
	r.GET("/channel/:channelId", func(c *gin.Context) { getChannelTagsHandler(c, db) })
	// Get all tags for a channel including tags assigned to the channel node
	r.GET("/channel/:channelId/node/:nodeId", func(c *gin.Context) { getChannelTagsHandler(c, db) })
	// Get all tags for a node
	r.GET("/node/:nodeId", func(c *gin.Context) { getNodeTagsHandler(c, db) })
}

func getTagHandler(c *gin.Context, db *sqlx.DB) {
	tagId, err := strconv.Atoi(c.Param("tagId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse tagId in the request.")
		return
	}
	tag, err := GetTag(db, tagId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting tag for tagId: %v", tagId))
		return
	}
	c.JSON(http.StatusOK, tag)
}

func deleteTagHandler(c *gin.Context, db *sqlx.DB) {
	tagId, err := strconv.Atoi(c.Param("tagId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse tagId in the request.")
		return
	}
	err = deleteTag(db, tagId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Deleting tag for tagId: %v", tagId))
		return
	}

	c.JSON(http.StatusOK, nil)
}

func getTagsHandler(c *gin.Context, db *sqlx.DB) {
	tags, err := GetTags(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting tags.")
		return
	}
	c.JSON(http.StatusOK, tags)
}

func getTagsByCategoryHandler(c *gin.Context, db *sqlx.DB) {
	categoryId, err := strconv.Atoi(c.Param("categoryId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse categoryId in the request.")
		return
	}
	tags, err := GetTagsByCategoryId(db, categoryId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting tags.")
		return
	}
	c.JSON(http.StatusOK, tags)
}

func createTagHandler(c *gin.Context, db *sqlx.DB) {
	var t Tag
	if err := c.BindJSON(&t); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	if t.Name == "" {
		server_errors.SendUnprocessableEntity(c, "Failed to find name in the request.")
		return
	}
	storedTag, err := createTag(db, t)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding tag.")
		return
	}
	c.JSON(http.StatusOK, storedTag)
}

func updateTagHandler(c *gin.Context, db *sqlx.DB) {
	var t Tag
	if err := c.BindJSON(&t); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	storedTag, err := updateTag(db, t)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Setting tag for tagId: %v", t.TagId))
		return
	}

	c.JSON(http.StatusOK, storedTag)
}

func tagEntityHandler(c *gin.Context, db *sqlx.DB) {
	var t TagEntityRequest
	if err := c.BindJSON(&t); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	err := TagEntity(db, t)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Setting tag for tagId: %v", t.TagId))
		return
	}

	c.JSON(http.StatusOK, t)
}

func untagEntityHandler(c *gin.Context, db *sqlx.DB) {
	var t TagEntityRequest
	if err := c.BindJSON(&t); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	err := UntagEntity(db, t)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Setting tag for tagId: %v", t.TagId))
		return
	}

	c.JSON(http.StatusOK, t)
}

func getChannelTagsHandler(c *gin.Context, db *sqlx.DB) {
	req := ChannelTagsRequest{}
	channelId, err := strconv.Atoi(c.Param("channelId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse channelId in the request.")
		return
	}
	req.ChannelId = channelId

	if c.Param("nodeId") != "" {
		nodeId, err := strconv.Atoi(c.Param("nodeId"))
		if err != nil {
			server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
			return
		}
		req.NodeId = &nodeId
	}

	var tagIds []int
	if req.NodeId != nil {
		tagIds = cache.GetTagIdsByNodeId(*req.NodeId)
	}
	tagIds = append(tagIds, cache.GetTagIdsByChannelId(req.ChannelId)...)
	if len(tagIds) == 0 {
		c.JSON(http.StatusOK, []Tag{})
		return
	}
	tags := GetTagsByTagIds(tagIds)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting tags.")
		return
	}
	c.JSON(http.StatusOK, tags)
}

func getNodeTagsHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	tagIds := cache.GetTagIdsByNodeId(nodeId)
	if len(tagIds) == 0 {
		c.JSON(http.StatusOK, []Tag{})
		return
	}
	tags := GetTagsByTagIds(tagIds)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting tags.")
		return
	}
	c.JSON(http.StatusOK, tags)
}
