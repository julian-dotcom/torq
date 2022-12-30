package tags

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/database"
)

func GetTagsByCategoryId(db *sqlx.DB, categoryId int) ([]Tag, error) {
	var tags []Tag
	err := db.Select(&tags, `SELECT t.* FROM tag t WHERE t.category_id = $1;`, categoryId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Tag{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return tags, nil
}

func GetTag(db *sqlx.DB, tagId int) (Tag, error) {
	var t Tag
	err := db.Get(&t, `SELECT * FROM tag WHERE tag_id=$1;`, tagId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tag{}, nil
		}
		return Tag{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return t, nil
}

func GetTags(db *sqlx.DB) ([]Tag, error) {
	var tags []Tag
	err := db.Select(&tags, `SELECT * FROM tag;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Tag{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return tags, nil
}

func addTag(db *sqlx.DB, tag Tag) (Tag, error) {
	tag.CreatedOn = time.Now().UTC()
	tag.UpdateOn = tag.CreatedOn
	err := db.QueryRowx(`INSERT INTO tag (name, style, created_on, updated_on, category_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING tag_id;`,
		tag.Name, tag.Style, tag.CreatedOn, tag.UpdateOn, tag.CategoryId).Scan(&tag.TagId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return Tag{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return Tag{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return tag, nil
}

// setTag you cannot update the category! If you want to reassign a tag to a new category you need to recreate the tag.
func setTag(db *sqlx.DB, tag Tag) (Tag, error) {
	tag.UpdateOn = time.Now().UTC()
	_, err := db.Exec(`UPDATE tag SET name=$1, style=$2, updated_on=$3, category_id=$4 WHERE tag_id=$5;`,
		tag.Name, tag.Style, tag.UpdateOn, tag.CategoryId, tag.TagId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return Tag{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return Tag{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return tag, nil
}
