package amboss_ping

import (
	"bytes"
	//"bytes"
	"context"
	"encoding/json"
	"net/http"
	//"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/commons"
)

const AMBOSS_SLEEP_SECONDS = 25

// Start runs the background server. It sends out a ping to Amboss every 25 seconds.
func Start(ctx context.Context, conn *grpc.ClientConn, nodeId int,
	serviceEventChannel chan<- commons.ServiceEvent) error {

	defer log.Info().Msgf("Amboss Ping Service terminated for nodeId: %v", nodeId)

	previousStatus := commons.ServicePending
	_, monitorCancel := context.WithCancel(context.Background())

	const ambossUrl = "https://api.amboss.space/graphql"
	client := lnrpc.NewLightningClient(conn)

	for {
		now := time.Now().UTC().Format("2006-01-02T15:04:05+0000")
		signMsgReq := lnrpc.SignMessageRequest{
			Msg: []byte(now),
		}
		signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
		if err != nil {
			monitorCancel()
			return errors.Wrapf(err, "Signing message: %v", now)
		}

		values := map[string]string{
			"query":     "mutation HealthCheck($signature: String!, $timestamp: String!) { healthCheck(signature: $signature, timestamp: $timestamp) }",
			"variables": "{\"signature\": \"" + signMsgResp.Signature + "\", \"timestamp\": \"" + now + "\"}"}
		jsonData, err := json.Marshal(values)
		if err != nil {
			monitorCancel()
			return errors.Wrapf(err, "Marshalling message: %v", values)
		}
		resp, err := http.Post(ambossUrl, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			monitorCancel()
			return errors.Wrapf(err, "Posting message: %v", values)
		}
		resp.Body.Close()
		if previousStatus == commons.ServiceInactive {
			commons.SendServiceEvent(nodeId, serviceEventChannel, previousStatus, commons.ServiceActive, commons.AmbossService, nil)
			previousStatus = commons.ServiceActive
		}
		log.Debug().Msgf("Amboss Ping Service %v", values)
		time.Sleep(AMBOSS_SLEEP_SECONDS * time.Second)
	}
}
