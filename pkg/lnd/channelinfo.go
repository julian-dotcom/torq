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
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/guregu/null.v4"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/pkg/commons"
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

// ImportRoutingPolicies imports routing policy information about all open channels if they don't already have
func ImportRoutingPolicies(client lnrpc.LightningClient, db *sqlx.DB) error {

	// Get all open channels from LND
	chanIdList, err := getOpenChanIds(client)
	if err != nil {
		return errors.Wrap(err, "Get open chan ids")
	}

	ctx := context.Background()
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
			torqNodeId := commons.GetActiveTorqNodeIdFromPublicKey(cu.AdvertisingNode)
			channelPoint, err := chanPointFromByte(cu.ChanPoint.GetFundingTxidBytes(), cu.ChanPoint.GetOutputIndex())
			if err != nil {
				return errors.Wrap(err, "Creating channel point from byte")
			}

			announcingNodeId := commons.GetNodeIdFromPublicKey(cu.AdvertisingNode)
			if announcingNodeId == 0 {
				announcingNode := nodes.Node{
					PublicKey: cu.AdvertisingNode,
				}
				_, err = nodes.AddNodeWhenNew(db, announcingNode)
				if err != nil {
					return errors.Wrap(err, "Adding new announcingNode")
				}
			}

			connectingNodeId := commons.GetNodeIdFromPublicKey(cu.ConnectingNode)
			if connectingNodeId == 0 {
				connectingNode := nodes.Node{
					PublicKey: cu.ConnectingNode,
				}
				_, err = nodes.AddNodeWhenNew(db, connectingNode)
				if err != nil {
					return errors.Wrap(err, "Adding new connectingNode")
				}
			}

			channelId := commons.GetChannelIdFromChannelPoint(channelPoint)
			if channelId == 0 {
				channel := channels.Channel{
					FirstNodeId:       announcingNodeId,
					SecondNodeId:      connectingNodeId,
					ShortChannelID:    channels.ConvertLNDShortChannelID(cu.ChanId),
					LNDShortChannelID: cu.ChanId,
					LNDChannelPoint:   null.StringFrom(channelPoint),
					Status:            channels.Open,
				}
				channelId, err = channels.AddChannelOrUpdateChannelStatus(db, channel)
				if err != nil {
					return errors.Wrap(err, "Adding new channel")
				}
			} else {
				channelStatusId := commons.GetChannelStatusIdFromChannelId(channelId)
				if channels.Status(channelStatusId) != channels.Open {
					err := channels.UpdateChannelStatus(db, channelId, channels.Open)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to update channel status for channelId: %v", channelId)
					}
				}
			}
			err = insertRoutingPolicy(db, time.Now().UTC(), torqNodeId != 0, channelId, cu)
			if err != nil {
				return errors.Wrap(err, "Insert routing policy")
			}

		}
	}
	return nil
}
