package categories

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterCategoryRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("get/:categoryId", func(c *gin.Context) { getCategoryHandler(c, db) })
	r.GET("all", func(c *gin.Context) { getCategoriesHandler(c, db) })
	r.POST("add", func(c *gin.Context) { addCategoryHandler(c, db) })
	r.PUT("set", func(c *gin.Context) { setCategoryHandler(c, db) })
}

func getCategoryHandler(c *gin.Context, db *sqlx.DB) {
	categoryId, err := strconv.Atoi(c.Param("categoryId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse categoryId in the request.")
		return
	}
	category, err := GetCategory(db, categoryId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting category for categoryId: %v", categoryId))
		return
	}
	c.JSON(http.StatusOK, category)
}

func getCategoriesHandler(c *gin.Context, db *sqlx.DB) {
	categories, err := GetCategories(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting categories.")
		return
	}
	c.JSON(http.StatusOK, categories)
}

func addCategoryHandler(c *gin.Context, db *sqlx.DB) {
	var t Category
	if err := c.BindJSON(&t); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	if t.Name == "" {
		server_errors.SendUnprocessableEntity(c, "Failed to find name in the request.")
		return
	}
	storedCategory, err := addCategory(db, t)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding category.")
		return
	}
	c.JSON(http.StatusOK, storedCategory)
}

func setCategoryHandler(c *gin.Context, db *sqlx.DB) {
	var t Category
	if err := c.BindJSON(&t); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	storedCategory, err := setCategory(db, t)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Setting category for categoryId: %v", t.CategoryId))
		return
	}

	c.JSON(http.StatusOK, storedCategory)
}
