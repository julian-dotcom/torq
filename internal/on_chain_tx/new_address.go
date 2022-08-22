package on_chain_tx

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type lndNewAddrClient interface {
	NewAddress(ctx context.Context, in *lnrpc.NewAddressRequest, opts ...grpc.CallOption) (*lnrpc.NewAddressResponse, error)
}

func newAddress(client lndNewAddrClient, addressType int32, account string) (r string, err error) {
	ctx := context.Background()
	lnAddressType := lnrpc.AddressType(addressType)
	newAddressReq := lnrpc.NewAddressRequest{
		Type:    lnAddressType,
		Account: account,
	}
	resp, err := client.NewAddress(ctx, &newAddressReq)

	if err != nil {
		log.Error().Msgf("Err creating new address: %v", err)
		return "Err creating new address", err
	}

	//log.Debug().Msgf("New address : %v", resp.Address)
	return resp.Address, nil
}
