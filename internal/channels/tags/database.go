package tags

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

func getTags(db *sqlx.DB, channelDBID int) (tags []tag, err error) {
	err = db.Select(&tags, "SELECT * FROM channel_tag WHERE channel_db_id = $1;", channelDBID)
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return make([]tag, 0), nil
		}
		return nil, errors.Wrap(err, "Unable to execute SQL query")
	}
	return tags, nil
}

func insertTag(db *sqlx.DB, tag tag) (tagID int, err error) {
	err = db.QueryRowx(`
INSERT INTO channel_tag (
  channel_db_id,
  tag,
  created_on
) values ($1, $2, $3)
RETURNING tag_id;`, tag.ChannelDBID, tag.Tag, time.Now().UTC()).Scan(&tagID)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to execute SQL statement")
	}

	return tagID, nil
}

func deleteTag(db *sqlx.DB, channelDBID int, tagID int) error {
	_, err := db.Exec("DELETE FROM channel_tag WHERE channel_db_id = $1 AND tag_id = $2;", channelDBID, tagID)
	if err != nil {
		return errors.Wrap(err, "Unable to execute SQL statement")
	}
	return nil
}
