package views

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"

	"github.com/lncapital/torq/internal/database"
)

// TODO: delete when tables are switched to v2
type TableViewJson struct {
	Id   int             `json:"id" db:"id"`
	View TableViewDetail `json:"view" db:"view"`
}

// TODO: delete when tables are switched to v2
type TableViewDetail struct {
	Title   string          `json:"title"`
	Saved   bool            `json:"saved"`
	Columns []ViewColumn    `json:"columns"`
	Page    string          `json:"page"`
	SortBy  []ViewOrder     `json:"sortBy"`
	Id      int             `json:"id"`
	Filters *types.JSONText `json:"filters,omitempty"`
}

// TODO: delete when tables are switched to v2
type ViewColumn struct {
	Key       string  `json:"key"`
	Key2      *string `json:"key2,omitempty"`
	Heading   string  `json:"heading"`
	Type      string  `json:"type"`
	ValueType string  `json:"valueType"`
	Suffix    *string `json:"suffix,omitempty"`
	Locked    *bool   `json:"locked,omitempty"`
}

// TODO: delete when tables are switched to v2
type ViewOrder struct {
	Key       string `json:"key"`
	Direction string `json:"direction"`
}

// TODO: delete when tables are switched to v2
type TableViewJsonLegacy struct {
	Id   int                   `json:"id" db:"id"`
	View TableViewDetailLegacy `json:"view" db:"view"`
}

// TODO: delete when tables are switched to v2
type TableViewDetailLegacy struct {
	Title   string              `json:"title"`
	Saved   bool                `json:"saved"`
	Columns []ViewColumn        `json:"columns"`
	Page    string              `json:"page"`
	SortBy  []ViewOrder         `json:"sortBy"`
	Id      int                 `json:"id"`
	Filters FilterClausesLegacy `json:"filters"`
}

// TODO: delete when tables are switched to v2
type FilterClausesLegacy struct {
	And    []FilterClausesLegacy `json:"$and,omitempty"`
	Or     []FilterClausesLegacy `json:"$or,omitempty"`
	Filter *FilterLegacy         `json:"$filter,omitempty"`
}

// TODO: delete when tables are switched to v2
type FilterLegacy struct {
	FuncName  string      `json:"funcName,omitempty"`
	Key       string      `json:"key,omitempty"`
	Parameter interface{} `json:"parameter,omitempty"`
	Category  string      `json:"category,omitempty"`
}

// TODO: delete when tables are switched to v2
func convertLegacyView(legacyTableViews []TableViewLayout) ([]TableViewLayout, error) {
	var tableViewLayoutV2s []TableViewLayout
	for _, legacyTableView := range legacyTableViews {
		if legacyTableView.Version == "v1" {
			var tableViewDetail TableViewDetailLegacy
			err := json.Unmarshal(legacyTableView.View, &tableViewDetail)
			if err != nil {
				return nil, errors.Wrap(err, "JSON unmarshal")
			}
			tableViewJson := TableViewJsonLegacy{
				legacyTableView.Id,
				tableViewDetail,
			}

			for index, column := range tableViewJson.View.Columns {
				tableViewJson.View.Columns[index].Key = strcase.ToLowerCamel(column.Key)
				if tableViewJson.View.Columns[index].Key2 != nil {
					newKey2 := strcase.ToLowerCamel(*tableViewJson.View.Columns[index].Key2)
					tableViewJson.View.Columns[index].Key2 = &newKey2
				}
			}

			for index, sorting := range tableViewJson.View.SortBy {
				tableViewJson.View.SortBy[index].Key = strcase.ToLowerCamel(sorting.Key)
			}

			if tableViewJson.View.Filters.And != nil {
				for index, andFilter := range tableViewJson.View.Filters.And {
					tableViewJson.View.Filters.And[index].Filter.Key = strcase.ToLowerCamel(andFilter.Filter.Key)
				}
			}

			if tableViewJson.View.Filters.Or != nil {
				for index, orFilter := range tableViewJson.View.Filters.Or {
					tableViewJson.View.Filters.Or[index].Filter.Key = strcase.ToLowerCamel(orFilter.Filter.Key)
				}
			}

			viewJson, err := json.Marshal(tableViewJson.View)
			if err != nil {
				return nil, errors.Wrap(err, "JSON marshal table view")
			}
			legacyTableView.View = viewJson
			legacyTableView.Version = "v2"
		}
		tableViewLayoutV2s = append(tableViewLayoutV2s, legacyTableView)
	}
	var resultingViewLayout []TableViewLayout
	for _, tableViewLayoutV2 := range tableViewLayoutV2s {
		if tableViewLayoutV2.Version == "v2" {
			var tableViewDetail TableViewDetailLegacy
			err := json.Unmarshal(tableViewLayoutV2.View, &tableViewDetail)
			if err != nil {
				return nil, errors.Wrap(err, "JSON unmarshal")
			}
			tableViewJson := TableViewJsonLegacy{
				tableViewLayoutV2.Id,
				tableViewDetail,
			}

			for index, column := range tableViewJson.View.Columns {
				tableViewJson.View.Columns[index].Key = convertKey(column.Key)
				if column.Key2 != nil {
					newKey2 := convertKey(*column.Key2)
					tableViewJson.View.Columns[index].Key2 = &newKey2
				}
			}

			for index, sorting := range tableViewJson.View.SortBy {
				tableViewJson.View.SortBy[index].Key = convertKey(sorting.Key)
			}

			if tableViewJson.View.Filters.And != nil {
				for index, andFilter := range tableViewJson.View.Filters.And {
					tableViewJson.View.Filters.And[index].Filter.Key = convertKey(andFilter.Filter.Key)
				}
			}

			if tableViewJson.View.Filters.Or != nil {
				for index, orFilter := range tableViewJson.View.Filters.Or {
					tableViewJson.View.Filters.Or[index].Filter.Key = convertKey(orFilter.Filter.Key)
				}
			}

			viewJson, err := json.Marshal(tableViewJson.View)
			if err != nil {
				return nil, errors.Wrap(err, "JSON marshal table view")
			}
			tableViewLayoutV2.View = viewJson
			tableViewLayoutV2.Version = "v3"
		}
		resultingViewLayout = append(resultingViewLayout, tableViewLayoutV2)
	}
	return resultingViewLayout, nil
}

