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

func sendCoins(client lndClientSendCoins, address string, amount int64, targetConf int32, satPerVbyte *uint64) (r string, err error) {
	ctx := context.Background()

	sendCoinsReq := lnrpc.SendCoinsRequest{
		Addr:   address,
		Amount: amount,
	}

	log.Debug().Msgf("before spvb in req: %v", sendCoinsReq.SatPerVbyte)

	switch {
	case satPerVbyte != nil:
		sendCoinsReq.SatPerVbyte = *satPerVbyte
	case targetConf != 0:
		sendCoinsReq.TargetConf = targetConf
	default:
	}

	//log.Debug().Msgf("after spvb in req: %v", sendCoinsReq.SatPerVbyte)

	resp, err := client.SendCoins(ctx, &sendCoinsReq)
	if err != nil {
		log.Error().Msgf("Err sending payment: %v", err)
		return "Err sending payment", err
	}

	//log.Debug().Msgf("Invoice : %v", resp.PaymentRequest)
	return resp.Txid, nil

}
