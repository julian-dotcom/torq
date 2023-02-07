package categories

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/database"
)

func GetCategory(db *sqlx.DB, categoryId int) (Category, error) {
	var c Category
	err := db.Get(&c, `SELECT * FROM category WHERE category_id=$1;`, categoryId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Category{}, nil
		}
		return Category{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return c, nil
}

func GetCategories(db *sqlx.DB) ([]Category, error) {
	var cs []Category
	err := db.Select(&cs, `SELECT * FROM category;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Category{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return cs, nil
}

func addCategory(db *sqlx.DB, category Category) (Category, error) {
	category.CreatedOn = time.Now().UTC()
	category.UpdateOn = category.CreatedOn
	err := db.QueryRowx(`INSERT INTO category (name, style, created_on, updated_on)
		VALUES ($1, $2, $3, $4) RETURNING category_id;`,
		category.Name, category.Style, category.CreatedOn, category.UpdateOn).Scan(&category.CategoryId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return Category{}, database.SqlUniqueConstraintError
			}
		}
		return Category{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return category, nil
}

func setCategory(db *sqlx.DB, category Category) (Category, error) {
	category.UpdateOn = time.Now().UTC()
	_, err := db.Exec(`UPDATE category SET name=$1, style=$2, updated_on=$3 WHERE category_id=$4;`,
		category.Name, category.Style, category.UpdateOn, category.CategoryId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return Category{}, database.SqlUniqueConstraintError
			}
		}
		return Category{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return category, nil
}
