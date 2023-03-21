package subscribe

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd"

	"google.golang.org/grpc"
)

const streamBootedCheckSeconds = 5
const streamPaymentsTickerSeconds = 10
const streamInflightPaymentsTickerSeconds = 60
const streamForwardsTickerSeconds = 10

const genericBootstrappingTimeSeconds = 60

// Start runs the background server. It subscribes to events, gossip and
// fetches data as needed and stores it in the database.
// It is meant to run as a background task / daemon and is the bases for all
// of Torqs data collection
func Start(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer,
	htlcEventChannel chan<- commons.HtlcEvent,
	forwardEventChannel chan<- commons.ForwardEvent,
	channelEventChannel chan<- commons.ChannelEvent,
	nodeGraphEventChannel chan<- commons.NodeGraphEvent,
	channelGraphEventChannel chan<- commons.ChannelGraphEvent,
	invoiceEventChannel chan<- commons.InvoiceEvent,
	paymentEventChannel chan<- commons.PaymentEvent,
	transactionEventChannel chan<- commons.TransactionEvent,
	peerEventChannel chan<- commons.PeerEvent,
	blockEventChannel chan<- commons.BlockEvent,
	lightningRequestChannel chan<- interface{}) error {

	active := commons.ServiceActive

	router := routerrpc.NewRouterClient(conn)
	client := lnrpc.NewLightningClient(conn)
	chain := chainrpc.NewChainNotifierClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	now := time.Now()
	responseChannel := make(chan commons.ImportResponse)
	lightningRequestChannel <- commons.ImportRequest{
		CommunicationRequest: commons.CommunicationRequest{
			RequestId:   fmt.Sprintf("%v", now.Unix()),
			RequestTime: &now,
			NodeId:      nodeSettings.NodeId,
		},
		ImportType:      commons.ImportAllChannels,
		ResponseChannel: responseChannel,
	}
	response := <-responseChannel
	if response.Error != nil {
		return errors.Wrapf(response.Error, "LND import Channels for nodeId: %v", nodeSettings.NodeId)
	}

	responseChannel = make(chan commons.ImportResponse)
	lightningRequestChannel <- commons.ImportRequest{
		CommunicationRequest: commons.CommunicationRequest{
			RequestId:   fmt.Sprintf("%v", now.Unix()),
			RequestTime: &now,
			NodeId:      nodeSettings.NodeId,
		},
		ImportType:      commons.ImportChannelRoutingPolicies,
		ResponseChannel: responseChannel,
	}
	response = <-responseChannel
	if response.Error != nil {
		return errors.Wrapf(response.Error, "LND import Channel routing policies for nodeId: %v", nodeSettings.NodeId)
	}

	responseChannel = make(chan commons.ImportResponse)
	lightningRequestChannel <- commons.ImportRequest{
		CommunicationRequest: commons.CommunicationRequest{
			RequestId:   fmt.Sprintf("%v", now.Unix()),
			RequestTime: &now,
			NodeId:      nodeSettings.NodeId,
		},
		ImportType:      commons.ImportNodeInformation,
		ResponseChannel: responseChannel,
	}
	response = <-responseChannel
	if response.Error != nil {
		return errors.Wrapf(response.Error, "LND import Node Information for nodeId: %v", nodeSettings.NodeId)
	}

	var wg sync.WaitGroup

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
		lnd.SubscribeAndStoreChannelEvents(ctx, client, db, nodeSettings, channelEventChannel, lightningRequestChannel)
	})()

	waitForReadyState(ctx, nodeSettings.NodeId, commons.ChannelEventStream, "ChannelEventStream")
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

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
		lnd.SubscribeAndStoreChannelGraph(ctx, client, db, nodeSettings, nodeGraphEventChannel, channelGraphEventChannel, lightningRequestChannel)
	})()

	waitForReadyState(ctx, nodeSettings.NodeId, commons.GraphEventStream, "GraphEventStream")
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

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
		lnd.SubscribeAndStoreHtlcEvents(ctx, router, db, nodeSettings, htlcEventChannel)
	})()

	waitForReadyState(ctx, nodeSettings.NodeId, commons.HtlcEventStream, "HtlcEventStream")
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

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
		lnd.SubscribePeerEvents(ctx, client, nodeSettings, peerEventChannel)
	})()

	waitForReadyState(ctx, nodeSettings.NodeId, commons.PeerEventStream, "PeerEventStream")
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

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
		lnd.ChannelBalanceCacheMaintenance(ctx, client, db, nodeSettings, broadcaster)
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
		lnd.SubscribeAndStoreTransactions(ctx, client, chain, db, nodeSettings, transactionEventChannel, blockEventChannel)
	})()

	waitForReadyState(ctx, nodeSettings.NodeId, commons.TransactionStream, "TransactionStream")
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

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
		lnd.SubscribeForwardingEvents(ctx, client, db, nodeSettings, forwardEventChannel, nil)
	})()

	waitForReadyState(ctx, nodeSettings.NodeId, commons.ForwardStream, "ForwardStream")
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

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
		lnd.SubscribeAndStorePayments(ctx, client, db, nodeSettings, paymentEventChannel, nil)
	})()

	waitForReadyState(ctx, nodeSettings.NodeId, commons.PaymentStream, "PaymentStream")
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

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
		lnd.SubscribeAndStoreInvoices(ctx, client, db, nodeSettings, invoiceEventChannel)
	})()

	waitForReadyState(ctx, nodeSettings.NodeId, commons.InvoiceStream, "InvoiceStream")
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

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
		lnd.UpdateInFlightPayments(ctx, client, db, nodeSettings, nil)
	})()

	// No need to waitForReadyState for UpdateInFlightPayments

	log.Info().Msgf("All LND specific streams are initializing for nodeId: %v", nodeId)

	wg.Wait()

	return nil
}

func waitForReadyState(ctx context.Context, nodeId int, subscriptionStream commons.SubscriptionStream, name string) {
	log.Info().Msgf("LND %v initialization started for nodeId: %v", name, nodeId)
	streamStartTime := time.Now()

	ticker := clock.New().Tick(streamBootedCheckSeconds * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
		}

		if errors.Is(ctx.Err(), context.Canceled) {
			return
		}
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
				lnd.SetStreamStatus(nodeId, subscriptionStream, commons.ServiceInitializing, commons.ServiceActive)
				return
			}
		}
	}
}
