package subscribe

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd"

	"google.golang.org/grpc"
)

// Start runs the background server. It subscribes to events, gossip and
// fetches data as needed and stores it in the database.
// It is meant to run as a background task / daemon and is the bases for all
// of Torqs data collection
func Start(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	eventChannel chan interface{},
	serviceEventChannel chan commons.ServiceEvent, serviceChannel chan commons.ServiceChannelMessage) error {
	router := routerrpc.NewRouterClient(conn)
	client := lnrpc.NewLightningClient(conn)
	chain := chainrpc.NewChainNotifierClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	var wg sync.WaitGroup

	importRequestChannel := make(chan commons.ImportRequest)
	go (func() {
		successTimes := make(map[commons.ImportType]time.Time, 0)
		for {
			select {
			case <-ctx.Done():
				return
			case importRequest := <-importRequestChannel:
				successTime, exists := successTimes[importRequest.ImportType]
				if exists && time.Since(successTime).Seconds() < commons.AVOID_CHANNEL_AND_POLICY_IMPORT_RERUN_TIME_SECONDS {
					if importRequest.ImportType == commons.ImportChannelAndRoutingPolicies {
						log.Info().Msgf("ImportChannelAndRoutingPolicies were imported very recently for nodeId: %v.", nodeSettings.NodeId)
					}
					if importRequest.ImportType == commons.ImportNodeInformation {
						log.Info().Msgf("ImportNodeInformation were imported very recently for nodeId: %v.", nodeSettings.NodeId)
					}
					importRequest.Out <- nil
					continue
				}
				if importRequest.ImportType == commons.ImportChannelAndRoutingPolicies {
					var err error
					//Import Open channels
					err = lnd.ImportChannelList(lnrpc.ChannelEventUpdate_OPEN_CHANNEL, db, client, nodeSettings)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to import open channels.")
						importRequest.Out <- err
						continue
					}

					// Import Closed channels
					err = lnd.ImportChannelList(lnrpc.ChannelEventUpdate_CLOSED_CHANNEL, db, client, nodeSettings)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to import closed channels.")
						importRequest.Out <- err
						continue
					}

					// TODO FIXME channels with short_channel_id = null and status IN (1,2,100,101,102,103) should be fixed somehow???
					//  Open                   = 1
					//  Closing                = 2
					//	CooperativeClosed      = 100
					//	LocalForceClosed       = 101
					//	RemoteForceClosed      = 102
					//	BreachClosed           = 103

					err = channels.InitializeManagedChannelCache(db)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to Initialize ManagedChannelCache.")
						importRequest.Out <- err
						continue
					}

					err = lnd.ImportRoutingPolicies(client, db, nodeSettings)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to import routing policies.")
						importRequest.Out <- err
						continue
					}
					log.Info().Msgf("ImportChannelAndRoutingPolicies was imported successfully for nodeId: %v.", nodeSettings.NodeId)
				}
				if importRequest.ImportType == commons.ImportNodeInformation {
					err := lnd.ImportNodeInfo(client, db, nodeSettings)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to import node information.")
						importRequest.Out <- err
						continue
					}
					log.Info().Msgf("ImportNodeInformation was imported successfully for nodeId: %v.", nodeSettings.NodeId)
				}
				successTimes[importRequest.ImportType] = time.Now()
				importRequest.Out <- nil
			}
		}
	})()

	responseChannel := make(chan error)
	importRequestChannel <- commons.ImportRequest{
		ImportType: commons.ImportChannelAndRoutingPolicies,
		Out:        responseChannel,
	}
	err := <-responseChannel
	if err != nil {
		return errors.Wrapf(err, "LND import Channel And Routing Policies for nodeId: %v", nodeSettings.NodeId)
	}

	importRequestChannel <- commons.ImportRequest{
		ImportType: commons.ImportNodeInformation,
		Out:        responseChannel,
	}
	err = <-responseChannel
	if err != nil {
		return errors.Wrapf(err, "LND import Node Information for nodeId: %v", nodeSettings.NodeId)
	}

	// Transactions
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				recoverPanic(panicError, serviceChannel, nodeId, commons.TransactionStream)
			}
		}()
		lnd.SubscribeAndStoreTransactions(ctx, client, chain, db, nodeSettings, eventChannel, serviceEventChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.TransactionStream, "TransactionStream", serviceEventChannel)

	// Forwarding history
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				recoverPanic(panicError, serviceChannel, nodeId, commons.ForwardStream)
			}
		}()
		lnd.SubscribeForwardingEvents(ctx, client, db, nodeSettings, eventChannel, serviceEventChannel, nil)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.ForwardStream, "ForwardStream", serviceEventChannel)

	// Payments
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				recoverPanic(panicError, serviceChannel, nodeId, commons.PaymentStream)
			}
		}()
		lnd.SubscribeAndStorePayments(ctx, client, db, nodeSettings, eventChannel, serviceEventChannel, nil)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.PaymentStream, "PaymentStream", serviceEventChannel)

	// Invoices
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				recoverPanic(panicError, serviceChannel, nodeId, commons.InvoiceStream)
			}
		}()
		lnd.SubscribeAndStoreInvoices(ctx, client, db, nodeSettings, eventChannel, serviceEventChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.InvoiceStream, "InvoiceStream", serviceEventChannel)

	// Channel events
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				recoverPanic(panicError, serviceChannel, nodeId, commons.ChannelEventStream)
			}
		}()
		lnd.SubscribeAndStoreChannelEvents(ctx, client, db, nodeSettings, eventChannel, serviceEventChannel, importRequestChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.ChannelEventStream, "ChannelEventStream", serviceEventChannel)

	// Graph (Node updates, fee updates etc.)
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				recoverPanic(panicError, serviceChannel, nodeId, commons.GraphEventStream)
			}
		}()
		lnd.SubscribeAndStoreChannelGraph(ctx, client, db, nodeSettings, eventChannel, serviceEventChannel, importRequestChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.GraphEventStream, "GraphEventStream", serviceEventChannel)

	// HTLC events
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				recoverPanic(panicError, serviceChannel, nodeId, commons.HtlcEventStream)
			}
		}()
		lnd.SubscribeAndStoreHtlcEvents(ctx, router, db, nodeSettings, eventChannel, serviceEventChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.HtlcEventStream, "HtlcEventStream", serviceEventChannel)

	// Peer Events
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				recoverPanic(panicError, serviceChannel, nodeId, commons.PeerEventStream)
			}
		}()
		lnd.SubscribePeerEvents(ctx, client, nodeSettings, eventChannel, serviceEventChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.PeerEventStream, "PeerEventStream", serviceEventChannel)

	// Update in flight payments
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				recoverPanic(panicError, serviceChannel, nodeId, commons.InFlightPaymentStream)
			}
		}()
		lnd.UpdateInFlightPayments(ctx, client, db, nodeSettings, eventChannel, serviceEventChannel, nil)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.InFlightPaymentStream, "InFlightPaymentStream", serviceEventChannel)

	if commons.RunningServices[commons.LndService].GetStatus(nodeId) == commons.Active {
		log.Info().Msgf("LND completely initialized for nodeId: %v", nodeId)
	} else {
		log.Error().Msgf("LND completely initialized but somehow a stream got out-of-sync for nodeId: %v", nodeId)
	}

	wg.Wait()

	return nil
}

