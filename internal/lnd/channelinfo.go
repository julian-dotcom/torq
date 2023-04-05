package lnd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
)

// For importing the latest routing policy at startup.

// Fetches the channel id form all open channels from LND
func getOpenChanIds(client lndClientChannelEvent) ([]uint64, error) {

	resp, err := client.ListChannels(context.Background(), &lnrpc.ListChannelsRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "LND List Channels")
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
		return &lnrpc.ChannelPoint{}, errors.Wrap(err, "fmt sscanf")
	}

	h, err := chainhash.NewHashFromStr(txId)
	if err != nil {
		return &lnrpc.ChannelPoint{}, errors.Wrap(err, "New hash from str")
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
		return [2]*lnrpc.ChannelEdgeUpdate{}, errors.Wrap(err, "Create channel point")
	}

	cp2, err := createChanPoint(chanEdge.ChanPoint)
	if err != nil {
		return [2]*lnrpc.ChannelEdgeUpdate{}, errors.Wrap(err, "Create channel point")
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

// ImportRoutingPolicies imports routing policy information about all channels if they don't already have
func ImportRoutingPolicies(ctx context.Context, client lndClientChannelEvent, db *sqlx.DB, nodeSettings cache.NodeSettingsCache) error {

	// Get all open channels from LND
	chanIdList, err := getOpenChanIds(client)
	if err != nil {
		return errors.Wrap(err, "Get open chan ids")
	}

	for _, cid := range chanIdList {
		ce, err := client.GetChanInfo(ctx, &lnrpc.ChanInfoRequest{ChanId: cid})
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "edge not found") {
				log.Debug().Err(err).Msgf("Edge wasn't found when importing routing policies for channel id: %v", cid)
				continue
			}
			if e, ok := status.FromError(err); ok {
				switch e.Code() {
				case codes.NotFound:
					log.Debug().Err(err).Msgf("Chan info not found error when importing routing policies for channel id: %v", cid)
					continue
				default:
					return errors.Wrap(err, "Get chan info")
				}
			}
		}
		ceu, err := constructChannelEdgeUpdates(ce)
		if err != nil {
			return errors.Wrap(err, "Construct Channel Edge Updates")
		}
		for _, cu := range ceu {
			// TODO FIXME shouldn't we check if announcingNodeId == nodeId || connectingNodeId == nodeId ???
			// We don't want our other torqNode updates in here???
			channelPoint, err := chanPointFromByte(cu.ChanPoint.GetFundingTxidBytes(), cu.ChanPoint.GetOutputIndex())
			if err != nil {
				return errors.Wrap(err, "Creating channel point from byte")
			}
			fundingTransactionHash, fundingOutputIndex := core.ParseChannelPoint(channelPoint)
			channelId := cache.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
			if channelId == 0 {
				return errors.New(fmt.Sprintf(
					"Importing routing policy for a channel that doesn't exist in our database? (fundingTransactionHash: %v, fundingOutputIndex: %v)",
					fundingTransactionHash, fundingOutputIndex))
			}
			channelStatus := cache.GetChannelStatusByChannelId(channelId)
			if channelStatus != core.Open {
				_, err := db.Exec(`
					UPDATE channel SET status_id=$1, updated_on=$2 WHERE channel_id=$3 AND status_id!=$1`,
					core.Open, time.Now().UTC(), channelId)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to update channel status for channelId: %v", channelId)
				}
				cache.SetChannelStatus(channelId, core.Open)
			}
			err = insertRoutingPolicy(db, time.Now().UTC(), channelId, nodeSettings, cu)
			if err != nil {
				return errors.Wrap(err, "Insert routing policy")
			}

		}
	}
	return nil
}
