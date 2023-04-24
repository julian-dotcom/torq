package vector_ping

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/vector"
	"github.com/lncapital/torq/proto/cln"
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
func Start(ctx context.Context, conn *grpc.ClientConn, implementation core.Implementation, nodeId int) {

	serviceType := services_helpers.LndServiceVectorService
	switch implementation {
	case core.CLN:
		serviceType = services_helpers.ClnServiceVectorService
	}

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetActiveNodeServiceState(serviceType, nodeId)

	ticker := time.NewTicker(vectorSleepSeconds * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeId)
			return
		case <-ticker.C:

			pingInfo := VectorPing{
				PingTime:    time.Now().UTC(),
				TorqVersion: build.ExtendedVersion(),
			}

			var pingInfoJsonByteArray []byte
			var signature string

			switch implementation {
			case core.LND:
				client := lnrpc.NewLightningClient(conn)
				getInfoRequest := lnrpc.GetInfoRequest{}
				info, err := client.GetInfo(ctx, &getInfoRequest)
				if err != nil {
					log.Error().Err(err).Msgf("%v obtaining info for nodeId: %v", serviceType.String(), nodeId)
					cache.SetFailedNodeServiceState(serviceType, nodeId)
					return
				}
				pingInfo.Implementation = "LND"
				pingInfo.Version = info.Version
				pingInfo.PublicKey = info.IdentityPubkey
				pingInfo.Alias = info.Alias
				pingInfo.Color = info.Color
				pingInfo.PendingChannelCount = int(info.NumPendingChannels)
				pingInfo.ActiveChannelCount = int(info.NumActiveChannels)
				pingInfo.InactiveChannelCount = int(info.NumInactiveChannels)
				pingInfo.PeerCount = int(info.NumPeers)
				pingInfo.BlockHeight = info.BlockHeight
				pingInfo.BlockHash = info.BlockHash
				pingInfo.BestHeaderTimestamp = time.Unix(info.BestHeaderTimestamp, 0)
				pingInfo.ChainSynced = info.SyncedToChain
				pingInfo.GraphSynced = info.SyncedToGraph
				pingInfo.Addresses = info.Uris
				pingInfo.HtlcInterceptorRequired = info.RequireHtlcInterceptor
				for _, chain := range info.Chains {
					pingInfo.Chains = append(pingInfo.Chains, VectorPingChain{Chain: chain.Chain, Network: chain.Network})
				}
				pingInfo.Features = make(map[int]VectorPingFeature)
				for number, feature := range info.Features {
					pingInfo.Features[int(number)] =
						VectorPingFeature{Name: feature.Name, Required: feature.IsRequired, Known: feature.IsKnown}
				}
				pingInfoJsonByteArray, err = json.Marshal(pingInfo)
				if err != nil {
					log.Error().Err(err).Msgf("%v marshalling message: %v", serviceType.String(), pingInfo)
					cache.SetFailedNodeServiceState(serviceType, nodeId)
					return
				}
				signMsgReq := lnrpc.SignMessageRequest{
					Msg: pingInfoJsonByteArray,
				}
				signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
				if err != nil {
					log.Error().Err(err).Msgf("%v Signing message: %v", serviceType.String(), string(pingInfoJsonByteArray))
					cache.SetFailedNodeServiceState(serviceType, nodeId)
					return
				}
				signature = signMsgResp.Signature
			case core.CLN:
				client := cln.NewNodeClient(conn)
				info, err := client.Getinfo(ctx, &cln.GetinfoRequest{})
				if err != nil {
					log.Error().Err(err).Msgf("%v obtaining info for nodeId: %v", serviceType.String(), nodeId)
					cache.SetFailedNodeServiceState(serviceType, nodeId)
					return
				}
				pingInfo.Implementation = "LND"
				pingInfo.Version = info.Version
				pingInfo.PublicKey = hex.EncodeToString(info.Id)
				pingInfo.Alias = info.Alias
				pingInfo.Color = hex.EncodeToString(info.Color)
				pingInfo.PendingChannelCount = int(info.NumPendingChannels)
				pingInfo.ActiveChannelCount = int(info.NumActiveChannels)
				pingInfo.InactiveChannelCount = int(info.NumInactiveChannels)
				pingInfo.PeerCount = int(info.NumPeers)
				pingInfo.BlockHeight = info.Blockheight
				pingInfo.ChainSynced = info.WarningLightningdSync == nil || *info.WarningLightningdSync == ""
				pingInfo.GraphSynced = info.WarningBitcoindSync == nil || *info.WarningBitcoindSync == ""
				pingInfo.Chains = append(pingInfo.Chains, VectorPingChain{Chain: core.Bitcoin.String(), Network: info.Network})
				for _, address := range info.Address {
					if address != nil && (*address).Address != nil {
						pingInfo.Addresses = append(pingInfo.Addresses, *address.Address)
					}
				}
				pingInfoJsonByteArray, err = json.Marshal(pingInfo)
				if err != nil {
					log.Error().Err(err).Msgf("%v marshalling message: %v", serviceType.String(), pingInfo)
					cache.SetFailedNodeServiceState(serviceType, nodeId)
					return
				}
				signMsgReq := cln.SignmessageRequest{
					Message: string(pingInfoJsonByteArray),
				}
				signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
				if err != nil {
					log.Error().Err(err).Msgf("%v Signing message: %v", serviceType.String(), string(pingInfoJsonByteArray))
					cache.SetFailedNodeServiceState(serviceType, nodeId)
					return
				}
				signature = string(signMsgResp.Signature)
			}

			b, err := json.Marshal(PeerEvent{Message: string(pingInfoJsonByteArray), Signature: signature})
			if err != nil {
				log.Error().Err(err).Msgf("%v marshalling message: %v", serviceType.String(), string(pingInfoJsonByteArray))
				cache.SetFailedNodeServiceState(serviceType, nodeId)
				return
			}

			req, err := http.NewRequest("POST", vector.GetVectorUrl(vectorPingUrlSuffix), bytes.NewBuffer(b))
			if err != nil {
				log.Error().Err(err).Msgf("%v creating new request for message: %v", serviceType.String(), string(pingInfoJsonByteArray))
				cache.SetFailedNodeServiceState(serviceType, nodeId)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Torq-Version", build.ExtendedVersion())
			req.Header.Set("Torq-UUID", cache.GetSettings().TorqUuid)
			httpClient := &http.Client{}
			resp, err := httpClient.Do(req)
			if err != nil {
				log.Error().Err(err).Msgf("%v posting message: %v", serviceType.String(), string(pingInfoJsonByteArray))
				cache.SetFailedNodeServiceState(serviceType, nodeId)
				return
			}
			err = resp.Body.Close()
			if err != nil {
				log.Error().Err(err).Msgf("%v closing response body.", serviceType.String())
				cache.SetFailedNodeServiceState(serviceType, nodeId)
				return
			}
			log.Debug().Msgf("Vector Ping Service %v (%v)", string(pingInfoJsonByteArray), signature)
		}
	}
}
