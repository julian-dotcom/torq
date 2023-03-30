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

	defer commons.SetInactiveLndServiceState(commons.LndServiceChannelEventStream, nodeId)
	commons.SetActiveLndServiceState(commons.LndServiceChannelEventStream, nodeId)

	client := lnrpc.NewLightningClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	successTimes, err := lightning.Import(db, commons.ImportAllChannels, false, nodeId, commons.GetSuccessTimes(nodeId))
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channels for nodeId: %v", nodeSettings.NodeId)
		return
	}

	successTimes, err = lightning.Import(db, commons.ImportChannelRoutingPolicies, false, nodeId, successTimes)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Channel routing policies for nodeId: %v", nodeSettings.NodeId)
		return
	}

	successTimes, err = lightning.Import(db, commons.ImportNodeInformation, false, nodeId, successTimes)
	if err != nil {
		log.Error().Err(err).Msgf("LND import Node Information for nodeId: %v", nodeSettings.NodeId)
		return
	}

	commons.SetSuccessTimes(nodeId, successTimes)
	lnd.SubscribeAndStoreChannelEvents(ctx, client, db, nodeSettings)
}

func StartGraphEventStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer commons.SetInactiveLndServiceState(commons.LndServiceGraphEventStream, nodeId)
	commons.SetActiveLndServiceState(commons.LndServiceGraphEventStream, nodeId)

	client := lnrpc.NewLightningClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	lnd.SubscribeAndStoreChannelGraph(ctx, client, db, nodeSettings)
}

func StartHtlcEvents(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer commons.SetInactiveLndServiceState(commons.LndServiceHtlcEventStream, nodeId)
	commons.SetActiveLndServiceState(commons.LndServiceHtlcEventStream, nodeId)

	router := routerrpc.NewRouterClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	lnd.SubscribeAndStoreHtlcEvents(ctx, router, db, nodeSettings)
}

func StartPeerEvents(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer commons.SetInactiveLndServiceState(commons.LndServicePeerEventStream, nodeId)
	commons.SetActiveLndServiceState(commons.LndServicePeerEventStream, nodeId)

	client := lnrpc.NewLightningClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	lnd.SubscribePeerEvents(ctx, client, nodeSettings)
}

func StartChannelBalanceCacheMaintenance(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer commons.SetInactiveLndServiceState(commons.LndServiceChannelBalanceCacheStream, nodeId)
	commons.SetActiveLndServiceState(commons.LndServiceChannelBalanceCacheStream, nodeId)

	client := lnrpc.NewLightningClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	lnd.ChannelBalanceCacheMaintenance(ctx, client, db, nodeSettings)
}

func StartTransactionStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer commons.SetInactiveLndServiceState(commons.LndServiceTransactionStream, nodeId)
	commons.SetActiveLndServiceState(commons.LndServiceTransactionStream, nodeId)

	client := lnrpc.NewLightningClient(conn)
	chain := chainrpc.NewChainNotifierClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	lnd.SubscribeAndStoreTransactions(ctx, client, chain, db, nodeSettings)
}

func StartForwardStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer commons.SetInactiveLndServiceState(commons.LndServiceForwardStream, nodeId)
	commons.SetActiveLndServiceState(commons.LndServiceForwardStream, nodeId)

	client := lnrpc.NewLightningClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	lnd.SubscribeForwardingEvents(ctx, client, db, nodeSettings, nil)
}

func StartPaymentStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer commons.SetInactiveLndServiceState(commons.LndServicePaymentStream, nodeId)
	commons.SetActiveLndServiceState(commons.LndServicePaymentStream, nodeId)

	client := lnrpc.NewLightningClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	lnd.SubscribeAndStorePayments(ctx, client, db, nodeSettings, nil)
}

func StartInvoiceStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer commons.SetInactiveLndServiceState(commons.LndServiceInvoiceStream, nodeId)
	commons.SetActiveLndServiceState(commons.LndServiceInvoiceStream, nodeId)

	client := lnrpc.NewLightningClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	lnd.SubscribeAndStoreInvoices(ctx, client, db, nodeSettings)
}

func StartInFlightPaymentStream(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer commons.SetInactiveLndServiceState(commons.LndServiceInFlightPaymentStream, nodeId)
	commons.SetActiveLndServiceState(commons.LndServiceInFlightPaymentStream, nodeId)

	client := lnrpc.NewLightningClient(conn)
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	lnd.UpdateInFlightPayments(ctx, client, db, nodeSettings, nil)
}
