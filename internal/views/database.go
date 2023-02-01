package views

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/database"
)

func getTableViewsStructured(db *sqlx.DB) ([]TableViewStructured, error) {
	var tableViews []TableView
	err := db.Select(&tableViews, `SELECT * FROM table_view ORDER BY "order";`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []TableViewStructured{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	var tableViewsStructured []TableViewStructured
	for _, tableView := range tableViews {
		tableViewColumns, err := getTableViewColumnsByTableViewId(db, tableView.TableViewId)
		if err != nil {
			return nil, err
		}
		tableViewFilters, err := getTableViewFiltersByTableViewId(db, tableView.TableViewId)
		if err != nil {
			return nil, err
		}
		tableViewSortings, err := getTableViewSortingsByTableViewId(db, tableView.TableViewId)
		if err != nil {
			return nil, err
		}
		tableViewsStructured = append(tableViewsStructured, TableViewStructured{
			TableViewId: tableView.TableViewId,
			Page:        tableView.Page,
			Title:       tableView.Title,
			Order:       tableView.Order,
			UpdateOn:    tableView.UpdateOn,
			Columns:     tableViewColumns,
			Filters:     tableViewFilters,
			Sortings:    tableViewSortings,
		})
	}
	return tableViewsStructured, nil
}

func getTableViewById(db *sqlx.DB, tableViewId int) (TableView, error) {
	var tableView TableView
	err := db.Get(&tableView, `SELECT * FROM table_view WHERE table_view_id=$1;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TableView{}, nil
		}
		return TableView{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return tableView, nil
}

func getTableViewColumnsByTableViewId(db *sqlx.DB, tableViewId int) ([]TableViewColumn, error) {
	var tableViewColumns []TableViewColumn
	err := db.Select(&tableViewColumns, `SELECT * FROM table_view_column WHERE table_view_id=$1 ORDER BY "order";`, tableViewId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []TableViewColumn{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return tableViewColumns, nil
}

func getTableViewFiltersByTableViewId(db *sqlx.DB, tableViewId int) ([]TableViewFilter, error) {
	var tableViewFilters []TableViewFilter
	err := db.Select(&tableViewFilters, `SELECT * FROM table_view_filter WHERE table_view_id=$1;`, tableViewId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []TableViewFilter{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return tableViewFilters, nil
}

func getTableViewSortingsByTableViewId(db *sqlx.DB, tableViewId int) ([]TableViewSorting, error) {
	var tableViewSortings []TableViewSorting
	err := db.Select(&tableViewSortings, `SELECT * FROM table_view_sorting WHERE table_view_id=$1 ORDER BY "order";`, tableViewId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []TableViewSorting{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return tableViewSortings, nil
}

func removeTableView(tx *sqlx.Tx, tableViewId int) error {
	err := removeTableViewColumns(tx, tableViewId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	err = removeTableViewFilters(tx, tableViewId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	err = removeTableViewSortings(tx, tableViewId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	_, err = tx.Exec("DELETE FROM table_view WHERE table_view_id = $1;", tableViewId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func removeTableViewColumns(tx *sqlx.Tx, tableViewId int) error {
	_, err := tx.Exec("DELETE FROM table_view_column WHERE table_view_id = $1;", tableViewId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func removeTableViewFilters(tx *sqlx.Tx, tableViewId int) error {
	_, err := tx.Exec("DELETE FROM table_view_filter WHERE table_view_id = $1;", tableViewId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func removeTableViewSortings(tx *sqlx.Tx, tableViewId int) error {
	_, err := tx.Exec("DELETE FROM table_view_sorting WHERE table_view_id = $1;", tableViewId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func addTableView(tx *sqlx.Tx, tableView TableView) (TableView, error) {
	tableView.CreatedOn = time.Now().UTC()
	tableView.UpdateOn = tableView.CreatedOn
	var err error
	if tableView.Order == 0 {
		err = tx.QueryRowx(`INSERT INTO table_view (page, title, "order", created_on, updated_on)
			SELECT $1, $2, COALESCE(MAX(t."order")+1, 1), $3, $4
			FROM table_view t
			WHERE t.page = $1
			RETURNING table_view_id, "order";`,
			tableView.Page, tableView.Title,
			tableView.CreatedOn, tableView.UpdateOn).Scan(&tableView.TableViewId, &tableView.Order)
	} else {
		err = tx.QueryRowx(`INSERT INTO table_view (page, title, "order", created_on, updated_on)
			VALUES ($1, $2, $3, $4, $5) RETURNING table_view_id;`,
			tableView.Page, tableView.Title, tableView.Order,
			tableView.CreatedOn, tableView.UpdateOn).Scan(&tableView.TableViewId)
	}
	if err != nil {
		return TableView{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return tableView, nil
}

func addTableViewColumn(tx *sqlx.Tx, tableViewColumn TableViewColumn) (TableViewColumn, error) {
	tableViewColumn.CreatedOn = time.Now().UTC()
	tableViewColumn.UpdateOn = tableViewColumn.CreatedOn
	err := tx.QueryRowx(`
		INSERT INTO table_view_column (key, key_second, "order", type, table_view_id, created_on, updated_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING table_view_column_id;`,
		tableViewColumn.Key, tableViewColumn.KeySecond, tableViewColumn.Order, tableViewColumn.Type,
		tableViewColumn.TableViewId, tableViewColumn.CreatedOn, tableViewColumn.UpdateOn).
		Scan(&tableViewColumn.TableViewColumnId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return TableViewColumn{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return TableViewColumn{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return tableViewColumn, nil
}

func addTableViewFilter(tx *sqlx.Tx, tableViewFilter TableViewFilter) (TableViewFilter, error) {
	tableViewFilter.CreatedOn = time.Now().UTC()
	tableViewFilter.UpdateOn = tableViewFilter.CreatedOn
	err := tx.QueryRowx(`
		INSERT INTO table_view_filter (filter, table_view_id, created_on, updated_on)
		VALUES ($1, $2, $3, $4) RETURNING table_view_filter_id;`,
		tableViewFilter.Filter,
		tableViewFilter.TableViewId, tableViewFilter.CreatedOn, tableViewFilter.UpdateOn).Scan(&tableViewFilter.TableViewFilterId)
	if err != nil {
		return TableViewFilter{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return tableViewFilter, nil
}

func addTableViewSorting(tx *sqlx.Tx, tableViewSorting TableViewSorting) (TableViewSorting, error) {
	tableViewSorting.CreatedOn = time.Now().UTC()
	tableViewSorting.UpdateOn = tableViewSorting.CreatedOn
	err := tx.QueryRowx(`
		INSERT INTO table_view_sorting (key, "order", ascending, table_view_id, created_on, updated_on)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING table_view_sorting_id;`,
		tableViewSorting.Key, tableViewSorting.Order, tableViewSorting.Ascending,
		tableViewSorting.TableViewId, tableViewSorting.CreatedOn, tableViewSorting.UpdateOn).Scan(&tableViewSorting.TableViewSortingId)
	if err != nil {
		return TableViewSorting{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return tableViewSorting, nil
}

func addTableViewLayout(tx *sqlx.Tx, newTableView NewTableView) (TableViewLayout, error) {
	tableViewLayout, err := convertLegacyTableView(tx, TableViewLayout{
		View:    newTableView.View,
		Page:    newTableView.Page,
		Version: "v3",
	})
	if err != nil {
		return TableViewLayout{}, err
	}
	return tableViewLayout, nil
}

func setTableViewLayout(db *sqlx.DB, updateTableView UpdateTableView) (TableViewLayout, error) {
	tableView, err := getTableViewById(db, updateTableView.Id)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "Obtaining existing tableView to update")
	}
	tx, err := db.Beginx()
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, database.SqlBeginTransactionError)
	}
	defer func(tx *sqlx.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Error().Err(err).Msgf("Failed to rollback add table view.")
		}
	}(tx)
	err = removeTableView(tx, updateTableView.Id)
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "Removing tableView (update)")
	}
	layout, err := addTableViewLayout(tx, NewTableView{
		View: updateTableView.View,
		Page: tableView.Page,
	})
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, "Adding tableView (update)")
	}
	err = tx.Commit()
	if err != nil {
		return TableViewLayout{}, errors.Wrap(err, database.SqlCommitTransactionError)
	}
	return layout, nil
}

func setTableViewLayoutOrder(db *sqlx.DB, tableViewOrder []TableViewOrder) error {
	now := time.Now().UTC()
	tx := db.MustBegin()
	for _, order := range tableViewOrder {
		_, err := tx.Exec(`UPDATE table_view SET "order"=$1, updated_on=$2 WHERE table_view_id=$3;`,
			order.ViewOrder, now, order.Id)
		if err != nil {
			return errors.Wrap(err, "Updating view order.")
		}
	}
	err := tx.Commit()
	if err != nil {
		return errors.Wrap(err, database.SqlCommitTransactionError)
	}
	return nil
}

func getTableViews(db *sqlx.DB) ([]TableViewLayout, error) {
	tableViewsStructured, err := getTableViewsStructured(db)
	if err != nil {
		return nil, err
	}
	var tableViews []TableViewLayout
	for _, tableView := range tableViewsStructured {
		tableViewJson := TableViewJson{
			Id: tableView.TableViewId,
			View: TableViewDetail{
				Title: tableView.Title,
				Saved: true,
				Page:  tableView.Page,
				Id:    tableView.TableViewId,
			},
		}

		for _, column := range tableView.Columns {
			columnDefinition := getTableViewColumnDefinition(column.Key)
			newColumn := ViewColumn{
				Key:       column.Key,
				Key2:      column.KeySecond,
				Heading:   columnDefinition.heading,
				Type:      column.Type,
				ValueType: columnDefinition.valueType,
			}
			locked := true
			if columnDefinition.locked {
				newColumn.Locked = &locked
			}
			if columnDefinition.suffix != "" {
				newColumn.Suffix = &columnDefinition.suffix
			}
			tableViewJson.View.Columns = append(tableViewJson.View.Columns, newColumn)
		}

		for _, sorting := range tableView.Sortings {
			viewOrder := ViewOrder{
				Key: sorting.Key,
			}
			if sorting.Ascending {
				viewOrder.Direction = "asc"
			} else {
				viewOrder.Direction = "desc"
			}
			tableViewJson.View.SortBy = append(tableViewJson.View.SortBy, viewOrder)
		}

		if len(tableView.Filters) != 0 {
			for filterIndex := range tableView.Filters {
				filter := tableView.Filters[filterIndex].Filter
				if filter.String() != "" && filter.String() != "{}" {
					tableViewJson.View.Filters = &filter
					break
				}
			}
		}

		viewJson, err := json.Marshal(tableViewJson.View)
		if err != nil {
			return nil, errors.Wrap(err, "JSON marshal table view")
		}
		tableViews = append(tableViews, TableViewLayout{
			Id:        tableView.TableViewId,
			View:      viewJson,
			Page:      tableView.Page,
			ViewOrder: tableView.Order,
			Version:   "v3",
		})
	}

	return tableViews, nil
}

// LEGACY
func getLegacyTableViews(db *sqlx.DB) ([]TableViewLayout, error) {
	rows, err := db.Query("SELECT id, view, page, view_order, version FROM table_view_legacy ORDER BY view_order;")
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get table views. SQL statement error")
	}
	var legacyTableViewLayouts []TableViewLayout
	for rows.Next() {
		v := &TableViewLayout{}
		err := rows.Scan(&v.Id, &v.View, &v.Page, &v.ViewOrder, &v.Version)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to scan table view response")
		}
		// Append to the result
		legacyTableViewLayouts = append(legacyTableViewLayouts, *v)
	}
	return legacyTableViewLayouts, nil
}
