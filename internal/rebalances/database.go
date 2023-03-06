package rebalances

import (
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/commons"
)

func AddRebalanceAndChannels(db *sqlx.DB, rebalancer Rebalance, rebalancerChannelIds []int) (int, error) {
	tx := db.MustBegin()
	err := tx.QueryRowx(`
			INSERT INTO rebalance (incoming_channel_id, outgoing_channel_id, status,
			                       origin, origin_id, origin_reference,
			                       amount_msat, maximum_concurrency, maximum_costmsat,
			                       schedule_target, created_on, updated_on)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING rebalance_id;`,
		rebalancer.IncomingChannelId, rebalancer.OutgoingChannelId, rebalancer.Status,
		rebalancer.Origin, rebalancer.OriginId, rebalancer.OriginReference,
		rebalancer.AmountMsat, rebalancer.MaximumConcurrency, rebalancer.MaximumCostMsat,
		rebalancer.ScheduleTarget, rebalancer.CreatedOn, rebalancer.UpdateOn).
		Scan(&rebalancer.RebalanceId)
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
		}
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	for _, rebalanceChannelId := range rebalancerChannelIds {
		_, err = tx.Exec(`
				INSERT INTO rebalance_channel (channel_id, status, rebalance_id, created_on, updated_on)
				VALUES ($1, $2, $3, $4, $5);`,
			rebalanceChannelId, commons.Active, rebalancer.RebalanceId, rebalancer.CreatedOn, rebalancer.UpdateOn)
		if err != nil {
			if rb := tx.Rollback(); rb != nil {
				log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
			}
			return 0, errors.Wrapf(err, "Storing rebalancer's (%v) channel (%v) ", rebalancer.RebalanceId, rebalanceChannelId)
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlCommitTransactionError)
	}
	return rebalancer.RebalanceId, nil
}

func SetRebalanceAndChannels(db *sqlx.DB, originReference string, amountMsat uint64, maximumConcurrency int,
	maximumCostMsat uint64, updateOn time.Time, rebalanceId int, rebalanceChannelIds []int) error {
	tx := db.MustBegin()
	_, err := tx.Exec(`
			UPDATE rebalance
			SET origin_reference=$1, amount_msat=$2, maximum_concurrency=$3, maximum_costmsat=$4, updated_on=$5
			WHERE rebalance_id=$6;`,
		originReference, amountMsat, maximumConcurrency, maximumCostMsat, updateOn, rebalanceId)
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
		}
		return errors.Wrapf(err, "Update rebalancer %v", rebalanceId)
	}
	_, err = tx.Exec(`UPDATE rebalance_channel SET status=$1 WHERE rebalance_id=$2;`, commons.Inactive, rebalanceId)
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
		}
		return errors.Wrapf(err, "Inactivate rebalancer's channels %v", rebalanceId)
	}
	for _, rebalanceChannelId := range rebalanceChannelIds {
		res, err := tx.Exec(`UPDATE rebalance_channel SET status=$1 WHERE rebalance_id=$2 AND channel_id=$3;`,
			commons.Active, rebalanceId, rebalanceChannelId)
		if err != nil {
			if rb := tx.Rollback(); rb != nil {
				log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
			}
			return errors.Wrapf(err, "Reactivate rebalancer's (%v) channel (%v) ", rebalanceId, rebalanceChannelId)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, database.SqlAffectedRowsCheckError)
		}
		if rowsAffected == 0 {
			_, err = db.Exec(`
				INSERT INTO rebalance_channel (channel_id, status, rebalance_id, created_on, updated_on)
				VALUES ($1, $2, $3, $4, $5);`,
				rebalanceChannelId, commons.Active, rebalanceId, updateOn, updateOn)
			if err != nil {
				if rb := tx.Rollback(); rb != nil {
					log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
				}
				return errors.Wrapf(err, "Activate rebalancer's (%v) channel (%v) ", rebalanceId, rebalanceChannelId)
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, database.SqlCommitTransactionError)
	}
	return nil
}

func AddRebalanceResult(db *sqlx.DB, rebalanceResult RebalanceResult) error {
	_, err := db.Exec(`INSERT INTO rebalance_log (incoming_channel_id, outgoing_channel_id, hops, status,
                           total_time_lock, total_fee_msat, total_amount_msat, error, rebalance_id, created_on, updated_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`,
		rebalanceResult.IncomingChannelId, rebalanceResult.OutgoingChannelId, rebalanceResult.Hops, rebalanceResult.Status,
		rebalanceResult.TotalTimeLock, rebalanceResult.TotalFeeMsat, rebalanceResult.TotalAmountMsat,
		rebalanceResult.Error, rebalanceResult.RebalanceId, rebalanceResult.CreatedOn, rebalanceResult.UpdateOn)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func GetLatestResult(db *sqlx.DB, incomingChannelId int, outgoingChannelId int, timeoutInMinutes int) (RebalanceResult, error) {
	var rebalanceResult RebalanceResult
	timeout := time.Duration(-1 * timeoutInMinutes)
	sqlString := `
		SELECT *
		FROM rebalance_log
		WHERE incoming_channel_id=$1 AND outgoing_channel_id=$2 AND created_on >= $3
		ORDER BY created_on DESC
		LIMIT 1;`
	err := db.Get(&rebalanceResult, sqlString, incomingChannelId, outgoingChannelId, time.Now().Add(timeout*time.Minute))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RebalanceResult{}, nil
		}
		return RebalanceResult{}, errors.Wrapf(err,
			"Getting latest rebalance result with incomingChannelId: %v, outgoingChannelId: %v", incomingChannelId, outgoingChannelId)
	}
	return rebalanceResult, nil
}

func GetLatestResultByOrigin(db *sqlx.DB,
	origin commons.RebalanceRequestOrigin, originId int,
	incomingChannelId int, outgoingChannelId int,
	status commons.Status, timeoutInMinutes int) (RebalanceResult, error) {
	var rebalanceResult RebalanceResult
	timeout := time.Duration(-1 * timeoutInMinutes)
	sqlString := `
		SELECT rr.*
		FROM rebalance_log rr
		JOIN rebalance r ON r.rebalance_id=rr.rebalance_id
		WHERE rr.incoming_channel_id=$1 AND rr.outgoing_channel_id=$2 AND rr.created_on>=$3 AND r.origin=$4 AND r.origin_id=$5 AND
		      rr.status=$6
		ORDER BY rr.created_on DESC
		LIMIT 1;`
	err := db.Get(&rebalanceResult, sqlString, incomingChannelId, outgoingChannelId, time.Now().Add(timeout*time.Minute),
		origin, originId, status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RebalanceResult{}, nil
		}
		return RebalanceResult{}, errors.Wrapf(err,
			"Getting latest rebalance result with incomingChannelId: %v, outgoingChannelId: %v", incomingChannelId, outgoingChannelId)
	}
	return rebalanceResult, nil
}
