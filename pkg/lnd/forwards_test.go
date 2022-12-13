package lnd

import (
	"context"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/mixer/clock"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/testutil"
)

// mockLightningClientForwardingHistory is used to moc responses from GetNodeInfo
type mockLightningClientForwardingHistory struct {
	CustomMaxEvents  int32
	ForwardingEvents []*lnrpc.ForwardingEvent
	LastOffsetIndex  uint32
	Error            error
	CancelFunc       func()
}

// TODO: Use fuzzy tests:
//	 https://go.dev/doc/fuzz/
//   https://go.dev/blog/fuzz-beta

// ForwardingHistory mocks the response of LNDs lnrpc.ForwardingHistory
func (c *mockLightningClientForwardingHistory) ForwardingHistory(ctx context.Context,
	in *lnrpc.ForwardingHistoryRequest,
	opts ...grpc.CallOption) (*lnrpc.ForwardingHistoryResponse, error) {

	if c.Error != nil {
		return nil, c.Error
	}

	if c.ForwardingEvents == nil {
		c.CancelFunc()
		return &lnrpc.ForwardingHistoryResponse{}, nil
	}

	r := lnrpc.ForwardingHistoryResponse{
		ForwardingEvents: c.ForwardingEvents,
		LastOffsetIndex:  c.LastOffsetIndex,
	}

	c.ForwardingEvents = nil

	return &r, nil
}

func TestSubscribeForwardingEvents(t *testing.T) {

	ctx := context.Background()
	ctx, stopSubFwE := context.WithCancel(ctx)
	c := clock.NewMockClock(time.Unix(0, 0))

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, err := srv.NewTestDatabase(true)
	defer cancel()
	if err != nil {
		t.Fatal(err)
	}
	//defer db.Close()

	err = settings.InitializeManagedSettingsCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing ManagedSettings cache: %v", err)
	}

	err = settings.InitializeManagedNodeCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing ManagedNode cache: %v", err)
	}

	err = channels.InitializeManagedChannelCache(db)
	if err != nil {
		cancel()
		log.Fatal().Err(err).Msgf("Problem initializing ManagedChannel cache: %v", err)
	}

	mockTickerInterval := 3000 * time.Second
	me := 1000
	opt := FwhOptions{
		MaxEvents: &me,
		Tick:      c.Tick(mockTickerInterval),
	}

	mclient := mockLightningClientForwardingHistory{
		CancelFunc: stopSubFwE,
		ForwardingEvents: []*lnrpc.ForwardingEvent{
			{
				ChanIdIn:    1111,
				ChanIdOut:   2222,
				AmtIn:       11,
				AmtOut:      10,
				Fee:         1,
				FeeMsat:     1000,
				AmtInMsat:   11000,
				AmtOutMsat:  10000,
				TimestampNs: uint64(c.Now().UnixNano()),
			},
			{
				ChanIdIn:    1111,
				ChanIdOut:   2222,
				AmtIn:       11,
				AmtOut:      10,
				Fee:         1,
				FeeMsat:     1000,
				AmtInMsat:   11000,
				AmtOutMsat:  10000,
				TimestampNs: uint64(c.Now().UnixNano()) + 500000000,
			},
			{ // Duplicate record used for testing
				ChanIdIn:    1111,
				ChanIdOut:   2222,
				AmtIn:       11,
				AmtOut:      10,
				Fee:         1,
				FeeMsat:     1000,
				AmtInMsat:   11000,
				AmtOutMsat:  10000,
				TimestampNs: uint64(c.Now().UnixNano()) + 500000000,
			},
			{
				ChanIdIn:    1111,
				ChanIdOut:   2222,
				AmtIn:       11,
				AmtOut:      10,
				Fee:         1,
				FeeMsat:     1000,
				AmtInMsat:   11000,
				AmtOutMsat:  10000,
				TimestampNs: uint64(c.Now().UnixNano()) + 1000000000,
			},
		},
		LastOffsetIndex: 0,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		SubscribeForwardingEvents(ctx, &mclient, db,
			commons.GetNodeSettingsByNodeId(
				commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet)), nil, &opt)
	}()
	// Simulate passing intervals, one more than required to process
	numbTicks := 2
	for i := 0; i < numbTicks; i++ {
		c.AddTime(mockTickerInterval)
	}
	wg.Wait()

	testutil.Given(t, "While running SubscribeForwardingEvents")

	testutil.WhenF(t, "We need to check that fetchLastForwardTime returns the expected nanosecond.")
	{
		var expected uint64 = 1000000000
		returned, err := fetchLastForwardTime(db)
		switch {
		case err != nil:
			testutil.Fatalf(t, "We get an error: %v", err)
		case returned != expected:
			testutil.Errorf(t, " "+
				"We expected %d got %d", expected, returned)
		case returned == expected:
			testutil.Successf(t, "We got the expected nanosecond ")
		}

	}

	testutil.WhenF(t, "We need to check that storeForwardingHistory only stores unique records.")
	{
		var expectedUnique = 3
		var returned int
		err := db.QueryRow("select count(*) from forward;").Scan(&returned)

		switch {
		case err != nil:
			testutil.Fatalf(t, "We get an error: %v", err)
		case returned != expectedUnique:
			testutil.Errorf(t, "We expected to store %d records but stored %d", expectedUnique,
				returned)
		case returned == expectedUnique:
			testutil.Successf(t, "We stored the expected number of records")
		}
	}

	var returnedDate time.Time
	err = db.QueryRow("select \"time\" from forward order by time_ns desc limit 1;").Scan(&returnedDate)
	if err != nil {
		t.Fatalf("Problem querying row: %v\n", err)
	}
	var expectedDate, _ = time.Parse("2006-01-02 15:04:05-0700 MST", "1970-01-01 00:00:01+0000 UTC")
	if returnedDate != expectedDate {
		t.Errorf("Time on the latest forward record (%v) isn't as expected (%v)", returnedDate, expectedDate)
	}

	db.Close()
	// err = srv.Cleanup()
	// if err != nil {
	// 	t.Fatal(err)
	// }

}
