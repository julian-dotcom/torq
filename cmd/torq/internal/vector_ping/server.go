package vector_ping

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/pkg/commons"
)

const vectorSleepSeconds = 20
const vectorPingUrlSuffix = "api/publicNodeEvents/ping"

type PeerEvent struct {
	Signature string `json:"signature"`
	Message   string `json:"message"`
}

type VectorPing struct {
	PingTime                time.Time                 `json:"pingTime"`
	TorqVersion             string                    `json:"torqVersion"`
	Implementation          string                    `json:"implementation"`
	Version                 string                    `json:"version"`
	PublicKey               string                    `json:"publicKey"`
	Alias                   string                    `json:"alias"`
	Color                   string                    `json:"color"`
	PendingChannelCount     int                       `json:"pendingChannelCount"`
	ActiveChannelCount      int                       `json:"activeChannelCount"`
	InactiveChannelCount    int                       `json:"inactiveChannelCount"`
	PeerCount               int                       `json:"peerCount"`
	BlockHeight             uint32                    `json:"blockHeight"`
	BlockHash               string                    `json:"blockHash"`
	BestHeaderTimestamp     time.Time                 `json:"bestHeaderTimestamp"`
	ChainSynced             bool                      `json:"chainSynced"`
	GraphSynced             bool                      `json:"graphSynced"`
	Chains                  []VectorPingChain         `json:"chains"`
	Addresses               []string                  `json:"addresses"`
	Features                map[int]VectorPingFeature `json:"features"`
	HtlcInterceptorRequired bool                      `json:"htlcInterceptorRequired"`
}
type VectorPingFeature struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Known    bool   `json:"known"`
}
type VectorPingChain struct {
	Chain   string `json:"chain"`
	Network string `json:"network"`
}

// Start runs the background server. It sends out a ping to Vector every 20 seconds.
func Start(ctx context.Context, conn *grpc.ClientConn, nodeId int) {

	defer log.Info().Msgf("Vector Ping Service terminated for nodeId: %v", nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in VectorService (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.VectorService, nodeId)
			return
		}
	}()

	commons.SetActiveLndServiceState(commons.VectorService, nodeId)

	client := lnrpc.NewLightningClient(conn)

	ticker := clock.New().Tick(vectorSleepSeconds * time.Second)

	for {
		select {
		case <-ctx.Done():
			commons.SetInactiveLndServiceState(commons.VectorService, nodeId)
			return
		case <-ticker:
			getInfoRequest := lnrpc.GetInfoRequest{}
			info, err := client.GetInfo(ctx, &getInfoRequest)
			if err != nil {
				log.Error().Err(err).Msgf("VectorService: Obtaining LND info for nodeId: %v", nodeId)
				commons.SetFailedLndServiceState(commons.VectorService, nodeId)
				return
			}

			pingInfo := VectorPing{
				PingTime:                time.Now().UTC(),
				TorqVersion:             build.ExtendedVersion(),
				Implementation:          "LND",
				Version:                 info.Version,
				PublicKey:               info.IdentityPubkey,
				Alias:                   info.Alias,
				Color:                   info.Color,
				PendingChannelCount:     int(info.NumPendingChannels),
				ActiveChannelCount:      int(info.NumActiveChannels),
				InactiveChannelCount:    int(info.NumInactiveChannels),
				PeerCount:               int(info.NumPeers),
				BlockHeight:             info.BlockHeight,
				BlockHash:               info.BlockHash,
				BestHeaderTimestamp:     time.Unix(info.BestHeaderTimestamp, 0),
				ChainSynced:             info.SyncedToChain,
				GraphSynced:             info.SyncedToGraph,
				Addresses:               info.Uris,
				HtlcInterceptorRequired: info.RequireHtlcInterceptor,
			}
			for _, chain := range info.Chains {
				pingInfo.Chains = append(pingInfo.Chains, VectorPingChain{Chain: chain.Chain, Network: chain.Network})
			}
			pingInfo.Features = make(map[int]VectorPingFeature)
			for number, feature := range info.Features {
				pingInfo.Features[int(number)] =
					VectorPingFeature{Name: feature.Name, Required: feature.IsRequired, Known: feature.IsKnown}
			}

			pingInfoJsonByteArray, err := json.Marshal(pingInfo)
			if err != nil {
				log.Error().Err(err).Msgf("VectorService: Marshalling message: %v", info)
				commons.SetFailedLndServiceState(commons.VectorService, nodeId)
				return
			}
			signMsgReq := lnrpc.SignMessageRequest{
				Msg: pingInfoJsonByteArray,
			}
			signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
			if err != nil {
				log.Error().Err(err).Msgf("VectorService: Signing message: %v", string(pingInfoJsonByteArray))
				commons.SetFailedLndServiceState(commons.VectorService, nodeId)
				return
			}
			b, err := json.Marshal(PeerEvent{Message: string(pingInfoJsonByteArray), Signature: signMsgResp.Signature})
			if err != nil {
				log.Error().Err(err).Msgf("VectorService: Marshalling message: %v", string(pingInfoJsonByteArray))
				commons.SetFailedLndServiceState(commons.VectorService, nodeId)
				return
			}

			req, err := http.NewRequest("POST", commons.GetVectorUrl(vectorPingUrlSuffix), bytes.NewBuffer(b))
			if err != nil {
				log.Error().Err(err).Msgf("VectorService: Creating new request for message: %v", string(pingInfoJsonByteArray))
				commons.SetFailedLndServiceState(commons.VectorService, nodeId)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Torq-Version", build.ExtendedVersion())
			req.Header.Set("Torq-UUID", commons.GetSettings().TorqUuid)
			httpClient := &http.Client{}
			resp, err := httpClient.Do(req)
			if err != nil {
				log.Error().Err(err).Msgf("VectorService: Posting message: %v", string(pingInfoJsonByteArray))
				commons.SetFailedLndServiceState(commons.VectorService, nodeId)
				return
			}
			err = resp.Body.Close()
			if err != nil {
				log.Error().Err(err).Msg("VectorService: Closing response body.")
				commons.SetFailedLndServiceState(commons.VectorService, nodeId)
				return
			}
			log.Debug().Msgf("Vector Ping Service %v (%v)", string(pingInfoJsonByteArray), signMsgResp)
		}
	}
}