func convertKey(key string) string {
	switch key {
	case "feeBaseMsat":
		return "feeBase"
	case "remoteFeeBaseMsat":
		return "remoteFeeBase"
	case "minHtlcMsat":
		return "minHtlc"
	case "remoteMinHtlcMsat":
		return "remoteMinHtlc"
	case "maxHtlcMsat":
		return "maxHtlc"
	case "remoteMaxHtlcMsat":
		return "remoteMaxHtlc"
	}
	return key
}

// TODO: delete when tables are switched to v3
func convertLegacyTableViews(tx *sqlx.Tx, tableViewLayouts []TableViewLayout) error {
	for index := range tableViewLayouts {
		_, err := convertLegacyTableView(tx, tableViewLayouts[index])
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: delete when tables are switched to v3
func convertLegacyTableView(tx *sqlx.Tx, tableViewLayout TableViewLayout) (TableViewLayout, error) {
	tableView := TableView{
		Page:  tableViewLayout.Page,
		Order: tableViewLayout.ViewOrder,
	}
	var tableViewDetail TableViewDetailLegacy
	err := json.Unmarshal(tableViewLayout.View, &tableViewDetail)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "JSON unmarshal")
	}
	tableView.Title = tableViewDetail.Title
	if tableView.Page == "" {
		tableView.Page = tableViewDetail.Page
	}

	tableView, err = addTableView(tx, tableView)
	if err != nil {
		return TableViewLayout{}, err
	}
	err = addTableViewDependencies(tx, tableViewDetail, tableView)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "JSON marshal table view")
	}
	_, err = tx.Exec("DELETE FROM table_view_legacy WHERE id = $1;", tableViewLayout.Id)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return TableViewLayout{
		Id:        tableView.TableViewId,
		View:      tableViewLayout.View,
		Page:      tableView.Page,
		ViewOrder: tableView.Order,
		Version:   "v3",
	}, nil
}

func updateLegacyTableView(tx *sqlx.Tx, tableViewId int, tableViewLayout TableViewLayout) (TableViewLayout, error) {
	err := removeTableViewColumns(tx, tableViewId)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "Removing tableView columns.")
	}
	err = removeTableViewFilters(tx, tableViewId)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "Removing tableView filters.")
	}
	err = removeTableViewSortings(tx, tableViewId)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "Removing tableView sortings.")
	}

	tableView := TableView{
		TableViewId: tableViewId,
		Page:        tableViewLayout.Page,
		Order:       tableViewLayout.ViewOrder,
	}
	var tableViewDetail TableViewDetailLegacy
	err = json.Unmarshal(tableViewLayout.View, &tableViewDetail)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "JSON unmarshal")
	}
	tableView.Title = tableViewDetail.Title
	if tableView.Page == "" {
		tableView.Page = tableViewDetail.Page
	}

	err = updateTableView(tx, tableView.TableViewId, tableView.Title)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "Updating tableView.")
	}
	err = addTableViewDependencies(tx, tableViewDetail, tableView)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "Updating tableView Dependencies.")
	}
	return TableViewLayout{
		Id:        tableView.TableViewId,
		View:      tableViewLayout.View,
		Page:      tableView.Page,
		ViewOrder: tableView.Order,
		Version:   "v3",
	}, nil
}

func addTableViewDependencies(tx *sqlx.Tx, tableViewDetail TableViewDetailLegacy, tableView TableView) error {
	for columnIndex, column := range tableViewDetail.Columns {
		_, err := addTableViewColumn(tx, TableViewColumn{
			Key:         column.Key,
			KeySecond:   column.Key2,
			Order:       columnIndex + 1,
			Type:        column.Type,
			TableViewId: tableView.TableViewId,
		})
		if err != nil {
			return errors.Wrap(err, "Add tableView column.")
		}
	}
	filtersJson, err := json.Marshal(tableViewDetail.Filters)
	if err != nil {
		return errors.Wrap(err, "JSON marshal table view")
	}
	_, err = addTableViewFilter(tx, TableViewFilter{
		Filter:      filtersJson,
		TableViewId: tableView.TableViewId,
	})
	if err != nil {
		return errors.Wrap(err, "Add tableView filter.")
	}
	for sortIndex, sorting := range tableViewDetail.SortBy {
		_, err = addTableViewSorting(tx, TableViewSorting{
			Key:         sorting.Key,
			Order:       sortIndex + 1,
			Ascending:   sorting.Direction == "asc",
			TableViewId: tableView.TableViewId,
		})
		if err != nil {
			return errors.Wrap(err, "Add tableView sorting.")
		}
	}
	return nil
}
