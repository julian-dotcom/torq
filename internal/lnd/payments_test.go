package lnd

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/mixer/clock"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/testutil"

	"google.golang.org/grpc"
)

type mockLightningClient_ListPayments struct {
	Payments        []*lnrpc.Payment
	LastIndexOffset uint64
	CancelFunc      func()
}

// _ListPayments mocks the response of LNDs lnrpc._ListPayments
func (c *mockLightningClient_ListPayments) ListPayments(ctx context.Context, in *lnrpc.ListPaymentsRequest,
	opts ...grpc.CallOption) (*lnrpc.ListPaymentsResponse, error) {

	r := lnrpc.ListPaymentsResponse{
		Payments:        c.Payments,
		LastIndexOffset: c.LastIndexOffset,
	}

	if len(c.Payments) > 0 {
		c.Payments = c.Payments[1:]
		return &r, nil
	}
	c.Payments = nil
	c.CancelFunc()
	return nil, context.Canceled

}

func TestSubscribePayments(t *testing.T) {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := clock.NewMockClock(time.Unix(0, 0))

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, dbCancel, err := srv.NewTestDatabase(true)
	defer dbCancel()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = settings.InitializeSettingsCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing SettingsCache cache: %v", err)
	}

	err = settings.InitializeNodesCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing NodeCache cache: %v", err)
	}

	err = settings.InitializeChannelsCache(db)
	if err != nil {
		cancel()
		log.Fatal().Err(err).Msgf("Problem initializing ChannelCache cache: %v", err)
	}

	mockTickerInterval := 1 * time.Millisecond
	opt := PayOptions{
		Tick: c.Tick(mockTickerInterval),
	}

	var hops []*lnrpc.Hop
	hops = append(hops, &lnrpc.Hop{
		ChanId: 1111,
		PubKey: testutil.TestPublicKey1,
	})
	hops = append(hops, &lnrpc.Hop{
		ChanId: 2222,
		PubKey: testutil.TestPublicKey1,
	})
	var htlcs []*lnrpc.HTLCAttempt
	htlcs = append(htlcs, &lnrpc.HTLCAttempt{
		AttemptId: 1,
		Route: &lnrpc.Route{
			Hops: hops,
		},
	})

	createdAt := time.Now().UTC()
	mclient := mockLightningClient_ListPayments{
		Payments: []*lnrpc.Payment{
			{
				PaymentIndex:    10,
				PaymentHash:     "SomeHash",
				Value:           0,
				CreationDate:    createdAt.Unix(),
				Fee:             10,
				PaymentPreimage: "SomePreimage",
				ValueSat:        1000,
				ValueMsat:       1000000,
				PaymentRequest: "lnbc142250n1psju2jfpp5jvgszhxwanal28cfxe7yr5tjmayfh4ehlk4ms504nyswx2qjf0psdz0235x2grsw" +
					"fjhqcted4jkuapqvehhygrpyp3ksctwdejkcgr0wpjku6twvusxzapqf38yy6289e3k7mgcqzpgxqrpxasp58zj7e3f4dadfsrz" +
					"wdv92e4j6vcst5ykvrxa47y9vp7x0h05r0fss9qy9qsq5dxkhqfj7ledlq8q7l9xfnlzwfvzwj2zv9u7sewumjxc0q2p4dmptah" +
					"n9xdkqcxumd0z6ks3ms7jf86hllm44hv7xkgmzys72xc4zqqp96ut9z",
				Status:         lnrpc.Payment_SUCCEEDED,
				FeeSat:         10,
				FeeMsat:        10000,
				CreationTimeNs: createdAt.UnixNano(),
				Htlcs:          htlcs,
				FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_NONE,
			},
			{
				PaymentIndex:    11,
				PaymentHash:     "InFLightTimeout",
				Value:           0,
				CreationDate:    1655644877,
				Fee:             10,
				PaymentPreimage: "InFLightTimeout",
				ValueSat:        1000,
				ValueMsat:       1000000,
				PaymentRequest: "lnbc142250n1psju2jfpp5jvgszhxwanal28cfxe7yr5tjmayfh4ehlk4ms504nyswx2qjf0psdz0235x2grsw" +
					"fjhqcted4jkuapqvehhygrpyp3ksctwdejkcgr0wpjku6twvusxzapqf38yy6289e3k7mgcqzpgxqrpxasp58zj7e3f4dadfsrz" +
					"wdv92e4j6vcst5ykvrxa47y9vp7x0h05r0fss9qy9qsq5dxkhqfj7ledlq8q7l9xfnlzwfvzwj2zv9u7sewumjxc0q2p4dmptah" +
					"n9xdkqcxumd0z6ks3ms7jf86hllm44hv7xkgmzys72xc4zqqp96ut9z",
				Status:         lnrpc.Payment_IN_FLIGHT,
				FeeSat:         10,
				FeeMsat:        10000,
				CreationTimeNs: time.Unix(1624108877, 0).UnixNano(),
				Htlcs:          htlcs,
				FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_NONE,
			},
			{
				PaymentIndex:    12,
				PaymentHash:     "SomeHash3",
				Value:           0,
				CreationDate:    createdAt.Unix(),
				Fee:             10,
				PaymentPreimage: "SomePreimage3",
				ValueSat:        1000,
				ValueMsat:       1000000,
				PaymentRequest: "lnbc142250n1psju2jfpp5jvgszhxwanal28cfxe7yr5tjmayfh4ehlk4ms504nyswx2qjf0psdz0235x2grsw" +
					"fjhqcted4jkuapqvehhygrpyp3ksctwdejkcgr0wpjku6twvusxzapqf38yy6289e3k7mgcqzpgxqrpxasp58zj7e3f4dadfsrz" +
					"wdv92e4j6vcst5ykvrxa47y9vp7x0h05r0fss9qy9qsq5dxkhqfj7ledlq8q7l9xfnlzwfvzwj2zv9u7sewumjxc0q2p4dmptah" +
					"n9xdkqcxumd0z6ks3ms7jf86hllm44hv7xkgmzys72xc4zqqp96ut9z",
				Status:         lnrpc.Payment_IN_FLIGHT,
				FeeSat:         10,
				FeeMsat:        10000,
				CreationTimeNs: createdAt.UnixNano(),
				Htlcs:          htlcs,
				FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_NO_ROUTE,
			},
			{
				PaymentIndex:    13,
				PaymentHash:     "SomeHash4",
				Value:           0,
				CreationDate:    createdAt.Unix(),
				Fee:             10,
				PaymentPreimage: "SomePreimage4",
				ValueSat:        1000,
				ValueMsat:       1000000,
				PaymentRequest: "lnbc142250n1psju2jfpp5jvgszhxwanal28cfxe7yr5tjmayfh4ehlk4ms504nyswx2qjf0psdz0235x2grsw" +
					"fjhqcted4jkuapqvehhygrpyp3ksctwdejkcgr0wpjku6twvusxzapqf38yy6289e3k7mgcqzpgxqrpxasp58zj7e3f4dadfsrz" +
					"wdv92e4j6vcst5ykvrxa47y9vp7x0h05r0fss9qy9qsq5dxkhqfj7ledlq8q7l9xfnlzwfvzwj2zv9u7sewumjxc0q2p4dmptah" +
					"n9xdkqcxumd0z6ks3ms7jf86hllm44hv7xkgmzys72xc4zqqp96ut9z",
				Status:         lnrpc.Payment_FAILED,
				FeeSat:         10,
				FeeMsat:        10000,
				CreationTimeNs: createdAt.UnixNano(),
				Htlcs:          htlcs,
				FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_TIMEOUT,
			},
			{
				PaymentIndex:    14,
				PaymentHash:     "SomeHash5",
				Value:           0,
				CreationDate:    createdAt.Unix(),
				Fee:             10,
				PaymentPreimage: "SomePreimage5",
				ValueSat:        1000,
				ValueMsat:       1000000,
				PaymentRequest: "lnbc142250n1psju2jfpp5jvgszhxwanal28cfxe7yr5tjmayfh4ehlk4ms504nyswx2qjf0psdz0235x2grsw" +
					"fjhqcted4jkuapqvehhygrpyp3ksctwdejkcgr0wpjku6twvusxzapqf38yy6289e3k7mgcqzpgxqrpxasp58zj7e3f4dadfsrz" +
					"wdv92e4j6vcst5ykvrxa47y9vp7x0h05r0fss9qy9qsq5dxkhqfj7ledlq8q7l9xfnlzwfvzwj2zv9u7sewumjxc0q2p4dmptah" +
					"n9xdkqcxumd0z6ks3ms7jf86hllm44hv7xkgmzys72xc4zqqp96ut9z",
				Status:         lnrpc.Payment_IN_FLIGHT,
				FeeSat:         10,
				FeeMsat:        10000,
				CreationTimeNs: createdAt.UnixNano(),
				Htlcs:          htlcs,
				FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_TIMEOUT,
			},
			{
				PaymentIndex:    15,
				PaymentHash:     "SomeHash6",
				Value:           0,
				CreationDate:    createdAt.Unix(),
				Fee:             10,
				PaymentPreimage: "SomePreimage6",
				ValueSat:        1000,
				ValueMsat:       1000000,
				PaymentRequest: "lnbc142250n1psju2jfpp5jvgszhxwanal28cfxe7yr5tjmayfh4ehlk4ms504nyswx2qjf0psdz0235x2grsw" +
					"fjhqcted4jkuapqvehhygrpyp3ksctwdejkcgr0wpjku6twvusxzapqf38yy6289e3k7mgcqzpgxqrpxasp58zj7e3f4dadfsrz" +
					"wdv92e4j6vcst5ykvrxa47y9vp7x0h05r0fss9qy9qsq5dxkhqfj7ledlq8q7l9xfnlzwfvzwj2zv9u7sewumjxc0q2p4dmptah" +
					"n9xdkqcxumd0z6ks3ms7jf86hllm44hv7xkgmzys72xc4zqqp96ut9z",
				Status:         lnrpc.Payment_IN_FLIGHT,
				FeeSat:         10,
				FeeMsat:        10000,
				CreationTimeNs: time.Unix(1630415433, 0).UnixNano(),
				Htlcs:          htlcs,
				FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_NONE,
			},
		},
		CancelFunc: cancel,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		SubscribeAndStorePayments(ctx, &mclient, db,
			cache.GetNodeSettingsByNodeId(
				cache.GetNodeIdByPublicKey(testutil.TestPublicKey1, core.Bitcoin, core.SigNet)), &opt)
	}()
	// Simulate passing intervals
	numbTicks := 4
	for i := 0; i < numbTicks; i++ {
		c.AddTime(mockTickerInterval)
	}
	wg.Wait()

	t.Run("Last payment index is stored correctly", func(t *testing.T) {
		var expected uint64 = 15
		returned, err := fetchLastPaymentIndex(db, 1)
		switch {
		case err != nil:
			testutil.Fatalf(t, "We get an error: %v", err)
		case returned != expected:
			testutil.Errorf(t, "We expected %d got %d", expected, returned)
		case returned == expected:
			testutil.Successf(t, "We got the expected payment index ")
		}
	})

	t.Run("List of in flight payments is correct.", func(t *testing.T) {
		var expected = []uint64{11, 12, 14, 15}
		returned, err := fetchInFlightPaymentIndexes(db, 1)
		switch {
		case err != nil:
			testutil.Fatalf(t, "We get an error: %v", err)
		case !reflect.DeepEqual(returned, expected):
			testutil.Errorf(t, "We expected %v got %v", expected, returned)
		case reflect.DeepEqual(returned, expected):
			testutil.Successf(t, "We got the expected list of payment indexes")
		}
	})

	// reset context
	ctx = context.Background()
	ctx, cancel = context.WithCancel(ctx)

	mclientUpdate := mockLightningClient_ListPayments{}
	mclientUpdate.CancelFunc = cancel
	mclientUpdate.Payments = []*lnrpc.Payment{
		{ // This payment is still in flight and should not be changed
			PaymentIndex:    11,
			PaymentHash:     "SomeHash2",
			Value:           0,
			CreationDate:    createdAt.Unix(),
			Fee:             10,
			PaymentPreimage: "SomePreimage2",
			ValueSat:        1000,
			ValueMsat:       1000000,
			PaymentRequest: "lnbc142250n1psju2jfpp5jvgszhxwanal28cfxe7yr5tjmayfh4ehlk4ms504nyswx2qjf0psdz0235x2grsw" +
				"fjhqcted4jkuapqvehhygrpyp3ksctwdejkcgr0wpjku6twvusxzapqf38yy6289e3k7mgcqzpgxqrpxasp58zj7e3f4dadfsrz" +
				"wdv92e4j6vcst5ykvrxa47y9vp7x0h05r0fss9qy9qsq5dxkhqfj7ledlq8q7l9xfnlzwfvzwj2zv9u7sewumjxc0q2p4dmptah" +
				"n9xdkqcxumd0z6ks3ms7jf86hllm44hv7xkgmzys72xc4zqqp96ut9z",
			Status:         lnrpc.Payment_IN_FLIGHT,
			FeeSat:         10,
			FeeMsat:        10000,
			CreationTimeNs: createdAt.UnixNano(),
			Htlcs:          nil,
			FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_NONE,
		},
		{ // This payment succeeded and should be updated
			PaymentIndex:    12,
			PaymentHash:     "SomeHash3",
			Value:           0,
			CreationDate:    createdAt.Unix(),
			Fee:             10,
			PaymentPreimage: "SomePreimage3",
			ValueSat:        1000,
			ValueMsat:       1000000,
			PaymentRequest: "lnbc142250n1psju2jfpp5jvgszhxwanal28cfxe7yr5tjmayfh4ehlk4ms504nyswx2qjf0psdz0235x2grsw" +
				"fjhqcted4jkuapqvehhygrpyp3ksctwdejkcgr0wpjku6twvusxzapqf38yy6289e3k7mgcqzpgxqrpxasp58zj7e3f4dadfsrz" +
				"wdv92e4j6vcst5ykvrxa47y9vp7x0h05r0fss9qy9qsq5dxkhqfj7ledlq8q7l9xfnlzwfvzwj2zv9u7sewumjxc0q2p4dmptah" +
				"n9xdkqcxumd0z6ks3ms7jf86hllm44hv7xkgmzys72xc4zqqp96ut9z",
			Status:         lnrpc.Payment_SUCCEEDED,
			FeeSat:         10,
			FeeMsat:        10000,
			CreationTimeNs: createdAt.UnixNano(),
			Htlcs:          nil,
			FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_NONE,
		},
		{ // This payment FAILED and should be updated
			PaymentIndex:    14,
			PaymentHash:     "SomeHash5",
			Value:           0,
			CreationDate:    createdAt.Unix(),
			Fee:             10,
			PaymentPreimage: "SomePreimage5",
			ValueSat:        1000,
			ValueMsat:       1000000,
			PaymentRequest: "lnbc142250n1psju2jfpp5jvgszhxwanal28cfxe7yr5tjmayfh4ehlk4ms504nyswx2qjf0psdz0235x2grsw" +
				"fjhqcted4jkuapqvehhygrpyp3ksctwdejkcgr0wpjku6twvusxzapqf38yy6289e3k7mgcqzpgxqrpxasp58zj7e3f4dadfsrz" +
				"wdv92e4j6vcst5ykvrxa47y9vp7x0h05r0fss9qy9qsq5dxkhqfj7ledlq8q7l9xfnlzwfvzwj2zv9u7sewumjxc0q2p4dmptah" +
				"n9xdkqcxumd0z6ks3ms7jf86hllm44hv7xkgmzys72xc4zqqp96ut9z",
			Status:         lnrpc.Payment_FAILED,
			FeeSat:         10,
			FeeMsat:        10000,
			CreationTimeNs: createdAt.UnixNano(),
			Htlcs:          nil,
			FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_TIMEOUT,
		},
		{ // This payment should timeout
			PaymentIndex:    15,
			PaymentHash:     "SomeHash6",
			Value:           0,
			CreationDate:    1630415433,
			Fee:             10,
			PaymentPreimage: "SomePreimage6",
			ValueSat:        1000,
			ValueMsat:       1000000,
			PaymentRequest: "lnbc142250n1psju2jfpp5jvgszhxwanal28cfxe7yr5tjmayfh4ehlk4ms504nyswx2qjf0psdz0235x2grsw" +
				"fjhqcted4jkuapqvehhygrpyp3ksctwdejkcgr0wpjku6twvusxzapqf38yy6289e3k7mgcqzpgxqrpxasp58zj7e3f4dadfsrz" +
				"wdv92e4j6vcst5ykvrxa47y9vp7x0h05r0fss9qy9qsq5dxkhqfj7ledlq8q7l9xfnlzwfvzwj2zv9u7sewumjxc0q2p4dmptah" +
				"n9xdkqcxumd0z6ks3ms7jf86hllm44hv7xkgmzys72xc4zqqp96ut9z",
			Status:         lnrpc.Payment_IN_FLIGHT,
			FeeSat:         10,
			FeeMsat:        10000,
			CreationTimeNs: time.Unix(1630415433, 0).UnixNano(),
			Htlcs:          nil,
			FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_NONE,
		},
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		UpdateInFlightPayments(ctx, &mclientUpdate, db,
			cache.GetNodeSettingsByNodeId(
				cache.GetNodeIdByPublicKey(testutil.TestPublicKey1, core.Bitcoin, core.SigNet)), &opt)
	}()
	// Simulate passing intervals
	numbTicks = 6
	for i := 0; i < numbTicks; i++ {
		c.AddTime(mockTickerInterval)
	}
	wg.Wait()

	t.Run("List of in flight payments is correct after update.", func(t *testing.T) {
		var expected = []uint64{11}
		returned, err := fetchInFlightPaymentIndexes(db, 1)
		switch {
		case err != nil:
			testutil.Fatalf(t, "We get an error: %v", err)
		case !reflect.DeepEqual(returned, expected):
			testutil.Errorf(t, "We expected %v got %v", expected, returned)
		case reflect.DeepEqual(returned, expected):
			testutil.Successf(t, "We got the expected list of payment indexes")
		}
	})

}
