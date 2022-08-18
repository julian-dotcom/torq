package channels

import (
	"context"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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
	errs, ctx := errgroup.WithContext(ctx)

	//Send open channel request
	openChanRes, err := client.OpenChannel(ctx, &openChanReq)
	if err != nil {
		log.Error().Msgf("Err opening channel: %v", err)
		return "Err opening channel", err
	}

	respChan := make(chan string)

	errs.Go(func() error {
		err = receiveOpenResponse(openChanRes, ctx, respChan)
		if err != nil {
			log.Error().Msgf("Error receiving response")
			return err
		}
		return nil
	})

	r = <-respChan

	return r, errs.Wait()

}

//Get response for open channel request
func receiveOpenResponse(req lnrpc.Lightning_OpenChannelClient, ctx context.Context, respChan chan string) error {
	for {
		select {
		case <-ctx.Done():
			log.Error().Msgf("%v", ctx.Err())
			respChan <- "Ctx Err"
			return ctx.Err()
		default:
		}

		resp, err := req.Recv()
		if err == io.EOF {
			log.Info().Msgf("Open channel EOF")
			return nil
		}
		if err != nil {
			log.Error().Msgf("Err receive %v", err.Error())
			respChan <- "Error"
			return err
		}
		//log.Debug().Msgf("Chan point pending: %v", resp.GetChanOpen().GetChannelPoint().GetFundingTxidStr())
		if resp.GetChanOpen().GetChannelPoint() != nil {
			//log.Debug().Msgf("Pending chan id: %v", resp.GetPendingChanId())
			fundingTxId := resp.GetChanOpen().GetChannelPoint().GetFundingTxidBytes()
			outputIndex := resp.GetChanOpen().GetChannelPoint().GetOutputIndex()
			channelPoint, err := translateChanPoint(fundingTxId, outputIndex)
			if err != nil {
				return err
			}
			//log.Debug().Msgf("Channel point: %v", channelPoint)
			respChan <- channelPoint
			return nil
		}
		//log.Debug().Msgf("Channel opening status: %v", resp.String())
	}
}

func translateChanPoint(cb []byte, oi uint32) (string, error) {
	ch, err := chainhash.NewHash(cb)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", ch.String(), oi), nil
}
