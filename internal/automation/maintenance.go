package automation

import (
	"context"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/cache"
	"github.com/lncapital/torq/pkg/core"
	"github.com/lncapital/torq/pkg/vector"
)

const maintenanceQueueTickerSeconds = 60 * 60
const maintenanceVectorDelayMilliseconds = 500

func MaintenanceServiceStart(ctx context.Context, db *sqlx.DB) {

	ticker := clock.New().Tick(maintenanceQueueTickerSeconds * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			// TODO get forwards/invoices/payments without firstNodeId/secondNodeId/nodeId and assign correctly
			processMissingChannelData(db)
		}
	}
}

func processMissingChannelData(db *sqlx.DB) {
	torqNodeIds := cache.GetAllTorqNodeIds()
	for _, torqNodeId := range torqNodeIds {
		nodeSettings := cache.GetNodeSettingsByNodeId(torqNodeId)
		if cache.GetVectorUrlBase() == vector.VectorUrl && (nodeSettings.Chain != core.Bitcoin || nodeSettings.Network != core.MainNet) {
			log.Info().Msgf("Skipping verification of funding and closing details from vector for nodeId: %v", nodeSettings.NodeId)
			continue
		}
		channelSettings := cache.GetChannelSettingsByNodeId(torqNodeId)
		for _, channelSetting := range channelSettings {
			if hasMissingClosingDetails(channelSetting) {
				transactionDetails := vector.GetTransactionDetailsFromVector(*channelSetting.ClosingTransactionHash, nodeSettings)
				err := updateClosingDetails(db, channelSetting, transactionDetails)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to update closing details from vector for channelId: %v", channelSetting.ChannelId)
				}
				time.Sleep(maintenanceVectorDelayMilliseconds * time.Millisecond)
			}
			if hasMissingFundingDetails(channelSetting) {
				transactionDetails := vector.GetTransactionDetailsFromVector(channelSetting.FundingTransactionHash, nodeSettings)
				err := updateFundingDetails(db, channelSetting, transactionDetails)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to update funding details from vector for channelId: %v", channelSetting.ChannelId)
				}
				time.Sleep(maintenanceVectorDelayMilliseconds * time.Millisecond)
			}
		}
	}
}

func hasMissingClosingDetails(channelSetting cache.ChannelSettingsCache) bool {
	if channelSetting.Status == core.Opening {
		return false
	}
	if channelSetting.Status == core.Open {
		return false
	}
	if channelSetting.Status == core.FundingCancelledClosed {
		return false
	}
	if channelSetting.Status == core.AbandonedClosed {
		return false
	}
	if channelSetting.ClosingTransactionHash != nil && *channelSetting.ClosingTransactionHash != "" {
		return !channelSetting.HasChannelFlags(core.ClosedOn)
	}
	return false
}

func updateClosingDetails(db *sqlx.DB, channel cache.ChannelSettingsCache, transactionDetails vector.TransactionDetailsHttpResponse) error {
	if transactionDetails.BlockHeight != 0 {
		channel.ClosedOn = &transactionDetails.BlockTimestamp
		channel.ClosingBlockHeight = &transactionDetails.BlockHeight
		channel.AddChannelFlags(core.ClosedOn)
		_, err := db.Exec(`
		UPDATE channel
		SET closing_block_height=$2, closed_on=$3, flags=$4, updated_on=$5
		WHERE channel_id=$1;`,
			channel.ChannelId, channel.ClosingBlockHeight, channel.ClosedOn, channel.Flags, time.Now().UTC())
		if err != nil {
			return errors.Wrap(err, database.SqlExecutionError)
		}
		cache.SetChannel(channel.ChannelId, &channel.ShortChannelId, &channel.LndShortChannelId, channel.Status,
			channel.FundingTransactionHash, channel.FundingOutputIndex,
			channel.FundingBlockHeight, channel.FundedOn,
			channel.Capacity, channel.Private, channel.FirstNodeId, channel.SecondNodeId,
			channel.InitiatingNodeId, channel.AcceptingNodeId,
			channel.ClosingTransactionHash, channel.ClosingNodeId, channel.ClosingBlockHeight, channel.ClosedOn,
			channel.Flags)
	}
	return nil
}

func hasMissingFundingDetails(channelSetting cache.ChannelSettingsCache) bool {
	if channelSetting.Status == core.Opening {
		return false
	}
	if channelSetting.Status == core.FundingCancelledClosed {
		return false
	}
	if channelSetting.Status == core.AbandonedClosed {
		return false
	}
	if channelSetting.FundingTransactionHash != "" {
		return !channelSetting.HasChannelFlags(core.FundedOn)
	}
	return false
}

func updateFundingDetails(db *sqlx.DB, channel cache.ChannelSettingsCache, transactionDetails vector.TransactionDetailsHttpResponse) error {
	if transactionDetails.BlockHeight != 0 {
		channel.FundedOn = &transactionDetails.BlockTimestamp
		channel.FundingBlockHeight = &transactionDetails.BlockHeight
		channel.AddChannelFlags(core.FundedOn)
		_, err := db.Exec(`
		UPDATE channel
		SET funding_block_height=$2, funded_on=$3, flags=$4, updated_on=$5
		WHERE channel_id=$1;`,
			channel.ChannelId, channel.FundingBlockHeight, channel.FundedOn, channel.Flags, time.Now().UTC())
		if err != nil {
			return errors.Wrap(err, database.SqlExecutionError)
		}
		cache.SetChannel(channel.ChannelId, &channel.ShortChannelId, &channel.LndShortChannelId, channel.Status,
			channel.FundingTransactionHash, channel.FundingOutputIndex,
			channel.FundingBlockHeight, channel.FundedOn,
			channel.Capacity, channel.Private, channel.FirstNodeId, channel.SecondNodeId,
			channel.InitiatingNodeId, channel.AcceptingNodeId,
			channel.ClosingTransactionHash, channel.ClosingNodeId, channel.ClosingBlockHeight, channel.ClosedOn,
			channel.Flags)
	}
	return nil
}
