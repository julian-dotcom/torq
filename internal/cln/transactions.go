package cln

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/internal/vector"
	"github.com/lncapital/torq/proto/cln"
)

const streamTransactionsTickerSeconds = 15 * 60

type client_ListTransactions interface {
	ListTransactions(ctx context.Context,
		in *cln.ListtransactionsRequest,
		opts ...grpc.CallOption) (*cln.ListtransactionsResponse, error)
}

func SubscribeAndStoreTransactions(ctx context.Context,
	client client_ListTransactions,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.ClnServiceTransactionsService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamTransactionsTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerChannel := ticker.C

	err := listAndProcessTransactions(ctx, db, client, serviceType, nodeSettings, true)
	if err != nil {
		processError(ctx, serviceType, nodeSettings, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		case <-tickerChannel:
			err = listAndProcessTransactions(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func listAndProcessTransactions(ctx context.Context, db *sqlx.DB, client client_ListTransactions,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	clnTransactions, err := client.ListTransactions(ctx, &cln.ListtransactionsRequest{})
	if err != nil {
		return errors.Wrapf(err, "listing transactions for nodeId: %v", nodeSettings.NodeId)
	}

	err = storeTransactions(db, clnTransactions.Transactions, nodeSettings)
	if err != nil {
		return errors.Wrapf(err, "storing transactions for nodeId: %v", nodeSettings.NodeId)
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of transactions is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func storeTransactions(db *sqlx.DB,
	clnTransactions []*cln.ListtransactionsTransactions,
	nodeSettings cache.NodeSettingsCache) error {

	var blockHeight int
	err := db.Get(&blockHeight, `SELECT COALESCE(MAX(block_height), 0) FROM tx WHERE node_id=$1;`, nodeSettings.NodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			blockHeight = 0
		}
		return errors.Wrapf(err, "obtaining maximum block height for transactions for nodeId: %v",
			nodeSettings.NodeId)
	}

	for _, clnTransaction := range clnTransactions {
		if clnTransaction == nil {
			continue
		}
		txBlockHeight := int(clnTransaction.Blockheight)
		if txBlockHeight <= blockHeight {
			continue
		}
		err = storeTransaction(db, clnTransaction, nodeSettings)
		if err != nil {
			return errors.Wrapf(err, "persisting transaction for nodeId: %v", nodeSettings.NodeId)
		}
	}
	return nil
}

func storeTransaction(db *sqlx.DB,
	clnTransaction *cln.ListtransactionsTransactions,
	nodeSettings cache.NodeSettingsCache) error {
	// Here we're only storing the output addresses, not the output index, amount or if these
	// transactions are ours or not. We might want to add this.
	var totalOutputsMsat uint64
	var destinationAddresses []string
	for _, output := range clnTransaction.Outputs {
		destinationAddresses = append(destinationAddresses, hex.EncodeToString(output.ScriptPubKey))
		if output.AmountMsat != nil {
			totalOutputsMsat += (*output.AmountMsat).Msat
		}
	}

	flags := 0
	transactionTime := time.Now().UTC()
	vectorData := vector.GetTransactionDetailsFromVector(hex.EncodeToString(clnTransaction.Hash), nodeSettings)
	if vectorData.BlockHeight != 0 {
		transactionTime = vectorData.BlockTimestamp
		flags = int(core.TransactionTime)
	}
	_, err := db.Exec(`INSERT INTO tx
    					(timestamp, tx_hash, amount, block_height, dest_addresses, raw_tx_hex, node_id, flags)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
					ON CONFLICT (timestamp, tx_hash) DO NOTHING;`,
		transactionTime,
		hex.EncodeToString(clnTransaction.Hash),
		totalOutputsMsat/1000,
		&clnTransaction.Blockheight,
		pq.Array(destinationAddresses),
		hex.EncodeToString(clnTransaction.Rawtx),
		nodeSettings.NodeId,
		flags,
	)
	if err != nil {
		return errors.Wrap(err, "Executing SQL")
	}
	return nil
}
