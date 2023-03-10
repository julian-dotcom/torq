package commons

import (
	"context"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/database"
)

const maintenanceQueueTickerSeconds = 60 * 60
const maintenanceVectorDelayMilliseconds = 500

func MaintenanceServiceStart(ctx context.Context, db *sqlx.DB) {

	defer log.Info().Msgf("MaintenanceService terminated")

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
	torqNodeIds := GetAllTorqNodeIds()
	for _, torqNodeId := range torqNodeIds {
		nodeSettings := GetNodeSettingsByNodeId(torqNodeId)
		if GetVectorUrlBase() == VectorUrl && (nodeSettings.Chain != Bitcoin || nodeSettings.Network != MainNet) {
			log.Info().Msgf("Skipping verification of funding and closing details from vector for nodeId: %v", nodeSettings.NodeId)
			continue
		}
		channelSettings := GetChannelSettingsByNodeId(torqNodeId)
		for _, channelSetting := range channelSettings {
			if hasMissingClosingDetails(channelSetting) {
				transactionDetails := GetTransactionDetailsFromVector(*channelSetting.ClosingTransactionHash, nodeSettings)
				err := updateClosingDetails(db, channelSetting, transactionDetails)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to update closing details from vector for channelId: %v", channelSetting.ChannelId)
				}
				time.Sleep(maintenanceVectorDelayMilliseconds * time.Millisecond)
			}
			if hasMissingFundingDetails(channelSetting) {
				transactionDetails := GetTransactionDetailsFromVector(channelSetting.FundingTransactionHash, nodeSettings)
				err := updateFundingDetails(db, channelSetting, transactionDetails)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to update funding details from vector for channelId: %v", channelSetting.ChannelId)
				}
				time.Sleep(maintenanceVectorDelayMilliseconds * time.Millisecond)
			}
		}
	}
}

func hasMissingClosingDetails(channelSetting ManagedChannelSettings) bool {
	if channelSetting.Status == Opening {
		return false
	}
	if channelSetting.Status == Open {
		return false
	}
	if channelSetting.Status == FundingCancelledClosed {
		return false
	}
	if channelSetting.Status == AbandonedClosed {
		return false
	}
	if channelSetting.ClosingTransactionHash != nil && *channelSetting.ClosingTransactionHash != "" {
		return !channelSetting.HasChannelFlags(ClosedOn)
	}
	return false
}

func updateClosingDetails(db *sqlx.DB, channel ManagedChannelSettings, transactionDetails TransactionDetailsHttpResponse) error {
	if transactionDetails.BlockHeight != 0 {
		channel.ClosedOn = &transactionDetails.BlockTimestamp
		channel.ClosingBlockHeight = &transactionDetails.BlockHeight
		channel.AddChannelFlags(ClosedOn)
		_, err := db.Exec(`
		UPDATE channel
		SET closing_block_height=$2, closed_on=$3, flags=$4, updated_on=$5
		WHERE channel_id=$1;`,
			channel.ChannelId, channel.ClosingBlockHeight, channel.ClosedOn, channel.Flags, time.Now().UTC())
		if err != nil {
			return errors.Wrap(err, database.SqlExecutionError)
		}
		SetChannel(channel.ChannelId, &channel.ShortChannelId, &channel.LndShortChannelId, channel.Status,
			channel.FundingTransactionHash, channel.FundingOutputIndex,
			channel.FundingBlockHeight, channel.FundedOn,
			channel.Capacity, channel.Private, channel.FirstNodeId, channel.SecondNodeId,
			channel.InitiatingNodeId, channel.AcceptingNodeId,
			channel.ClosingTransactionHash, channel.ClosingNodeId, channel.ClosingBlockHeight, channel.ClosedOn,
			channel.Flags)
	}
	return nil
}

func hasMissingFundingDetails(channelSetting ManagedChannelSettings) bool {
	if channelSetting.Status == Opening {
		return false
	}
	if channelSetting.Status == FundingCancelledClosed {
		return false
	}
	if channelSetting.Status == AbandonedClosed {
		return false
	}
	if channelSetting.FundingTransactionHash != "" {
		return !channelSetting.HasChannelFlags(FundedOn)
	}
	return false
}

func updateFundingDetails(db *sqlx.DB, channel ManagedChannelSettings, transactionDetails TransactionDetailsHttpResponse) error {
	if transactionDetails.BlockHeight != 0 {
		channel.FundedOn = &transactionDetails.BlockTimestamp
		channel.FundingBlockHeight = &transactionDetails.BlockHeight
		channel.AddChannelFlags(FundedOn)
		_, err := db.Exec(`
		UPDATE channel
		SET funding_block_height=$2, funded_on=$3, flags=$4, updated_on=$5
		WHERE channel_id=$1;`,
			channel.ChannelId, channel.FundingBlockHeight, channel.FundedOn, channel.Flags, time.Now().UTC())
		if err != nil {
			return errors.Wrap(err, database.SqlExecutionError)
		}
		SetChannel(channel.ChannelId, &channel.ShortChannelId, &channel.LndShortChannelId, channel.Status,
			channel.FundingTransactionHash, channel.FundingOutputIndex,
			channel.FundingBlockHeight, channel.FundedOn,
			channel.Capacity, channel.Private, channel.FirstNodeId, channel.SecondNodeId,
			channel.InitiatingNodeId, channel.AcceptingNodeId,
			channel.ClosingTransactionHash, channel.ClosingNodeId, channel.ClosingBlockHeight, channel.ClosedOn,
			channel.Flags)
	}
	return nil
}
