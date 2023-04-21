package lightning

import (
	"reflect"
	"testing"

	"github.com/lncapital/torq/proto/lnrpc"
)

func Test_prepareOpenRequest(t *testing.T) {
	var satpvb uint64 = 10
	var tgConf int32 = 12
	var pushSat int64 = 12
	var targetConf int32 = 12
	var private = true
	var minHtlcMsat uint64 = 12
	var remoteCsvDelay uint32 = 12
	var minConfs int32 = 12
	var spendUnconfirmed = true
	var closeAddress = "test"
	var pubKeyStr = "024bf894b017051472911cb3db5097a825e2fc9a5602c824ff7bbea2a625f40972"
	var pubKeyByte = []byte{2, 75, 248, 148, 176, 23, 5, 20, 114, 145, 28, 179, 219, 80, 151, 168, 37, 226, 252, 154, 86, 2, 200, 36, 255, 123, 190, 162, 166, 37, 244, 9, 114}

	tests := []struct {
		name    string
		input   OpenChannelRequest
		want    *lnrpc.OpenChannelRequest
		wantErr bool
	}{
		{"Node id not provided",
			OpenChannelRequest{
				NodePubKey:         pubKeyStr,
				LocalFundingAmount: 12,
				PushSat:            nil,
				SatPerVbyte:        &satpvb,
				TargetConf:         &tgConf,
			},
			&lnrpc.OpenChannelRequest{},
			true},
		{"Just mandatory params",
			OpenChannelRequest{
				NodeId:             1,
				NodePubKey:         pubKeyStr,
				LocalFundingAmount: 12,
			},
			&lnrpc.OpenChannelRequest{
				NodePubkey:         pubKeyByte,
				LocalFundingAmount: 12,
			},
			false},
		{"Both targetConf & satPerVbyte provided",
			OpenChannelRequest{
				NodePubKey:         pubKeyStr,
				LocalFundingAmount: 12,
				PushSat:            nil,
				SatPerVbyte:        &satpvb,
				TargetConf:         &tgConf,
			},
			&lnrpc.OpenChannelRequest{},
			true},
		{"Just mandatory params",
			OpenChannelRequest{
				NodeId:             1,
				NodePubKey:         pubKeyStr,
				LocalFundingAmount: 12,
			},
			&lnrpc.OpenChannelRequest{
				NodePubkey:         pubKeyByte,
				LocalFundingAmount: 12,
			},
			false},
		{"All params provided",
			OpenChannelRequest{
				NodeId:             1,
				NodePubKey:         pubKeyStr,
				LocalFundingAmount: 12,
				PushSat:            &pushSat,
				TargetConf:         &targetConf,
				Private:            &private,
				MinHtlcMsat:        &minHtlcMsat,
				RemoteCsvDelay:     &remoteCsvDelay,
				MinConfs:           &minConfs,
				SpendUnconfirmed:   &spendUnconfirmed,
				CloseAddress:       &closeAddress,
			},
			&lnrpc.OpenChannelRequest{
				NodePubkey:         pubKeyByte,
				LocalFundingAmount: 12,
				PushSat:            12,
				TargetConf:         int32(12),
				Private:            true,
				MinHtlcMsat:        12,
				RemoteCsvDelay:     uint32(12),
				MinConfs:           int32(12),
				SpendUnconfirmed:   true,
				CloseAddress:       "test",
			},
			false},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := prepareOpenRequest(test.input)

			if err != nil {
				if test.wantErr {
					return
				}
				t.Errorf("prepareOpenRequest error: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: newSendPaymentRequest()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}

		})
	}
}

func Test_translateChanPoint(t *testing.T) {
	want := "72c09cee568b5637d1c6730a83fb4576689fd8fec32a8b8a8bab9e25d2267cdc:0"
	t.Run("Translate channel point", func(t *testing.T) {
		txid := []byte{
			220, 124, 38, 210, 37, 158, 171, 139, 138, 139, 42, 195, 254, 216, 159, 104, 118, 69, 251,
			131, 10, 115, 198, 209, 55, 86, 139, 86, 238, 156, 192, 114,
		}
		got, err := translateChanPoint(txid, uint32(0))
		if err != nil {
			t.Errorf("translateChanPoint error: %v", err)
		}

		if got != want {
			t.Errorf("translateChanPoint()\nGot:\n%v\nWant:\n%v\n", got, want)
		}
	})
}
