package settings

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"gopkg.in/guregu/null.v4"
	// "strconv"
)

type settings struct {
	DefaultDateRange  string `json:"defaultDateRange" db:"default_date_range"`
	PreferredTimezone string `json:"preferredTimezone" db:"preferred_timezone"`
	WeekStartsOn      string `json:"weekStartsOn" db:"week_starts_on"`
}

type timeZone struct {
	Name string `json:"name" db:"name"`
}

func RegisterSettingRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getSettingsHandler(c, db) })
	r.PUT("", func(c *gin.Context) { updateSettingsHandler(c, db) })
	r.GET("local-node", func(c *gin.Context) { getLocalNodeHandler(c, db) })
	r.PUT("local-node", func(c *gin.Context) { updateLocalNodeHandler(c, db) })
}
func RegisterUnauthorisedRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("timezones", func(c *gin.Context) { getTimeZonesHandler(c, db) })
}

func getTimeZonesHandler(c *gin.Context, db *sqlx.DB) {
	timeZones, err := getTimeZones(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, timeZones)
}
func getSettingsHandler(c *gin.Context, db *sqlx.DB) {
	settings, err := getSettings(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func updateSettingsHandler(c *gin.Context, db *sqlx.DB) {
	var settings settings
	if err := c.BindJSON(&settings); err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	err := updateSettings(db, settings)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

type localNode struct {
	LocalNodeId      int         `json:"localNodeId" db:"local_node_id"`
	Implementation   string      `json:"implementation" db:"implementation"`
	GRPCAddress      null.String `json:"GRPCAddress" db:"grpc_address"`
	TLSFileName      null.String `json:"TLSFileName" db:"tls_file_name"`
	TLSDATA          []byte      `json:"TLSData" db:"tls_data"`
	MacaroonFileName null.String `json:"macaroonFileName" db:"macaroon_file_name"`
	MacaroonData     []byte      `json:"macaroonData" db:"macaroon_data"`
	CreateOn         time.Time   `json:"createdOn" db:"created_on"`
	UpdatedOn        null.Time   `json:"updatedOn" db:"updated_on"`
}

func getLocalNodeHandler(c *gin.Context, db *sqlx.DB) {
	localNode, err := getLocalNode(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, localNode)
}

func updateLocalNodeHandler(c *gin.Context, db *sqlx.DB) {
	var localNode localNode
	if err := c.BindJSON(&localNode); err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	err := updateLocalNode(db, localNode)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, localNode)
}
