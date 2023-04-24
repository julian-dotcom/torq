package subscribe

import (
	"context"
	"runtime/debug"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	cln2 "github.com/lncapital/torq/internal/cln"
	"github.com/lncapital/torq/internal/lnd"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/cln"
	"github.com/lncapital/torq/proto/lnrpc"
	"github.com/lncapital/torq/proto/lnrpc/chainrpc"
	"github.com/lncapital/torq/proto/lnrpc/routerrpc"

	"google.golang.org/grpc"
)

func StartChannelEventStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.LndServiceChannelEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	err := lnd.ImportAllChannels(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channels for nodeId: %v", nodeId)
		cache.SetFailedNodeServiceState(serviceType, nodeId)
		return
	}

	err = lnd.ImportChannelRoutingPolicies(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channel routing policies for nodeId: %v", nodeId)
		cache.SetFailedNodeServiceState(serviceType, nodeId)
		return
	}

	err = lnd.ImportNodeInformation(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Node Information for nodeId: %v", nodeId)
		cache.SetFailedNodeServiceState(serviceType, nodeId)
		return
	}

	lnd.SubscribeAndStoreChannelEvents(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartGraphEventStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.LndServiceGraphEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	err := lnd.ImportAllChannels(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channels for nodeId: %v", nodeId)
		cache.SetFailedNodeServiceState(serviceType, nodeId)
		return
	}

	err = lnd.ImportChannelRoutingPolicies(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channel routing policies for nodeId: %v", nodeId)
		cache.SetFailedNodeServiceState(serviceType, nodeId)
		return
	}

	err = lnd.ImportNodeInformation(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Node Information for nodeId: %v", nodeId)
		cache.SetFailedNodeServiceState(serviceType, nodeId)
		return
	}

	lnd.SubscribeAndStoreChannelGraph(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartHtlcEvents(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.LndServiceHtlcEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	lnd.SubscribeAndStoreHtlcEvents(ctx, routerrpc.NewRouterClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartPeerEvents(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.LndServicePeerEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	err := lnd.ImportPeerStatus(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import peer status for nodeId: %v", nodeId)
		cache.SetFailedNodeServiceState(serviceType, nodeId)
		return
	}

	lnd.SubscribePeerEvents(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartTransactionStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.LndServiceTransactionStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	lnd.SubscribeAndStoreTransactions(ctx,
		lnrpc.NewLightningClient(conn),
		chainrpc.NewChainNotifierClient(conn),
		db,
		cache.GetNodeSettingsByNodeId(nodeId))
}

func StartForwardsService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.LndServiceForwardsService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	lnd.SubscribeForwardingEvents(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId), nil)
}

func StartPaymentsService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.LndServicePaymentsService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	lnd.SubscribeAndStorePayments(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId), nil)
}

func StartInvoiceStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.LndServiceInvoiceStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	lnd.SubscribeAndStoreInvoices(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartInFlightPaymentsService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.LndServiceInFlightPaymentsService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	lnd.UpdateInFlightPayments(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId), nil)
}

func StartPeersService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.ClnServicePeersService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	cln2.SubscribeAndStorePeers(ctx, cln.NewNodeClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartChannelsService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.ClnServiceChannelsService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	cln2.SubscribeAndStoreChannels(ctx, cln.NewNodeClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartFundsService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.ClnServiceFundsService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	cln2.SubscribeAndStoreFunds(ctx, cln.NewNodeClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartNodesService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := services_helpers.ClnServiceNodesService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	cln2.SubscribeAndStoreNodes(ctx, cln.NewNodeClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartChannelBalanceCacheMaintenance(ctx context.Context,
	conn *grpc.ClientConn,
	db *sqlx.DB,
	nodeId int) {

	serviceType := services_helpers.LndServiceChannelBalanceCacheService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingNodeServiceState(serviceType, nodeId)

	lnd.ChannelBalanceCacheMaintenance(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}
