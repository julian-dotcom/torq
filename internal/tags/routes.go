package tags

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterTagRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET(":tagId", func(c *gin.Context) { getTagHandler(c, db) })
	r.GET("", func(c *gin.Context) { getTagsHandler(c, db) })
	r.POST("", func(c *gin.Context) { addTagHandler(c, db) })
	// setTagHandler you cannot reassign a tag to a new category!
	r.PUT("", func(c *gin.Context) { setTagHandler(c, db) })
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
