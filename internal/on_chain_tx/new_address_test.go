package on_chain_tx

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"testing"
)

type MockNewAddressLClnt struct {
}

func (m MockNewAddressLClnt) WalletBalance(ctx context.Context, in *lnrpc.WalletBalanceRequest, opts ...grpc.CallOption) (*lnrpc.WalletBalanceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ChannelBalance(ctx context.Context, in *lnrpc.ChannelBalanceRequest, opts ...grpc.CallOption) (*lnrpc.ChannelBalanceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) GetTransactions(ctx context.Context, in *lnrpc.GetTransactionsRequest, opts ...grpc.CallOption) (*lnrpc.TransactionDetails, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) EstimateFee(ctx context.Context, in *lnrpc.EstimateFeeRequest, opts ...grpc.CallOption) (*lnrpc.EstimateFeeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SendCoins(ctx context.Context, in *lnrpc.SendCoinsRequest, opts ...grpc.CallOption) (*lnrpc.SendCoinsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ListUnspent(ctx context.Context, in *lnrpc.ListUnspentRequest, opts ...grpc.CallOption) (*lnrpc.ListUnspentResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SubscribeTransactions(ctx context.Context, in *lnrpc.GetTransactionsRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeTransactionsClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SendMany(ctx context.Context, in *lnrpc.SendManyRequest, opts ...grpc.CallOption) (*lnrpc.SendManyResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SignMessage(ctx context.Context, in *lnrpc.SignMessageRequest, opts ...grpc.CallOption) (*lnrpc.SignMessageResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) VerifyMessage(ctx context.Context, in *lnrpc.VerifyMessageRequest, opts ...grpc.CallOption) (*lnrpc.VerifyMessageResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ConnectPeer(ctx context.Context, in *lnrpc.ConnectPeerRequest, opts ...grpc.CallOption) (*lnrpc.ConnectPeerResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) DisconnectPeer(ctx context.Context, in *lnrpc.DisconnectPeerRequest, opts ...grpc.CallOption) (*lnrpc.DisconnectPeerResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ListPeers(ctx context.Context, in *lnrpc.ListPeersRequest, opts ...grpc.CallOption) (*lnrpc.ListPeersResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SubscribePeerEvents(ctx context.Context, in *lnrpc.PeerEventSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribePeerEventsClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) GetInfo(ctx context.Context, in *lnrpc.GetInfoRequest, opts ...grpc.CallOption) (*lnrpc.GetInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) GetRecoveryInfo(ctx context.Context, in *lnrpc.GetRecoveryInfoRequest, opts ...grpc.CallOption) (*lnrpc.GetRecoveryInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) PendingChannels(ctx context.Context, in *lnrpc.PendingChannelsRequest, opts ...grpc.CallOption) (*lnrpc.PendingChannelsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ListChannels(ctx context.Context, in *lnrpc.ListChannelsRequest, opts ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SubscribeChannelEvents(ctx context.Context, in *lnrpc.ChannelEventSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelEventsClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ClosedChannels(ctx context.Context, in *lnrpc.ClosedChannelsRequest, opts ...grpc.CallOption) (*lnrpc.ClosedChannelsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) OpenChannelSync(ctx context.Context, in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (*lnrpc.ChannelPoint, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) OpenChannel(ctx context.Context, in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (lnrpc.Lightning_OpenChannelClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) BatchOpenChannel(ctx context.Context, in *lnrpc.BatchOpenChannelRequest, opts ...grpc.CallOption) (*lnrpc.BatchOpenChannelResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) FundingStateStep(ctx context.Context, in *lnrpc.FundingTransitionMsg, opts ...grpc.CallOption) (*lnrpc.FundingStateStepResp, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ChannelAcceptor(ctx context.Context, opts ...grpc.CallOption) (lnrpc.Lightning_ChannelAcceptorClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) CloseChannel(ctx context.Context, in *lnrpc.CloseChannelRequest, opts ...grpc.CallOption) (lnrpc.Lightning_CloseChannelClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) AbandonChannel(ctx context.Context, in *lnrpc.AbandonChannelRequest, opts ...grpc.CallOption) (*lnrpc.AbandonChannelResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SendPayment(ctx context.Context, opts ...grpc.CallOption) (lnrpc.Lightning_SendPaymentClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SendPaymentSync(ctx context.Context, in *lnrpc.SendRequest, opts ...grpc.CallOption) (*lnrpc.SendResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SendToRoute(ctx context.Context, opts ...grpc.CallOption) (lnrpc.Lightning_SendToRouteClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SendToRouteSync(ctx context.Context, in *lnrpc.SendToRouteRequest, opts ...grpc.CallOption) (*lnrpc.SendResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) AddInvoice(ctx context.Context, in *lnrpc.Invoice, opts ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ListInvoices(ctx context.Context, in *lnrpc.ListInvoiceRequest, opts ...grpc.CallOption) (*lnrpc.ListInvoiceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) LookupInvoice(ctx context.Context, in *lnrpc.PaymentHash, opts ...grpc.CallOption) (*lnrpc.Invoice, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SubscribeInvoices(ctx context.Context, in *lnrpc.InvoiceSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeInvoicesClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) DecodePayReq(ctx context.Context, in *lnrpc.PayReqString, opts ...grpc.CallOption) (*lnrpc.PayReq, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ListPayments(ctx context.Context, in *lnrpc.ListPaymentsRequest, opts ...grpc.CallOption) (*lnrpc.ListPaymentsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) DeletePayment(ctx context.Context, in *lnrpc.DeletePaymentRequest, opts ...grpc.CallOption) (*lnrpc.DeletePaymentResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) DeleteAllPayments(ctx context.Context, in *lnrpc.DeleteAllPaymentsRequest, opts ...grpc.CallOption) (*lnrpc.DeleteAllPaymentsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) DescribeGraph(ctx context.Context, in *lnrpc.ChannelGraphRequest, opts ...grpc.CallOption) (*lnrpc.ChannelGraph, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) GetNodeMetrics(ctx context.Context, in *lnrpc.NodeMetricsRequest, opts ...grpc.CallOption) (*lnrpc.NodeMetricsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) GetChanInfo(ctx context.Context, in *lnrpc.ChanInfoRequest, opts ...grpc.CallOption) (*lnrpc.ChannelEdge, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) GetNodeInfo(ctx context.Context, in *lnrpc.NodeInfoRequest, opts ...grpc.CallOption) (*lnrpc.NodeInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) QueryRoutes(ctx context.Context, in *lnrpc.QueryRoutesRequest, opts ...grpc.CallOption) (*lnrpc.QueryRoutesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) GetNetworkInfo(ctx context.Context, in *lnrpc.NetworkInfoRequest, opts ...grpc.CallOption) (*lnrpc.NetworkInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) StopDaemon(ctx context.Context, in *lnrpc.StopRequest, opts ...grpc.CallOption) (*lnrpc.StopResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SubscribeChannelGraph(ctx context.Context, in *lnrpc.GraphTopologySubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelGraphClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) DebugLevel(ctx context.Context, in *lnrpc.DebugLevelRequest, opts ...grpc.CallOption) (*lnrpc.DebugLevelResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) FeeReport(ctx context.Context, in *lnrpc.FeeReportRequest, opts ...grpc.CallOption) (*lnrpc.FeeReportResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) UpdateChannelPolicy(ctx context.Context, in *lnrpc.PolicyUpdateRequest, opts ...grpc.CallOption) (*lnrpc.PolicyUpdateResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ForwardingHistory(ctx context.Context, in *lnrpc.ForwardingHistoryRequest, opts ...grpc.CallOption) (*lnrpc.ForwardingHistoryResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ExportChannelBackup(ctx context.Context, in *lnrpc.ExportChannelBackupRequest, opts ...grpc.CallOption) (*lnrpc.ChannelBackup, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ExportAllChannelBackups(ctx context.Context, in *lnrpc.ChanBackupExportRequest, opts ...grpc.CallOption) (*lnrpc.ChanBackupSnapshot, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) VerifyChanBackup(ctx context.Context, in *lnrpc.ChanBackupSnapshot, opts ...grpc.CallOption) (*lnrpc.VerifyChanBackupResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) RestoreChannelBackups(ctx context.Context, in *lnrpc.RestoreChanBackupRequest, opts ...grpc.CallOption) (*lnrpc.RestoreBackupResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SubscribeChannelBackups(ctx context.Context, in *lnrpc.ChannelBackupSubscription, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelBackupsClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) BakeMacaroon(ctx context.Context, in *lnrpc.BakeMacaroonRequest, opts ...grpc.CallOption) (*lnrpc.BakeMacaroonResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ListMacaroonIDs(ctx context.Context, in *lnrpc.ListMacaroonIDsRequest, opts ...grpc.CallOption) (*lnrpc.ListMacaroonIDsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) DeleteMacaroonID(ctx context.Context, in *lnrpc.DeleteMacaroonIDRequest, opts ...grpc.CallOption) (*lnrpc.DeleteMacaroonIDResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) ListPermissions(ctx context.Context, in *lnrpc.ListPermissionsRequest, opts ...grpc.CallOption) (*lnrpc.ListPermissionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) CheckMacaroonPermissions(ctx context.Context, in *lnrpc.CheckMacPermRequest, opts ...grpc.CallOption) (*lnrpc.CheckMacPermResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) RegisterRPCMiddleware(ctx context.Context, opts ...grpc.CallOption) (lnrpc.Lightning_RegisterRPCMiddlewareClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SendCustomMessage(ctx context.Context, in *lnrpc.SendCustomMessageRequest, opts ...grpc.CallOption) (*lnrpc.SendCustomMessageResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) SubscribeCustomMessages(ctx context.Context, in *lnrpc.SubscribeCustomMessagesRequest, opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeCustomMessagesClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockNewAddressLClnt) NewAddress(ctx context.Context, in *lnrpc.NewAddressRequest, opts ...grpc.CallOption) (*lnrpc.NewAddressResponse, error) {
	resp := lnrpc.NewAddressResponse{Address: "test"}
	return &resp, nil
}

func TestNewInvoice(t *testing.T) {
	addressType := 1
	account := "test"

	client := MockNewAddressLClnt{}
	resp, err := newAddress(client, int32(addressType), account)
	if err != nil {
		t.Fail()
	}

	if resp != "test" {
		t.Fail()
	}

}
