package payments

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lncapital/torq/internal/logging"
	"github.com/rzajac/zltest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

var payCounter int64
var payCtxMain = context.Background()

type MockrrpcClientSendPayment struct {
}

func (m MockrrpcClientSendPayment) SendPayment(ctx context.Context, in *routerrpc.SendPaymentRequest, opts ...grpc.CallOption) (routerrpc.Router_SendPaymentClient, error) {
	req := MockSendPaymentClientRecv{}
	return req, nil
}

type MockSendPaymentClientRecv struct {
	eof bool
	err bool
}

func (ml MockSendPaymentClientRecv) Header() (metadata.MD, error) {
	//TODO implement me
	panic("implement me")
}

func (ml MockSendPaymentClientRecv) Trailer() metadata.MD {
	//TODO implement me
	panic("implement me")
}

func (ml MockSendPaymentClientRecv) CloseSend() error {
	//TODO implement me
	panic("implement me")
}

func (ml MockSendPaymentClientRecv) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (ml MockSendPaymentClientRecv) SendMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (ml MockSendPaymentClientRecv) RecvMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (ml MockSendPaymentClientRecv) Recv() (*routerrpc.PaymentStatus, error) {
	paymentState := routerrpc.PaymentStatus{
		State:    1,
		Preimage: nil,
		Htlcs:    nil,
	}
	if ml.eof {
		return nil, io.EOF
	}

	if ml.err {
		return nil, errors.New("error")
	}
	atomic.AddInt64(&payCounter, 1)

	return &paymentState, nil
}

func TestSendNewPayment(t *testing.T) {
	dest := []byte("test")
	paymentHash := []byte("test")
	invoice := "invoice"
	client := MockrrpcClientSendPayment{}
	resp, _ := SendNewPayment(dest, 25000, 0, paymentHash, invoice, 10, client)
	if resp != "Payment sending" {
		t.Fatalf("Failed")
	}
}

func TestSendRecvCalled(t *testing.T) {
	payCounter = 0

	recv := MockSendPaymentClientRecv{eof: false, err: false}

	go func() {
		err := receivePayResponse(&recv, payCtxMain)
		if err != nil {

		}
	}()
	time.Sleep(100 * time.Millisecond)

	if payCounter < 1 {
		t.Fatalf("SendPaymentClientRecv recv not called")
	}
}

func TestSendRecvEOF(t *testing.T) {
	tst := zltest.New(t)
	logging.InitLogTest(tst)

	req := MockSendPaymentClientRecv{eof: true, err: false}

	go func() {
		err := receivePayResponse(&req, payCtxMain)
		if err != nil {

		}
	}()
	time.Sleep(10 * time.Millisecond)

	ent := tst.LastEntry()
	ent.ExpStr("message", "New payment EOF")
}

func TestPayRecvErr(t *testing.T) {
	tst := zltest.New(t)
	logging.InitLogTest(tst)
	req := MockSendPaymentClientRecv{eof: false, err: true}

	go func() {
		err := receivePayResponse(&req, payCtxMain)
		if err != nil {

		}
	}()
	time.Sleep(5 * time.Millisecond)

	ent := tst.LastEntry()
	ent.ExpStr("message", "Err receive error")
}

func TestPayContextCanceled(t *testing.T) {
	tst := zltest.New(t)
	logging.InitLogTest(tst)
	ctx, cancel := context.WithCancel(payCtxMain)

	req := MockSendPaymentClientRecv{eof: false, err: false}
	go func() {
		err := receivePayResponse(&req, ctx)
		if err != nil {

		}
	}()

	cancel()
	time.Sleep(10 * time.Millisecond)

	ent := tst.LastEntry()
	ent.ExpStr("message", "context canceled")
}
