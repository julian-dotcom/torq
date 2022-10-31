package on_chain_tx

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"google.golang.org/grpc"
)

const (
	Unknown int32 = 0
	P2WPKH  int32 = 1
	P2WKH   int32 = 2
	NP2WKH  int32 = 3
	P2TR    int32 = 4
)

type NewAddressRequest struct {
	LocalNodeId int   `json:"localNodeId"`
	Type        int32 `json:"type"`
	//The name of the account to generate a new address for. If empty, the default wallet account is used.
	Account string `json:"account"`
}

type NewAddressResponse struct {
	ReqId   string `json:"reqId"`
	Type    string `json:"type"`
	Address string `json:"address"`
}

type rpcClientNewAddress interface {
	NextAddr(ctx context.Context, in *walletrpc.AddrRequest, opts ...grpc.CallOption) (*walletrpc.AddrResponse, error)
}

func NewAddress(
	wChan chan interface{},
	db *sqlx.DB,
	context *gin.Context,
	newAddressRequest NewAddressRequest,
	reqId string,
) (err error) {

	if newAddressRequest.LocalNodeId == 0 {
		return errors.New("Node id is missing")
	}

	connectionDetails, err := settings.GetNodeConnectionDetailsById(db, newAddressRequest.LocalNodeId)
	if err != nil {
		return errors.Wrap(err, "Getting node connection details from the db")
	}
	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return errors.Wrap(err, "Getting node connection details from the db")
	}
	defer conn.Close()
	client := walletrpc.NewWalletKitClient(conn)
	return newAddress(client, newAddressRequest, wChan, reqId)
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
	}
	return lndAddressRequest, nil
}

func newAddress(client rpcClientNewAddress, newAddressRequest NewAddressRequest, wChan chan interface{}, reqId string) (err error) {
	// Create and validate payment request details
	lndAddressRequest, err := createLndAddressRequest(newAddressRequest)
	if err != nil {
		return err
	}

	ctx := context.Background()
	lndResponse, err := client.NextAddr(ctx, lndAddressRequest)
	if err != nil {
		return errors.Wrap(err, "New address")
	}

	wChan <- processResponse(lndResponse, reqId)
	return nil
}

func processResponse(lndResponse *walletrpc.AddrResponse, reqId string) (r NewAddressResponse) {
	r.ReqId = reqId
	r.Type = "newAddress"
	r.Address = lndResponse.GetAddr()
	return r
}
