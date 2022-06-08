package lnd

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"go.uber.org/ratelimit"
	"io"
	"time"
)

func ImportTransactions(client lnrpc.LightningClient, db *sqlx.DB) error {

	ctx := context.Background()

	req := lnrpc.GetTransactionsRequest{}
	res, err := client.GetTransactions(ctx, &req)

	for _, tx := range res.Transactions {
		err = storeTransaction(db, tx)
		if err != nil {
			return errors.Wrapf(err, "ImportTransactions -> storeTransaction(%v, %v)", db, tx)
		}
	}

	return nil
}

// SubscribeAndStoreTransactions Subscribes to on-chain transaction events from LND and stores them in the
// database as a time series. It will also import unregistered transactions on startup.
func SubscribeAndStoreTransactions(ctx context.Context, client lnrpc.LightningClient, db *sqlx.DB) error {

	err := ImportTransactions(client, db)
	if err != nil {
		return errors.Wrapf(err, "ImportTransactions(%v, %v)", client, db)
	}

	req := lnrpc.GetTransactionsRequest{}
	stream, err := client.SubscribeTransactions(ctx, &req)
	if err != nil {
		return errors.Wrapf(err, "SubscribeAndStoreTransactions -> client.SubscribeTransactions(%v, %v)",
			ctx, req)
	}
	rl := ratelimit.New(1) // 1 per second maximum rate limit

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		tx, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Printf("Subscribe transactions stream receive error: %v", err)
			// rate limited resubscribe
			rl.Take()
			stream, err = client.SubscribeTransactions(ctx, &req)
			continue
		}

		err = storeTransaction(db, tx)
		if err != nil {
			fmt.Printf("Subscribe transaction events store transaction error: %v", err)
			// rate limit for caution but hopefully not needed
			rl.Take()
			continue
		}
	}

	return nil
}

var insertTx = `INSERT INTO tx (timestamp, tx_hash, amount, num_confirmations, block_hash, block_height, 
                total_fees, dest_addresses, raw_tx_hex, label) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
                ON CONFLICT (timestamp, tx_hash) DO NOTHING;`

func storeTransaction(db *sqlx.DB, tx *lnrpc.Transaction) error {

	_, err := db.Exec(insertTx,
		time.Unix(tx.TimeStamp, 0).UTC(),
		tx.TxHash,
		tx.Amount,
		tx.NumConfirmations,
		tx.BlockHash,
		tx.BlockHeight,
		tx.TotalFees,
		pq.Array(tx.DestAddresses),
		tx.RawTxHex,
		tx.Label,
	)

	if err != nil {
		return errors.Wrapf(err, `storeTransaction -> db.Exec(%s, %s, %s, %d, %d, %s, %d, %d, %v, %s, %s)`, insertTx,
			time.Unix(tx.TimeStamp, 0).UTC(),
			tx.TxHash,
			tx.Amount,
			tx.NumConfirmations,
			tx.BlockHash,
			tx.BlockHeight,
			tx.TotalFees,
			pq.Array(tx.DestAddresses),
			tx.RawTxHex,
			tx.Label)
	}
	return nil

}
