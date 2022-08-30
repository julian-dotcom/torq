package payments

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"reflect"
	"testing"
	"time"
)

//{
//  "reqId": "something",
//  "type": "newPayment",
//  "status": "SUCCEEDED",
//  "hash": "4552e8bd1a8c5d0fe490c33a15a5a6946912d3c50fafc2106549f702965f6d8c",
//  "preimage": "fee347b7a00b3247b48312b0a16ad4ab46de2ba30bb61269caeff43c0798e87e",
//  "paymentRequest": "",
//  "amountMsat": 1000,
//  "creationDate": "2022-08-23T11:33:44.365328469+02:00",
//  "path": {
//    "AttemptId": 259230,
//    "Status": "SUCCEEDED",
//    "Route": {
//      "TotalTimeLock": 750762,
//      "Hops": [
//        {
//          "ChanId": "708152x2971x1",
//          "Expiry": 750762,
//          "AmtToForwardMsat": 1000,
//          "PubKey": "035e4ff418fc8b5554c5d9eea66396c227bd429a3251c8cbc711002ba215bfc226",
//          "MppRecord": {
//            "PaymentAddr": "51b48e61abb0e59ba790ef90b1d1bc8734a93d5db35906cb837f497a2db99e09",
//            "TotalAmtMsat": 1000
//          }
//        }
//      ],
//      "TotalAmtMsat": 1000
//    },
//    "AttemptTimeNs": "2022-08-23T11:33:44.465361286+02:00",
//    "ResolveTimeNs": "2022-08-23T11:33:46.766148987+02:00",
//    "Preimage": "fee347b7a00b3247b48312b0a16ad4ab46de2ba30bb61269caeff43c0798e87e"
//  }
//}
//
//
//{
//  "reqId": "something",
//  "type": "newPayment",
//  "status": "IN_FLIGHT",
//  "hash": "4552e8bd1a8c5d0fe490c33a15a5a6946912d3c50fafc2106549f702965f6d8c",
//  "preimage": "0000000000000000000000000000000000000000000000000000000000000000",
//  "paymentRequest": "",
//  "amountMsat": 1000,
//  "creationDate": "2022-08-23T11:33:44.365328469+02:00",
//  "path": {
//    "AttemptId": 259230,
//    "Status": "IN_FLIGHT",
//    "Route": {
//      "TotalTimeLock": 750762,
//      "Hops": [
//        {
//          "ChanId": "708152x2971x1",
//          "Expiry": 750762,
//          "AmtToForwardMsat": 1000,
//          "PubKey": "035e4ff418fc8b5554c5d9eea66396c227bd429a3251c8cbc711002ba215bfc226",
//          "MppRecord": {
//            "PaymentAddr": "51b48e61abb0e59ba790ef90b1d1bc8734a93d5db35906cb837f497a2db99e09",
//            "TotalAmtMsat": 1000
//          }
//        }
//      ],
//      "TotalAmtMsat": 1000
//    },
//    "AttemptTimeNs": "2022-08-23T11:33:44.465361286+02:00",
//    "ResolveTimeNs": "1970-01-01T01:00:00+01:00",
//    "Preimage": ""
//  }
//}

