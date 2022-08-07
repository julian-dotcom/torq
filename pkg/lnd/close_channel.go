package lnd

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"io"
)

type lndClientCloseChannel interface {
	CloseChannel(ctx context.Context, in *lnrpc.CloseChannelRequest, opts ...grpc.CallOption) (lnrpc.Lightning_CloseChannelClient, error)
}

func closeChannel(fundingTxid *lnrpc.ChannelPoint_FundingTxidStr, outputIndex uint32, client lndClientCloseChannel) (r Response, err error) {
	ctx := context.Background()
	channelPoint := lnrpc.ChannelPoint{
		FundingTxid: fundingTxid,
		OutputIndex: outputIndex,
	}

	//open channel request
	closeChanReq := lnrpc.CloseChannelRequest{
		ChannelPoint:    &channelPoint,
		Force:           false,
		TargetConf:      0,
		DeliveryAddress: "",
		SatPerVbyte:     0,
	}

	closeChanRes, err := client.CloseChannel(ctx, &closeChanReq)
	if err != nil {
		log.Error().Msgf("Err closing channel: %v", err)
	}

	go receiveCloseResponse(closeChanRes, ctx)

	r.Response = "Channel closing"
	return r, nil
}

func receiveCloseResponse(req lnrpc.Lightning_CloseChannelClient, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msgf("%v", ctx.Err())
			return
		default:
		}

		resp, err := req.Recv()
		if err == io.EOF {
			log.Debug().Msgf("Close channel EOF")
			return
		}
		if err != nil {
			log.Error().Msgf("Err receive %v", err.Error())
			return
		}
		_ = resp
		//log.Debug().Msgf("Channel closing status: %v", resp.String())

	}
}
