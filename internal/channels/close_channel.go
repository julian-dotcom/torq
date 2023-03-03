package channels

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/rs/zerolog/log"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

func CloseChannel(eventChannel chan<- interface{}, db *sqlx.DB, ccReq commons.CloseChannelRequest, requestId string) (err error) {
	connectionDetails, err := settings.GetConnectionDetailsById(db, ccReq.NodeId)
	if err != nil {
		return errors.New("Getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return errors.Wrap(err, "Connecting to LND")
	}
	defer conn.Close()
	client := lnrpc.NewLightningClient(conn)

	closeChanReq, err := prepareCloseRequest(ccReq)
	if err != nil {
		return errors.Wrap(err, "Preparing close request")
	}

	return closeChannelResp(client, closeChanReq, eventChannel, ccReq, requestId, db)
}

func prepareCloseRequest(ccReq commons.CloseChannelRequest) (r *lnrpc.CloseChannelRequest, err error) {

	if ccReq.NodeId == 0 {
		return &lnrpc.CloseChannelRequest{}, errors.New("Node id is missing")
	}

	if ccReq.SatPerVbyte != nil && ccReq.TargetConf != nil {
		return &lnrpc.CloseChannelRequest{}, errors.New("Cannot set both SatPerVbyte and TargetConf")
	}

	channelSettings := commons.GetChannelSettingByChannelId(ccReq.ChannelId)

	//Make the close channel request
	closeChanReq := &lnrpc.CloseChannelRequest{
		ChannelPoint: &lnrpc.ChannelPoint{
			FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
				FundingTxidStr: channelSettings.FundingTransactionHash,
			},
			OutputIndex: uint32(channelSettings.FundingOutputIndex),
		},
	}

	if ccReq.Force != nil {
		closeChanReq.Force = *ccReq.Force
	}

	if ccReq.TargetConf != nil {
		closeChanReq.TargetConf = *ccReq.TargetConf
	}

	if ccReq.SatPerVbyte != nil {
		closeChanReq.SatPerVbyte = *ccReq.SatPerVbyte
	}

	if ccReq.DeliveryAddress != nil {
		closeChanReq.DeliveryAddress = *ccReq.DeliveryAddress
	}

	return closeChanReq, nil
}

func closeChannelResp(client lnrpc.LightningClient, closeChanReq *lnrpc.CloseChannelRequest, eventChannel chan<- interface{},
	ccReq commons.CloseChannelRequest, requestId string, db *sqlx.DB) error {

	ctx := context.Background()
	closeChanRes, err := client.CloseChannel(ctx, closeChanReq)
	if err != nil {
		err = errors.Wrap(err, "Problem sending closing channel request to LND")
		log.Error().Err(err).Send()
		return err
	}

	for {
		select {
		case <-ctx.Done():
			//log.Debug().Msgf("%v", ctx.Err())
			return nil
		default:
		}

		resp, err := closeChanRes.Recv()
		if err == io.EOF {
			//log.Debug().Msgf("Close channel EOF")
			return nil
		}
		if err != nil {
			log.Error().Err(errors.Wrap(err, "LND close channel")).Send()
			// return unwrapped error as it will give more context to the user
			return err
		}

		r := processCloseResponse(resp, ccReq, requestId)

		pendingChannel, err := client.PendingChannels(ctx, &lnrpc.PendingChannelsRequest{})
		if err != nil {
			err = errors.Wrap(err, "Channel closed but problem getting pending channels")
			log.Error().Err(err).Send()
			return err
		}

		for _, closing := range pendingChannel.WaitingCloseChannels {
			stringOutputIndex := strconv.FormatUint(uint64(closeChanReq.ChannelPoint.GetOutputIndex()), 10)
			if closeChanReq.ChannelPoint.GetFundingTxidStr()+":"+stringOutputIndex == closing.Channel.ChannelPoint {
				r.ClosePendingChannelPoint.TxId = []byte(closing.ClosingTxid)
			}
		}

		if eventChannel != nil {
			eventChannel <- r
		}

		if r.Status == commons.CooperativeClosed {
			// channel has closed, exiting
			return nil
		}

		err = updateChannelToClosingByChannelId(db, r.Request.ChannelId, closing.ClosingTxid)
		if err != nil {
			return errors.Wrap(err, "Updating channel to closing status in the db")
		}
	}
}

func convertChannelPoint(chanPointStr string) (chanPoint *lnrpc.ChannelPoint, err error) {
	splitChanPoint := strings.Split(chanPointStr, ":")
	if len(splitChanPoint) != 2 {
		return chanPoint, errors.New("Channel point missing a colon")
	}

	fundingTxid := &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: splitChanPoint[0]}

	oIndxUint, err := strconv.ParseUint(splitChanPoint[1], 10, 1)
	if err != nil {
		return chanPoint, errors.New("Parsing channel point output index")
	}

	outputIndex := uint32(oIndxUint)
	chanPoint = &lnrpc.ChannelPoint{
		FundingTxid: fundingTxid,
		OutputIndex: outputIndex,
	}

	return chanPoint, nil
}

func processCloseResponse(resp *lnrpc.CloseStatusUpdate, req commons.CloseChannelRequest, requestId string) commons.CloseChannelResponse {
	switch resp.GetUpdate().(type) {
	case *lnrpc.CloseStatusUpdate_ClosePending:
		//log.Info().Msgf("Pending closing channel")
		ccPending := resp.GetClosePending()
		return commons.CloseChannelResponse{
			RequestId:                requestId,
			Request:                  req,
			Status:                   commons.Closing,
			ClosePendingChannelPoint: commons.ChannelPoint{TxId: ccPending.Txid, OutputIndex: ccPending.OutputIndex},
		}
	case *lnrpc.CloseStatusUpdate_ChanClose:
		//log.Info().Msgf("Channel closed/or not")
		ccClose := resp.GetChanClose()
		//if !ccClose.Success {
		//	log.Error().Msgf("Channel closing failed")
		//	return nil, errors.New("Channel failed to close")
		//}
		return commons.CloseChannelResponse{
			RequestId:          requestId,
			Request:            req,
			Status:             commons.CooperativeClosed,
			CloseChannelStatus: commons.CloseChannelStatus{ClosingTxId: ccClose.ClosingTxid, Success: ccClose.Success},
		}
	default:
	}

	return commons.CloseChannelResponse{}
}
