package lnd

import (
	"reflect"
	"testing"

	"github.com/lncapital/torq/internal/lightning_helpers"
	"github.com/lncapital/torq/proto/lnrpc"
)

func Test_checkPrepareReq(t *testing.T) {
	var satpvb int64 = 12
	var tgConf int32 = 12
	var pushSat int64 = 12
	var private = true
	bobPKbyte := []byte{2, 190, 169, 250, 229, 164, 252, 104, 90, 205,
		95, 89, 4, 113, 105, 9, 71, 116, 213, 31, 173, 13, 47, 59, 70, 193, 190, 225, 220, 35, 166, 206, 45}
	davePKbyte := []byte{3, 0, 58, 60, 77, 240, 60, 90, 152, 5, 137, 98,
		106, 105, 201, 85, 18, 108, 130, 141, 81, 165, 143, 112, 14, 241, 198, 78, 3, 191, 48, 48, 176}

	tests := []struct {
		name    string
		input   lightning_helpers.BatchOpenChannelRequest
		want    *lnrpc.BatchOpenChannelRequest
		wantErr bool
	}{
		{
			"Node ID is missing",
			lightning_helpers.BatchOpenChannelRequest{
				Channels: []lightning_helpers.BatchOpenChannel{
					{NodePublicKey: "02bea9fae5a4fc685acd5f59047169094774d51fad0d2f3b46c1bee1dc23a6ce2d", LocalFundingAmount: 250000},
					{NodePublicKey: "03003a3c4df03c5a980589626a69c955126c828d51a58f700ef1c64e03bf3030b0", LocalFundingAmount: 250000},
				},
			},
			&lnrpc.BatchOpenChannelRequest{Channels: []*lnrpc.BatchOpenChannel{
				{NodePubkey: bobPKbyte, LocalFundingAmount: 250000},
				{NodePubkey: davePKbyte, LocalFundingAmount: 250000},
			}},
			true,
		},
		{
			"Channels array empty",
			lightning_helpers.BatchOpenChannelRequest{
				CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: 1},
				Channels:             []lightning_helpers.BatchOpenChannel{},
				TargetConf:           nil,
				SatPerVbyte:          nil,
			},
			&lnrpc.BatchOpenChannelRequest{},
			true,
		},
		{
			"Both satpervbyte and targetconf set",
			lightning_helpers.BatchOpenChannelRequest{
				CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: 1},
				Channels: []lightning_helpers.BatchOpenChannel{
					{NodePublicKey: "03003a3c4df03c5a980589626a69c955126c828d51a58f700ef1c64e03bf3030b0"},
				},
				TargetConf:  &tgConf,
				SatPerVbyte: &satpvb,
			},
			&lnrpc.BatchOpenChannelRequest{},
			true,
		},
		{
			"LocalFundingAmount 0",
			lightning_helpers.BatchOpenChannelRequest{
				CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: 1},
				Channels: []lightning_helpers.BatchOpenChannel{
					{NodePublicKey: "03003a3c4df03c5a980589626a69c955126c828d51a58f700ef1c64e03bf3030b0", LocalFundingAmount: 0},
				},
				TargetConf: &tgConf,
			},
			&lnrpc.BatchOpenChannelRequest{},
			true,
		},
		{
			"Only mandatory params",
			lightning_helpers.BatchOpenChannelRequest{
				CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: 1},
				Channels: []lightning_helpers.BatchOpenChannel{
					{NodePublicKey: "02bea9fae5a4fc685acd5f59047169094774d51fad0d2f3b46c1bee1dc23a6ce2d", LocalFundingAmount: 250000},
					{NodePublicKey: "03003a3c4df03c5a980589626a69c955126c828d51a58f700ef1c64e03bf3030b0", LocalFundingAmount: 250000},
				},
			},
			&lnrpc.BatchOpenChannelRequest{Channels: []*lnrpc.BatchOpenChannel{
				{NodePubkey: bobPKbyte, LocalFundingAmount: 250000},
				{NodePubkey: davePKbyte, LocalFundingAmount: 250000},
			}},
			false,
		},
		{
			"All optional params",
			lightning_helpers.BatchOpenChannelRequest{
				CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: 1},
				Channels: []lightning_helpers.BatchOpenChannel{
					{
						NodePublicKey:      "02bea9fae5a4fc685acd5f59047169094774d51fad0d2f3b46c1bee1dc23a6ce2d",
						LocalFundingAmount: 250000,
						PushSat:            &pushSat,
						Private:            &private,
					},
					{
						NodePublicKey:      "03003a3c4df03c5a980589626a69c955126c828d51a58f700ef1c64e03bf3030b0",
						LocalFundingAmount: 250000,
						PushSat:            &pushSat,
						Private:            &private,
					},
				},
				TargetConf: &tgConf,
			},
			&lnrpc.BatchOpenChannelRequest{
				Channels: []*lnrpc.BatchOpenChannel{
					{
						NodePubkey:         bobPKbyte,
						LocalFundingAmount: 250000,
						PushSat:            int64(12),
						Private:            true,
						MinHtlcMsat:        12,
					},
					{
						NodePubkey:         davePKbyte,
						LocalFundingAmount: 250000,
						PushSat:            int64(12),
						Private:            true,
						MinHtlcMsat:        12,
					},
				},
				TargetConf: int32(12),
			},
			false,
		},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := checkPrepareReq(test.input)

			if err != nil {
				if test.wantErr {
					return
				}
				t.Errorf("checkPrepareReq: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: checkPrepareReq()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}
}

func Test_processBocResponse(t *testing.T) {
	txid := []byte{5, 16, 112, 218, 117, 193, 7, 32, 150, 201, 187, 247, 116, 47,
		115, 136, 49, 172, 219, 164, 121, 189, 10, 158, 184, 57, 176, 187, 21, 210, 113, 102}
	test := struct {
		name  string
		req   lightning_helpers.BatchOpenChannelRequest
		input lnrpc.BatchOpenChannelResponse
		want  lightning_helpers.BatchOpenChannelResponse
	}{
		"Test response",
		lightning_helpers.BatchOpenChannelRequest{},
		lnrpc.BatchOpenChannelResponse{
			PendingChannels: []*lnrpc.PendingUpdate{
				{
					Txid:        txid,
					OutputIndex: 0,
				},
				{
					Txid:        txid,
					OutputIndex: 1,
				},
			},
		},
		lightning_helpers.BatchOpenChannelResponse{
			PendingChannelPoints: []string{
				"6671d215bbb039b89e0abd79a4dbac3188732f74f7bbc9962007c175da701005:0",
				"6671d215bbb039b89e0abd79a4dbac3188732f74f7bbc9962007c175da701005:1",
			},
		},
	}

	t.Run(test.name, func(t *testing.T) {
		got, err := processBatchOpenChannelResponse(&test.input)

		if err != nil {
			t.Errorf("checkPrepareReq: %v", err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("processBocResponse()\nGot:\n%v\nWant:\n%v\n", got, test.want)
		}
	})
}
