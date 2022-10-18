package on_chain_tx

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/rs/zerolog/log"
)

type PayOnChainRequest struct {
	NodeId           int     `json:"nodeId"`
	Addr             string  `json:"addr"`
	AmountSat        int64   `json:"amountSat"`
	TargetConf       *int32  `json:"targetConf"`
	SatPerVbyte      *uint64 `json:"satPerVbyte"`
	SendAll          *bool   `json:"sendAll"`
	Label            *string `json:"label"`
	MinConfs         *int32  `json:"minConfs"`
	SpendUnconfirmed *bool   `json:"spendUnconfirmed"`
}

type PayOnChainResponse struct {
	TxId string `json:"txId"`
}

func PayOnChain(db *sqlx.DB, req PayOnChainRequest) (r string, err error) {

	sendCoinsReq, err := processSendRequest(req)
	if err != nil {
		return "", errors.Wrap(err, "Process send request")
	}

	connectionDetails, err := settings.GetNodeConnectionDetailsById(db, req.NodeId)
	if err != nil {
		return "", errors.New("Error getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return "", errors.Wrap(err, "Connecting to LND")
	}

	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)
	ctx := context.Background()

	resp, err := client.SendCoins(ctx, &sendCoinsReq)
	if err != nil {
		return "", errors.Wrap(err, "Sending coins")
	}

	return resp.Txid, nil

}

func processSendRequest(req PayOnChainRequest) (r lnrpc.SendCoinsRequest, err error) {
	if req.NodeId == 0 {
		return r, errors.New("Node id is missing")
	}

	if req.Addr == "" {
		log.Error().Msgf("Address must be provided")
		return r, errors.New("Address must be provided")
	}

	if req.AmountSat <= 0 {
		log.Error().Msgf("Invalid amount")
		return r, errors.New("Invalid amount")
	}

	if req.TargetConf != nil && req.SatPerVbyte != nil {
		log.Error().Msgf("Either targetConf or satPerVbyte accepted")
		return r, errors.New("Either targetConf or satPerVbyte accepted")
	}

	r.Addr = req.Addr
	r.Amount = req.AmountSat

	if req.TargetConf != nil {
		r.TargetConf = *req.TargetConf
	}

	if req.SatPerVbyte != nil {
		r.SatPerVbyte = *req.SatPerVbyte
	}

	if req.SendAll != nil {
		r.SendAll = *req.SendAll
	}

	if req.Label != nil {
		r.Label = *req.Label
	}

	if req.MinConfs != nil {
		r.MinConfs = *req.MinConfs
	}

	if req.SpendUnconfirmed != nil {
		r.SpendUnconfirmed = *req.SpendUnconfirmed
	}

	return r, nil
}
