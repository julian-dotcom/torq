package subscribe

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd"

	"google.golang.org/grpc"
)

const streamBootedCheckSeconds = 5
const streamPaymentsTickerSeconds = 10
const streamInflightPaymentsTickerSeconds = 60
const streamForwardsTickerSeconds = 10

// 70 because a reconnection is attempted every 60 seconds
const avoidChannelAndPolicyImportRerunTimeSeconds = 70

const genericBootstrappingTimeSeconds = 60

// Start runs the background server. It subscribes to events, gossip and
// fetches data as needed and stores it in the database.
// It is meant to run as a background task / daemon and is the bases for all
// of Torqs data collection
func Start(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, vectorUrl string, nodeId int, broadcaster broadcast.BroadcastServer,
	serviceEventChannel chan<- commons.ServiceEvent, htlcEventChannel chan<- commons.HtlcEvent, forwardEventChannel chan<- commons.ForwardEvent,
	channelEventChannel chan<- commons.ChannelEvent, nodeGraphEventChannel chan<- commons.NodeGraphEvent, channelGraphEventChannel chan<- commons.ChannelGraphEvent,
	invoiceEventChannel chan<- commons.InvoiceEvent, paymentEventChannel chan<- commons.PaymentEvent, transactionEventChannel chan<- commons.TransactionEvent,
	peerEventChannel chan<- commons.PeerEvent, blockEventChannel chan<- commons.BlockEvent, lightningRequestChannel chan<- interface{}) error {

	router := routerrpc.NewRouterClient(conn)
	client := lnrpc.NewLightningClient(conn)
	chain := chainrpc.NewChainNotifierClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	active := commons.ServiceActive

	var wg sync.WaitGroup

	importRequestChannel := make(chan lnd.ImportRequest)
	go (func() {
		successTimes := make(map[lnd.ImportType]time.Time, 0)
		for {
			select {
			case <-ctx.Done():
				return
			case importRequest := <-importRequestChannel:
				successTime, exists := successTimes[importRequest.ImportType]
				if exists && time.Since(successTime).Seconds() < avoidChannelAndPolicyImportRerunTimeSeconds {
					if importRequest.ImportType == lnd.ImportChannelAndRoutingPolicies {
						log.Info().Msgf("ImportChannelAndRoutingPolicies were imported very recently for nodeId: %v.", nodeSettings.NodeId)
					}
					if importRequest.ImportType == lnd.ImportNodeInformation {
						log.Info().Msgf("ImportNodeInformation were imported very recently for nodeId: %v.", nodeSettings.NodeId)
					}
					importRequest.Out <- nil
					continue
				}
				if importRequest.ImportType == lnd.ImportChannelAndRoutingPolicies {
					var err error
					//Import Pending channels
					err = lnd.ImportPendingChannels(db, vectorUrl, client, nodeSettings, lightningRequestChannel)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to import pending channels.")
						importRequest.Out <- err
						continue
					}

					//Import Open channels
					err = lnd.ImportOpenChannels(db, vectorUrl, client, nodeSettings, lightningRequestChannel)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to import open channels.")
						importRequest.Out <- err
						continue
					}

					// Import Closed channels
					err = lnd.ImportClosedChannels(db, vectorUrl, client, nodeSettings, lightningRequestChannel)
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
				if importRequest.ImportType == lnd.ImportNodeInformation {
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
	importRequestChannel <- lnd.ImportRequest{
		ImportType: lnd.ImportChannelAndRoutingPolicies,
		Out:        responseChannel,
	}
	err := <-responseChannel
	if err != nil {
		return errors.Wrapf(err, "LND import Channel And Routing Policies for nodeId: %v", nodeSettings.NodeId)
	}

	importRequestChannel <- lnd.ImportRequest{
		ImportType: lnd.ImportNodeInformation,
		Out:        responseChannel,
	}
	err = <-responseChannel
	if err != nil {
		return errors.Wrapf(err, "LND import Node Information for nodeId: %v", nodeSettings.NodeId)
	}

	// Channel events
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in ChannelEventStream (nodeId: %v) %v with stack: %v", nodeId, panicError, string(debug.Stack()))
				commons.RunningServices[commons.LndService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.SubscribeAndStoreChannelEvents(ctx, client, db, nodeSettings, channelEventChannel, importRequestChannel, serviceEventChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.ChannelEventStream, "ChannelEventStream", serviceEventChannel)

	// Graph (Node updates, fee updates etc.)
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in GraphEventStream (nodeId: %v) %v with stack: %v", nodeId, panicError, string(debug.Stack()))
				commons.RunningServices[commons.LndService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.SubscribeAndStoreChannelGraph(ctx, client, db, nodeSettings, nodeGraphEventChannel, channelGraphEventChannel, importRequestChannel, serviceEventChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.GraphEventStream, "GraphEventStream", serviceEventChannel)

	// HTLC events
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in HtlcEventStream (nodeId: %v) %v with stack: %v", nodeId, panicError, string(debug.Stack()))
				commons.RunningServices[commons.LndService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.SubscribeAndStoreHtlcEvents(ctx, router, db, nodeSettings, htlcEventChannel, serviceEventChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.HtlcEventStream, "HtlcEventStream", serviceEventChannel)

	// Peer Events
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in PeerEventStream (nodeId: %v) %v with stack: %v", nodeId, panicError, string(debug.Stack()))
				commons.RunningServices[commons.LndService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.SubscribePeerEvents(ctx, client, nodeSettings, peerEventChannel, serviceEventChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.PeerEventStream, "PeerEventStream", serviceEventChannel)

	// Channel Balance Cache Maintenance
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in ChannelBalanceCacheStream (nodeId: %v) %v with stack: %v", nodeId, panicError, string(debug.Stack()))
				commons.RunningServices[commons.LndService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.ChannelBalanceCacheMaintenance(ctx, client, db, nodeSettings, broadcaster, serviceEventChannel)
	})()
	// No need to waitForReadyState for ChannelBalanceCacheMaintenance

	// Transactions
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in TransactionStream (nodeId: %v) %v with stack: %v", nodeId, panicError, string(debug.Stack()))
				commons.RunningServices[commons.LndService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.SubscribeAndStoreTransactions(ctx, client, chain, db, nodeSettings, transactionEventChannel, blockEventChannel, serviceEventChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.TransactionStream, "TransactionStream", serviceEventChannel)

	// Forwarding history
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in ForwardStream (nodeId: %v) %v with stack: %v", nodeId, panicError, string(debug.Stack()))
				commons.RunningServices[commons.LndService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.SubscribeForwardingEvents(ctx, client, db, nodeSettings, forwardEventChannel, serviceEventChannel, nil)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.ForwardStream, "ForwardStream", serviceEventChannel)

	// Payments
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in PaymentStream (nodeId: %v) %v with stack: %v", nodeId, panicError, string(debug.Stack()))
				commons.RunningServices[commons.LndService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.SubscribeAndStorePayments(ctx, client, db, nodeSettings, paymentEventChannel, serviceEventChannel, nil)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.PaymentStream, "PaymentStream", serviceEventChannel)

	// Invoices
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in InvoiceStream (nodeId: %v) %v with stack: %v", nodeId, panicError, string(debug.Stack()))
				commons.RunningServices[commons.LndService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.SubscribeAndStoreInvoices(ctx, client, db, nodeSettings, invoiceEventChannel, serviceEventChannel)
	})()

	waitForReadyState(nodeSettings.NodeId, commons.InvoiceStream, "InvoiceStream", serviceEventChannel)

	// Update in flight payments
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in InFlightPaymentStream (nodeId: %v) %v with stack: %v", nodeId, panicError, string(debug.Stack()))
				commons.RunningServices[commons.LndService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.UpdateInFlightPayments(ctx, client, db, nodeSettings, serviceEventChannel, nil)
	})()

	// No need to waitForReadyState for UpdateInFlightPayments

	log.Info().Msgf("All LND specific streams are initializing for nodeId: %v", nodeId)

	wg.Wait()

	return nil
}

func waitForReadyState(nodeId int, subscriptionStream commons.SubscriptionStream, name string, serviceEventChannel chan<- commons.ServiceEvent) {
	log.Info().Msgf("LND %v initialization started for nodeId: %v", name, nodeId)
	streamStartTime := time.Now()
	time.Sleep(1 * time.Second)
	for {
		if commons.RunningServices[commons.LndService].GetStreamStatus(nodeId, subscriptionStream) == commons.ServiceActive {
			log.Info().Msgf("LND %v initial download done (in less then %s) for nodeId: %v", name, time.Since(streamStartTime).Round(1*time.Second), nodeId)
			return
		}
		if commons.RunningServices[commons.LndService].GetStreamStatus(nodeId, subscriptionStream) == commons.ServiceDeleted {
			log.Info().Msgf("LND %v skipped (in less then %s) for nodeId: %v", name, time.Since(streamStartTime).Round(1*time.Second), nodeId)
			return
		}
		if time.Since(streamStartTime).Seconds() > genericBootstrappingTimeSeconds {
			lastInitializationPing := commons.RunningServices[commons.LndService].GetStreamInitializationPingTime(nodeId, subscriptionStream)
			if lastInitializationPing == nil {
				log.Error().Msgf("LND %v could not be initialized for nodeId: %v", name, nodeId)
				return
			}
			pingTimeOutInSeconds := genericBootstrappingTimeSeconds
			switch subscriptionStream {
			case commons.ForwardStream:
				pingTimeOutInSeconds = pingTimeOutInSeconds + streamForwardsTickerSeconds
			case commons.PaymentStream:
				pingTimeOutInSeconds = pingTimeOutInSeconds + streamPaymentsTickerSeconds
			case commons.InFlightPaymentStream:
				pingTimeOutInSeconds = pingTimeOutInSeconds + streamInflightPaymentsTickerSeconds
			}
			if time.Since(*lastInitializationPing).Seconds() > float64(pingTimeOutInSeconds) {
				log.Info().Msgf("LND %v idle for over %v seconds for nodeId: %v", name, pingTimeOutInSeconds, nodeId)
				lnd.SendStreamEvent(serviceEventChannel, nodeId, subscriptionStream, commons.ServiceActive, commons.ServiceInitializing)
				return
			}
		}
		time.Sleep(streamBootedCheckSeconds * time.Second)
	}
}
