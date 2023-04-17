package channels

import (
	"reflect"
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/testutil"
)

const FundingTransactionHash = "c946aad8ea807099f2f4eaf2f92821024c9d8a79afd465573e924dacddfa490c"
const FundingOutputIndex = 1

func Test_prepareCloseRequest(t *testing.T) {

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, err := srv.NewTestDatabase(true)
	defer cancel()
	if err != nil {
		t.Fatal(err)
	}

	err = settings.InitializeSettingsCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing SettingsCache cache: %v", err)
	}

	err = settings.InitializeNodesCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing NodeCache cache: %v", err)
	}

	lndShortChannelId := uint64(9999)
	shortChannelId := core.ConvertLNDShortChannelID(lndShortChannelId)
	channel, err := addChannel(db, Channel{
		ShortChannelID:         &shortChannelId,
		Status:                 core.Open,
		Private:                false,
		Capacity:               10_000_000,
		FirstNodeId:            cache.GetChannelPeerNodeIdByPublicKey(testutil.TestPublicKey1, core.Bitcoin, core.SigNet),
		SecondNodeId:           cache.GetChannelPeerNodeIdByPublicKey(testutil.TestPublicKey2, core.Bitcoin, core.SigNet),
		LNDShortChannelID:      &lndShortChannelId,
		FundingOutputIndex:     FundingOutputIndex,
		FundingTransactionHash: FundingTransactionHash,
	})
	if err != nil {
		log.Fatal().Err(err).Msgf("Problem initializing channel: %v", err)
	}
	log.Info().Msgf("Created OPEN channel to be closed with channelId: %v", channel.ChannelID)

	err = settings.InitializeChannelsCache(db)
	if err != nil {
		cancel()
		log.Fatal().Err(err).Msgf("Problem initializing ChannelCache cache: %v", err)
	}

	fundingTxid := &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: FundingTransactionHash}
	var channelPoint = &lnrpc.ChannelPoint{FundingTxid: fundingTxid, OutputIndex: FundingOutputIndex}
	var force = true
	var targetConf int32 = 12
	var deliveryAddress = "test"
	var satPerVbyte uint64 = 12

	tests := []struct {
		name    string
		input   CloseChannelRequest
		want    *lnrpc.CloseChannelRequest
		wantErr bool
	}{
		{
			"Node ID not provided",
			CloseChannelRequest{
				ChannelId: channel.ChannelID,
			},
			&lnrpc.CloseChannelRequest{
				ChannelPoint: channelPoint,
			},
			true,
		},
		{
			"Both targetConf & satPerVbyte provided",
			CloseChannelRequest{
				NodeId:          cache.GetChannelPeerNodeIdByPublicKey(testutil.TestPublicKey1, core.Bitcoin, core.SigNet),
				ChannelId:       channel.ChannelID,
				Force:           nil,
				TargetConf:      &targetConf,
				DeliveryAddress: nil,
				SatPerVbyte:     &satPerVbyte,
			},
			&lnrpc.CloseChannelRequest{
				ChannelPoint:    nil,
				Force:           false,
				TargetConf:      0,
				DeliveryAddress: "",
				SatPerVbyte:     0,
			},
			true,
		},
		{
			"Just mandatory params",
			CloseChannelRequest{
				NodeId:    cache.GetChannelPeerNodeIdByPublicKey(testutil.TestPublicKey1, core.Bitcoin, core.SigNet),
				ChannelId: channel.ChannelID,
			},
			&lnrpc.CloseChannelRequest{
				ChannelPoint: channelPoint,
			},
			false,
		},
		{
			"All params provide",
			CloseChannelRequest{
				NodeId:          cache.GetChannelPeerNodeIdByPublicKey(testutil.TestPublicKey1, core.Bitcoin, core.SigNet),
				ChannelId:       channel.ChannelID,
				Force:           &force,
				TargetConf:      &targetConf,
				DeliveryAddress: &deliveryAddress,
			},
			&lnrpc.CloseChannelRequest{
				ChannelPoint:    channelPoint,
				Force:           true,
				TargetConf:      12,
				DeliveryAddress: "test",
			},
			false,
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := prepareCloseRequest(test.input)

			if err != nil {
				if test.wantErr {
					return
				}
				t.Errorf("prepareOpenRequest error: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%d: newSendPaymentRequest()\nGot:\n%v\nWant:\n%v\n", i, got, test.want)
			}
		})
	}
}
