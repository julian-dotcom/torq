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

		if feePolicy.FeePolicyStrategy == policyStrategyStep {
			var steps []FeePolicyStep
			err = db.Select(&steps, "SELECT * FROM fee_policy_step where fee_policy_id = $1;", feePolicy.FeePolicyId)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}
				return nil, errors.Wrap(err, database.SqlExecutionError)
			}
			feePolicies[ti].Steps = steps
		}

	}

	return feePolicies, nil

}

func updateFeePolicy(db *sqlx.DB, fp FeePolicy) (FeePolicy, error) {
	_, err := db.Exec(`
UPDATE fee_policy SET fee_policy_strategy = $1, name = $2, include_pending_htlcs = $3, aggregate_on_peer = $4,
 active = $5, interval = $6, updated_on = $7
WHERE fee_policy_id = $8;`,
		fp.FeePolicyStrategy, fp.Name, fp.IncludePendingHTLCs, fp.AggregateOnPeer, fp.Active, fp.Interval,
		time.Now(), fp.FeePolicyId)
	if err != nil {
		return FeePolicy{}, errors.Wrap(err, "Updating fee_policy in the db")
	}

	targetIds := []int{}
	for _, target := range fp.Targets {
		if target.FeePolicyTargetId == 0 {
			err = db.QueryRowx(`
INSERT INTO fee_policy_target (fee_policy_id, tag_id, category_id, node_id, channel_id, created_on, updated_on)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING fee_policy_target_id;`, fp.FeePolicyId, target.TagId, target.CategoryId, target.NodeId, target.ChannelId,
				time.Now(), time.Now()).
				Scan(&target.FeePolicyTargetId)
			if err != nil {
				return FeePolicy{}, errors.Wrap(err, "Inserting new fee_policy_target")
			}
			targetIds = append(targetIds, target.FeePolicyTargetId)
			continue
		}
		_, err := db.Exec(`
UPDATE fee_policy_target SET fee_policy_id = $1, tag_id = $2, category_id = $3, node_id = $4, channel_id = $5, updated_on = $6
WHERE fee_policy_target_id = $7;`, fp.FeePolicyId, target.TagId, target.CategoryId, target.NodeId, target.ChannelId,
			time.Now(), target.FeePolicyTargetId)
		if err != nil {
			return FeePolicy{}, errors.Wrap(err, "Updating fee_policy_target")
		}
		targetIds = append(targetIds, target.FeePolicyTargetId)
	}

	query, args, err := sqlx.In("DELETE FROM fee_policy_target WHERE fee_policy_id = ? and fee_policy_target_id NOT IN (?);", fp.FeePolicyId, targetIds)
	if err != nil {
		return FeePolicy{}, errors.Wrap(err, "Making SQL In query")
	}
	_, err = db.Exec(db.Rebind(query), args...)
	if err != nil {
		return FeePolicy{}, errors.Wrap(err, "Deleting fee_policy_targets")
	}

	stepIds := []int{}
	if fp.FeePolicyStrategy == policyStrategyStep {
		for _, step := range fp.Steps {
			if step.FeePolicyStepId == 0 {
				err := db.QueryRowx(`
INSERT INTO fee_policy_step (fee_policy_id, filter_max_ratio, filter_min_ratio, filter_max_balance, filter_min_balance,
  set_min_htlc, set_max_htlc, set_fee_ppm, set_base_fee, created_on, updated_on)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING fee_policy_step_id;`, fp.FeePolicyId,
					step.FilterMaxRatio, step.FilterMinRatio, step.FilterMaxBalance, step.FilterMinBalance,
					step.SetMinHTLC, step.SetMaxHTLC, step.SetFeePPM, step.SetBaseFee,
					time.Now(), time.Now()).
					Scan(&step.FeePolicyStepId)
				if err != nil {
					return FeePolicy{}, errors.Wrap(err, "Inserting new fee_policy_step")
				}
				stepIds = append(stepIds, step.FeePolicyStepId)
				continue
			}
			_, err := db.Exec(`
UPDATE fee_policy_step SET fee_policy_id = $1, filter_max_ratio = $2, filter_min_ratio = $3, filter_max_balance = $4, filter_min_balance = $5,
  set_min_htlc = $6, set_max_htlc = $7, set_fee_ppm = $8, set_base_fee = $9, updated_on = $10
WHERE fee_policy_step_id = $11;`, fp.FeePolicyId,
				step.FilterMaxRatio, step.FilterMinRatio, step.FilterMaxBalance, step.FilterMinBalance,
				step.SetMinHTLC, step.SetMaxHTLC, step.SetFeePPM, step.SetBaseFee,
				time.Now(), step.FeePolicyStepId)
			if err != nil {
				return FeePolicy{}, errors.Wrap(err, "Updating fee_policy_step")
			}
			stepIds = append(stepIds, step.FeePolicyStepId)
		}

		query, args, err := sqlx.In("DELETE FROM fee_policy_step WHERE fee_policy_id = ? and fee_policy_step_id NOT IN (?);", fp.FeePolicyId, stepIds)
		if err != nil {
			return FeePolicy{}, errors.Wrap(err, "Making SQL In query")
		}
		_, err = db.Exec(db.Rebind(query), args...)
		if err != nil {
			return FeePolicy{}, errors.Wrap(err, "Deleting fee_policy_step")
		}

	}
	return fp, nil
}

func addFeePolicy(db *sqlx.DB, fp FeePolicy) (FeePolicy, error) {
	err := db.QueryRowx(`
INSERT INTO fee_policy (fee_policy_strategy, name, include_pending_htlcs, aggregate_on_peer,
 active, interval, created_on, updated_on)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING fee_policy_id;`,
		fp.FeePolicyStrategy, fp.Name, fp.IncludePendingHTLCs, fp.AggregateOnPeer, fp.Active, fp.Interval,
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

	if fp.FeePolicyStrategy == policyStrategyStep {
		for _, step := range fp.Steps {
			_, err := db.Exec(`
INSERT INTO fee_policy_step (fee_policy_id, filter_max_ratio, filter_min_ratio, filter_max_balance, filter_min_balance,
  set_min_htlc, set_max_htlc, set_fee_ppm, set_base_fee, created_on, updated_on)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`, fp.FeePolicyId,
				step.FilterMaxRatio, step.FilterMinRatio, step.FilterMaxBalance, step.FilterMinBalance,
				step.SetMinHTLC, step.SetMaxHTLC, step.SetFeePPM, step.SetBaseFee,
				time.Now(), time.Now())
			if err != nil {
				return FeePolicy{}, errors.Wrap(err, database.SqlExecutionError)
			}
		}
	}

	return fp, nil
}
