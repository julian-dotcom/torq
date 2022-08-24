package channels

//import (
//	"context"
//	"github.com/btcsuite/btcd/btcec"
//	"github.com/cockroachdb/errors"
//	"github.com/lightningnetwork/lnd/lnrpc"
//	"github.com/lncapital/torq/internal/logging"
//	"github.com/rzajac/zltest"
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/metadata"
//	"io"
//	"sync/atomic"
//	"testing"
//	"time"
//)
//
//var counter int64
//var ctxMain = context.Background()
//
//type MockOpenChannelLC struct {
//}
//
//func (m MockOpenChannelLC) OpenChannel(ctx context.Context, in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (lnrpc.Lightning_OpenChannelClient, error) {
//	req := MockLOpenChannelClientRecv{}
//	return req, nil
//}
//
//type MockLOpenChannelClientRecv struct {
//	eof bool
//	err bool
//}
//
//func (ml MockLOpenChannelClientRecv) Header() (metadata.MD, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (ml MockLOpenChannelClientRecv) Trailer() metadata.MD {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (ml MockLOpenChannelClientRecv) CloseSend() error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (ml MockLOpenChannelClientRecv) Context() context.Context {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (ml MockLOpenChannelClientRecv) SendMsg(m interface{}) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (ml MockLOpenChannelClientRecv) RecvMsg(m interface{}) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (ml MockLOpenChannelClientRecv) Recv() (*lnrpc.OpenStatusUpdate, error) {
//	txid := []byte{
//		220, 124, 38, 210, 37, 158, 171, 139, 138, 139, 42, 195, 254, 216, 159, 104, 118, 69, 251,
//		131, 10, 115, 198, 209, 55, 86, 139, 86, 238, 156, 192, 114,
//	}
//	fundingTxid := lnrpc.ChannelPoint_FundingTxidBytes{FundingTxidBytes: txid}
//	channelPoint := lnrpc.ChannelPoint{
//		FundingTxid: &fundingTxid,
//		OutputIndex: 0,
//	}
//	chanOpenUpd := lnrpc.ChannelOpenUpdate{ChannelPoint: &channelPoint}
//	//
//
//	update := lnrpc.OpenStatusUpdate_ChanOpen{ChanOpen: &chanOpenUpd}
//	statusUpdate := lnrpc.OpenStatusUpdate{
//		Update:        &update,
//		PendingChanId: []byte("1"),
//	}
//
//	if ml.eof {
//		return nil, io.EOF
//	}
//
//	if ml.err {
//		return nil, errors.New("error")
//	}
//	atomic.AddInt64(&counter, 1)
//
//	return &statusUpdate, nil
//}
//
//// randPubKey generates a new secp keypair, and returns the public key.
//func randPubKey(t *testing.T) *btcec.PublicKey {
//	t.Helper()
//
//	sk, err := btcec.NewPrivateKey(btcec.S256())
//	if err != nil {
//		t.Fatalf("unable to generate pubkey: %v", err)
//	}
//
//	return sk.PubKey()
//}
//
//func TestOpenChannel(t *testing.T) {
//	testPubKey := randPubKey(t)
//	testAmt := int64(1)
//	client := MockOpenChannelLC{}
//	resp, _ := OpenChannel(client, testPubKey.SerializeCompressed(), testAmt, nil)
//
//	if resp != "72c09cee568b5637d1c6730a83fb4576689fd8fec32a8b8a8bab9e25d2267cdc:0" {
//		t.Fatalf("Failed")
//	}
//}
//
//func TestOpenRecvCalled(t *testing.T) {
//	counter = 0
//
//	recv := MockLOpenChannelClientRecv{eof: false, err: false}
//	respChan := make(chan string)
//	go func() {
//		err := receiveOpenResponse(&recv, ctxMain, respChan)
//		if err != nil {
//
//		}
//	}()
//	time.Sleep(5 * time.Millisecond)
//
//	if counter < 1 {
//		t.Fatalf("Lightning_OpenChannelClient recv not called")
//	}
//}
//
//func TestOpenRecvEOF(t *testing.T) {
//	tst := zltest.New(t)
//	logging.InitLogTest(tst)
//
//	req := MockLOpenChannelClientRecv{eof: true, err: false}
//
//	respChan := make(chan string)
//	go func() {
//		err := receiveOpenResponse(&req, ctxMain, respChan)
//		if err != nil {
//
//		}
//	}()
//	time.Sleep(10 * time.Millisecond)
//
//	ent := tst.LastEntry()
//	ent.ExpStr("message", "Open channel EOF")
//}
//
//func TestOpenRecvErr(t *testing.T) {
//	tst := zltest.New(t)
//	logging.InitLogTest(tst)
//
//	req := MockLOpenChannelClientRecv{eof: false, err: true}
//	respChan := make(chan string)
//	go func() {
//		err := receiveOpenResponse(&req, ctxMain, respChan)
//		if err != nil {
//
//		}
//	}()
//	time.Sleep(5 * time.Millisecond)
//
//	ent := tst.LastEntry()
//	ent.ExpStr("message", "Err receive error")
//}
//
//func TestOpenContextCanceled(t *testing.T) {
//	tst := zltest.New(t)
//	logging.InitLogTest(tst)
//	ctx, cancel := context.WithCancel(ctxMain)
//
//	req := MockLOpenChannelClientRecv{eof: false, err: false}
//	respChan := make(chan string)
//	go func() {
//		err := receiveOpenResponse(&req, ctx, respChan)
//		if err != nil {
//
//		}
//	}()
//
//	cancel()
//	time.Sleep(100 * time.Millisecond)
//
//	ent := tst.LastEntry()
//	ent.ExpStr("message", "context canceled")
//}
