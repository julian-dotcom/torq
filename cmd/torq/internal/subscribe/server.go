package subscribe

import (
	"context"
	"runtime/debug"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/lightning"
	"github.com/lncapital/torq/internal/lnd"
	"github.com/lncapital/torq/proto/lnrpc"
	"github.com/lncapital/torq/proto/lnrpc/chainrpc"
	"github.com/lncapital/torq/proto/lnrpc/routerrpc"

	"google.golang.org/grpc"
)

func StartChannelEventStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.LndServiceChannelEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	err := lightning.ImportAllChannels(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channels for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	err = lightning.ImportChannelRoutingPolicies(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channel routing policies for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	err = lightning.ImportNodeInformation(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Node Information for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	lnd.SubscribeAndStoreChannelEvents(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartGraphEventStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.LndServiceGraphEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	err := lightning.ImportAllChannels(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channels for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	err = lightning.ImportChannelRoutingPolicies(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channel routing policies for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	err = lightning.ImportNodeInformation(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Node Information for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	lnd.SubscribeAndStoreChannelGraph(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartHtlcEvents(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.LndServiceHtlcEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.SubscribeAndStoreHtlcEvents(ctx, routerrpc.NewRouterClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartPeerEvents(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.LndServicePeerEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	err := lightning.ImportPeerStatus(db, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import peer status for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	lnd.SubscribePeerEvents(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartChannelBalanceCacheMaintenance(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.LndServiceChannelBalanceCacheStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.ChannelBalanceCacheMaintenance(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartTransactionStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.LndServiceTransactionStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.SubscribeAndStoreTransactions(ctx,
		lnrpc.NewLightningClient(conn),
		chainrpc.NewChainNotifierClient(conn),
		db,
		cache.GetNodeSettingsByNodeId(nodeId))
}

func StartForwardStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.LndServiceForwardStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.SubscribeForwardingEvents(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId), nil)
}

func StartPaymentStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.LndServicePaymentStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.SubscribeAndStorePayments(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId), nil)
}

func StartInvoiceStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.LndServiceInvoiceStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.SubscribeAndStoreInvoices(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId))
}

func StartInFlightPaymentStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.LndServiceInFlightPaymentStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.UpdateInFlightPayments(ctx, lnrpc.NewLightningClient(conn), db, cache.GetNodeSettingsByNodeId(nodeId), nil)
}
