package channel_history

import (
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/lncapital/torq/pkg/commons"
)

func getTotalOnChainCost(db *sqlx.DB, nodeIds []int, from time.Time, to time.Time) (*uint64, error) {
	var Cost uint64

	q := `
		select coalesce(sum(total_fees), 0) as cost
		from tx
		where timestamp::timestamp AT TIME ZONE ($4) >= $1::timestamp
			and timestamp::timestamp AT TIME ZONE ($4) <= $2::timestamp
			AND node_id = ANY ($3)`

	row := db.QueryRowx(q, from, to, pq.Array(nodeIds), commons.GetSettings().PreferredTimeZone)
	err := row.Scan(&Cost)

	if err != nil {
		return nil, errors.Wrap(err, "SQL row scan for cost")
	}

	return &Cost, nil
}

func getChannelOnChainCost(db *sqlx.DB, lndShortChannelIdStrings []string) (cost *uint64, err error) {

	q := `select coalesce(sum(total_fees), 0) as on_chain_cost
		from tx
		where split_part(label, '-', 2) = ANY ($1)`

	row := db.QueryRowx(q, pq.Array(lndShortChannelIdStrings))
	err = row.Scan(&cost)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrap(err, "SQL row scan for cost")
	}

	return cost, nil
}
