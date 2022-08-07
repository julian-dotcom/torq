package lnd

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/logging"
	"github.com/rzajac/zltest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

var closeCounter int64
var closeCtxMain = context.Background()

type MockCloseChannelLC struct {
}

func (m MockCloseChannelLC) CloseChannel(ctx context.Context, in *lnrpc.CloseChannelRequest, opts ...grpc.CallOption) (lnrpc.Lightning_CloseChannelClient, error) {
	req := MockCloseChannelClientRecv{}
	return req, nil
}

type MockCloseChannelClientRecv struct {
	eof bool
	err bool
}

func (ml MockCloseChannelClientRecv) Recv() (*lnrpc.CloseStatusUpdate, error) {
	if ml.eof {
		return nil, io.EOF
	}

	if ml.err {
		return nil, errors.New("error")
	}
	atomic.AddInt64(&closeCounter, 1)

	return nil, nil
}

func (ml MockCloseChannelClientRecv) Header() (metadata.MD, error) {
	//TODO implement me
	panic("implement me")
}

func (ml MockCloseChannelClientRecv) Trailer() metadata.MD {
	//TODO implement me
	panic("implement me")
}

func (ml MockCloseChannelClientRecv) CloseSend() error {
	//TODO implement me
	panic("implement me")
}

func (ml MockCloseChannelClientRecv) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (ml MockCloseChannelClientRecv) SendMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (ml MockCloseChannelClientRecv) RecvMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}

func TestCloseChannel(t *testing.T) {
	fundingTxid := lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: "test"}
	outputIndex := uint32(1)
	client := MockCloseChannelLC{}
	resp, _ := closeChannel(client, &fundingTxid, outputIndex, nil)
	//fmt.Fprintf(os.Stderr, "%v\n", resp)
	if resp.Response != "Channel closing" {
		t.Fatalf("Failed")
	}
}

func TestCloseRecvCalled(t *testing.T) {
	closeCounter = 0

	recv := MockCloseChannelClientRecv{eof: false, err: false}

	go receiveCloseResponse(&recv, closeCtxMain)
	time.Sleep(100 * time.Millisecond)

	if closeCounter < 1 {
		t.Fatalf("Lightning_CloseChannelClient recv not called")
	}
}

func TestCloseRecvEOF(t *testing.T) {
	tst := zltest.New(t)
	logging.InitLogTest(tst)
	//ctx := context.Background()
	req := MockCloseChannelClientRecv{eof: true, err: false}

	go receiveCloseResponse(&req, closeCtxMain)
	time.Sleep(10 * time.Millisecond)

	ent := tst.LastEntry()
	ent.ExpStr("message", "Close channel EOF")
}

func TestCloseRecvErr(t *testing.T) {
	tst := zltest.New(t)
	logging.InitLogTest(tst)
	//ctx := context.Background()
	req := MockCloseChannelClientRecv{eof: false, err: true}

	go receiveCloseResponse(&req, closeCtxMain)
	time.Sleep(5 * time.Millisecond)

	ent := tst.LastEntry()
	ent.ExpStr("message", "Err receive error")
}

func TestCloseContextCanceled(t *testing.T) {
	tst := zltest.New(t)
	logging.InitLogTest(tst)
	ctx, cancel := context.WithCancel(closeCtxMain)

	req := MockCloseChannelClientRecv{eof: false, err: false}
	go receiveCloseResponse(&req, ctx)

	time.Sleep(10 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)

	ent := tst.LastEntry()
	ent.ExpStr("message", "context canceled")
}
