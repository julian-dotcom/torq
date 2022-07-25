package channel_history

import (
	"database/sql"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"time"
)

func getTotalOnChainCost(db *sqlx.DB, from time.Time, to time.Time) (*uint64, error) {
	var Cost uint64

	q := `WITH tz AS (select preferred_timezone as tz from settings)
		select coalesce(sum(total_fees), 0) as cost
		from tx
		where timestamp::timestamp AT TIME ZONE (table tz) >= $1::timestamp
			and timestamp::timestamp AT TIME ZONE (table tz) <= $2::timestamp`

	row := db.QueryRow(q, from, to)
	err := row.Scan(&Cost)

	if err != nil {
		return &Cost, err
	}

	return &Cost, nil
}

func getChannelOnChainCost(db *sqlx.DB, chanIds []string) (cost *uint64, err error) {

	q := `select coalesce(sum(total_fees), 0) as on_chain_cost
		from tx
		where split_part(label, '-', 2) in (?)`

	qs, args, err := sqlx.In(q, chanIds)
	if err != nil {
		return nil, errors.Wrapf(err, "sqlx.In(%s, %v)", q, chanIds)
	}

	qsr := db.Rebind(qs)

	row := db.QueryRow(qsr, args...)
	err = row.Scan(&cost)

	if err == sql.ErrNoRows {
		return cost, nil
	}

	if err != nil {
		return cost, err
	}

	return cost, nil
}
