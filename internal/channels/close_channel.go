package channels

import (
	"context"
	"io"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

const closeChannelTimeoutInSeconds = 60

type CloseChannelRequest struct {
	NodeId          int     `json:"nodeId"`
	ChannelId       int     `json:"channelId"`
	Force           *bool   `json:"force"`
	TargetConf      *int32  `json:"targetConf"`
	DeliveryAddress *string `json:"deliveryAddress"`
	SatPerVbyte     *uint64 `json:"satPerVbyte"`
}

type CloseChannelResponse struct {
	Request                CloseChannelRequest `json:"request"`
	Status                 core.ChannelStatus  `json:"status"`
	ClosingTransactionHash string              `json:"closingTransactionHash"`
}

func CloseChannel(db *sqlx.DB, req CloseChannelRequest) (response CloseChannelResponse, err error) {
	connectionDetails, err := settings.GetConnectionDetailsById(db, req.NodeId)
	if err != nil {
		return CloseChannelResponse{}, errors.New("Getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return CloseChannelResponse{}, errors.Wrap(err, "Connecting to LND")
	}
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("failed to close connection to lnd")
		}
	}()
	client := lnrpc.NewLightningClient(conn)

	closeChanReq, err := prepareCloseRequest(req)
	if err != nil {
		return CloseChannelResponse{}, errors.Wrap(err, "Preparing close request")
	}

	return closeChannelResp(db, client, closeChanReq, req)
}

func prepareCloseRequest(ccReq CloseChannelRequest) (r *lnrpc.CloseChannelRequest, err error) {

	if ccReq.NodeId == 0 {
		return &lnrpc.CloseChannelRequest{}, errors.New("Node id is missing")
	}

	if ccReq.SatPerVbyte != nil && ccReq.TargetConf != nil {
		return &lnrpc.CloseChannelRequest{}, errors.New("Cannot set both SatPerVbyte and TargetConf")
	}

	channelSettings := cache.GetChannelSettingByChannelId(ccReq.ChannelId)

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

func closeChannelResp(db *sqlx.DB,
	client lnrpc.LightningClient,
	closeChanReq *lnrpc.CloseChannelRequest,
	ccReq CloseChannelRequest) (CloseChannelResponse, error) {

	// Create a context with a timeout.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), closeChannelTimeoutInSeconds*time.Second)
	defer cancel()

	// Call CloseChannel with the timeout context.
	closeChanRes, err := client.CloseChannel(timeoutCtx, closeChanReq)
	if err != nil {
		err = errors.Wrap(err, "Problem sending closing channel request to LND")
		log.Error().Err(err).Send()
		return CloseChannelResponse{}, err
	}

	// Loop until we receive a close channel response or the context times out.
	for {
		select {
		case <-timeoutCtx.Done():
			return CloseChannelResponse{}, errors.New("Close channel request timeout")
		default:
		}

		// Receive the next close channel response message.
		resp, err := closeChanRes.Recv()
		if err != nil {
			if err == io.EOF {
				// No more messages to receive, the channel is closed.
				return CloseChannelResponse{}, nil
			}
			log.Error().Err(errors.Wrap(err, "LND close channel")).Send()
			return CloseChannelResponse{}, errors.Wrap(err, "LND Close channel")
		}

		// Process the close channel response and see if the channel is pending closure.
		r := CloseChannelResponse{
			Request: ccReq,
		}
		if resp.Update == nil {
			continue
		}

		switch resp.GetUpdate().(type) {
		case *lnrpc.CloseStatusUpdate_ClosePending:
			r.Status = core.Closing
			ch, err := chainhash.NewHash(resp.GetClosePending().Txid)
			if err != nil {
				return CloseChannelResponse{}, errors.Wrap(err, "Getting closing transaction hash")
			}
			r.ClosingTransactionHash = ch.String()

			err = updateChannelToClosingByChannelId(db, ccReq.ChannelId, ch.String())
			if err != nil {
				return CloseChannelResponse{}, errors.Wrap(err, "Updating channel to closing status in the db")
			}
			return r, nil
		}
	}
}
