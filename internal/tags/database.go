package tags

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
)

func getTagsForChannel(db *sqlx.DB, channelId int) ([]Tag, error) {
	var tags []Tag
	err := db.Select(&tags, `
		SELECT t.*
		FROM tag t
		JOIN channel_tag ct ON t.tag_id = ct.tag_id
        WHERE ct.channel_id = $1;`, channelId)
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
	err := db.QueryRowx(`INSERT INTO tag (name, style, created_on, updated_on)
		VALUES ($1, $2, $3, $4) RETURNING tag_id;`,
		tag.Name, tag.Style, tag.CreatedOn, tag.UpdateOn).Scan(&tag.TagId)
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

func setTag(db *sqlx.DB, tag Tag) (Tag, error) {
	tag.UpdateOn = time.Now().UTC()
	_, err := db.Exec(`UPDATE tag SET name=$1, style=$2, updated_on=$3 WHERE tag_id=$4;`,
		tag.Name, tag.Style, tag.UpdateOn, tag.TagId)
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

func removeTag(db *sqlx.DB, tagId int) (int64, error) {
	referencingCorridors, err := corridors.GetCorridorsReferencingTag(db, tagId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	if len(referencingCorridors) > 0 {
		return 0, errors.New(fmt.Sprintf("Could not remove tag since it's in use. %v", referencingCorridors))
	}
	res, err := db.Exec(`DELETE FROM tag WHERE tag_id = $1;`, tagId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}
