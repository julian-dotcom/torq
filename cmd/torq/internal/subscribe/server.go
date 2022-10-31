package subscribe

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd"

	// "github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// Start runs the background server. It subscribes to events, gossip and
// fetches data as needed and stores it in the database.
// It is meant to run as a background task / daemon and is the bases for all
// of Torqs data collection
func Start(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int, wsChan chan interface{}) error {

	_, monitorCancel := context.WithCancel(context.Background())

	router := routerrpc.NewRouterClient(conn)
	client := lnrpc.NewLightningClient(conn)

	// Create an error group to catch errors from go routines.
	// TODO: Improve this by using the context to propogate the error,
	//   shutting down the if one of the subscribe go routines fail.
	//   https://www.fullstory.com/blog/why-errgroup-withcontext-in-golang-server-handlers/
	// TODO: Also consider using the same context used by the gRPC connection from Golang and the
	//   gRPC server of Torq
	errs, ctx := errgroup.WithContext(ctx)

	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	//Import Open channels
	err := lnd.ImportChannelList(lnrpc.ChannelEventUpdate_OPEN_CHANNEL, db, client, nodeSettings)
	if err != nil {
		monitorCancel()
		return errors.Wrap(err, "LND import channels list - open chanel")
	}

	// Import Closed channels
	err = lnd.ImportChannelList(lnrpc.ChannelEventUpdate_CLOSED_CHANNEL, db, client, nodeSettings)
	if err != nil {
		monitorCancel()
		return errors.Wrap(err, "LND import channels list - closed chanel")
	}

	// Import routing policies from open channels
	err = lnd.ImportRoutingPolicies(client, db, nodeSettings)
	if err != nil {
		monitorCancel()
		return errors.Wrap(err, "LND import routing policies")
	}

	// Transactions
	errs.Go(func() error {
		err := lnd.SubscribeAndStoreTransactions(ctx, client, db, nodeSettings, wsChan)
		if err != nil {
			return errors.Wrap(err, "LND subscribe and store transactions")
		}
		return nil
	})

	// // HTLC events
	errs.Go(func() error {
		err := lnd.SubscribeAndStoreHtlcEvents(ctx, router, db, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "LND subscribe and store HTLC events")
		}
		return nil
	})

	// // Channel Events
	errs.Go(func() error {
		err := lnd.SubscribeAndStoreChannelEvents(ctx, client, db, nodeSettings, wsChan)
		if err != nil {
			return errors.Wrap(err, "LND subscribe and store channel events")
		}
		return nil
	})

	// Graph (Node updates, fee updates etc.)
	errs.Go(func() error {
		err := lnd.SubscribeAndStoreChannelGraph(ctx, client, db, nodeSettings)
		if err != nil {
			return errors.Wrap(err, "LND subscribe and store channel graph")
		}
		return nil
	})

	// Forwarding history
	errs.Go(func() error {
		err := lnd.SubscribeForwardingEvents(ctx, client, db, nodeSettings, nil)
		if err != nil {
			return errors.Wrap(err, "LND subscribe forwarding events")
		}
		return nil
	})

	// Invoices
	errs.Go(func() error {
		err := lnd.SubscribeAndStoreInvoices(ctx, client, db, nodeSettings, wsChan)
		if err != nil {
			return errors.Wrap(err, "LND subscribe and store invoices")
		}
		return nil
	})

	// Payments
	errs.Go(func() error {
		err := lnd.SubscribeAndStorePayments(ctx, client, db, nodeSettings, nil)
		if err != nil {
			return errors.Wrap(err, "LND subscribe and store payments")
		}
		return nil
	})

	// Update in flight payments
	errs.Go(func() error {
		err := lnd.UpdateInFlightPayments(ctx, client, db, nodeSettings, nil)
		if err != nil {
			return errors.Wrap(err, "LND subscribe and update payments")
		}
		return nil
	})

	// Peer Events
	errs.Go(func() error {
		err := lnd.SubscribePeerEvents(ctx, client, nodeSettings, wsChan)
		if err != nil {
			return errors.Wrap(err, "LND subscribe peer events")
		}
		return nil
	})

	err = errs.Wait()

	// Everything that will write to the PeerPubKeyList and ChanPointList has finised so we can cancel the monitor functions
	monitorCancel()

	return err
}
