package automation

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

type RoutingPolicyHistory struct {
	ChannelId     int
	ExecutionTime time.Time
	RoutingPolicy lnrpc.RoutingPolicy
}

func RoutingPolicyService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer, eventChannel chan interface{}) {

	client := lnrpc.NewLightningClient(conn)
	router := routerrpc.NewRouterClient(conn)

	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	// TODO FIXME IMPLEMENT SAME CHANNEL POLICY UPDATE RATE LIMITER

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		listener := broadcaster.Subscribe()
		for event := range listener {
			select {
			case <-ctx.Done():
				broadcaster.CancelSubscription(listener)
				return
			default:
			}
			if updateEvent, ok := event.(commons.ChannelStatusUpdateRequest); ok {
				if updateEvent.NodeId != nodeSettings.NodeId {
					return
				}
				if updateEvent.ChannelId == 0 {
					return
				}
				if updateEvent.ChannelStatus != commons.Active &&
					updateEvent.ChannelStatus != commons.Inactive {
					return
				}
				channelState := commons.GetChannelState(updateEvent.NodeId, updateEvent.ChannelId, true)
				hasChanges := false
				action := routerrpc.ChanStatusAction_ENABLE
				if updateEvent.ChannelStatus == commons.Active && channelState.LocalDisabled {
					hasChanges = true
				}
				if updateEvent.ChannelStatus == commons.Inactive && !channelState.LocalDisabled {
					action = routerrpc.ChanStatusAction_DISABLE
					hasChanges = true
				}
				if hasChanges {
					channelSettings := commons.GetChannelSettingByChannelId(updateEvent.ChannelId)
					_, err := router.UpdateChanStatus(ctx, &routerrpc.UpdateChanStatusRequest{
						ChanPoint: &lnrpc.ChannelPoint{
							FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: channelSettings.FundingTransactionHash},
							OutputIndex: uint32(channelSettings.FundingOutputIndex)},
						Action: action,
					})
					if err != nil {
						log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v", updateEvent.ChannelId, updateEvent.NodeId)
					}
					if eventChannel != nil {
						if err == nil {
							eventChannel <- commons.ChannelStatusUpdateResponse{
								Request: updateEvent,
								Status:  commons.Active,
							}
						} else {
							eventChannel <- commons.ChannelStatusUpdateResponse{
								Request: updateEvent,
								Status:  commons.Inactive,
							}
						}
					}
				}
			}
			if updateEvent, ok := event.(commons.RoutingPolicyUpdateRequest); ok {
				if updateEvent.NodeId != nodeSettings.NodeId {
					return
				}
				if updateEvent.ChannelId == 0 {
					return
				}
				if updateEvent.FeeRateMilliMsat == nil &&
					updateEvent.FeeBaseMsat == nil &&
					updateEvent.MaxHtlcMsat == nil &&
					updateEvent.MinHtlcMsat == nil &&
					updateEvent.TimeLockDelta == nil {
					return
				}
				policyUpdateRequest := &lnrpc.PolicyUpdateRequest{}
				channelState := commons.GetChannelState(updateEvent.NodeId, updateEvent.ChannelId, true)
				hasChanges := false
				if updateEvent.TimeLockDelta != nil &&
					*updateEvent.TimeLockDelta != channelState.LocalTimeLockDelta {
					if *updateEvent.TimeLockDelta < 18 {
						if channelState.LocalTimeLockDelta != 18 {
							policyUpdateRequest.TimeLockDelta = 18
							hasChanges = true
						}
					} else {
						policyUpdateRequest.TimeLockDelta = *updateEvent.TimeLockDelta
						hasChanges = true
					}
				} else {
					policyUpdateRequest.TimeLockDelta = *updateEvent.TimeLockDelta
				}
				if updateEvent.FeeRateMilliMsat != nil &&
					*updateEvent.FeeRateMilliMsat != channelState.LocalFeeRateMilliMsat {
					policyUpdateRequest.FeeRatePpm = uint32(*updateEvent.FeeRateMilliMsat)
					hasChanges = true
				}
				if !hasChanges && updateEvent.FeeBaseMsat != nil &&
					*updateEvent.FeeBaseMsat != channelState.LocalFeeBaseMsat {
					policyUpdateRequest.BaseFeeMsat = int64(*updateEvent.FeeBaseMsat)
					hasChanges = true
				}
				if !hasChanges && updateEvent.MinHtlcMsat != nil &&
					*updateEvent.MinHtlcMsat != channelState.LocalMinHtlcMsat {
					policyUpdateRequest.MinHtlcMsat = *updateEvent.MinHtlcMsat
					policyUpdateRequest.MinHtlcMsatSpecified = true
					hasChanges = true
				}
				if !hasChanges && updateEvent.MaxHtlcMsat != nil &&
					*updateEvent.MaxHtlcMsat != channelState.LocalMinHtlcMsat {
					policyUpdateRequest.MaxHtlcMsat = *updateEvent.MaxHtlcMsat
					hasChanges = true
				}

				if hasChanges {
					channelSettings := commons.GetChannelSettingByChannelId(updateEvent.ChannelId)
					policyUpdateRequest.Scope = &lnrpc.PolicyUpdateRequest_ChanPoint{
						ChanPoint: &lnrpc.ChannelPoint{
							FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: channelSettings.FundingTransactionHash},
							OutputIndex: uint32(channelSettings.FundingOutputIndex)}}
					resp, err := client.UpdateChannelPolicy(ctx, policyUpdateRequest)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v", updateEvent.ChannelId, updateEvent.NodeId)
					}
					if len(resp.GetFailedUpdates()) > 0 {
						for _, failUpdate := range resp.GetFailedUpdates() {
							log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (%v)",
								updateEvent.ChannelId, updateEvent.NodeId, failUpdate.UpdateError)
						}
					}
					if eventChannel != nil {
						if err == nil {
							eventChannel <- commons.RoutingPolicyUpdateResponse{
								Request: updateEvent,
								Status:  commons.Active,
							}
						} else {
							var failedUpdSlice []commons.FailedRequest
							if len(resp.GetFailedUpdates()) > 0 {
								for _, failUpdate := range resp.GetFailedUpdates() {
									failedUpd := commons.FailedRequest{
										Reason: failUpdate.UpdateError,
										Error:  failUpdate.UpdateError,
									}
									failedUpdSlice = append(failedUpdSlice, failedUpd)
								}
							}
							eventChannel <- commons.RoutingPolicyUpdateResponse{
								Request:       updateEvent,
								Status:        commons.Inactive,
								FailedUpdates: failedUpdSlice,
							}
						}
					}
				}
			}
		}
	}
}
