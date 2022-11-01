package lnd

import (
	"reflect"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/testutil"
)

type Transaction struct {
	Amount           int64          `db:"amount"`
	NumConfirmations int32          `db:"num_confirmations"`
	BlockHash        string         `db:"block_hash"`
	BlockHeight      int32          `db:"block_height"`
	TotalFees        int64          `db:"total_fees"`
	DestAddresses    pq.StringArray `db:"dest_addresses"`
	Label            string         `db:"label"`
}

func TestStoreTransaction(t *testing.T) {

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, err := srv.NewTestDatabase(true)
	// TODO FIXME WHY?
	defer time.Sleep(5 * time.Second)
	defer cancel()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expected := Transaction{
		Amount:           100000,
		NumConfirmations: 1,
		BlockHash:        "0000000000000000000000000000000000000000000000000000000000000000",
		BlockHeight:      1,
		TotalFees:        1000,
		DestAddresses: []string{
			"sb1q3e0rpuq04nknd9gzd7kfp5tqqfuvmxd3v9aaax",
			"sb1qzfw8yz3ays09rztc9vcvpey2l2tzf2kefclmap",
		},
	}

	err = storeTransaction(db, &lnrpc.Transaction{
		TxHash:           "test",
		Amount:           expected.Amount,
		NumConfirmations: expected.NumConfirmations,
		BlockHash:        expected.BlockHash,
		BlockHeight:      expected.BlockHeight,
		TimeStamp:        time.Now().Unix(),
		TotalFees:        expected.TotalFees,
		OutputDetails: []*lnrpc.OutputDetail{
			{
				OutputType:   lnrpc.OutputScriptType(1),
				Address:      expected.DestAddresses[0],
				PkScript:     "testScript1",
				OutputIndex:  0,
				Amount:       60,
				IsOurAddress: false,
			},
			{
				OutputType:   lnrpc.OutputScriptType(1),
				Address:      expected.DestAddresses[1],
				PkScript:     "testScript2",
				OutputIndex:  1,
				Amount:       40,
				IsOurAddress: true,
			},
		},
		RawTxHex:          "",
		Label:             expected.Label,
		PreviousOutpoints: nil,
	}, commons.GetNodeIdFromPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet))

	if err != nil {
		testutil.Fatalf(t, "storeTransaction", err)
	}

	row := db.QueryRowx(`select
		amount,
		num_confirmations,
		block_hash,
		block_height,
		total_fees,
		dest_addresses,
		label
	from tx LIMIT 1;`)

	if row.Err() != nil {
		testutil.Fatalf(t, "querying tx table", err)
	}

	got := Transaction{}
	err = row.StructScan(&got)
	if err != nil {
		testutil.Fatalf(t, "scanning row", err)
	}

	if !reflect.DeepEqual(got, expected) {
		testutil.Errorf(t, "Got:\n%v\nWant:\n%v\n", got, expected)
	}

}
