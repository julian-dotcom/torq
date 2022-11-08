package lnd

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lncapital/torq/pkg/commons"
)

// For importing the latest node info at startup.

// ImportNodeInfo imports node information about all channels if they don't already have
func ImportNodeInfo(client lnrpc.LightningClient, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings) error {
	// Get all node public keys with channels
	publicKeys := commons.GetAllChannelPublicKeys(nodeSettings.Chain, nodeSettings.Network)

	ctx := context.Background()
	for _, publicKey := range publicKeys {
		ni, err := client.GetNodeInfo(ctx, &lnrpc.NodeInfoRequest{PubKey: publicKey, IncludeChannels: false})
		if err != nil {
			if e, ok := status.FromError(err); ok {
				switch e.Code() {
				case codes.NotFound:
					log.Debug().Err(err).Msgf("Node info not found error when importing node info for public key: %v", publicKey)
					continue
				default:
					return errors.Wrap(err, "Get node info")
				}
			}
		}
		err = insertNodeEvent(db, time.Now().UTC(),
			commons.GetNodeIdFromPublicKey(publicKey, nodeSettings.Chain, nodeSettings.Network),
			ni.Node.Alias, ni.Node.Color, ni.Node.Addresses, ni.Node.Features, nodeSettings.NodeId, nil)
		if err != nil {
			return errors.Wrap(err, "Insert node event")
		}
	}
	return nil
}