func recoverPanic(panicError any, serviceChannel chan commons.ServiceChannelMessage, nodeId int, subscriptionStream commons.SubscriptionStream) {
	log.Error().Msgf("Panic occurred in %v (nodeId: %v) %v", subscriptionStream, nodeId, panicError)
	log.Error().Msgf("Killing the LND Service for nodeId: %v", nodeId)
	resultChannel := make(chan commons.Status)
	serviceChannel <- commons.ServiceChannelMessage{
		NodeId:         nodeId,
		ServiceType:    commons.LndService,
		ServiceCommand: commons.Kill,
		NoDelay:        true,
		Out:            resultChannel,
	}
	switch <-resultChannel {
	case commons.Active:
		log.Error().Msgf("Killed LND service after Panic in %v (nodeId: %v)", subscriptionStream, nodeId)
	case commons.Pending:
		log.Error().Msgf("Failed to kill LND service (it's booting) after Panic in %v (nodeId: %v)", subscriptionStream, nodeId)
	case commons.Inactive:
		log.Error().Msgf("Killed LND service after Panic in %v (nodeId: %v)", subscriptionStream, nodeId)
	}
}

func waitForReadyState(nodeId int, subscriptionStream commons.SubscriptionStream, name string, serviceEventChannel chan commons.ServiceEvent) {
	log.Info().Msgf("LND %v initialization started for nodeId: %v", name, nodeId)
	streamStartTime := time.Now()
	time.Sleep(1 * time.Second)
	for {
		if commons.RunningServices[commons.LndService].GetStreamStatus(nodeId, subscriptionStream) == commons.Active {
			log.Info().Msgf("LND %v initial download done (in less then %s) for nodeId: %v", name, time.Since(streamStartTime).Round(1*time.Second), nodeId)
			return
		}
		if time.Since(streamStartTime).Seconds() > commons.GENERIC_BOOTSTRAPPING_TIME_SECONDS {
			lastInitializationPing := commons.RunningServices[commons.LndService].GetStreamInitializationPingTime(nodeId, subscriptionStream)
			if lastInitializationPing == nil {
				log.Error().Msgf("LND %v could not be initialized for nodeId: %v", name, nodeId)
				return
			} else {
				pingTimeOutInSeconds := commons.GENERIC_BOOTSTRAPPING_TIME_SECONDS
				switch subscriptionStream {
				case commons.ForwardStream:
					pingTimeOutInSeconds = pingTimeOutInSeconds + commons.STREAM_FORWARDS_TICKER_SECONDS
				case commons.PaymentStream:
					pingTimeOutInSeconds = pingTimeOutInSeconds + commons.STREAM_PAYMENTS_TICKER_SECONDS
				case commons.InFlightPaymentStream:
					pingTimeOutInSeconds = pingTimeOutInSeconds + commons.STREAM_INFLIGHT_PAYMENTS_TICKER_SECONDS
				case commons.InvoiceStream:
					pingTimeOutInSeconds = commons.INVOICE_BOOTSTRAPPING_TIME_SECONDS
				}
				if time.Since(*lastInitializationPing).Seconds() > float64(pingTimeOutInSeconds) {
					log.Info().Msgf("LND %v idle for over %v seconds for nodeId: %v", name, pingTimeOutInSeconds, nodeId)
					lnd.SendStreamEvent(serviceEventChannel, nodeId, subscriptionStream, commons.Active, commons.Initializing)
					return
				}
			}
		}
		time.Sleep(commons.STREAM_BOOTED_CHECK_SECONDS * time.Second)
	}
}
