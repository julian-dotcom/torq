package lnd

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

type Tx struct {
	Timestamp             time.Time `json:"timestamp" db:"timestamp"`
	TransactionHash       *string   `json:"transactionHash" db:"tx_hash"`
	Amount                *int64    `json:"amount" db:"amount"`
	NumberOfConfirmations *int32    `json:"numberOfConfirmations" db:"num_confirmations"`
	BlockHash             *string   `json:"blockHash" db:"block_hash"`
	BlockHeight           *int32    `json:"blockHeight" db:"block_height"`
	TotalFees             *int64    `json:"totalFees" db:"total_fees"`
	DestinationAddresses  *[]string `json:"destinationAddresses" db:"dest_addresses"`
	RawTransactionHex     *string   `json:"rawTransactionHex" db:"raw_tx_hex"`
	Label                 *string   `json:"label" db:"label"`
	NodeId                int       `json:"nodeId" db:"node_id"`
}

func fetchLastTxHeight(db *sqlx.DB) (txHeight int32, err error) {

	sqlLatest := `select coalesce(max(block_height),1) from tx;`

	row := db.QueryRow(sqlLatest)
	err = row.Scan(&txHeight)

	if err != nil {
		return 1, err
	}

	return txHeight, nil
}

// SubscribeAndStoreTransactions Subscribes to on-chain transaction events from LND and stores them in the
// database as a time series. It will also import unregistered transactions on startup.
func SubscribeAndStoreTransactions(ctx context.Context, client lnrpc.LightningClient, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}) {
	var transactionHeight int32
	var err error
	var transactionDetails *lnrpc.TransactionDetails
	var storedTx Tx

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		transactionHeight, err = fetchLastTxHeight(db)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to obtain last know transaction, will retry in 10 seconds")
			time.Sleep(10 * time.Second)
			continue
		}

		transactionDetails, err = client.GetTransactions(ctx, &lnrpc.GetTransactionsRequest{
			StartHeight: transactionHeight,
		})
		if err != nil {
			log.Error().Err(err).Msgf("Failed to obtain last transaction details, will retry in 10 seconds")
			time.Sleep(10 * time.Second)
			continue
		}

		for _, transaction := range transactionDetails.Transactions {
			storedTx, err = storeTransaction(db, transaction, nodeSettings.NodeId)
			if err != nil {
				// TODO FIXME THIS WILL CAUSE AN INFINITE LOOP???
				// It's either an infinite loop or missing a transaction.
				log.Error().Err(err).Msg("Failed to store the transaction, will retry in 10 seconds")
				time.Sleep(10 * time.Second)
				continue
			}
			if eventChannel != nil {
				eventChannel <- broadcast.TransactionEvent{
					EventData: broadcast.EventData{
						EventTime: time.Now().UTC(),
						NodeId:    nodeSettings.NodeId,
					},
					Timestamp:             storedTx.Timestamp,
					TransactionHash:       storedTx.TransactionHash,
					Amount:                storedTx.Amount,
					NumberOfConfirmations: storedTx.NumberOfConfirmations,
					BlockHash:             storedTx.BlockHash,
					BlockHeight:           storedTx.BlockHeight,
					TotalFees:             storedTx.TotalFees,
					DestinationAddresses:  storedTx.DestinationAddresses,
					RawTransactionHex:     storedTx.RawTransactionHex,
					Label:                 storedTx.Label,
					NodeId:                storedTx.NodeId,
				}
			}
		}
	}
}

func storeTransaction(db *sqlx.DB, tx *lnrpc.Transaction, nodeId int) (Tx, error) {
	if tx == nil {
		return Tx{}, nil
	}

	// Here we're only storing the output addresses, not the output index, amount or if these
	// transactions are ours or not. We might want to add this.
	var destinationAddresses []string
	for _, output := range tx.OutputDetails {
		destinationAddresses = append(destinationAddresses, output.Address)
	}

	storedTx := Tx{
		Timestamp:             time.Unix(tx.TimeStamp, 0).UTC(),
		TransactionHash:       &tx.TxHash,
		Amount:                &tx.Amount,
		NumberOfConfirmations: &tx.NumConfirmations,
		BlockHash:             &tx.BlockHash,
		BlockHeight:           &tx.BlockHeight,
		TotalFees:             &tx.TotalFees,
		DestinationAddresses:  &destinationAddresses,
		RawTransactionHex:     &tx.RawTxHex,
		Label:                 &tx.Label,
		NodeId:                nodeId,
	}

	var insertTx = `INSERT INTO tx (timestamp, tx_hash, amount, num_confirmations, block_hash, block_height,
                total_fees, dest_addresses, raw_tx_hex, label, node_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
                ON CONFLICT (timestamp, tx_hash) DO NOTHING;`

	_, err := db.Exec(insertTx,
		time.Unix(tx.TimeStamp, 0).UTC(),
		tx.TxHash,
		tx.Amount,
		tx.NumConfirmations,
		tx.BlockHash,
		tx.BlockHeight,
		tx.TotalFees,
		pq.Array(destinationAddresses),
		tx.RawTxHex,
		tx.Label,
		nodeId,
	)

	if err != nil {
		return Tx{}, errors.Wrapf(err, `inserting transaction`)
	}

	return storedTx, nil
}
