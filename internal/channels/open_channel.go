package channels

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/pkg/cln_connect"
	"github.com/lncapital/torq/proto/cln"
	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/core"
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
	Request                OpenChannelRequest `json:"request"`
	Status                 core.ChannelStatus `json:"status"`
	ChannelPoint           string             `json:"channelPoint"`
	FundingTransactionHash string             `json:"fundingTransactionHash,omitempty"`
	FundingOutputIndex     uint32             `json:"fundingOutputIndex,omitempty"`
}

func OpenChannel(req OpenChannelRequest) (OpenChannelResponse, error) {
	ctx := context.Background()

	connectionDetails := cache.GetNodeConnectionDetails(req.NodeId)
	switch connectionDetails.Implementation {
	case core.LND:
		openChanReq, err := prepareLndOpenRequest(req)
		if err != nil {
			return OpenChannelResponse{}, err
		}

		conn, err := lnd_connect.Connect(
			connectionDetails.GRPCAddress,
			connectionDetails.TLSFileBytes,
			connectionDetails.MacaroonFileBytes)
		if err != nil {
			return OpenChannelResponse{}, errors.Wrap(err, "connecting to LND")
		}
		defer func(conn *grpc.ClientConn) {
			err := conn.Close()
			if err != nil {
				log.Error().Msgf("Error closing grpc connection: %v", err)
			}
		}(conn)

		client := lnrpc.NewLightningClient(conn)

		//If host provided - check if node is connected to peer and if not, connect peer
		if req.NodePubKey != "" && req.Host != nil {
			if err := checkConnectPeer(client, ctx, req.NodeId, req.NodePubKey, *req.Host); err != nil {
				return OpenChannelResponse{}, errors.Wrap(err, "could not connect to peer")
			}
		}

		// Send open channel request
		cp, err := openChannelProcess(client, openChanReq, req)
		if err != nil {
			return OpenChannelResponse{}, errors.Wrap(err, "LND Open channel")
		}

		return cp, nil
	case core.CLN:
		openChanReq, err := prepareClnOpenRequest(req)
		if err != nil {
			return OpenChannelResponse{}, err
		}

		conn, err := cln_connect.Connect(
			connectionDetails.GRPCAddress,
			connectionDetails.CertificateFileBytes,
			connectionDetails.KeyFileBytes)
		if err != nil {
			return OpenChannelResponse{}, errors.Wrap(err, "connecting to CLN")
		}
		defer func(conn *grpc.ClientConn) {
			err := conn.Close()
			if err != nil {
				log.Error().Msgf("Error closing grpc connection: %v", err)
			}
		}(conn)

		client := cln.NewNodeClient(conn)
		channel, err := client.FundChannel(ctx, openChanReq)
		if err != nil {
			return OpenChannelResponse{}, errors.Wrap(err, "LND Open channel")
		}

		response := OpenChannelResponse{}
		response.ChannelPoint = fmt.Sprintf("%v:%v", hex.EncodeToString(channel.Txid), channel.Outnum)
		response.Status = core.Opening
		response.Request = req
		response.FundingTransactionHash = hex.EncodeToString(channel.Txid)
		response.FundingOutputIndex = channel.Outnum
		return response, nil
	}
	return OpenChannelResponse{}, errors.New("implementation not supported")
}

func prepareLndOpenRequest(ocReq OpenChannelRequest) (r *lnrpc.OpenChannelRequest, err error) {
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

func prepareClnOpenRequest(request OpenChannelRequest) (*cln.FundchannelRequest, error) {
	if request.NodeId == 0 {
		return nil, errors.New("nodeId is missing")
	}

	if request.SatPerVbyte != nil && request.TargetConf != nil {
		return nil, errors.New("Cannot set both SatPerVbyte and TargetConf")
	}

	pubKeyHex, err := hex.DecodeString(request.NodePubKey)
	if err != nil {
		return nil, errors.New("error decoding public key hex")
	}

	//open channel request
	openChanReq := &cln.FundchannelRequest{
		Id: pubKeyHex,

		// This is the amount we are putting into the channel (channel size)
		Amount: &cln.AmountOrAll{Value: &cln.AmountOrAll_Amount{Amount: &cln.Amount{Msat: uint64(request.LocalFundingAmount * 1_000)},
		}},
	}

	// The amount to give the other node in the opening process.
	// NB: This means you will give the other node this amount of sats
	if request.PushSat != nil {
		openChanReq.PushMsat = &cln.Amount{Msat: uint64((*request.PushSat) * 1_000)}
	}

	if request.SatPerVbyte != nil {
		// TODO FIXME CLN verify
		openChanReq.Feerate = &cln.Feerate{Style: &cln.Feerate_Perkw{Perkw: uint32(*request.SatPerVbyte)}}
	}

	if request.TargetConf != nil {
		minDept := uint32(*request.TargetConf)
		openChanReq.Mindepth = &minDept
	}

	if request.Private != nil {
		announced := !*request.Private
		openChanReq.Announce = &announced
	}

	// TODO FIXME CLN verify
	//if request.MinHtlcMsat != nil {
	//	openChanReq.MinHtlcMsat = int64(*request.MinHtlcMsat)
	//}

	// TODO FIXME CLN verify
	//if request.RemoteCsvDelay != nil {
	//	openChanReq.RemoteCsvDelay = *request.RemoteCsvDelay
	//}

	if request.MinConfs != nil {
		minConf := uint32(*request.MinConfs)
		openChanReq.Minconf = &minConf
	}

	// TODO FIXME CLN verify
	//if request.SpendUnconfirmed != nil {
	//	openChanReq.SpendUnconfirmed = *request.SpendUnconfirmed
	//}

	if request.CloseAddress != nil {
		openChanReq.CloseTo = request.CloseAddress
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

	peerList, err := ListPeers(client, ctx, true)
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

	req := ConnectPeerRequest{
		NodeId: nodeId,
		LndAddress: LndAddress{
			PubKey: remotePubkey,
			Host:   host,
		},
	}

	_, err = ConnectPeer(client, ctx, req)
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
			r.Status = core.Opening
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
