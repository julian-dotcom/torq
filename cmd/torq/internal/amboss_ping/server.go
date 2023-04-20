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

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/services_core"
	"github.com/lncapital/torq/proto/lnrpc"

	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/proto/cln"
)

const ambossSleepSeconds = 25

// Start runs the background server. It sends out a ping to Amboss every 25 seconds.
func Start(ctx context.Context, conn *grpc.ClientConn, implementation core.Implementation, nodeId int) {

	serviceType := services_core.LndServiceAmbossService
	switch implementation {
	case core.CLN:
		serviceType = services_core.ClnServiceAmbossService
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

	const ambossUrl = "https://api.amboss.space/graphql"

	ticker := time.NewTicker(ambossSleepSeconds * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeId)
			return
		case <-ticker.C:
			now := time.Now().UTC().Format("2006-01-02T15:04:05+0000")
			var signature string
			switch implementation {
			case core.LND:
				client := lnrpc.NewLightningClient(conn)
				signMsgReq := lnrpc.SignMessageRequest{
					Msg: []byte(now),
				}
				signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
				if err != nil {
					log.Error().Err(err).Msgf("%v signing message: %v", serviceType.String(), now)
					cache.SetFailedNodeServiceState(serviceType, nodeId)
					return
				}
				signature = signMsgResp.Signature
			case core.CLN:
				client := cln.NewNodeClient(conn)
				signMsgReq := cln.SignmessageRequest{
					Message: now,
				}
				signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
				if err != nil {
					log.Error().Err(err).Msgf("%v signing message: %v", serviceType.String(), now)
					cache.SetFailedNodeServiceState(serviceType, nodeId)
					return
				}
				signature = string(signMsgResp.Signature)
			}

			values := map[string]string{
				"query":     "mutation HealthCheck($signature: String!, $timestamp: String!) { healthCheck(signature: $signature, timestamp: $timestamp) }",
				"variables": "{\"signature\": \"" + signature + "\", \"timestamp\": \"" + now + "\"}"}
			jsonData, err := json.Marshal(values)
			if err != nil {
				log.Error().Err(err).Msgf("%v marshalling message: %v", serviceType.String(), values)
				cache.SetFailedNodeServiceState(serviceType, nodeId)
				return
			}
			resp, err := http.Post(ambossUrl, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Error().Err(err).Msgf("%v posting message: %v", serviceType.String(), values)
				cache.SetFailedNodeServiceState(serviceType, nodeId)
				return
			}
			err = resp.Body.Close()
			if err != nil {
				log.Error().Err(err).Msgf("%v closing body", serviceType.String())
				cache.SetFailedNodeServiceState(serviceType, nodeId)
				return
			}
			log.Debug().Msgf("Amboss Ping Service %v", values)
		}
	}
}
