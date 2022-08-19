package on_chain_tx

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type lndClientSendCoins interface {
	SendCoins(ctx context.Context, in *lnrpc.SendCoinsRequest, opts ...grpc.CallOption) (*lnrpc.SendCoinsResponse, error)
}

func sendCoins(client lndClientSendCoins, address string, amount int64, satPerVbyte *uint64) (r string, err error) {
	ctx := context.Background()

	sendCoinsReq := lnrpc.SendCoinsRequest{
		Addr:   address,
		Amount: amount,
	}

	if satPerVbyte != nil {
		sendCoinsReq.SatPerVbyte = *satPerVbyte
	}

	resp, err := client.SendCoins(ctx, &sendCoinsReq)
	if err != nil {
		log.Error().Msgf("Err sending payment: %v", err)
		return "Err sending payment", err
	}

	//log.Debug().Msgf("Invoice : %v", resp.PaymentRequest)
	return resp.Txid, nil

}
