package channels

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"io"
	"time"

	"github.com/lncapital/torq/internal/peers"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

type PsbtDetails struct {
	FundingAddress string `json:"funding_address,omitempty"`
	FundingAmount  int64  `json:"funding_amount,omitempty"`
	Psbt           []byte `json:"psbt,omitempty"`
}

type OpenChannelRequest struct {
	NodeId             int     `json:"nodeId"`
	SatPerVbyte        *uint64 `json:"satPerVbyte"`
	NodePubKey         string  `json:"nodePubKey"`
	Host               *string `json:"host"`
	LocalFundingAmount int64   `json:"localFundingAmount"`
	PushSat            *int64  `json:"pushSat"`
	TargetConf         *int32  `json:"targetConf"`
	Private            *bool   `json:"private"`
	MinHtlcMsat        *uint64 `json:"minHtlcMsat"`
	RemoteCsvDelay     *uint32 `json:"remoteCsvDelay"`
	MinConfs           *int32  `json:"minConfs"`
	SpendUnconfirmed   *bool   `json:"spendUnconfirmed"`
	CloseAddress       *string `json:"closeAddress"`
}

type OpenChannelResponse struct {
	Request                OpenChannelRequest    `json:"request"`
	Status                 commons.ChannelStatus `json:"status"`
	ChannelPoint           string                `json:"channelPoint"`
	FundingTransactionHash string                `json:"fundingTransactionHash,omitempty"`
	FundingOutputIndex     uint32                `json:"fundingOutputIndex,omitempty"`
}

func OpenChannel(db *sqlx.DB, req OpenChannelRequest) (response OpenChannelResponse, err error) {
	openChanReq, err := prepareOpenRequest(req)
	if err != nil {
		return OpenChannelResponse{}, err
	}

	connectionDetails, err := settings.GetConnectionDetailsById(db, req.NodeId)
	if err != nil {
		return OpenChannelResponse{}, errors.Wrap(err, "Getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return OpenChannelResponse{}, errors.Wrap(err, "Connecting to LND")
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Error().Msgf("Error closing grpc connection: %v", err)
		}
	}(conn)

	client := lnrpc.NewLightningClient(conn)

	ctx := context.Background()

	//If host provided - check if node is connected to peer and if not, connect peer
	if req.NodePubKey != "" && req.Host != nil {
		if err := checkConnectPeer(client, ctx, req.NodeId, req.NodePubKey, *req.Host); err != nil {
			return OpenChannelResponse{}, errors.Wrap(err, "Could not connect to peer")
		}
	}

	// Send open channel request
	cp, err := openChannelProcess(client, openChanReq, req)
	if err != nil {
		return OpenChannelResponse{}, errors.Wrap(err, "LND Open channel")
	}

	return cp, nil

}

func prepareOpenRequest(ocReq OpenChannelRequest) (r *lnrpc.OpenChannelRequest, err error) {
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

func openChannelProcess(client lnrpc.LightningClient, openChannelReq *lnrpc.OpenChannelRequest,
	ccReq OpenChannelRequest) (OpenChannelResponse, error) {

	// Create a context with a timeout.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), closeChannelTimeoutInSeconds*time.Second)
	defer cancel()

	// Call OpenChannel with the timeout context.
	openReq, err := client.OpenChannel(timeoutCtx, openChannelReq)
	if err != nil {
		return OpenChannelResponse{}, errors.Wrap(err, "Close channel request")
	}

	// Loop until we receive an open channel response or the context times out.
	for {
		select {
		case <-timeoutCtx.Done():
			return OpenChannelResponse{}, errors.New("Close channel request timeout")
		default:
		}

		// Receive the next close channel response message.
		resp, err := openReq.Recv()
		if err != nil {
			if err == io.EOF {
				// No more messages to receive, the channel is open.
				return OpenChannelResponse{}, nil
			}
			return OpenChannelResponse{}, errors.Wrap(err, "LND Open channel")
		}

		r := OpenChannelResponse{
			Request: ccReq,
		}
		if resp.Update == nil {
			continue
		}

		switch resp.GetUpdate().(type) {
		case *lnrpc.OpenStatusUpdate_ChanPending:
			r.Status = commons.Opening
			ch, err := chainhash.NewHash(resp.GetChanPending().Txid)
			if err != nil {
				return OpenChannelResponse{}, errors.Wrap(err, "Getting closing transaction hash")
			}
			r.FundingTransactionHash = ch.String()
			r.FundingOutputIndex = resp.GetChanPending().OutputIndex
			r.ChannelPoint = fmt.Sprintf("%s:%d", ch.String(), resp.GetChanPending().OutputIndex)
			return r, nil
		}
	}
}