func Test_processResponse(t *testing.T) {
	type args struct {
		response []byte
	}
	tests := []struct {
		name    string
		reqId   string
		input   *lnrpc.Payment
		want    NewPaymentResponse
		wantErr bool
	}{{
		name:  "Successful payment",
		reqId: "Unique ID here",
		input: &lnrpc.Payment{
			PaymentHash:     "4552e8bd1a8c5d0fe490c33a15a5a6946912d3c50fafc2106549f702965f6d8c",
			PaymentPreimage: "fee347b7a00b3247b48312b0a16ad4ab46de2ba30bb61269caeff43c0798e87e",
			ValueSat:        12,
			ValueMsat:       12000,
			PaymentRequest:  "",
			Status:          lnrpc.Payment_SUCCEEDED,
			FeeSat:          0,
			FeeMsat:         100,
			CreationTimeNs:  time.Unix(1661252258, 0).UnixNano(),
			Htlcs: []*lnrpc.HTLCAttempt{
				{
					AttemptId: 1234,
					Status:    lnrpc.HTLCAttempt_SUCCEEDED,
					Route: &lnrpc.Route{
						TotalTimeLock: 10,
						Hops: []*lnrpc.Hop{
							{
								ChanId:           778621358427537409,
								Expiry:           0,
								AmtToForwardMsat: 0,
								FeeMsat:          0,
								PubKey:           "",
								MppRecord: &lnrpc.MPPRecord{
									PaymentAddr:  []byte{254, 227, 71, 183, 160, 11, 50, 71, 180, 131, 18, 176, 161, 106, 212, 171, 70, 222, 43, 163, 11, 182, 18, 105, 202, 239, 244, 60, 7, 152, 232, 126},
									TotalAmtMsat: 1200,
								},
							},
						},
						TotalFeesMsat: 100,
						TotalAmtMsat:  12000,
					},
					AttemptTimeNs: time.Unix(1661252259, 0).UnixNano(),
					ResolveTimeNs: time.Unix(1661252260, 0).UnixNano(),
					Failure:       nil,
					Preimage:      []byte{254, 227, 71, 183, 160, 11, 50, 71, 180, 131, 18, 176, 161, 106, 212, 171, 70, 222, 43, 163, 11, 182, 18, 105, 202, 239, 244, 60, 7, 152, 232, 126},
				},
			},
			PaymentIndex:  234,
			FailureReason: 0,
		},
		want: NewPaymentResponse{
			ReqId:          "Unique ID here",
			Type:           "newPayment",
			Status:         "SUCCEEDED",
			Hash:           "4552e8bd1a8c5d0fe490c33a15a5a6946912d3c50fafc2106549f702965f6d8c",
			Preimage:       "fee347b7a00b3247b48312b0a16ad4ab46de2ba30bb61269caeff43c0798e87e",
			PaymentRequest: "",
			AmountMsat:     12000,
			CreationDate:   time.Unix(1661252258, 0),
			Attempt: attempt{
				AttemptId: 1234,
				Status:    "SUCCEEDED",
				Route: route{
					TotalTimeLock: 10,
					Hops: []hops{
						{
							ChanId:           "708152x2971x1",
							Expiry:           0,
							AmtToForwardMsat: 0,
							PubKey:           "",
							MppRecord: MppRecord{
								PaymentAddr:  "fee347b7a00b3247b48312b0a16ad4ab46de2ba30bb61269caeff43c0798e87e",
								TotalAmtMsat: 1200,
							},
						},
					},
					TotalAmtMsat: 12000,
				},
				AttemptTimeNs: time.Unix(1661252259, 0),
				ResolveTimeNs: time.Unix(1661252260, 0),
				Preimage:      "fee347b7a00b3247b48312b0a16ad4ab46de2ba30bb61269caeff43c0798e87e",
				Failure:       failureDetails{},
			},
		},
	},
		{
			name:  "Failed payment",
			reqId: "Unique ID here",
			input: &lnrpc.Payment{
				PaymentHash:     "4552e8bd1a8c5d0fe490c33a15a5a6946912d3c50fafc2106549f702965f6d8c",
				PaymentPreimage: "00000",
				ValueSat:        12,
				ValueMsat:       12000,
				PaymentRequest:  "",
				Status:          lnrpc.Payment_FAILED,
				FeeSat:          0,
				FeeMsat:         100,
				CreationTimeNs:  time.Unix(1661252258, 0).UnixNano(),
				Htlcs: []*lnrpc.HTLCAttempt{
					{
						AttemptId: 12345,
						Status:    lnrpc.HTLCAttempt_FAILED,
						Route: &lnrpc.Route{
							TotalTimeLock: 10,
							Hops: []*lnrpc.Hop{
								{
									ChanId:           778621358427537409,
									Expiry:           0,
									AmtToForwardMsat: 0,
									FeeMsat:          0,
									PubKey:           "",
									MppRecord: &lnrpc.MPPRecord{
										PaymentAddr:  []byte{254, 227, 71, 183, 160, 11, 50, 71, 180, 131, 18, 176, 161, 106, 212, 171, 70, 222, 43, 163, 11, 182, 18, 105, 202, 239, 244, 60, 7, 152, 232, 126},
										TotalAmtMsat: 1200,
									},
								},
							},
							TotalFeesMsat: 100,
							TotalAmtMsat:  12000,
						},
						AttemptTimeNs: time.Unix(1661252259, 0).UnixNano(),
						ResolveTimeNs: 0,
						Failure: &lnrpc.Failure{
							Code:               lnrpc.Failure_INCORRECT_OR_UNKNOWN_PAYMENT_DETAILS,
							FailureSourceIndex: 1,
							Height:             11,
						},
						Preimage: []byte{254, 227, 71, 183, 160, 11, 50, 71, 180, 131, 18, 176, 161, 106, 212, 171, 70, 222, 43, 163, 11, 182, 18, 105, 202, 239, 244, 60, 7, 152, 232, 126},
					},
				},
				PaymentIndex:  234,
				FailureReason: 0,
			},
			want: NewPaymentResponse{
				ReqId:          "Unique ID here",
				Type:           "newPayment",
				Status:         "FAILED",
				Hash:           "4552e8bd1a8c5d0fe490c33a15a5a6946912d3c50fafc2106549f702965f6d8c",
				Preimage:       "00000",
				PaymentRequest: "",
				AmountMsat:     12000,
				CreationDate:   time.Unix(1661252258, 0),
				Attempt: attempt{
					AttemptId: 12345,
					Status:    "FAILED",
					Route: route{
						TotalTimeLock: 10,
						Hops: []hops{
							{
								ChanId:           "708152x2971x1",
								Expiry:           0,
								AmtToForwardMsat: 0,
								PubKey:           "",
								MppRecord: MppRecord{
									PaymentAddr:  "fee347b7a00b3247b48312b0a16ad4ab46de2ba30bb61269caeff43c0798e87e",
									TotalAmtMsat: 1200,
								},
							},
						},
						TotalAmtMsat: 12000,
					},
					AttemptTimeNs: time.Unix(1661252259, 0),
					ResolveTimeNs: time.Unix(0, 0),
					Preimage:      "fee347b7a00b3247b48312b0a16ad4ab46de2ba30bb61269caeff43c0798e87e",
					Failure: failureDetails{
						Reason:             "INCORRECT_OR_UNKNOWN_PAYMENT_DETAILS",
						FailureSourceIndex: 1,
						Height:             11,
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := processResponse(test.input, test.reqId)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: processResponse()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}

}

func Test_newSendPaymentRequest(t *testing.T) {
	var amount int64 = 2000
	var allowSelfPayment bool = true
	var feeLimitMsat int64 = 100
	tests := []struct {
		name  string
		reqId string
		input NewPaymentRequest
		want  routerrpc.SendPaymentRequest
	}{
		{
			name: "with allow self and fee limit",
			input: NewPaymentRequest{
				Invoice:          "abcd",
				TimeOutSecs:      3600,
				Dest:             nil,
				AmtMSat:          &amount,
				FeeLimitMsat:     &feeLimitMsat,
				AllowSelfPayment: &allowSelfPayment,
			},
			want: routerrpc.SendPaymentRequest{
				AmtMsat:          amount,
				PaymentRequest:   "abcd",
				TimeoutSeconds:   3600,
				CltvLimit:        0,
				AllowSelfPayment: allowSelfPayment,
				FeeLimitMsat:     feeLimitMsat,
			},
		},
		{
			name: "with out any optional params",
			input: NewPaymentRequest{
				Invoice:     "bcde",
				TimeOutSecs: 3600,
			},
			want: routerrpc.SendPaymentRequest{
				PaymentRequest: "bcde",
				TimeoutSeconds: 3600,
			},
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := newSendPaymentRequest(test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: newSendPaymentRequest()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}
}
