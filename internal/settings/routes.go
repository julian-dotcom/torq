package settings

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	// "github.com/lncapital/torq/pkg/server_errors"
	"net/http"
	// "strconv"
)

type settings struct {
	DefaultDateRange  string `json:"defaultDateRange" db:"default_date_range"`
	PreferredTimezone int    `json:"preferredTimezone" db:"preferred_timezone"`
	WeekStartsOn      string `json:"weekStartsOn" db:"week_starts_on"`
}

func RegisterSettingRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getSettingsHandler(c, db) })
}
func getSettingsHandler(c *gin.Context, db *sqlx.DB) {
	currentSettings := settings{
		DefaultDateRange:  "last28days",
		PreferredTimezone: 3,
		WeekStartsOn:      "saturday",
	}
	c.JSON(http.StatusOK, currentSettings)
}
