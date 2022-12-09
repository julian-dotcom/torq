package views

import (
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"net/http"
	"strconv"

	"github.com/lncapital/torq/pkg/server_errors"
)

type TableView struct {
	Id        int            `json:"id" db:"id"`
	View      types.JSONText `json:"view" db:"view"`
	Page      string         `json:"page" db:"page"`
	Uuid      string         `json:"uuid" db:"uuid"`
	ViewOrder *int32         `json:"viewOrder" db:"view_order"`
	Version   string         `json:"version" db:"version"`
}

type TableViewResponse struct {
	Forwards []TableView `json:"forwards"`
	Channel  []TableView `json:"channel"`
	Payments []TableView `json:"payments"`
	Invoices []TableView `json:"invoices"`
	OnChain  []TableView `json:"onChain"`
}

// TODO: delete when tables are switched to v2
type TableViewJson struct {
	Id   int             `json:"id" db:"id"`
	View TableViewDetail `json:"view" db:"view"`
}

// TODO: delete when tables are switched to v2
type TableViewDetail struct {
	Title   string        `json:"title"`
	Saved   bool          `json:"saved"`
	Columns []ViewColumn  `json:"columns"`
	Page    string        `json:"page"`
	SortBy  []ViewOrder   `json:"sortBy"`
	Id      int           `json:"id"`
	Filters FilterClauses `json:"filters"`
}

// TODO: delete when tables are switched to v2
type ViewColumn struct {
	Key       string `json:"key"`
	Heading   string `json:"heading"`
	Type      string `json:"type"`
	ValueType string `json:"valueType"`
}

// TODO: delete when tables are switched to v2
type ViewOrder struct {
	Value     string `json:"value"`
	Direction string `json:"direction"`
	Label     string `json:"label"`
}

// TODO: delete when tables are switched to v2
type FilterClauses struct {
	And    []FilterClauses `json:"$and,omitempty"`
	Or     []FilterClauses `json:"$or,omitempty"`
	Filter *Filter         `json:"$filter,omitempty"`
}

// TODO: delete when tables are switched to v2
type Filter struct {
	FuncName  string      `json:"funcName,omitempty"`
	Key       string      `json:"key,omitempty"`
	Parameter interface{} `json:"parameter,omitempty"`
	Category  string      `json:"category,omitempty"`
}

// TODO: delete when tables are switched to v2
func convertView(r []*TableView, db *sqlx.DB, c *gin.Context) ([]*TableView, error) {
	var tableViewDetail TableViewDetail
	for i, view := range r {
		if view.Version == "v1" {
			err := json.Unmarshal(view.View, &tableViewDetail)
			if err != nil {
				return nil, errors.Wrap(err, "JSON unmarshal")
			}
			tableViewJson := TableViewJson{
				view.Id,
				tableViewDetail,
			}

			for j, viewColumn := range tableViewJson.View.Columns {
				tableViewJson.View.Columns[j].Key = strcase.ToLowerCamel(viewColumn.Key)
			}

			for j, viewOrder := range tableViewJson.View.SortBy {
				tableViewJson.View.SortBy[j].Value = strcase.ToLowerCamel(viewOrder.Value)
			}

			if tableViewJson.View.Filters.And != nil {
				for j, viewAndfilter := range tableViewJson.View.Filters.And {
					tableViewJson.View.Filters.And[j].Filter.Key = strcase.ToLowerCamel(viewAndfilter.Filter.Key)
				}
			}

			if tableViewJson.View.Filters.Or != nil {
				for j, viewOrfilter := range tableViewJson.View.Filters.Or {
					tableViewJson.View.Filters.Or[j].Filter.Key = strcase.ToLowerCamel(viewOrfilter.Filter.Key)
				}
			}

			viewJson, err := json.Marshal(tableViewJson.View)
			if err != nil {
				return nil, errors.Wrap(err, "JSON marshal table view")
			}
			r[i].View = viewJson
			r[i].Version = "v2"
			_, err = updateTableView(db, *r[i])
			if err != nil {
				return nil, err
			}
		}
	}
	return r, nil
}

func getTableViewsHandler(c *gin.Context, db *sqlx.DB) {
	r, err := getTableViews(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	// Temporary function
	// We converted the API responses from snake_case to cameCase. The old views needs to be converted as well
	// This function will be deleted when all table_views will be on version "v2"
	views, err := convertView(r, db, c)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	response := TableViewResponse{}
	for _, view := range views {
		switch view.Page {
		case "forwards":
			fmt.Println(view)
			response.Forwards = append(response.Forwards, *view)
		case "channel":
			response.Channel = append(response.Channel, *view)
		case "payments":
			response.Payments = append(response.Payments, *view)
		case "invoices":
			response.Invoices = append(response.Invoices, *view)
		case "onChain":
			response.OnChain = append(response.OnChain, *view)
		}
	}

	c.JSON(http.StatusOK, response)
}

func insertTableViewsHandler(c *gin.Context, db *sqlx.DB) {
	view := NewTableView{}
	if err := c.BindJSON(&view); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	r, err := insertTableView(db, view)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func updateTableViewHandler(c *gin.Context, db *sqlx.DB) {

	view := TableView{}
	if err := c.BindJSON(&view); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	r, err := updateTableView(db, view)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func deleteTableViewsHandler(c *gin.Context, db *sqlx.DB) {

	id, err := strconv.Atoi(c.Param("viewId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse viewId in the request.")
		return
	}

	err = deleteTableView(db, id)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func updateTableViewOrderHandler(c *gin.Context, db *sqlx.DB) {
	var viewOrders []TableViewOrder
	if err := c.BindJSON(&viewOrders); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	err := updateTableViewOrder(db, viewOrders)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
