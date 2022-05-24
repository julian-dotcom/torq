package settings

import (
	// "io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	// "gopkg.in/guregu/null.v4"
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
	LocalNodeId       int                   `json:"localNodeId" form:"localNodeId" db:"local_node_id"`
	Implementation    string                `json:"implementation" form:"implementation" db:"implementation"`
	GRPCAddress       *string               `json:"grpcAddress" form:"grpcAddress" db:"grpc_address"`
	TLSFileName       *string               `json:"tlsFileName" form:"tlsFileName" db:"tls_file_name"`
	TLSData           *multipart.FileHeader `json:"tlsData" form:"tlsData" db:"tls_data"`
	MacaroonFileName  *string               `json:"macaroonFileName" form:"macaroonFileName" db:"macaroon_file_name"`
	MacaroonData      *multipart.FileHeader `json:"macaroonData" form:"macaroonData" db:"macaroon_data"`
	CreateOn          time.Time             `json:"createdOn" form:"createdOn" db:"created_on"`
	UpdatedOn         *time.Time            `json:"updatedOn" form:"updatedOn" db:"updated_on"`
	TLSDataBytes      []byte
	MacaroonDataBytes []byte
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
	// localNode.Implementation = c.PostForm("implementation")
	// localNode.GRPCAddress.String = c.PostForm("GRPCAddress")
	// localNode.TLSFileName = null.StringFrom(c.PostForm("TLSFileName"))
	// tlsDataFileHeader, err := c.FormFile("tlsData")
	// tlsDataFile, err := tlsDataFileHeader.Open()
	// tlsData, err := io.ReadAll(tlsDataFile)
	// if err != nil {
	// 	return
	// }
	// localNode.TLSDataBytes = tlsData
	// localNode.TLSData = c.Request.ParseMultipartForm(c.Request.FormValue("TLSData"))

	if err := c.Bind(&localNode); err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	log.Printf("%v", localNode)
	err := updateLocalNode(db, localNode)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, localNode)
}
