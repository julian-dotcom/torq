package channels

import (
	"reflect"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/testutil"
)

func Test_processChannelPoint(t *testing.T) {
	tests := []struct {
		name          string
		txOpenHash    string
		txOutputIndex uint32
		want          *lnrpc.PolicyUpdateRequest_ChanPoint
		wantErr       bool
	}{
		{
			"Test correct chann point",
			"e43bf0d5f03e179c2d107a1a8e4303bca066e883f4dbc0d9394f0c5b0721c7ce",
			0,
			&lnrpc.PolicyUpdateRequest_ChanPoint{
				ChanPoint: &lnrpc.ChannelPoint{
					FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
						FundingTxidStr: "e43bf0d5f03e179c2d107a1a8e4303bca066e883f4dbc0d9394f0c5b0721c7ce",
					},
					OutputIndex: 0,
				},
			},
			false,
		},
		//{
		//	"Test wrong chann point format",
		//	"e43bf0d5f03e179c2d107a1a8e4303bca066e883f4dbc0d9394f0c5b0721c7ce0",
		//	0,
		//	&lnrpc.PolicyUpdateRequest_ChanPoint{},
		//	true,
		//},
		//{
		//	"Output Index not int",
		//	"e43bf0d5f03e179c2d107a1a8e4303bca066e883f4dbc0d9394f0c5b0721c7ce",
		//	1,
		//	&lnrpc.PolicyUpdateRequest_ChanPoint{},
		//	true,
		//},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := processChannelPoint(test.txOpenHash, test.txOutputIndex)
			if err != nil {
				if test.wantErr {
					return
				}
				t.Errorf("processChannelPoint error: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: processChannelPoint()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}
}

func Test_createPolicyRequest(t *testing.T) {

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, err := srv.NewTestDatabase(true)
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

	err = InitializeManagedChannelCache(db)
	if err != nil {
		cancel()
		log.Fatal().Err(err).Msgf("Problem initializing ManagedChannel cache: %v", err)
	}

	lndShortChannelId := uint64(9999)
	shortChannelId := commons.ConvertLNDShortChannelID(lndShortChannelId)
	channel, err := addChannel(db, Channel{
		ShortChannelID:         &shortChannelId,
		Status:                 commons.Open,
		Private:                false,
		Capacity:               10_000_000,
		FirstNodeId:            commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet),
		SecondNodeId:           commons.GetNodeIdByPublicKey(testutil.TestPublicKey2, commons.Bitcoin, commons.SigNet),
		LNDShortChannelID:      &lndShortChannelId,
		FundingOutputIndex:     FundingOutputIndex,
		FundingTransactionHash: FundingTransactionHash,
	})
	if err != nil {
		log.Fatal().Err(err).Msgf("Problem initializing channel: %v", err)
	}
	log.Info().Msgf("Created OPEN channel to be closed with channelId: %v", channel.ChannelID)

	var timeLockDelta0 uint32 = 0
	var timeLockDelta uint32 = 18
	var feeRateMilliMsat uint64 = 11
	var feeBaseMsat uint64 = 12
	var maxHtlcMsat uint64 = 13
	var minHtlc int64 = 14

	tests := []struct {
		name    string
		input   commons.UpdateChannelRequest
		want    *lnrpc.PolicyUpdateRequest
		wantErr bool
	}{
		{
			"Missing node Id",
			commons.UpdateChannelRequest{
				NodeId:        commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet),
				ChannelId:     &channel.ChannelID,
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
				TimeLockDelta: 18,
			},
			true,
		},
		{
			"Channel point not provided - update all",
			commons.UpdateChannelRequest{
				NodeId:        commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet),
				ChannelId:     &channel.ChannelID,
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
				TimeLockDelta: 18,
			},
			false,
		},
		{
			"Channel point provided - update one",
			commons.UpdateChannelRequest{
				NodeId:        commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet),
				ChannelId:     &channel.ChannelID,
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
				TimeLockDelta: timeLockDelta,
			},
			false,
		},
		{
			"TimeLockDelta < 18",
			commons.UpdateChannelRequest{
				NodeId:        commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet),
				ChannelId:     &channel.ChannelID,
				TimeLockDelta: &timeLockDelta0,
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
				TimeLockDelta: timeLockDelta,
			},
			false,
		},
		{
			"All params provided",
			commons.UpdateChannelRequest{
				NodeId:           commons.GetNodeIdByPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet),
				ChannelId:        &channel.ChannelID,
				FeeRateMilliMsat: &feeRateMilliMsat,
				FeeBaseMsat:      &feeBaseMsat,
				MaxHtlcMsat:      &maxHtlcMsat,
				MinHtlc:          &minHtlc,
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
				MinHtlcMsat:          uint64(14000),
				MinHtlcMsatSpecified: true,
			},
			false,
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := createPolicyRequest(test.input)
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
		req   commons.UpdateChannelRequest
		input *lnrpc.PolicyUpdateResponse
		want  commons.UpdateChannelResponse
	}{
		{
			"Update succeeded",
			commons.UpdateChannelRequest{},
			&lnrpc.PolicyUpdateResponse{FailedUpdates: []*lnrpc.FailedUpdate{}},
			commons.UpdateChannelResponse{
				Status:        commons.Active,
				FailedUpdates: noFailedUpdSlice,
			},
		},
		{
			"Update failed",
			commons.UpdateChannelRequest{},
			&lnrpc.PolicyUpdateResponse{FailedUpdates: []*lnrpc.FailedUpdate{{
				Outpoint: &lnrpc.OutPoint{
					TxidBytes:   txidByte,
					TxidStr:     "e43bf0d5f03e179c2d107a1a8e4303bca066e883f4dbc0d9394f0c5b0721c7ce",
					OutputIndex: 0,
				},
				Reason:      2,
				UpdateError: "not found",
			}}},
			commons.UpdateChannelResponse{
				Status:        commons.Inactive,
				FailedUpdates: failedUpdSlice,
			},
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := processUpdateResponse(test.req, test.input, nil)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: processUpdateResponse()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}
}
