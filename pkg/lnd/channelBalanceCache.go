package lnd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
)

func ChannelBalanceCacheMaintenance(ctx context.Context, lndClient lnrpc.LightningClient, nodeSettings commons.ManagedNodeSettings,
	broadcaster broadcast.BroadcastServer, eventChannel chan interface{}, serviceEventChannel chan commons.ServiceEvent) {

	var err error
	serviceStatus := commons.Inactive
	bootStrapping := true
	subscriptionStream := commons.ChannelBalanceCacheStream

	// Create the default ticker used to fetch the latest information from LND
	c := clock.New()
	lndSyncTicker := c.Tick(60 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-lndSyncTicker:
		case event := <-broadcaster.Subscribe():
			if serviceEvent, ok := event.(commons.ServiceEvent); ok {
				if serviceEvent.Type == commons.LndService && serviceEvent.NodeId == nodeSettings.NodeId &&
					serviceEvent.Status == commons.Active && serviceEvent.PreviousStatus != commons.Active {
					if commons.RunningServices[commons.LndService].GetStatus(serviceEvent.NodeId) == commons.Active {
						if bootStrapping {
							serviceStatus = sendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Initializing, serviceStatus)
						}
						initializeChannelBalanceFromLnd(lndClient, serviceEvent.NodeId)
						bootStrapping = false
					}
				}
			}
		}
	}
}

func initializeChannelBalanceFromLnd(lndClient lnrpc.LightningClient, nodeId int) error {
	r, err := lndClient.ListChannels(context.Background(), &lnrpc.ListChannelsRequest{})
	if err != nil {
		return errors.Wrapf(err, "Obtaining channels from LND for nodeId: %v", nodeId)
	}
	for _, lndChannel := range r.Channels {
		fundingTransactionHash, fundingOutputIndex := channels.ParseChannelPoint(lndChannel.ChannelPoint)
		channelId := commons.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
		channelStateSettings := commons.ManagedChannelStateSettings{
			NodeId: nodeId,
			ChannelId: channelId,
			LocalBalance: lndChannel.LocalBalance,
			LocalDisabled: lndChannel.,
			LocalFeeBaseMsat: lndChannel.,
			LocalFeeRateMilliMsat: lndChannel.,
			LocalMinHtlc: lndChannel.,
			LocalMaxHtlcMsat: lndChannel.,
			LocalTimeLockDelta: lndChannel.,
			RemoteBalance: lndChannel.RemoteBalance,
			RemoteDisabled: lndChannel.,
			RemoteFeeBaseMsat: lndChannel.,
			RemoteFeeRateMilliMsat: lndChannel.,
			RemoteMinHtlc: lndChannel.,
			RemoteMaxHtlcMsat: lndChannel.,
			RemoteTimeLockDelta: lndChannel.,
			// STALE INFORMATION ONLY OBTAINED VIA LND REGULAR CHECKINS SO NOT MAINTAINED
			CommitFee: lndChannel.CommitFee,
			CommitWeight: lndChannel.CommitWeight,
			FeePerKw: lndChannel.FeePerKw,
			NumUpdates: lndChannel.NumUpdates,
			ChanStatusFlags: lndChannel.ChanStatusFlags,
			LocalChanReserveSat: lndChannel.LocalChanReserveSat,
			RemoteChanReserveSat: lndChannel.RemoteChanReserveSat,
			CommitmentType: lndChannel.CommitmentType,
			Lifetime: lndChannel.Lifetime,
			TotalSatoshisSent: lndChannel.TotalSatoshisSent,
			TotalSatoshisReceived: lndChannel.TotalSatoshisReceived,
		}

		if len(lndChannel.PendingHtlcs) > 0 {
			pendingPaymentHTLCsCount:=0
			pendingPaymentHTLCsAmount:=int64(0)
			pendingInvoiceHTLCsCount:=0
			pendingInvoiceHTLCsAmount:=int64(0)
			pendingDecreasingForwardHTLCsCount:=0
			pendingDecreasingForwardHTLCsAmount:=int64(0)
			pendingIncreasingForwardHTLCsCount:=0
			pendingIncreasingForwardHTLCsAmount:=int64(0)
			for _, pendingHtlc := range lndChannel.PendingHtlcs {
				if pendingHtlc.ForwardingHtlcIndex == 0 {
					continue
				}
				pendingHtlc += pendingHtlc.Amount
			}
			channelStateSettings.PendingPaymentHTLCsCount = pendingPaymentHTLCsCount
			channelStateSettings.PendingPaymentHTLCsAmount = pendingPaymentHTLCsAmount
			channelStateSettings.PendingInvoiceHTLCsCount = pendingInvoiceHTLCsCount
			channelStateSettings.PendingInvoiceHTLCsAmount = pendingInvoiceHTLCsAmount
			channelStateSettings.PendingDecreasingForwardHTLCsCount = pendingDecreasingForwardHTLCsCount
			channelStateSettings.PendingDecreasingForwardHTLCsAmount = pendingDecreasingForwardHTLCsAmount
			channelStateSettings.PendingIncreasingForwardHTLCsCount = pendingIncreasingForwardHTLCsCount
			channelStateSettings.PendingIncreasingForwardHTLCsAmount = pendingIncreasingForwardHTLCsAmount
		}
		commons.SetChannelState(nodeId, channelId, channelStateSettings)
	}
}
