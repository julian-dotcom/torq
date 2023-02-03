package views

import (
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterTableViewRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getTableViewsHandler(c, db) })
	r.POST("", func(c *gin.Context) { addTableViewsHandler(c, db) })
	r.PUT("", func(c *gin.Context) { setTableViewHandler(c, db) }) // TODO: Change to PATCH
	r.PATCH("/order", func(c *gin.Context) { setTableViewOrderHandler(c, db) })
	r.DELETE(":tableViewId", func(c *gin.Context) { removeTableViewsHandler(c, db) })
}

func getTableViewsHandler(c *gin.Context, db *sqlx.DB) {
	legacyTableViews, err := getLegacyTableViews(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	if len(legacyTableViews) != 0 {
		// Temporary function
		// We converted the API responses from snake_case to cameCase. The old views needs to be converted as well
		// This function will be deleted when all table_views will be on version "v2"
		legacyTableViews, err = convertLegacyView(legacyTableViews)
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
	}

	if len(legacyTableViews) != 0 {
		tx, err := db.Beginx()
		if err != nil {
			server_errors.LogAndSendServerError(c, errors.Wrap(err, database.SqlBeginTransactionError))
			return
		}
		err = convertLegacyTableViews(tx, legacyTableViews)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				log.Error().Err(rollbackErr).Msgf("Failed to rollback table view migration.")
			}
			server_errors.LogAndSendServerError(c, err)
			return
		}
		err = tx.Commit()
		if err != nil {
			server_errors.LogAndSendServerError(c, errors.Wrap(err, database.SqlCommitTransactionError))
			return
		}
	}

	tableViews, err := getTableViews(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	response := TableViewResponses{}
	for _, view := range tableViews {
		switch view.Page {
		case "forwards":
			response.Forwards = append(response.Forwards, view)
		case "channel":
			response.Channel = append(response.Channel, view)
		case "payments":
			response.Payments = append(response.Payments, view)
		case "invoices":
			response.Invoices = append(response.Invoices, view)
		case "onChain":
			response.OnChain = append(response.OnChain, view)
		}
	}
	c.JSON(http.StatusOK, response)
}

func addTableViewsHandler(c *gin.Context, db *sqlx.DB) {
	req := NewTableView{}
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	tx, err := db.Beginx()
	if err != nil {
		server_errors.LogAndSendServerError(c, errors.Wrap(err, database.SqlBeginTransactionError))
		return
	}
	tableView, err := addTableViewLayout(tx, req)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Error().Err(rollbackErr).Msgf("Failed to rollback add table view.")
		}
		server_errors.LogAndSendServerError(c, err)
		return
	}
	err = tx.Commit()
	if err != nil {
		server_errors.LogAndSendServerError(c, errors.Wrap(err, database.SqlCommitTransactionError))
	}
	c.JSON(http.StatusOK, tableView)
}

func setTableViewHandler(c *gin.Context, db *sqlx.DB) {
	req := UpdateTableView{}
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	r, err := setTableViewLayout(db, req)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func setTableViewOrderHandler(c *gin.Context, db *sqlx.DB) {
	var viewOrders []TableViewOrder
	if err := c.BindJSON(&viewOrders); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	err := setTableViewLayoutOrder(db, viewOrders)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func removeTableViewsHandler(c *gin.Context, db *sqlx.DB) {
	tableViewId, err := strconv.Atoi(c.Param("tableViewId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse tableViewId in the request.")
		return
	}
	tx, err := db.Beginx()
	if err != nil {
		server_errors.LogAndSendServerError(c, errors.Wrap(err, database.SqlBeginTransactionError))
		return
	}
	err = removeTableView(tx, tableViewId)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Error().Err(rollbackErr).Msgf("Failed to rollback removing table view.")
		}
		server_errors.LogAndSendServerError(c, err)
		return
	}
	err = tx.Commit()
	if err != nil {
		server_errors.LogAndSendServerError(c, errors.Wrap(err, database.SqlCommitTransactionError))
	}
	c.Status(http.StatusOK)
}
