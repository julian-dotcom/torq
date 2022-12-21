package channels

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/peers"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

type PsbtDetails struct {
	FundingAddress string `json:"funding_address,omitempty"`
	FundingAmount  int64  `json:"funding_amount,omitempty"`
	Psbt           []byte `json:"psbt,omitempty"`
}

func OpenChannel(eventChannel chan interface{}, db *sqlx.DB, req commons.OpenChannelRequest, requestId string) (err error) {
	openChanReq, err := prepareOpenRequest(req)
	if err != nil {
		return errors.Wrap(err, "Preparing open request")
	}

	connectionDetails, err := settings.GetConnectionDetailsById(db, req.NodeId)
	if err != nil {
		return errors.Wrap(err, "Getting node connection details from the db")
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

	ctx := context.Background()

	//If host provided - check if node is connected to peer and if not, connect peer
	if req.NodePubKey != "" && req.Host != nil {
		//log.Debug().Msgf("Host provided. connect peer")
		if err := checkConnectPeer(client, ctx, req.NodeId, req.NodePubKey, *req.Host); err != nil {
			return err
		}
	}

	//Send open channel request
	openChanRes, err := client.OpenChannel(ctx, openChanReq)

	if err != nil {
		return errors.Wrap(err, "LND Open channel")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		resp, err := openChanRes.Recv()

		if err == io.EOF {
			//log.Info().Msgf("Open channel EOF")
			return nil
		}

		if err != nil {
			return errors.Wrapf(err, "Opening channel")
		}

		r, err := processOpenResponse(resp, req, requestId)
		if err != nil {
			return errors.Wrap(err, "Processing open response")
		}
		if eventChannel != nil {
			eventChannel <- r
		}
	}
}

func prepareOpenRequest(ocReq commons.OpenChannelRequest) (r *lnrpc.OpenChannelRequest, err error) {
	if ocReq.NodeId == 0 {
		return &lnrpc.OpenChannelRequest{}, errors.New("Node id is missing")
	}

	if ocReq.SatPerVbyte != nil && ocReq.TargetConf != nil {
		return &lnrpc.OpenChannelRequest{}, errors.New("Cannot set both SatPerVbyte and TargetConf")
	}

	pubKeyHex, err := hex.DecodeString(ocReq.NodePubKey)
	if err != nil {
		return &lnrpc.OpenChannelRequest{}, errors.New("error decoding public key hex")
	}

	//open channel request
	openChanReq := &lnrpc.OpenChannelRequest{
		NodePubkey: pubKeyHex,

		// This is the amount we are putting into the channel (channel size)
		LocalFundingAmount: ocReq.LocalFundingAmount,
	}

	// The amount to give the other node in the opening process.
	// NB: This means you will give the other node this amount of sats
	if ocReq.PushSat != nil {
		openChanReq.PushSat = *ocReq.PushSat
	}

	if ocReq.SatPerVbyte != nil {
		openChanReq.SatPerVbyte = *ocReq.SatPerVbyte
	}

	if ocReq.TargetConf != nil {
		openChanReq.TargetConf = *ocReq.TargetConf
	}

	if ocReq.Private != nil {
		openChanReq.Private = *ocReq.Private
	}

	if ocReq.MinHtlcMsat != nil {
		openChanReq.MinHtlcMsat = int64(*ocReq.MinHtlcMsat)
	}

	if ocReq.RemoteCsvDelay != nil {
		openChanReq.RemoteCsvDelay = *ocReq.RemoteCsvDelay
	}

	if ocReq.MinConfs != nil {
		openChanReq.MinConfs = *ocReq.MinConfs
	}

	if ocReq.SpendUnconfirmed != nil {
		openChanReq.SpendUnconfirmed = *ocReq.SpendUnconfirmed
	}

	if ocReq.CloseAddress != nil {
		openChanReq.CloseAddress = *ocReq.CloseAddress
	}
	return openChanReq, nil
}

func processOpenResponse(resp *lnrpc.OpenStatusUpdate, req commons.OpenChannelRequest, requestId string) (commons.OpenChannelResponse, error) {
	switch resp.GetUpdate().(type) {
	case *lnrpc.OpenStatusUpdate_ChanPending:
		log.Info().Msgf("Channel pending")

		pc := resp.GetChanPending()
		pcp, err := translateChanPoint(pc.Txid, pc.OutputIndex)
		if err != nil {
			log.Error().Msgf("Error translating pending channel point")
			return commons.OpenChannelResponse{}, err
		}

		return commons.OpenChannelResponse{
			RequestId:           requestId,
			Request:             req,
			Status:              commons.Opening,
			PendingChannelPoint: pcp,
		}, nil

	case *lnrpc.OpenStatusUpdate_ChanOpen:
		log.Info().Msgf("Channel open")

		oc := resp.GetChanOpen()
		ocp, err := translateChanPoint(oc.ChannelPoint.GetFundingTxidBytes(), oc.ChannelPoint.OutputIndex)
		if err != nil {
			log.Error().Msgf("Error translating channel point")
			return commons.OpenChannelResponse{}, err
		}

		return commons.OpenChannelResponse{
			RequestId:    requestId,
			Request:      req,
			Status:       commons.Open,
			ChannelPoint: ocp,
		}, nil

	case *lnrpc.OpenStatusUpdate_PsbtFund:
		log.Error().Msg("Channel psbt fund response received. Can't process this response")
		return commons.OpenChannelResponse{}, errors.New("Channel psbt fund response received. Can't process this response")
	default:
	}

	return commons.OpenChannelResponse{}, nil
}

func translateChanPoint(cb []byte, oi uint32) (string, error) {
	ch, err := chainhash.NewHash(cb)
	if err != nil {
		return "", errors.Wrap(err, "Chainhash new hash")
	}

	return fmt.Sprintf("%s:%d", ch.String(), oi), nil
}

func checkConnectPeer(client lnrpc.LightningClient, ctx context.Context, nodeId int, remotePubkey string, host string) (err error) {

	peerList, err := peers.ListPeers(client, ctx, "true")
	if err != nil {
		return errors.Wrap(err, "List peers")
	}

	for _, peer := range peerList {
		if peer.PubKey == remotePubkey {
			// peer found
			//log.Debug().Msgf("Peer is connected")
			return nil
		}
	}

	req := peers.ConnectPeerRequest{
		NodeId: nodeId,
		LndAddress: peers.LndAddress{
			PubKey: remotePubkey,
			Host:   host,
		},
	}

	_, err = peers.ConnectPeer(client, ctx, req)
	if err != nil {
		return errors.Wrap(err, "Connect peer")
	}

	return nil
}
