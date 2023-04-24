package cln_connect

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

// Connect connects to CLN using gRPC.
func Connect(host string, certificate []byte, key []byte, caCertificate []byte) (*grpc.ClientConn, error) {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(io.Discard, os.Stderr, os.Stderr))

	clientCrt, err := tls.X509KeyPair(certificate, key)
	if err != nil {
		return nil, errors.New("CLN credentials: failed to create X509 KeyPair")
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCertificate)

	serverName := "localhost"
	if strings.Contains(host, "cln") {
		serverName = "cln"
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		ClientAuth:   tls.RequestClientCert,
		Certificates: []tls.Certificate{clientCrt},
		RootCAs:      certPool,
		ServerName:   serverName,
	}

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
