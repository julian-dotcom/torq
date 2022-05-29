package settings

import (
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
)

type settings struct {
	DefaultDateRange  string `json:"defaultDateRange" db:"default_date_range"`
	PreferredTimezone string `json:"preferredTimezone" db:"preferred_timezone"`
	WeekStartsOn      string `json:"weekStartsOn" db:"week_starts_on"`
}

type timeZone struct {
	Name string `json:"name" db:"name"`
}

func RegisterSettingRoutes(r *gin.RouterGroup, db *sqlx.DB, restartLNDSub func()) {
	r.GET("", func(c *gin.Context) { getSettingsHandler(c, db) })
	r.PUT("", func(c *gin.Context) { updateSettingsHandler(c, db) })
	r.GET("local-node", func(c *gin.Context) { getLocalNodeHandler(c, db) })
	r.PUT("local-node", func(c *gin.Context) { updateLocalNodeHandler(c, db, restartLNDSub) })
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
	LocalNodeId       int                   `json:"localNodeId" form:"localNodeId" db:"local_node_id"`
	Implementation    string                `json:"implementation" form:"implementation" db:"implementation"`
	GRPCAddress       *string               `json:"grpcAddress" form:"grpcAddress" db:"grpc_address"`
	TLSFileName       *string               `json:"tlsFileName" db:"tls_file_name"`
	TLSFile           *multipart.FileHeader `form:"tlsFile"`
	MacaroonFileName  *string               `json:"macaroonFileName" db:"macaroon_file_name"`
	MacaroonFile      *multipart.FileHeader `form:"macaroonFile"`
	CreateOn          time.Time             `json:"createdOn" db:"created_on"`
	UpdatedOn         *time.Time            `json:"updatedOn"  db:"updated_on"`
	TLSDataBytes      []byte                `db:"tls_data"`
	MacaroonDataBytes []byte                `db:"macaroon_data"`
}

func getLocalNodeHandler(c *gin.Context, db *sqlx.DB) {
	localNode, err := getLocalNode(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, localNode)
}

func updateLocalNodeHandler(c *gin.Context, db *sqlx.DB, restartLNDSub func()) {
	var localNode localNode

	if err := c.Bind(&localNode); err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	err := updateLocalNodeDetails(db, localNode)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	if localNode.TLSFile != nil {
		localNode.TLSFileName = &localNode.TLSFile.Filename
		tlsDataFile, err := localNode.TLSFile.Open()
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
		tlsData, err := io.ReadAll(tlsDataFile)
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
		localNode.TLSDataBytes = tlsData

		err = updateLocalNodeTLS(db, localNode)
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
	}

	if localNode.MacaroonFile != nil {
		localNode.MacaroonFileName = &localNode.MacaroonFile.Filename
		macaroonDataFile, err := localNode.MacaroonFile.Open()
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
		macaroonData, err := io.ReadAll(macaroonDataFile)
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
		localNode.MacaroonDataBytes = macaroonData
		err = updateLocalNodeMacaroon(db, localNode)
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
	}

	restartLNDSub()

	c.JSON(http.StatusOK, localNode)
}
