package lnd

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

type peerEventsClient interface {
	SubscribePeerEvents(ctx context.Context, in *lnrpc.PeerEventSubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribePeerEventsClient, error)
}

func SubscribePeerEvents(ctx context.Context, client peerEventsClient,
	nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}) error {

	peerEventStream, err := client.SubscribePeerEvents(ctx, &lnrpc.PeerEventSubscription{})

	if err != nil {
		return errors.Wrap(err, "lnrpc subscribe invoices")
	}

	rl := ratelimit.New(1) // 1 per second maximum rate limit

	for {

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		peerEvent, err := peerEventStream.Recv()

		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				log.Info().Msgf("Peer events subscription - Context canceled")
				break
			}
			log.Error().Err(err).Msg("Problem with peer events subscription")
			// rate limited resubscribe
			for {
				rl.Take()
				peerEventStream, err = client.SubscribePeerEvents(ctx, &lnrpc.PeerEventSubscription{})
				if err == nil {
					//log.Debug().Msgf("Reconnected to invoice subscription")
					break
				}
			}
			continue
		}

		if eventChannel != nil {
			eventChannel <- broadcast.PeerEvent{
				EventData: broadcast.EventData{
					EventTime: time.Now().UTC(),
					NodeId:    nodeSettings.NodeId,
				},
				Type:           peerEvent.Type,
				EventPublicKey: peerEvent.PubKey,
			}
		}
	}

	return nil
}
