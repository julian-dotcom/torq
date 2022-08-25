package channels

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"reflect"
	"testing"
)

func Test_processResponse(t *testing.T) {

	tests := []struct {
		name    string
		reqId   string
		input   *lnrpc.CloseStatusUpdate
		want    *CloseChannelResponse
		wantErr bool
	}{
		{
			name:  "Close Pending",
			reqId: "Test",
			input: &lnrpc.CloseStatusUpdate{
				Update: &lnrpc.CloseStatusUpdate_ClosePending{
					ClosePending: &lnrpc.PendingUpdate{
						Txid:        []byte("test"),
						OutputIndex: 0,
					},
				},
			},

			want: &CloseChannelResponse{
				ReqId:        "Test",
				Status:       "PENDING",
				ClosePending: pendingUpdate{[]byte("test"), 0},
				ChanClose:    channelCloseUpdate{},
			},
		},
		{
			name:  "Closed",
			reqId: "Test",
			input: &lnrpc.CloseStatusUpdate{
				Update: &lnrpc.CloseStatusUpdate_ChanClose{
					ChanClose: &lnrpc.ChannelCloseUpdate{
						ClosingTxid: []byte("test"),
						Success:     false,
					},
				},
			},
			want: &CloseChannelResponse{
				ReqId:        "Test",
				Status:       "CLOSED",
				ClosePending: pendingUpdate{},
				ChanClose:    channelCloseUpdate{ClosingTxId: []byte("test"), Success: false},
			},
		},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := processCloseResponse(test.input, test.reqId)
			if err != nil {
				t.Errorf("processCloseResponse error: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: processResponse()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}
}

func Test_convertChannelPoint(t *testing.T) {
	fundTxidStr := "c946aad8ea807099f2f4eaf2f92821024c9d8a79afd465573e924dacddfa490c"
	fundingTxid := &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: fundTxidStr}
	want := &lnrpc.ChannelPoint{
		FundingTxid: fundingTxid,
		OutputIndex: 1,
	}
	t.Run("converChanPoint", func(t *testing.T) {
		chanPointStr := "c946aad8ea807099f2f4eaf2f92821024c9d8a79afd465573e924dacddfa490c:1"
		got, err := convertChannelPoint(chanPointStr)
		if err != nil {
			t.Errorf("convertChannelPoint error: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("converChannelPoint()\nGot:\n%v\nWant:\n%v\n", got, want)
		}
	})
}
