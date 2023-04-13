package channels

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/core"
)

func ListPeers(client lnrpc.LightningClient, ctx context.Context, latestErr bool) ([]core.Peer, error) {
	resp, err := client.ListPeers(ctx, &lnrpc.ListPeersRequest{LatestError: latestErr})
	if err != nil {
		return []core.Peer{}, errors.Wrap(err, "Listing peers")
	}

	var peers []core.Peer
	for _, peer := range resp.Peers {
		peers = append(peers, core.GetPeer(peer))
	}
	return peers, nil
}
