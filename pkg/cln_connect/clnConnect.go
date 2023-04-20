package cln_connect

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

// Connect connects to CLN using gRPC.
func Connect(host string, certificate []byte, key []byte) (*grpc.ClientConn, error) {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(io.Discard, os.Stderr, os.Stderr))

	clientCrt, err := tls.X509KeyPair(certificate, key)
	if err != nil {
		return nil, errors.New("CLN credentials: failed to create X509 KeyPair")
	}
	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = []tls.Certificate{clientCrt}
	tlsConfig.ClientAuth = tls.RequestClientCert
	tlsConfig.InsecureSkipVerify = true

	opts := []grpc.DialOption{
		grpc.WithReturnConnectionError(),
		grpc.FailOnNonTempDialError(true),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		// max size to 25mb
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(25 << (10 * 2))),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	conn, err := grpc.DialContext(ctx, host, opts...)
	if err != nil {
		return nil, errors.New("CLN dial up")
	}

	return conn, nil
}
