package lnd

import (
	"context"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// For importing the latest routing policy at startup.

// Fetches the channel id form all open channels from LND
func getOpenChanIds(client lnrpc.LightningClient) ([]uint64, error) {

	resp, err := client.ListChannels(context.Background(), &lnrpc.ListChannelsRequest{})
	if err != nil {
		return nil, err
	}

	var chanIdList []uint64
	for _, channel := range resp.Channels {
		chanIdList = append(chanIdList, channel.ChanId)
	}

	return chanIdList, nil
}

func createChanPoint(scp string) (*lnrpc.ChannelPoint, error) {
	var txId string
	outIndex := uint32(0)
	_, err := fmt.Sscanf(scp, "%64s:%d", &txId, &outIndex)
	if err != nil {
		return &lnrpc.ChannelPoint{}, errors.Wrapf(err, "fmt.Sscanf(scp, ..., %s, %d)", txId, outIndex)
	}

	h, err := chainhash.NewHashFromStr(txId)
	if err != nil {
		return &lnrpc.ChannelPoint{}, errors.Wrapf(err, "chainhash.NewHashFromStr(%s)", txId)
	}

	cp := &lnrpc.ChannelPoint{
		FundingTxid: &lnrpc.ChannelPoint_FundingTxidBytes{
			FundingTxidBytes: h.CloneBytes(),
		},
		OutputIndex: outIndex,
	}

	return cp, nil
}

func constructChannelEdgeUpdates(chanEdge *lnrpc.ChannelEdge) ([2]*lnrpc.ChannelEdgeUpdate, error) {

	// Create the channel point struct
	cp1, err := createChanPoint(chanEdge.ChanPoint)
	if err != nil {
		return [2]*lnrpc.ChannelEdgeUpdate{}, errors.Wrapf(err, "ImportRoutingPolicies -> createChanPoint(%s)", chanEdge.ChanPoint)
	}

	cp2, err := createChanPoint(chanEdge.ChanPoint)
	if err != nil {
		return [2]*lnrpc.ChannelEdgeUpdate{}, errors.Wrapf(err, "ImportRoutingPolicies -> createChanPoint(%s)", chanEdge.ChanPoint)
	}

	r := [2]*lnrpc.ChannelEdgeUpdate{
		{
			ChanId:          chanEdge.ChannelId,
			ChanPoint:       cp1,
			Capacity:        chanEdge.Capacity,
			RoutingPolicy:   chanEdge.Node1Policy,
			AdvertisingNode: chanEdge.Node1Pub,
		},
		{
			ChanId:          chanEdge.ChannelId,
			ChanPoint:       cp2,
			Capacity:        chanEdge.Capacity,
			RoutingPolicy:   chanEdge.Node2Policy,
			AdvertisingNode: chanEdge.Node2Pub,
		},
	}

	return r, nil
}

// ImportRoutingPolicies imports routing policy information about all open channels if they don't already have
func ImportRoutingPolicies(client lnrpc.LightningClient, db *sqlx.DB, ourNodePubKeys []string) error {

	// Get all open channels from LND
	chanIdList, err := getOpenChanIds(client)
	if err != nil {
		return errors.Wrapf(err, "ImportRoutingPolicies -> getOpenChanIds(client)")
	}

	ctx := context.Background()
	for _, cid := range chanIdList {

		ce, err := client.GetChanInfo(ctx, &lnrpc.ChanInfoRequest{ChanId: cid})
		if err != nil {
			if e, ok := status.FromError(err); ok {
				switch e.Code() {
				case codes.NotFound:
					continue
				default:
					return errors.Wrapf(err, "ImportRoutingPolicies -> "+
						"client.GetChanInfo(ctx, &lnrpc.ChanInfoRequest{ChanId: %d})",
						cid)
				}
			}
		}

		ceu, err := constructChannelEdgeUpdates(ce)
		if err != nil {
			return errors.Wrapf(err, "ImportRoutingPolicies -> "+
				"constructChannelEdgeUpdates(%v)", ce)
		}

		var ts time.Time
		var outbound bool

		for _, cu := range ceu {

			ts = time.Now().UTC()
			outbound = slices.Contains(ourNodePubKeys, cu.AdvertisingNode)

			err := insertRoutingPolicy(db, ts, outbound, cu)
			if err != nil {
				return errors.Wrap(err, "Insert routing policy")
			}

		}

	}

	return nil
}
