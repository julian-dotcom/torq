package lnd

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

func openChannel(pubkey []byte, amt int64, client lndClientOpenChannel) (r Response, err error) {

	//open channel request
	openChanReq := lnrpc.OpenChannelRequest{
		NodePubkey:         pubkey,
		LocalFundingAmount: amt,
	}

	ctx := context.Background()

	//Send open channel request
	openChanRes, err := client.OpenChannel(ctx, &openChanReq)
	if err != nil {
		log.Error().Msgf("Err opening channel: %v", err)
		r.Response = "Err opening channel"
		return r, err
	}
	go receiveOpenResponse(openChanRes, ctx)

	r.Response = "Channel opening"
	return r, nil
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
		//log.Debug().Msgf("Chan point pending: %v", resp.GetChanOpen().String())
		if resp.GetChanOpen() != nil {
			log.Info().Msgf("Chan point: %v", resp.GetChanOpen().GetChannelPoint().GetFundingTxidStr())
		}
		//log.Debug().Msgf("Channel opening status: %v", resp.String())
	}
}
