package channels

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"io"
)

type lndClientOpenChannel interface {
	OpenChannel(ctx context.Context, in *lnrpc.OpenChannelRequest, opts ...grpc.CallOption) (lnrpc.Lightning_OpenChannelClient, error)
}

func OpenChannel(client lndClientOpenChannel, pubkey []byte, amt int64, satPerVbyte *uint64) (r string, err error) {

	//open channel request
	openChanReq := lnrpc.OpenChannelRequest{
		NodePubkey:         pubkey,
		LocalFundingAmount: amt,
	}

	if satPerVbyte != nil {
		openChanReq.SatPerVbyte = *satPerVbyte
	}

	ctx := context.Background()

	//Send open channel request
	openChanRes, err := client.OpenChannel(ctx, &openChanReq)
	if err != nil {
		log.Error().Msgf("Err opening channel: %v", err)
		return "Err opening channel", err
	}
	go receiveOpenResponse(openChanRes, ctx)

	return "Channel opening", nil
}

//Get response for open channel request
func receiveOpenResponse(req lnrpc.Lightning_OpenChannelClient, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Error().Msgf("%v", ctx.Err())
			return
		default:
		}

		resp, err := req.Recv()
		if err == io.EOF {
			log.Info().Msgf("Open channel EOF")
			return
		}
		if err != nil {
			log.Error().Msgf("Err receive %v", err.Error())
			return
		}

		if resp.GetChanOpen() != nil {
			log.Info().Msgf("Chan point: %v", resp.GetChanOpen().GetChannelPoint().GetFundingTxidStr())
		}
	}
}
