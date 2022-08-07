package lnd

import (
	"context"
	"github.com/btcsuite/btcd/btcec"
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

var counter int64
var ctxMain = context.Background()

type MockOpenChannelLC struct {
}

func (m MockOpenChannelLC) OpenChannel(ctx context.Context, in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (lnrpc.Lightning_OpenChannelClient, error) {
	req := MockLOpenChannelClientRecv{}
	return req, nil
}

type MockLOpenChannelClientRecv struct {
	eof bool
	err bool
}

func (ml MockLOpenChannelClientRecv) Header() (metadata.MD, error) {
	//TODO implement me
	panic("implement me")
}

func (ml MockLOpenChannelClientRecv) Trailer() metadata.MD {
	//TODO implement me
	panic("implement me")
}

func (ml MockLOpenChannelClientRecv) CloseSend() error {
	//TODO implement me
	panic("implement me")
}

func (ml MockLOpenChannelClientRecv) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (ml MockLOpenChannelClientRecv) SendMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (ml MockLOpenChannelClientRecv) RecvMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (ml MockLOpenChannelClientRecv) Recv() (*lnrpc.OpenStatusUpdate, error) {

	if ml.eof {
		return nil, io.EOF
	}

	if ml.err {
		return nil, errors.New("error")
	}
	atomic.AddInt64(&counter, 1)

	return nil, nil
}

// randPubKey generates a new secp keypair, and returns the public key.
func randPubKey(t *testing.T) *btcec.PublicKey {
	t.Helper()

	sk, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Fatalf("unable to generate pubkey: %v", err)
	}

	return sk.PubKey()
}

func TestOpenChannel(t *testing.T) {
	testPubKey := randPubKey(t)
	testAmt := int64(1)
	client := MockOpenChannelLC{}
	resp, _ := openChannel(client, testPubKey.SerializeCompressed(), testAmt, nil)
	//fmt.Fprintf(os.Stderr, "%v\n", resp)
	if resp.Response != "Channel opening" {
		t.Fatalf("Failed")
	}
}

func TestOpenRecvCalled(t *testing.T) {
	counter = 0

	recv := MockLOpenChannelClientRecv{eof: false, err: false}

	go receiveOpenResponse(&recv, ctxMain)
	time.Sleep(5 * time.Millisecond)

	if counter < 1 {
		t.Fatalf("Lightning_OpenChannelClient recv not called")
	}
}

func TestOpenRecvEOF(t *testing.T) {
	tst := zltest.New(t)
	logging.InitLogTest(tst)
	//ctx := context.Background()
	req := MockLOpenChannelClientRecv{eof: true, err: false}

	go receiveOpenResponse(&req, ctxMain)
	time.Sleep(10 * time.Millisecond)

	ent := tst.LastEntry()
	ent.ExpStr("message", "Open channel EOF")
}

func TestOpenRecvErr(t *testing.T) {
	tst := zltest.New(t)
	logging.InitLogTest(tst)
	//ctx := context.Background()
	req := MockLOpenChannelClientRecv{eof: false, err: true}

	go receiveOpenResponse(&req, ctxMain)
	time.Sleep(5 * time.Millisecond)

	ent := tst.LastEntry()
	ent.ExpStr("message", "Err receive error")
}

func TestOpenContextCanceled(t *testing.T) {
	tst := zltest.New(t)
	logging.InitLogTest(tst)
	ctx, cancel := context.WithCancel(ctxMain)

	req := MockLOpenChannelClientRecv{eof: false, err: false}
	go receiveOpenResponse(&req, ctx)

	time.Sleep(10 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)

	ent := tst.LastEntry()
	ent.ExpStr("message", "context canceled")
}
