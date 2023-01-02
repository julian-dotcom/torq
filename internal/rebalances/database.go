package rebalances

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/commons"
)

func AddRebalanceAndChannels(db *sqlx.DB, rebalance Rebalance, rebalanceChannels []RebalanceChannel) (int, error) {
	err := db.QueryRowx(`
			INSERT INTO rebalance (incoming_channel_id, outgoing_channel_id, status,
			                       origin, origin_id, origin_reference,
			                       amount_msat, maximum_concurrency, maximum_costmsat,
			                       created_on, updated_on)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING rebalance_id;`,
		rebalance.IncomingChannelId, rebalance.OutgoingChannelId, rebalance.Status,
		rebalance.Origin, rebalance.OriginId, rebalance.OriginReference,
		rebalance.AmountMsat, rebalance.MaximumConcurrency, rebalance.MaximumCostMsat,
		rebalance.CreatedOn, rebalance.UpdateOn).
		Scan(&rebalance.RebalanceId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	for _, rebalanceChannel := range rebalanceChannels {
		rebalanceChannel.RebalanceId = rebalance.RebalanceId
		_, err = db.Exec(`
				INSERT INTO rebalance_channel (channel_id, status, rebalance_id, created_on, updated_on)
				VALUES ($1, $2, $3, $4, $5);`,
			rebalanceChannel.ChannelId, rebalanceChannel.Status, rebalance.RebalanceId, rebalanceChannel.CreatedOn, rebalanceChannel.UpdateOn)
		if err != nil {
			return 0, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return rebalance.RebalanceId, nil
}

func SetRebalanceAndChannels(db *sqlx.DB, rebalance Rebalance, rebalanceChannels []RebalanceChannel) error {
	tx := db.MustBegin()
	_, err := tx.Exec(`
			UPDATE rebalance
			SET origin_reference=$1, amount_msat=$2, maximum_concurrency=$3, maximum_costmsat=$4, updated_on=$5
			WHERE rebalance_id=$6;`,
		rebalance.OriginReference, rebalance.AmountMsat, rebalance.MaximumConcurrency, rebalance.MaximumCostMsat,
		rebalance.UpdateOn, rebalance.RebalanceId)
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
		}
		return errors.Wrap(err, database.SqlExecutionError)
	}
	_, err = tx.Exec(`UPDATE rebalance_channel SET status=$1 WHERE rebalance_id=$2;`, commons.Inactive, rebalance.RebalanceId)
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
		}
		return errors.Wrap(err, database.SqlExecutionError)
	}
	for _, rebalanceChannel := range rebalanceChannels {
		res, err := tx.Exec(`UPDATE rebalance_channel SET status=$1 WHERE rebalance_id=$2 AND channel_id=$3;`,
			rebalanceChannel.Status, rebalance.RebalanceId, rebalanceChannel.ChannelId)
		if err != nil {
			if rb := tx.Rollback(); rb != nil {
				log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
			}
			return errors.Wrap(err, database.SqlExecutionError)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, database.SqlAffectedRowsCheckError)
		}
		if rowsAffected == 0 {
			_, err = db.Exec(`
				INSERT INTO rebalance_channel (channel_id, status, rebalance_id, created_on, updated_on)
				VALUES ($1, $2, $3, $4, $5);`,
				rebalanceChannel.ChannelId, rebalanceChannel.Status, rebalance.RebalanceId, rebalanceChannel.CreatedOn, rebalanceChannel.UpdateOn)
			if err != nil {
				if rb := tx.Rollback(); rb != nil {
					log.Error().Err(rb).Msg(database.SqlRollbackTransactionError)
				}
				return errors.Wrap(err, database.SqlExecutionError)
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, database.SqlCommitTransactionError)
	}
	return nil
}

func AddRebalanceLog(db *sqlx.DB, rebalanceLog RebalanceLog) error {
	_, err := db.Exec(`INSERT INTO rebalance_log (incoming_channel_id, outgoing_channel_id, hops, status,
                           total_time_lock, total_fee_msat, total_amount_msat, error, rebalance_id, created_on, updated_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`,
		rebalanceLog.IncomingChannelId, rebalanceLog.OutgoingChannelId, rebalanceLog.Hops, rebalanceLog.Status,
		rebalanceLog.TotalTimeLock, rebalanceLog.TotalFeeMsat, rebalanceLog.TotalAmountMsat,
		rebalanceLog.Error, rebalanceLog.RebalanceId, rebalanceLog.CreatedOn, rebalanceLog.UpdateOn)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}
