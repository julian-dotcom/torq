package tags

import (
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
	"strconv"
)

func RegisterTagRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("get/:tagId", func(c *gin.Context) { getTagHandler(c, db) })
	r.GET("all", func(c *gin.Context) { getTagsHandler(c, db) })
	r.GET("forChannel/:channelId", func(c *gin.Context) { getTagsForChannelHandler(c, db) })
	r.POST("add", func(c *gin.Context) { addTagHandler(c, db) })
	r.PUT("set", func(c *gin.Context) { setTagHandler(c, db) })
	r.DELETE(":tagId", func(c *gin.Context) { removeTagHandler(c, db) })
}

func getTagsForChannelHandler(c *gin.Context, db *sqlx.DB) {
	channelId, err := strconv.Atoi(c.Param("channelId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse channelId in the request.")
		return
	}
	tags, err := getTagsForChannel(db, channelId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting tags for channelId: %v", channelId))
		return
	}
	c.JSON(http.StatusOK, tags)
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

func getTagsHandler(c *gin.Context, db *sqlx.DB) {
	tags, err := GetTags(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting tags.")
		return
	}
	c.JSON(http.StatusOK, tags)
}

func addTagHandler(c *gin.Context, db *sqlx.DB) {
	var t Tag
	if err := c.BindJSON(&t); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	if t.Name == "" {
		server_errors.SendUnprocessableEntity(c, "Failed to find name in the request.")
		return
	}
	storedTag, err := addTag(db, t)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding tag.")
		return
	}
	c.JSON(http.StatusOK, storedTag)
}

func setTagHandler(c *gin.Context, db *sqlx.DB) {
	var t Tag
	if err := c.BindJSON(&t); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	storedTag, err := setTag(db, t)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Setting tag for tagId: %v", t.TagId))
		return
	}

	c.JSON(http.StatusOK, storedTag)
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
