package views

import (
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"time"
)

type NewTableView struct {
	View types.JSONText `json:"view" db:"view"`
	Page string         `json:"page" db:"page"`
}

type UpdateTableView struct {
	Id      int            `json:"id" db:"id"`
	View    types.JSONText `json:"view" db:"view"`
	Version string         `json:"version" db:"version"`
}

func getTableViews(db *sqlx.DB) (r []*TableView, err error) {
	sql := `SELECT id, view, page, view_order, version FROM table_view ORDER BY view_order;`

	rows, err := db.Query(sql)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get table views. SQL statement error")
	}

	for rows.Next() {
		v := &TableView{}

		err := rows.Scan(&v.Id, &v.View, &v.Page, &v.ViewOrder, &v.Version)
		if err != nil {
			return r, errors.Wrapf(err, "Unable to scan table view response")
		}

		// Append to the result
		r = append(r, v)
	}
	return r, nil
}

func insertTableView(db *sqlx.DB, view NewTableView) (r TableView, err error) {
	nextViewOrder := 0
	err = db.QueryRowx(`SELECT coalesce(MAX(view_order)+1, 1) FROM table_view where page = $1;`,
		&view.Page).Scan(&nextViewOrder)
	if err != nil {
		return TableView{}, errors.Wrap(err, "Unable to get highest view order. SQL statement error")
	}

	sql := `
		INSERT INTO table_view (view, page, view_order, created_on) values ($1, $2, $3, $4)
			RETURNING id, view, page, view_order, version
	`
	err = db.QueryRowx(sql, &view.View, &view.Page, &nextViewOrder, time.Now().UTC()).
		Scan(&r.Id, &r.View, &r.Page, &r.ViewOrder, &r.Version)
	if err != nil {
		return r, errors.Wrap(err, "Unable to create view. SQL statement error")
	}

	return r, nil
}

func updateTableView(db *sqlx.DB, view UpdateTableView) (r TableView, err error) {
	sql := `UPDATE table_view SET view = $1, updated_on = $2 WHERE id = $3 RETURNING id, view, page, view_order,
version;`

	err = db.QueryRowx(sql, &view.View, time.Now().UTC(), &view.Id).Scan(&r.Id, &r.View, &r.Page, &r.ViewOrder, &r.Version)
	if err != nil {
		return TableView{}, errors.Wrap(err, "Unable to create view. SQL statement error")
	}

	return r, nil
}

func deleteTableView(db *sqlx.DB, id int) error {

	sql := `DELETE FROM table_view WHERE id = $1;`

	_, err := db.Exec(sql, id)
	if err != nil {
		return errors.Wrap(err, "Unable to create view. SQL statement error")
	}

	return nil
}

type TableViewOrder struct {
	Id        int `json:"id" db:"id"`
	ViewOrder int `json:"viewOrder" db:"view_order"`
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
