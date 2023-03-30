package subscribe

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lightning"
	"github.com/lncapital/torq/pkg/lnd"

	"google.golang.org/grpc"
)

func StartChannelEventStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in LndServiceChannelEventStream (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.LndServiceChannelEventStream, nodeId)
			return
		}
	}()
	commons.SetActiveLndServiceState(commons.LndServiceChannelEventStream, nodeId)

	client := lnrpc.NewLightningClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	err := lightning.Import(db, commons.ImportAllChannels, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channels for nodeId: %v", nodeSettings.NodeId)
		commons.SetFailedLndServiceState(commons.LndServiceChannelEventStream, nodeId)
		return
	}

	err = lightning.Import(db, commons.ImportChannelRoutingPolicies, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channel routing policies for nodeId: %v", nodeSettings.NodeId)
		commons.SetFailedLndServiceState(commons.LndServiceChannelEventStream, nodeId)
		return
	}

	err = lightning.Import(db, commons.ImportNodeInformation, false, nodeId)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Node Information for nodeId: %v", nodeSettings.NodeId)
		commons.SetFailedLndServiceState(commons.LndServiceChannelEventStream, nodeId)
		return
	}

	lnd.SubscribeAndStoreChannelEvents(ctx, client, db, nodeSettings)

	commons.SetInactiveLndServiceState(commons.LndServiceChannelEventStream, nodeId)
}

func StartGraphEventStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in LndServiceGraphEventStream (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.LndServiceGraphEventStream, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(commons.LndServiceGraphEventStream, nodeId)
	}()
	commons.SetActiveLndServiceState(commons.LndServiceGraphEventStream, nodeId)

	lnd.SubscribeAndStoreChannelGraph(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId))
}

func StartHtlcEvents(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in LndServiceHtlcEventStream (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.LndServiceHtlcEventStream, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(commons.LndServiceHtlcEventStream, nodeId)
	}()
	commons.SetActiveLndServiceState(commons.LndServiceHtlcEventStream, nodeId)

	lnd.SubscribeAndStoreHtlcEvents(ctx, routerrpc.NewRouterClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId))
}

func StartPeerEvents(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in LndServicePeerEventStream (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.LndServicePeerEventStream, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(commons.LndServicePeerEventStream, nodeId)
	}()
	commons.SetActiveLndServiceState(commons.LndServicePeerEventStream, nodeId)

	lnd.SubscribePeerEvents(ctx, lnrpc.NewLightningClient(conn), commons.GetNodeSettingsByNodeId(nodeId))
}

func StartChannelBalanceCacheMaintenance(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in LndServiceChannelBalanceCacheStream (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.LndServiceChannelBalanceCacheStream, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(commons.LndServiceChannelBalanceCacheStream, nodeId)
	}()
	commons.SetActiveLndServiceState(commons.LndServiceChannelBalanceCacheStream, nodeId)

	lnd.ChannelBalanceCacheMaintenance(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId))
}

func StartTransactionStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in LndServiceTransactionStream (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.LndServiceTransactionStream, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(commons.LndServiceTransactionStream, nodeId)
	}()
	commons.SetActiveLndServiceState(commons.LndServiceTransactionStream, nodeId)

	lnd.SubscribeAndStoreTransactions(ctx,
		lnrpc.NewLightningClient(conn),
		chainrpc.NewChainNotifierClient(conn),
		db,
		commons.GetNodeSettingsByNodeId(nodeId))
}

func StartForwardStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in LndServiceForwardStream (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.LndServiceForwardStream, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(commons.LndServiceForwardStream, nodeId)
	}()
	commons.SetActiveLndServiceState(commons.LndServiceForwardStream, nodeId)

	lnd.SubscribeForwardingEvents(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId), nil)
}

func StartPaymentStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in LndServicePaymentStream (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.LndServicePaymentStream, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(commons.LndServicePaymentStream, nodeId)
	}()
	commons.SetActiveLndServiceState(commons.LndServicePaymentStream, nodeId)

	lnd.SubscribeAndStorePayments(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId), nil)
}

func StartInvoiceStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in LndServiceInvoiceStream (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.LndServiceInvoiceStream, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(commons.LndServiceInvoiceStream, nodeId)
	}()
	commons.SetActiveLndServiceState(commons.LndServiceInvoiceStream, nodeId)

	lnd.SubscribeAndStoreInvoices(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId))
}

func StartInFlightPaymentStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in LndServiceInFlightPaymentStream (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.LndServiceInFlightPaymentStream, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(commons.LndServiceInFlightPaymentStream, nodeId)
	}()
	commons.SetActiveLndServiceState(commons.LndServiceInFlightPaymentStream, nodeId)

	lnd.UpdateInFlightPayments(ctx, lnrpc.NewLightningClient(conn), db, commons.GetNodeSettingsByNodeId(nodeId), nil)
}
