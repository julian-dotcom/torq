package on_chain_tx

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/proto/lnrpc/walletrpc"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

type AddressType int32

const (
	Unknown = AddressType(iota)
	P2WPKH  = 1
	P2WKH   = 2
	NP2WKH  = 3
	P2TR    = 4
)

type rpcClientNewAddress interface {
	NextAddr(ctx context.Context, in *walletrpc.AddrRequest, opts ...grpc.CallOption) (*walletrpc.AddrResponse, error)
}

// NEW ADDRESS
type NewAddressRequest struct {
	NodeId int         `json:"nodeId"`
	Type   AddressType `json:"type"`
	//The name of the account to generate a new address for. If empty, the default wallet account is used.
	Account string `json:"account"`
}

type NewAddressResponse struct {
	RequestId string            `json:"requestId"`
	Request   NewAddressRequest `json:"request"`
	Address   string            `json:"address"`
}

func NewAddress(db *sqlx.DB, newAddressRequest NewAddressRequest) (NewAddressResponse, error) {

	if newAddressRequest.NodeId == 0 {
		return NewAddressResponse{}, errors.New("Node id is missing")
	}

	connectionDetails, err := settings.GetConnectionDetailsById(db, newAddressRequest.NodeId)
	if err != nil {
		return NewAddressResponse{}, errors.Wrap(err, "Getting node connection details from the db")
	}
	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return NewAddressResponse{}, errors.Wrap(err, "Getting node connection details from the db")
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Error().Msgf("Error closing grpc connection: %v", err)
		}
	}(conn)

	client := walletrpc.NewWalletKitClient(conn)
	resp, err := newAddress(client, newAddressRequest)
	if err != nil {
		return NewAddressResponse{}, err
	}
	return resp, nil
}

func createLndAddressRequest(newAddressRequest NewAddressRequest) (r *walletrpc.AddrRequest, err error) {
	lndAddressRequest := &walletrpc.AddrRequest{}
	if newAddressRequest.Account != "" {
		lndAddressRequest.Account = newAddressRequest.Account
	}
	switch newAddressRequest.Type {
	case P2WPKH:
		lndAddressRequest.Type = walletrpc.AddressType_WITNESS_PUBKEY_HASH
	case P2WKH:
		lndAddressRequest.Type = walletrpc.AddressType_NESTED_WITNESS_PUBKEY_HASH
	case NP2WKH:
		lndAddressRequest.Type = walletrpc.AddressType_HYBRID_NESTED_WITNESS_PUBKEY_HASH
	case P2TR:
		lndAddressRequest.Type = walletrpc.AddressType_TAPROOT_PUBKEY
	default:
		return nil, errors.New("Unknown address type")
	}
	return lndAddressRequest, nil
}

func newAddress(client rpcClientNewAddress, newAddressRequest NewAddressRequest) (NewAddressResponse, error) {
	lndAddressRequest, err := createLndAddressRequest(newAddressRequest)
	if err != nil {
		return NewAddressResponse{}, errors.Wrap(err, "Creating lnd address request")
	}

	ctx := context.Background()
	lndResponse, err := client.NextAddr(ctx, lndAddressRequest)
	if err != nil {
		return NewAddressResponse{}, errors.Wrap(err, "Getting new address from lnd")
	}

	return NewAddressResponse{
		Request: newAddressRequest,
		Address: lndResponse.Addr,
	}, nil
}
