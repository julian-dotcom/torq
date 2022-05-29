package tags

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
	"strconv"
)

func RegisterTagRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getTagsHandler(c, db) })
	r.POST("", func(c *gin.Context) { postTagHandler(c, db) })
	r.DELETE(":tagId", func(c *gin.Context) { deleteTagHandler(c, db) })
}

func getTagsHandler(c *gin.Context, db *sqlx.DB) {
	channelDBID, err := strconv.Atoi(c.Param("channelDbId"))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	tags, err := getTags(db, channelDBID)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, tags)
}

func postTagHandler(c *gin.Context, db *sqlx.DB) {
	channelDBID, err := strconv.Atoi(c.Param("channelDbId"))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	var tag tag
	if err := c.BindJSON(&tag); err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	tag.ChannelDBID = channelDBID
	tagID, err := insertTag(db, tag)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"tagId": tagID})
}

func deleteTagHandler(c *gin.Context, db *sqlx.DB) {
	channelDBID, err := strconv.Atoi(c.Param("channelDbId"))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	tagID, err := strconv.Atoi(c.Param("tagId"))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	err = deleteTag(db, channelDBID, tagID)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully deleted tag"})
}
