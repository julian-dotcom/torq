package rebalances

import (
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/core"
)

func AddRebalance(db *sqlx.DB, rebalancer Rebalance) (int, error) {
	err := db.QueryRowx(`
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
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	return rebalancer.RebalanceId, nil
}

func SetRebalance(db *sqlx.DB, originReference string, amountMsat uint64, maximumConcurrency int,
	maximumCostMsat uint64, updateOn time.Time, rebalanceId int) error {
	_, err := db.Exec(`
			UPDATE rebalance
			SET origin_reference=$1, amount_msat=$2, maximum_concurrency=$3, maximum_costmsat=$4, updated_on=$5
			WHERE rebalance_id=$6;`,
		originReference, amountMsat, maximumConcurrency, maximumCostMsat, updateOn, rebalanceId)
	if err != nil {
		return errors.Wrapf(err, "Update rebalancer %v", rebalanceId)
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

func GetLatestResultByOrigin(db *sqlx.DB,
	origin core.RebalanceRequestOrigin, originId int,
	incomingChannelId int, outgoingChannelId int,
	status core.Status, timeoutInMinutes int) (RebalanceResult, error) {
	if incomingChannelId == 0 && outgoingChannelId == 0 {
		return RebalanceResult{}, nil
	}
	var rebalanceResult RebalanceResult
	timeout := time.Duration(-1 * timeoutInMinutes)
	sqlString := `
		SELECT rr.*
		FROM rebalance_log rr
		JOIN rebalance r ON r.rebalance_id=rr.rebalance_id
		WHERE ($1=0 OR rr.incoming_channel_id=$1) AND ($2=0 OR rr.outgoing_channel_id=$2) AND
		      rr.created_on>=$3 AND
		      r.origin=$4 AND r.origin_id=$5 AND
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
