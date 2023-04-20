package channels

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/internal/lightning_requests"
	"github.com/lncapital/torq/proto/lnrpc"
)

func ListPeers(client lnrpc.LightningClient, ctx context.Context, latestErr bool) ([]lightning_requests.Peer, error) {
	resp, err := client.ListPeers(ctx, &lnrpc.ListPeersRequest{LatestError: latestErr})
	if err != nil {
		return []lightning_requests.Peer{}, errors.Wrap(err, "Listing peers")
	}

	var peers []lightning_requests.Peer
	for _, peer := range resp.Peers {
		peers = append(peers, lightning_requests.GetPeerLND(peer))
	}
	return peers, nil
}
