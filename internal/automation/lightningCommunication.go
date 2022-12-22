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

func LightningCommunicationService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
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
			var response interface{}
			var responseChannel chan interface{}
			if updateRequest, ok := event.(commons.ChannelStatusUpdateRequest); ok {
				if updateRequest.NodeId != nodeSettings.NodeId {
					return
				}
				if updateRequest.ResponseChannel != nil {
					responseChannel = updateRequest.ResponseChannel
				}
				if updateRequest.ChannelId == 0 {
					response = &commons.ChannelStatusUpdateResponse{
						Request: updateRequest,
						CommunicationResponse: commons.CommunicationResponse{
							Status: commons.Inactive,
							Error:  "ChannelId is 0",
						},
					}
				}
				if updateRequest.ChannelStatus != commons.Active &&
					updateRequest.ChannelStatus != commons.Inactive {
					response = &commons.ChannelStatusUpdateResponse{
						Request: updateRequest,
						CommunicationResponse: commons.CommunicationResponse{
							Status: commons.Inactive,
							Error:  "ChannelStatus is not Active nor Inactive",
						},
					}
				}
				if response == nil {
					hasChanges := false
					action := routerrpc.ChanStatusAction_ENABLE
					channelState := commons.GetChannelState(updateRequest.NodeId, updateRequest.ChannelId, true)
					if updateRequest.ChannelStatus == commons.Active && channelState.LocalDisabled {
						hasChanges = true
					}
					if updateRequest.ChannelStatus == commons.Inactive && !channelState.LocalDisabled {
						action = routerrpc.ChanStatusAction_DISABLE
						hasChanges = true
					}
					if hasChanges {
						channelSettings := commons.GetChannelSettingByChannelId(updateRequest.ChannelId)
						_, err := router.UpdateChanStatus(ctx, &routerrpc.UpdateChanStatusRequest{
							ChanPoint: &lnrpc.ChannelPoint{
								FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: channelSettings.FundingTransactionHash},
								OutputIndex: uint32(channelSettings.FundingOutputIndex)},
							Action: action,
						})
						if err != nil {
							log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v", updateRequest.ChannelId, updateRequest.NodeId)
						}
						if err == nil {
							response = &commons.ChannelStatusUpdateResponse{
								Request: updateRequest,
								CommunicationResponse: commons.CommunicationResponse{
									Status: commons.Active,
								},
							}
						} else {
							response = &commons.ChannelStatusUpdateResponse{
								Request: updateRequest,
								CommunicationResponse: commons.CommunicationResponse{
									Status: commons.Inactive,
									Error:  err.Error(),
								},
							}
						}
					} else {
						response = &commons.ChannelStatusUpdateResponse{
							Request: updateRequest,
							CommunicationResponse: commons.CommunicationResponse{
								Status:  commons.Active,
								Message: "Nothing changed so update is ignored",
							},
						}
					}
				}
			}
			if updateRequest, ok := event.(commons.RoutingPolicyUpdateRequest); ok {
				if updateRequest.NodeId != nodeSettings.NodeId {
					return
				}
				if updateRequest.ResponseChannel != nil {
					responseChannel = updateRequest.ResponseChannel
				}
				if updateRequest.FeeRateMilliMsat == nil &&
					updateRequest.FeeBaseMsat == nil &&
					updateRequest.MaxHtlcMsat == nil &&
					updateRequest.MinHtlcMsat == nil &&
					updateRequest.TimeLockDelta == nil {
					response = &commons.RoutingPolicyUpdateResponse{
						Request: updateRequest,
						CommunicationResponse: commons.CommunicationResponse{
							Status:  commons.Active,
							Message: "Nothing changed so update is ignored",
						},
					}
				}
				if updateRequest.ChannelId == nil && (updateRequest.TimeLockDelta == nil || *updateRequest.TimeLockDelta == 0) {
					response = &commons.RoutingPolicyUpdateResponse{
						Request: updateRequest,
						CommunicationResponse: commons.CommunicationResponse{
							Status: commons.Inactive,
							Error:  "TimeLockDelta cannot be empty unless channelId is also provided",
						},
					}
				}
				if response == nil {
					hasChanges := false
					policyUpdateRequest := &lnrpc.PolicyUpdateRequest{}
					if updateRequest.ChannelId == nil {
						hasChanges = true
						if *updateRequest.TimeLockDelta < 18 {
							policyUpdateRequest.TimeLockDelta = 18
						} else {
							policyUpdateRequest.TimeLockDelta = *updateRequest.TimeLockDelta
						}

						if updateRequest.FeeRateMilliMsat != nil {
							policyUpdateRequest.FeeRatePpm = uint32(*updateRequest.FeeRateMilliMsat)
						}
						if updateRequest.FeeBaseMsat != nil {
							policyUpdateRequest.BaseFeeMsat = int64(*updateRequest.FeeBaseMsat)
						}

						if updateRequest.MinHtlcMsat != nil {
							policyUpdateRequest.MinHtlcMsat = *updateRequest.MinHtlcMsat
							policyUpdateRequest.MinHtlcMsatSpecified = true
						}
						if updateRequest.MaxHtlcMsat != nil {
							policyUpdateRequest.MaxHtlcMsat = *updateRequest.MaxHtlcMsat
						}
					} else {
						channelState := commons.GetChannelState(updateRequest.NodeId, *updateRequest.ChannelId, true)
						if updateRequest.TimeLockDelta != nil &&
							*updateRequest.TimeLockDelta != channelState.LocalTimeLockDelta {
							if *updateRequest.TimeLockDelta < 18 {
								if channelState.LocalTimeLockDelta != 18 {
									policyUpdateRequest.TimeLockDelta = 18
									hasChanges = true
								}
							} else {
								policyUpdateRequest.TimeLockDelta = *updateRequest.TimeLockDelta
								hasChanges = true
							}
						} else {
							policyUpdateRequest.TimeLockDelta = *updateRequest.TimeLockDelta
						}
						if updateRequest.FeeRateMilliMsat != nil &&
							*updateRequest.FeeRateMilliMsat != channelState.LocalFeeRateMilliMsat {
							policyUpdateRequest.FeeRatePpm = uint32(*updateRequest.FeeRateMilliMsat)
							hasChanges = true
						}
						if !hasChanges && updateRequest.FeeBaseMsat != nil &&
							*updateRequest.FeeBaseMsat != channelState.LocalFeeBaseMsat {
							policyUpdateRequest.BaseFeeMsat = int64(*updateRequest.FeeBaseMsat)
							hasChanges = true
						}
						if !hasChanges && updateRequest.MinHtlcMsat != nil &&
							*updateRequest.MinHtlcMsat != channelState.LocalMinHtlcMsat {
							policyUpdateRequest.MinHtlcMsat = *updateRequest.MinHtlcMsat
							policyUpdateRequest.MinHtlcMsatSpecified = true
							hasChanges = true
						}
						if !hasChanges && updateRequest.MaxHtlcMsat != nil &&
							*updateRequest.MaxHtlcMsat != channelState.LocalMinHtlcMsat {
							policyUpdateRequest.MaxHtlcMsat = *updateRequest.MaxHtlcMsat
							hasChanges = true
						}
					}

					if hasChanges {
						if updateRequest.ChannelId == nil {
							policyUpdateRequest.Scope = &lnrpc.PolicyUpdateRequest_Global{Global: true}
						} else {
							channelSettings := commons.GetChannelSettingByChannelId(*updateRequest.ChannelId)
							policyUpdateRequest.Scope = &lnrpc.PolicyUpdateRequest_ChanPoint{
								ChanPoint: &lnrpc.ChannelPoint{
									FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: channelSettings.FundingTransactionHash},
									OutputIndex: uint32(channelSettings.FundingOutputIndex)}}
						}
						resp, err := client.UpdateChannelPolicy(ctx, policyUpdateRequest)
						if err != nil {
							log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v", updateRequest.ChannelId, updateRequest.NodeId)
						}
						if len(resp.GetFailedUpdates()) > 0 {
							for _, failUpdate := range resp.GetFailedUpdates() {
								log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (%v)",
									updateRequest.ChannelId, updateRequest.NodeId, failUpdate.UpdateError)
							}
						}
						if err == nil {
							response = &commons.RoutingPolicyUpdateResponse{
								Request: updateRequest,
								CommunicationResponse: commons.CommunicationResponse{
									Status: commons.Active,
								},
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
							response = &commons.RoutingPolicyUpdateResponse{
								Request: updateRequest,
								CommunicationResponse: commons.CommunicationResponse{
									Status: commons.Inactive,
								},
								FailedUpdates: failedUpdSlice,
							}
						}
					} else {
						// Nothing changed so update is ignored
						response = &commons.RoutingPolicyUpdateResponse{
							Request: updateRequest,
							CommunicationResponse: commons.CommunicationResponse{
								Status: commons.Active,
							},
						}
					}
				}
			}

			if response != nil {
				if responseChannel != nil {
					responseChannel <- response
				}
				if eventChannel != nil {
					eventChannel <- response
				}
			}
		}
	}
}
