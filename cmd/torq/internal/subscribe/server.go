package subscribe

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd"

	"google.golang.org/grpc"
)

// Start runs the background server. It subscribes to events, gossip and
// fetches data as needed and stores it in the database.
// It is meant to run as a background task / daemon and is the bases for all
// of Torqs data collection
func Start(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int, eventChannel chan interface{}) error {
	router := routerrpc.NewRouterClient(conn)
	client := lnrpc.NewLightningClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	var wg sync.WaitGroup

	//Import Open channels
	err := lnd.ImportChannelList(lnrpc.ChannelEventUpdate_OPEN_CHANNEL, db, client, nodeSettings)
	if err != nil {
		return errors.Wrap(err, "LND import open channels list")
	}

	// Import Closed channels
	err = lnd.ImportChannelList(lnrpc.ChannelEventUpdate_CLOSED_CHANNEL, db, client, nodeSettings)
	if err != nil {
		return errors.Wrap(err, "LND import closed channels list")
	}

	// TODO FIXME channels with short_channel_id = null and status IN (1,2,100,101,102,103) should be fixed somehow???
	//  Open                   = 1
	//  Closing                = 2
	//	CooperativeClosed      = 100
	//	LocalForceClosed       = 101
	//	RemoteForceClosed      = 102
	//	BreachClosed           = 103

	// Transactions
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in SubscribeAndStoreTransactions: %v", panicError)
				log.Error().Msg("Cancelling the subscription.")
				err = ctx.Err()
				if err != nil {
					log.Error().Err(err).Msgf("Failed to cancel context after Panic in SubscribeAndStoreTransactions")
				}
			}
		}()
		lnd.SubscribeAndStoreTransactions(ctx, client, db, nodeSettings, eventChannel)
	})()

	// HTLC events
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in SubscribeAndStoreHtlcEvents: %v", panicError)
				log.Error().Msg("Cancelling the subscription.")
				err = ctx.Err()
				if err != nil {
					log.Error().Err(err).Msgf("Failed to cancel context after Panic in SubscribeAndStoreHtlcEvents")
				}
			}
		}()
		lnd.SubscribeAndStoreHtlcEvents(ctx, router, db, nodeSettings, eventChannel)
	})()

	// Channel events
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in SubscribeAndStoreChannelEvents: %v", panicError)
				log.Error().Msg("Cancelling the subscription.")
				err = ctx.Err()
				if err != nil {
					log.Error().Err(err).Msgf("Failed to cancel context after Panic in SubscribeAndStoreChannelEvents")
				}
			}
		}()
		lnd.SubscribeAndStoreChannelEvents(ctx, client, db, nodeSettings, eventChannel)
	})()

	// Graph (Node updates, fee updates etc.)
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in SubscribeAndStoreChannelGraph: %v", panicError)
				log.Error().Msg("Cancelling the subscription.")
				err = ctx.Err()
				if err != nil {
					log.Error().Err(err).Msgf("Failed to cancel context after Panic in SubscribeAndStoreChannelGraph")
				}
			}
		}()
		lnd.SubscribeAndStoreChannelGraph(ctx, client, db, nodeSettings, eventChannel)
	})()

	// Forwarding history
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in SubscribeForwardingEvents: %v", panicError)
				log.Error().Msg("Cancelling the subscription.")
				err = ctx.Err()
				if err != nil {
					log.Error().Err(err).Msgf("Failed to cancel context after Panic in SubscribeForwardingEvents")
				}
			}
		}()
		lnd.SubscribeForwardingEvents(ctx, client, db, nodeSettings, eventChannel, nil)
	})()

	// Invoices
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in SubscribeAndStoreInvoices: %v", panicError)
				log.Error().Msg("Cancelling the subscription.")
				err = ctx.Err()
				if err != nil {
					log.Error().Err(err).Msgf("Failed to cancel context after Panic in SubscribeAndStoreInvoices")
				}
			}
		}()
		lnd.SubscribeAndStoreInvoices(ctx, client, db, nodeSettings, eventChannel)
	})()

	// Payments
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in SubscribeAndStorePayments: %v", panicError)
				log.Error().Msg("Cancelling the subscription.")
				err = ctx.Err()
				if err != nil {
					log.Error().Err(err).Msgf("Failed to cancel context after Panic in SubscribeAndStorePayments")
				}
			}
		}()
		lnd.SubscribeAndStorePayments(ctx, client, db, nodeSettings, eventChannel, nil)
	})()

	// Update in flight payments
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in UpdateInFlightPayments: %v", panicError)
				log.Error().Msg("Cancelling the subscription.")
				err = ctx.Err()
				if err != nil {
					log.Error().Err(err).Msgf("Failed to cancel context after Panic in UpdateInFlightPayments")
				}
			}
		}()
		lnd.UpdateInFlightPayments(ctx, client, db, nodeSettings, eventChannel, nil)
	})()

	// Peer Events
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in SubscribePeerEvents: %v", panicError)
				log.Error().Msg("Cancelling the subscription.")
				err = ctx.Err()
				if err != nil {
					log.Error().Err(err).Msgf("Failed to cancel context after Panic in SubscribePeerEvents")
				}
			}
		}()
		lnd.SubscribePeerEvents(ctx, client, nodeSettings, eventChannel)
	})()

	wg.Wait()

	return nil
}
