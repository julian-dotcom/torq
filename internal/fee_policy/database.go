package fee_policy

import (
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/internal/database"
)

func getFeePolices(db *sqlx.DB) (feePolicies []FeePolicy, err error) {
	err = db.Select(&feePolicies, "SELECT * FROM fee_policy;")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []FeePolicy{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}

	for ti, feePolicy := range feePolicies {
		var targets []FeePolicyTarget
		err = db.Select(&targets, "SELECT * FROM fee_policy_target where fee_policy_id = $1;", feePolicy.FeePolicyId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
		feePolicies[ti].Targets = targets
	}

	return feePolicies, nil

}

func addFeePolicy(db *sqlx.DB, fp FeePolicy) (FeePolicy, error) {
	err := db.QueryRowx(`
INSERT INTO fee_policy (fee_policy_type, name, include_pending_htlcs, aggregate_on_peer, max_ratio,
  min_ratio, max_balance, min_balance, active, interval, created_on, updated_on)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING fee_policy_id;`,
		fp.FeePolicyType, fp.Name, fp.IncludePendingHTLCs, fp.AggregateOnPeer, fp.MaxRatio,
		fp.MinRatio, fp.MaxBalance, fp.MinBalance, fp.Active, fp.Interval,
		time.Now(), time.Now()).
		Scan(&fp.FeePolicyId)
	if err != nil {
		return FeePolicy{}, errors.Wrap(err, database.SqlExecutionError)
	}

	for _, target := range fp.Targets {
		_, err := db.Exec(`
INSERT INTO fee_policy_target (fee_policy_id, tag_id, category_id, node_id, channel_id, created_on, updated_on)
VALUES ($1, $2, $3, $4, $5, $6, $7)`, fp.FeePolicyId, target.TagId, target.CategoryId, target.NodeId, target.ChannelId,
			time.Now(), time.Now())
		if err != nil {
			return FeePolicy{}, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return fp, nil
}
