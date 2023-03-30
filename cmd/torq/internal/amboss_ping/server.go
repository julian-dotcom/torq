package amboss_ping

import (
	"bytes"
	//"bytes"
	"context"
	"encoding/json"
	"net/http"
	//"net/http"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/commons"
)

const ambossSleepSeconds = 25

// Start runs the background server. It sends out a ping to Amboss every 25 seconds.
func Start(ctx context.Context, conn *grpc.ClientConn, nodeId int) {

	defer log.Info().Msgf("Amboss Ping Service terminated for nodeId: %v", nodeId)

	defer commons.SetInactiveLndServiceState(commons.AmbossService, nodeId)
	commons.SetActiveLndServiceState(commons.AmbossService, nodeId)

	const ambossUrl = "https://api.amboss.space/graphql"
	client := lnrpc.NewLightningClient(conn)

	ticker := clock.New().Tick(ambossSleepSeconds * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			now := time.Now().UTC().Format("2006-01-02T15:04:05+0000")
			signMsgReq := lnrpc.SignMessageRequest{
				Msg: []byte(now),
			}
			signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
			if err != nil {
				log.Error().Err(err).Msgf("AmbossService: Signing message: %v", now)
				return
			}

			values := map[string]string{
				"query":     "mutation HealthCheck($signature: String!, $timestamp: String!) { healthCheck(signature: $signature, timestamp: $timestamp) }",
				"variables": "{\"signature\": \"" + signMsgResp.Signature + "\", \"timestamp\": \"" + now + "\"}"}
			jsonData, err := json.Marshal(values)
			if err != nil {
				log.Error().Err(err).Msgf("AmbossService: Marshalling message: %v", values)
				return
			}
			resp, err := http.Post(ambossUrl, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Error().Err(err).Msgf("AmbossService: Posting message: %v", values)
				return
			}
			err = resp.Body.Close()
			if err != nil {
				log.Error().Err(err).Msg("AmbossService: Closing body")
				return
			}
			log.Debug().Msgf("Amboss Ping Service %v", values)
		}
	}
}
