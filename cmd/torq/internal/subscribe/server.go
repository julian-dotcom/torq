package subscribe

import (
	"context"
	"runtime/debug"

	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/cache"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lightning"
	"github.com/lncapital/torq/pkg/lnd"

	"google.golang.org/grpc"
)

func StartChannelEventStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.LndServiceChannelEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	err := lightning.Import(db, commons.ImportAllChannels, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channels for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	err = lightning.Import(db, commons.ImportChannelRoutingPolicies, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channel routing policies for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	err = lightning.Import(db, commons.ImportNodeInformation, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Node Information for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	lnd.SubscribeAndStoreChannelEvents(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId))
}

func StartGraphEventStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.LndServiceGraphEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	err := lightning.Import(db, commons.ImportAllChannels, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channels for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	err = lightning.Import(db, commons.ImportChannelRoutingPolicies, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channel routing policies for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	err = lightning.Import(db, commons.ImportNodeInformation, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Node Information for nodeId: %v", nodeId)
		cache.SetFailedLndServiceState(serviceType, nodeId)
		return
	}

	lnd.SubscribeAndStoreChannelGraph(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId))
}

func StartHtlcEvents(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.LndServiceHtlcEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.SubscribeAndStoreHtlcEvents(ctx, routerrpc.NewRouterClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId))
}

func StartPeerEvents(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.LndServicePeerEventStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.SubscribePeerEvents(ctx, lnrpc.NewLightningClient(conn), commons.GetNodeSettingsByNodeId(nodeId))
}

func StartChannelBalanceCacheMaintenance(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.LndServiceChannelBalanceCacheStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.ChannelBalanceCacheMaintenance(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId))
}

func StartTransactionStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.LndServiceTransactionStream

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
		commons.GetNodeSettingsByNodeId(nodeId))
}

func StartForwardStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.LndServiceForwardStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.SubscribeForwardingEvents(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId), nil)
}

func StartPaymentStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.LndServicePaymentStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.SubscribeAndStorePayments(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId), nil)
}

func StartInvoiceStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.LndServiceInvoiceStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.SubscribeAndStoreInvoices(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId))
}

func StartInFlightPaymentStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.LndServiceInFlightPaymentStream

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetPendingLndServiceState(serviceType, nodeId)

	lnd.UpdateInFlightPayments(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId), nil)
}
