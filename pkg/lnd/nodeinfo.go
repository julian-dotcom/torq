package lnd

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// getMissingNodePubKeys creates a string slice with all the PubKey of all nodes
// we have a channel with but where we do not have any node event records.
func getMissingNodePubKeys(db *sqlx.DB) ([]string, error) {

	// Fetch a slice of all public keys related to both open and closed channels
	q := `select array_agg(distinct pub_key) from channel_event where pub_key != '';`

	var res []string
	err := db.QueryRowx(q).Scan(pq.Array(&res))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ImportMissingNodeEvents imports information about all nodes that we have had a channel with.
func ImportMissingNodeEvents(client lnrpc.LightningClient, db *sqlx.DB) error {

	pubKeyList, err := getMissingNodePubKeys(db)
	if err != nil {
		return errors.Wrapf(err, "ImportMissingNodeEvents -> getMissingNodePubKeys(db)")
	}

	ctx := context.Background()
	for _, p := range pubKeyList {
		rsp, err := client.GetNodeInfo(ctx, &lnrpc.NodeInfoRequest{PubKey: p, IncludeChannels: false})
		if err != nil {
			if e, ok := status.FromError(err); ok {
				switch e.Code() {
				case codes.NotFound:
					continue
				default:
					return errors.Wrapf(err, "failed to get alias for node with pubkey %s", p)
				}
			}
		}
		ts := time.Now().UTC()
		err = insertNodeEvent(db, ts, rsp.Node.PubKey, rsp.Node.Alias, rsp.Node.Color,
			rsp.Node.Addresses, rsp.Node.Features)
		if err != nil {
			return errors.Wrapf(err, "ImportMissingNodeEvents -> insertNodeEvent(db, %s, %s, %s, %s, %v, %v)",
				ts, rsp.Node.PubKey, rsp.Node.Alias, rsp.Node.Color, rsp.Node.Addresses, rsp.Node.Features)
		}
	}

	return nil
}
