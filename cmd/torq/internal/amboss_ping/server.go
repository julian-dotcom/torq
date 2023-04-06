package amboss_ping

import (
	"bytes"
	//"bytes"
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	//"net/http"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
)

const ambossSleepSeconds = 25

// Start runs the background server. It sends out a ping to Amboss every 25 seconds.
func Start(ctx context.Context, conn *grpc.ClientConn, nodeId int) {

	serviceType := core.AmbossService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetActiveLndServiceState(serviceType, nodeId)

	const ambossUrl = "https://api.amboss.space/graphql"
	client := lnrpc.NewLightningClient(conn)

	ticker := time.NewTicker(ambossSleepSeconds * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveLndServiceState(serviceType, nodeId)
			return
		case <-ticker.C:
			now := time.Now().UTC().Format("2006-01-02T15:04:05+0000")
			signMsgReq := lnrpc.SignMessageRequest{
				Msg: []byte(now),
			}
			signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
			if err != nil {
				log.Error().Err(err).Msgf("AmbossService: Signing message: %v", now)
				cache.SetFailedLndServiceState(serviceType, nodeId)
				return
			}

			values := map[string]string{
				"query":     "mutation HealthCheck($signature: String!, $timestamp: String!) { healthCheck(signature: $signature, timestamp: $timestamp) }",
				"variables": "{\"signature\": \"" + signMsgResp.Signature + "\", \"timestamp\": \"" + now + "\"}"}
			jsonData, err := json.Marshal(values)
			if err != nil {
				log.Error().Err(err).Msgf("AmbossService: Marshalling message: %v", values)
				cache.SetFailedLndServiceState(serviceType, nodeId)
				return
			}
			resp, err := http.Post(ambossUrl, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Error().Err(err).Msgf("AmbossService: Posting message: %v", values)
				cache.SetFailedLndServiceState(serviceType, nodeId)
				return
			}
			err = resp.Body.Close()
			if err != nil {
				log.Error().Err(err).Msg("AmbossService: Closing body")
				cache.SetFailedLndServiceState(serviceType, nodeId)
				return
			}
			log.Debug().Msgf("Amboss Ping Service %v", values)
		}
	}
}
