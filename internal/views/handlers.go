package views

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"

	"github.com/lncapital/torq/pkg/server_errors"
)

type TableView struct {
	Id      int            `json:"id" db:"id"`
	View    types.JSONText `json:"view" db:"view"`
	Version string         `json:"version" db:"version"`
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
				return nil, err
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
				return nil, err
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
	page := c.Query("page")
	r, err := getTableViews(db, page)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	// Temporary function
	// We converted the API responses from snake_case to cameCase. The old views needs to be converted as well
	// This function will be deleted when all table_views will be on version "v2"
	r, err = convertView(r, db, c)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func getTableViews(db *sqlx.DB, page string) (r []*TableView, err error) {
	sql := `SELECT id, view, version FROM table_view WHERE page = $1 ORDER BY view_order;`

	rows, err := db.Query(sql, page)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		v := &TableView{}

		err = rows.Scan(&v.Id, &v.View, &v.Version)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, v)
	}
	return r, nil
}

type NewTableView struct {
	View types.JSONText `json:"view" db:"view"`
	Page string         `json:"page" db:"page"`
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

func insertTableView(db *sqlx.DB, view NewTableView) (r TableView, err error) {

	sql := `
		INSERT INTO table_view (view, page, created_on) values ($1, $2, $3)
		RETURNING id, view;
	`
	err = db.QueryRowx(sql, &view.View, &view.Page, time.Now().UTC()).Scan(&r.Id, &r.View)
	if err != nil {
		return TableView{}, errors.Wrap(err, "Unable to create view. SQL statement error")
	}

	return r, nil
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

func updateTableView(db *sqlx.DB, view TableView) (TableView, error) {
	sql := `UPDATE table_view SET view = $1, updated_on = $2, version =$3 WHERE id = $4;`

	_, err := db.Exec(sql, &view.View, time.Now().UTC(), "v2", &view.Id)
	if err != nil {
		return TableView{}, errors.Wrap(err, "Unable to create view. SQL statement error")
	}

	return view, nil
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

func deleteTableView(db *sqlx.DB, id int) error {

	sql := `DELETE FROM table_view WHERE id = $1;`

	_, err := db.Exec(sql, id)
	if err != nil {
		return errors.Wrap(err, "Unable to create view. SQL statement error")
	}

	return nil
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

type TableViewOrder struct {
	Id        int `json:"id" db:"id"`
	ViewOrder int `json:"view_order" db:"view_order"`
}

func updateTableViewOrder(db *sqlx.DB, viewOrders []TableViewOrder) error {

	// TODO: Switch tp updating using this and add Unique constraint
	//sql := `
	//	update table_view set view_order = temp_table.view_order
	//	from (values
	//		(78,  1),
	//		(79,  3),
	//		(81,  2)
	//	) as temp_table(id, view_order)
	//	where temp_table.id = table_view.id;
	//`

	sql := `
		update table_view set view_order = $1
		where id = $2;
	`
	tx := db.MustBegin()
	for _, order := range viewOrders {
		_, err := tx.Exec(sql, order.ViewOrder, order.Id)
		if err != nil {
			return errors.Wrap(err, "Unable to update view order. SQL statement error")
		}
	}

	err := tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Unable to commit update view order. SQL statement error")
	}

	return nil
}
