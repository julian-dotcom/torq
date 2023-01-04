package lnd

import (
	"reflect"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/testutil"
)

const FundingTransactionHash = "c946aad8ea807099f2f4eaf2f92821024c9d8a79afd465573e924dacddfa490c"
const FundingOutputIndex = 1

func Test_createPolicyRequest(t *testing.T) {

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, _, _, err := srv.NewTestDatabase(true)
	defer cancel()
	if err != nil {
		t.Fatal(err)
	}

	err = settings.InitializeManagedSettingsCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing ManagedSettings cache: %v", err)
	}

	err = settings.InitializeManagedNodeCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing ManagedNode cache: %v", err)
	}

	lndShortChannelId := uint64(9999)
	shortChannelId := commons.ConvertLNDShortChannelID(lndShortChannelId)
	nodeId := commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet)
	remoteNodeId := commons.GetNodeIdByPublicKey(testutil.TestPublicKey2, commons.Bitcoin, commons.SigNet)
	err = testutil.AddChannel(db, shortChannelId, lndShortChannelId, FundingTransactionHash, FundingOutputIndex, nodeId, remoteNodeId, cancel)
	if err != nil {
		log.Fatal().Err(err).Msgf("Problem initializing channel: %v", err)
	}

	err = channels.InitializeManagedChannelCache(db)
	if err != nil {
		cancel()
		log.Fatal().Err(err).Msgf("Problem initializing ManagedChannel cache: %v", err)
	}

	channelId := commons.GetChannelIdByFundingTransaction(FundingTransactionHash, FundingOutputIndex)
	log.Info().Msgf("Created OPEN channel to be closed with channelId: %v", channelId)

	var timeLockDelta uint32 = 18
	var feeRateMilliMsat uint64 = 11
	var feeBaseMsat uint64 = 12
	var maxHtlcMsat uint64 = 13
	var minHtlcMsat uint64 = 14

	tests := []struct {
		name    string
		input   commons.RoutingPolicyUpdateRequest
		want    *lnrpc.PolicyUpdateRequest
		wantErr bool
	}{
		{
			"Missing node Id",
			commons.RoutingPolicyUpdateRequest{
				CommunicationRequest: commons.CommunicationRequest{
					NodeId: nodeId,
				},
				ChannelId:     channelId,
				TimeLockDelta: &timeLockDelta,
			},
			&lnrpc.PolicyUpdateRequest{
				Scope: &lnrpc.PolicyUpdateRequest_ChanPoint{
					ChanPoint: &lnrpc.ChannelPoint{
						FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
							FundingTxidStr: FundingTransactionHash,
						},
						OutputIndex: FundingOutputIndex,
					},
				},
				TimeLockDelta:        18,
				BaseFeeMsat:          int64(feeBaseMsat),
				FeeRatePpm:           uint32(feeRateMilliMsat),
				MaxHtlcMsat:          maxHtlcMsat,
				MinHtlcMsat:          minHtlcMsat,
				MinHtlcMsatSpecified: true,
			},
			true,
		},
		{
			"Channel point not provided - update all",
			commons.RoutingPolicyUpdateRequest{
				CommunicationRequest: commons.CommunicationRequest{
					NodeId: nodeId,
				},
				ChannelId:     channelId,
				TimeLockDelta: &timeLockDelta,
			},
			&lnrpc.PolicyUpdateRequest{
				Scope: &lnrpc.PolicyUpdateRequest_ChanPoint{
					ChanPoint: &lnrpc.ChannelPoint{
						FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
							FundingTxidStr: FundingTransactionHash,
						},
						OutputIndex: FundingOutputIndex,
					},
				},
				TimeLockDelta:        18,
				BaseFeeMsat:          int64(feeBaseMsat),
				FeeRatePpm:           uint32(feeRateMilliMsat),
				MaxHtlcMsat:          maxHtlcMsat,
				MinHtlcMsat:          minHtlcMsat,
				MinHtlcMsatSpecified: true,
			},
			false,
		},
		{
			"Channel point provided - update one",
			commons.RoutingPolicyUpdateRequest{
				CommunicationRequest: commons.CommunicationRequest{
					NodeId: nodeId,
				},
				ChannelId:     channelId,
				TimeLockDelta: &timeLockDelta,
			},
			&lnrpc.PolicyUpdateRequest{
				Scope: &lnrpc.PolicyUpdateRequest_ChanPoint{
					ChanPoint: &lnrpc.ChannelPoint{
						FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
							FundingTxidStr: FundingTransactionHash,
						},
						OutputIndex: FundingOutputIndex,
					},
				},
				TimeLockDelta:        timeLockDelta,
				BaseFeeMsat:          int64(feeBaseMsat),
				FeeRatePpm:           uint32(feeRateMilliMsat),
				MaxHtlcMsat:          maxHtlcMsat,
				MinHtlcMsat:          minHtlcMsat,
				MinHtlcMsatSpecified: true,
			},
			false,
		},
		{
			"All params provided",
			commons.RoutingPolicyUpdateRequest{
				CommunicationRequest: commons.CommunicationRequest{
					NodeId: nodeId,
				},
				ChannelId:        channelId,
				FeeRateMilliMsat: &feeRateMilliMsat,
				FeeBaseMsat:      &feeBaseMsat,
				MaxHtlcMsat:      &maxHtlcMsat,
				MinHtlcMsat:      &minHtlcMsat,
				TimeLockDelta:    &timeLockDelta,
			},
			&lnrpc.PolicyUpdateRequest{
				Scope: &lnrpc.PolicyUpdateRequest_ChanPoint{
					ChanPoint: &lnrpc.ChannelPoint{
						FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
							FundingTxidStr: FundingTransactionHash,
						},
						OutputIndex: FundingOutputIndex,
					},
				},
				BaseFeeMsat:          int64(12),
				FeeRatePpm:           uint32(11),
				TimeLockDelta:        18,
				MaxHtlcMsat:          uint64(13),
				MinHtlcMsat:          uint64(14),
				MinHtlcMsatSpecified: true,
			},
			false,
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dummyManagedChannelStateSettings := commons.ManagedChannelStateSettings{
				NodeId:                nodeId,
				RemoteNodeId:          remoteNodeId,
				ChannelId:             channelId,
				LocalFeeRateMilliMsat: feeRateMilliMsat,
				LocalFeeBaseMsat:      feeBaseMsat,
				LocalMinHtlcMsat:      minHtlcMsat,
				LocalMaxHtlcMsat:      maxHtlcMsat,
				LocalTimeLockDelta:    timeLockDelta,
			}
			got := constructPolicyUpdateRequest(test.input, &dummyManagedChannelStateSettings)
			if err != nil {
				if test.wantErr {
					return
				}
				t.Errorf("createPolicyRequest error: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: createPolicyRequest()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}
}

func Test_validatePolicyRequest(t *testing.T) {
	var timeLockDelta0 uint32 = 0
	//var timeLockDelta uint32 = 18

	tests := []struct {
		name  string
		input commons.RoutingPolicyUpdateRequest
		want  *commons.RoutingPolicyUpdateResponse
	}{
		{
			"TimeLockDelta < 18",
			commons.RoutingPolicyUpdateRequest{
				CommunicationRequest: commons.CommunicationRequest{
					NodeId: 1,
				},
				ChannelId:     1,
				TimeLockDelta: &timeLockDelta0,
			},
			&commons.RoutingPolicyUpdateResponse{
				Request: commons.RoutingPolicyUpdateRequest{
					CommunicationRequest: commons.CommunicationRequest{
						NodeId: 1,
					},
					ChannelId:     1,
					TimeLockDelta: &timeLockDelta0,
				},
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Inactive,
					Error:  "TimeLockDelta is < 18",
				},
			},
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validateRoutingPolicyUpdateRequest(test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: createPolicyRequest()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}
}

func Test_processUpdateResponse(t *testing.T) {
	var noFailedUpdSlice []commons.FailedRequest
	txidByte := []byte{206, 199, 33, 7, 91, 12, 79, 57, 217, 192, 219, 244, 131, 232, 102, 160, 188,
		3, 67, 142, 26, 122, 16, 45, 156, 23, 62, 240, 213, 240, 59, 228}
	failedUpdSlice := []commons.FailedRequest{
		{
			Reason: "not found",
			Error:  "not found",
		},
	}

	tests := []struct {
		name  string
		req   commons.RoutingPolicyUpdateRequest
		input *lnrpc.PolicyUpdateResponse
		want  commons.RoutingPolicyUpdateResponse
	}{
		{
			"Update succeeded",
			commons.RoutingPolicyUpdateRequest{},
			&lnrpc.PolicyUpdateResponse{FailedUpdates: []*lnrpc.FailedUpdate{}},
			commons.RoutingPolicyUpdateResponse{
				Request: commons.RoutingPolicyUpdateRequest{},
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Active,
				},
				FailedUpdates: noFailedUpdSlice,
			},
		},
		{
			"Update failed",
			commons.RoutingPolicyUpdateRequest{},
			&lnrpc.PolicyUpdateResponse{FailedUpdates: []*lnrpc.FailedUpdate{{
				Outpoint: &lnrpc.OutPoint{
					TxidBytes:   txidByte,
					TxidStr:     "e43bf0d5f03e179c2d107a1a8e4303bca066e883f4dbc0d9394f0c5b0721c7ce",
					OutputIndex: 0,
				},
				Reason:      2,
				UpdateError: "not found",
			}}},
			commons.RoutingPolicyUpdateResponse{
				Request: commons.RoutingPolicyUpdateRequest{},
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Inactive,
				},
				FailedUpdates: failedUpdSlice,
			},
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := processRoutingPolicyUpdateResponse(test.req, test.input, nil)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: processUpdateResponse()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}
}
