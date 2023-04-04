package channels

import (
	"time"

	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/pkg/cache"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/core"
)

// GetClosureStatus returns Closing when our API is outdated and a new lnrpc.ChannelCloseSummary_ClosureType is added
func GetClosureStatus(lndClosureType lnrpc.ChannelCloseSummary_ClosureType) core.ChannelStatus {
	switch lndClosureType {
	case lnrpc.ChannelCloseSummary_COOPERATIVE_CLOSE:
		return core.CooperativeClosed
	case lnrpc.ChannelCloseSummary_LOCAL_FORCE_CLOSE:
		return core.LocalForceClosed
	case lnrpc.ChannelCloseSummary_REMOTE_FORCE_CLOSE:
		return core.RemoteForceClosed
	case lnrpc.ChannelCloseSummary_BREACH_CLOSE:
		return core.BreachClosed
	case lnrpc.ChannelCloseSummary_FUNDING_CANCELED:
		return core.FundingCancelledClosed
	case lnrpc.ChannelCloseSummary_ABANDONED:
		return core.AbandonedClosed
	}
	return core.Closing
}

type Channel struct {
	// ChannelID A database primary key. NOT a channel_id as specified in BOLT 2
	ChannelID              int                `json:"channelId" db:"channel_id"`
	Tags                   []tags.Tag         `json:"tags" db:"tags"`
	ShortChannelID         *string            `json:"shortChannelId" db:"short_channel_id"`
	FundingTransactionHash string             `json:"fundingTransactionHash" db:"funding_transaction_hash"`
	FundingOutputIndex     int                `json:"fundingOutputIndex" db:"funding_output_index"`
	ClosingTransactionHash *string            `json:"closingTransactionHash" db:"closing_transaction_hash"`
	LNDShortChannelID      *uint64            `json:"lndShortChannelId" db:"lnd_short_channel_id"`
	Capacity               int64              `json:"capacity" db:"capacity"`
	Private                bool               `json:"private" db:"private"`
	FirstNodeId            int                `json:"firstNodeId" db:"first_node_id"`
	SecondNodeId           int                `json:"secondNodeId" db:"second_node_id"`
	InitiatingNodeId       *int               `json:"initiatingNodeId" db:"initiating_node_id"`
	AcceptingNodeId        *int               `json:"acceptingNodeId" db:"accepting_node_id"`
	ClosingNodeId          *int               `json:"closingNodeId" db:"closing_node_id"`
	Status                 core.ChannelStatus `json:"status" db:"status_id"`
	CreatedOn              time.Time          `json:"createdOn" db:"created_on"`
	UpdateOn               *time.Time         `json:"updatedOn" db:"updated_on"`
	FundingBlockHeight     *uint32            `json:"fundingBlockHeight" db:"funding_block_height"`
	FundedOn               *time.Time         `json:"fundedOn" db:"funded_on"`
	ClosingBlockHeight     *uint32            `json:"closingBlockHeight" db:"closing_block_height"`
	ClosedOn               *time.Time         `json:"closedOn" db:"closed_on"`
	Flags                  core.ChannelFlags  `json:"flags" db:"flags"`
}

func (channel *Channel) AddChannelFlags(flags core.ChannelFlags) {
	channel.Flags = channel.Flags.AddChannelFlag(flags)
}
func (channel *Channel) HasChannelFlags(flags core.ChannelFlags) bool {
	return channel.Flags.HasChannelFlag(flags)
}
func (channel *Channel) RemoveChannelFlags(flags core.ChannelFlags) {
	channel.Flags = channel.Flags.RemoveChannelFlag(flags)
}

func AddChannelOrUpdateChannelStatus(db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache,
	channel Channel) (int, error) {

	var err error
	var existingChannelId int
	if channel.FundingTransactionHash == "" {
		if channel.ShortChannelID == nil || *channel.ShortChannelID == "" || *channel.ShortChannelID == "0x0x0" {
			existingChannelId = 0
		} else {
			existingChannelId = cache.GetChannelIdByShortChannelId(*channel.ShortChannelID)
			if existingChannelId == 0 {
				existingChannelId, err = getChannelIdByShortChannelId(db, channel.ShortChannelID)
				if err != nil {
					return 0, errors.Wrapf(err, "Getting channelId by ShortChannelID %v", channel.ShortChannelID)
				}
			}
		}
		if existingChannelId == 0 {
			return 0, errors.Wrapf(err, "No valid FundingTransactionHash %v, FundingOutputIndex %v",
				channel.FundingTransactionHash, channel.FundingOutputIndex)
		}
	} else {
		existingChannelId = cache.GetChannelIdByFundingTransaction(channel.FundingTransactionHash, channel.FundingOutputIndex)
		if existingChannelId == 0 {
			existingChannelId, err = getChannelIdByFundingTransaction(db, channel.FundingTransactionHash, channel.FundingOutputIndex)
			if err != nil {
				return 0, errors.Wrapf(err, "Getting channelId by FundingTransactionHash %v, FundingOutputIndex %v",
					channel.FundingTransactionHash, channel.FundingOutputIndex)
			}
		}
		if existingChannelId == 0 {
			storedChannel, err := addChannel(db, channel)
			if err != nil {
				return 0, errors.Wrapf(err, "Adding channel FundingTransactionHash %v, FundingOutputIndex %v",
					channel.FundingTransactionHash, channel.FundingOutputIndex)
			}
			return storedChannel.ChannelID, nil
		}
	}
	// existingChannelId is known or it would have aborted/returned already
	existingChannelSettings := cache.GetChannelSettingByChannelId(existingChannelId)
	if existingChannelSettings.ChannelId == 0 {
		existingChannel, err := GetChannel(db, existingChannelId)
		if err != nil {
			return 0, errors.Wrapf(err, "Obtaining existing channel for channelId: %v.", existingChannelId)
		}
		vectorActive := cache.GetVectorUrlBase() != cache.VectorUrl ||
			(nodeSettings.Chain == core.Bitcoin && nodeSettings.Network == core.MainNet)
		switch channel.Status {
		case core.CooperativeClosed, core.LocalForceClosed, core.RemoteForceClosed, core.BreachClosed:
			if channel.ClosingTransactionHash != nil && *channel.ClosingTransactionHash != "" &&
				!existingChannel.HasChannelFlags(core.ClosedOn) &&
				vectorActive {

				vectorResponse := cache.GetTransactionDetailsFromVector(*channel.ClosingTransactionHash, nodeSettings)
				if vectorResponse.BlockHeight != 0 {
					channel.ClosingBlockHeight = &vectorResponse.BlockHeight
					channel.ClosedOn = &vectorResponse.BlockTimestamp
					channel.AddChannelFlags(core.ClosedOn)
				}
			}
			if existingChannel.ClosingBlockHeight == nil || *existingChannel.ClosingBlockHeight == 0 {
				currentBlockHeight := cache.GetBlockHeight()
				channel.ClosingBlockHeight = &currentBlockHeight
				channel.RemoveChannelFlags(core.ClosedOn)
			}
			if existingChannel.ClosedOn == nil {
				now := time.Now().UTC()
				channel.ClosedOn = &now
				channel.RemoveChannelFlags(core.ClosedOn)
			}
			fallthrough
		case core.Open, core.Closing:
			if channel.FundingTransactionHash != "" &&
				!existingChannel.HasChannelFlags(core.FundedOn) &&
				vectorActive {

				vectorResponse := cache.GetTransactionDetailsFromVector(channel.FundingTransactionHash, nodeSettings)
				if vectorResponse.BlockHeight != 0 {
					channel.FundingBlockHeight = &vectorResponse.BlockHeight
					channel.FundedOn = &vectorResponse.BlockTimestamp
					channel.AddChannelFlags(core.FundedOn)
				}
			}
			if existingChannel.FundingBlockHeight == nil || *existingChannel.FundingBlockHeight == 0 {
				currentBlockHeight := cache.GetBlockHeight()
				channel.FundingBlockHeight = &currentBlockHeight
				channel.RemoveChannelFlags(core.FundedOn)
			}
			if existingChannel.FundedOn == nil {
				now := time.Now().UTC()
				channel.FundedOn = &now
				channel.RemoveChannelFlags(core.FundedOn)
			}
		}
		if existingChannel.Status != channel.Status ||
			(existingChannel.ShortChannelID == nil) != (channel.ShortChannelID == nil) ||
			existingChannel.ShortChannelID != nil && channel.ShortChannelID != nil && *existingChannel.ShortChannelID != *channel.ShortChannelID ||
			(existingChannel.LNDShortChannelID == nil) != (channel.LNDShortChannelID == nil) ||
			existingChannel.LNDShortChannelID != nil && channel.LNDShortChannelID != nil && *existingChannel.LNDShortChannelID != *channel.LNDShortChannelID {
			err = updateChannelStatusAndLndIds(db, existingChannelId, channel.Status, channel.ShortChannelID,
				channel.LNDShortChannelID)
			if err != nil {
				return 0, errors.Wrapf(err, "Updating existing channel with FundingTransactionHash %v, FundingOutputIndex %v",
					channel.FundingTransactionHash, channel.FundingOutputIndex)
			}
		}
		if existingChannel.FundingBlockHeight == nil && channel.FundingBlockHeight != nil ||
			existingChannel.FundedOn == nil && channel.FundedOn != nil ||
			!existingChannel.HasChannelFlags(core.FundedOn) && channel.HasChannelFlags(core.FundedOn) {
			err := updateChannelFunding(db, existingChannelId, channel.FundingBlockHeight, channel.FundedOn, channel.Flags)
			if err != nil {
				return 0, errors.Wrapf(err, "Updating channel status and closing transaction hash %v.", existingChannelId)
			}
		}
		if channel.ClosingTransactionHash != nil && (
			existingChannel.ClosingTransactionHash == nil ||
				*existingChannel.ClosingTransactionHash != *channel.ClosingTransactionHash ||
				existingChannel.ClosingBlockHeight == nil && channel.ClosingBlockHeight != nil ||
				existingChannel.ClosedOn == nil && channel.ClosedOn != nil ||
				existingChannel.ClosingNodeId == nil && channel.ClosingNodeId != nil ||
				!existingChannel.HasChannelFlags(core.ClosedOn) && channel.HasChannelFlags(core.ClosedOn)) {
			err := updateChannelClosing(db, existingChannelId,
				*channel.ClosingTransactionHash, channel.ClosingBlockHeight, channel.ClosedOn, channel.ClosingNodeId,
				channel.Flags)
			if err != nil {
				return 0, errors.Wrapf(err, "Updating channel status and closing transaction hash %v.", existingChannelId)
			}
		}
	} else {
		newShortChannelId := ""
		if channel.ShortChannelID != nil {
			newShortChannelId = *channel.ShortChannelID
		}
		newLndShortChannelId := uint64(0)
		if channel.LNDShortChannelID != nil {
			newLndShortChannelId = *channel.LNDShortChannelID
		}
		vectorActive := cache.GetVectorUrlBase() != cache.VectorUrl ||
			(nodeSettings.Chain == core.Bitcoin && nodeSettings.Network == core.MainNet)
		switch channel.Status {
		case core.CooperativeClosed, core.LocalForceClosed, core.RemoteForceClosed, core.BreachClosed:
			if channel.ClosingTransactionHash != nil && *channel.ClosingTransactionHash != "" &&
				!existingChannelSettings.HasChannelFlags(core.ClosedOn) &&
				vectorActive {

				vectorResponse := cache.GetTransactionDetailsFromVector(*channel.ClosingTransactionHash, nodeSettings)
				if vectorResponse.BlockHeight != 0 {
					channel.ClosingBlockHeight = &vectorResponse.BlockHeight
					channel.ClosedOn = &vectorResponse.BlockTimestamp
					channel.AddChannelFlags(core.ClosedOn)
				}
			}
			if existingChannelSettings.ClosingBlockHeight == nil || *existingChannelSettings.ClosingBlockHeight == 0 &&
				(channel.FundingBlockHeight == nil || *channel.FundingBlockHeight == 0) {
				currentBlockHeight := cache.GetBlockHeight()
				channel.ClosingBlockHeight = &currentBlockHeight
				channel.RemoveChannelFlags(core.ClosedOn)
			}
			if existingChannelSettings.ClosedOn == nil && channel.ClosedOn == nil {
				now := time.Now().UTC()
				channel.ClosedOn = &now
				channel.RemoveChannelFlags(core.ClosedOn)
			}
			fallthrough
		case core.Open, core.Closing:
			if channel.FundingTransactionHash != "" &&
				!existingChannelSettings.HasChannelFlags(core.FundedOn) &&
				vectorActive {

				vectorResponse := cache.GetTransactionDetailsFromVector(channel.FundingTransactionHash, nodeSettings)
				if vectorResponse.BlockHeight != 0 {
					channel.FundingBlockHeight = &vectorResponse.BlockHeight
					channel.FundedOn = &vectorResponse.BlockTimestamp
					channel.AddChannelFlags(core.FundedOn)
				}
			}
			if (existingChannelSettings.FundingBlockHeight == nil || *existingChannelSettings.FundingBlockHeight == 0) &&
				(channel.FundingBlockHeight == nil || *channel.FundingBlockHeight == 0) {
				currentBlockHeight := cache.GetBlockHeight()
				channel.FundingBlockHeight = &currentBlockHeight
				channel.RemoveChannelFlags(core.FundedOn)
			}
			if existingChannelSettings.FundedOn == nil && channel.FundedOn == nil {
				now := time.Now().UTC()
				channel.FundedOn = &now
				channel.RemoveChannelFlags(core.FundedOn)
			}
		}
		if existingChannelSettings.Status != channel.Status ||
			existingChannelSettings.ShortChannelId != newShortChannelId ||
			existingChannelSettings.LndShortChannelId != newLndShortChannelId {
			err = updateChannelStatusAndLndIds(db, existingChannelId, channel.Status, channel.ShortChannelID,
				channel.LNDShortChannelID)
			if err != nil {
				return 0, errors.Wrapf(err, "Updating existing channel with FundingTransactionHash %v, FundingOutputIndex %v",
					channel.FundingTransactionHash, channel.FundingOutputIndex)
			}
		}
		if existingChannelSettings.FundingBlockHeight == nil && channel.FundingBlockHeight != nil ||
			existingChannelSettings.FundedOn == nil && channel.FundedOn != nil ||
			!existingChannelSettings.HasChannelFlags(core.FundedOn) && channel.HasChannelFlags(core.FundedOn) {
			err := updateChannelFunding(db, existingChannelId, channel.FundingBlockHeight, channel.FundedOn, channel.Flags)
			if err != nil {
				return 0, errors.Wrapf(err, "Updating channel status and closing transaction hash %v.", existingChannelId)
			}
		}
		if channel.ClosingTransactionHash != nil && (
			existingChannelSettings.ClosingTransactionHash == nil ||
				*existingChannelSettings.ClosingTransactionHash != *channel.ClosingTransactionHash ||
				existingChannelSettings.ClosingBlockHeight == nil && channel.ClosingBlockHeight != nil ||
				existingChannelSettings.ClosedOn == nil && channel.ClosedOn != nil ||
				existingChannelSettings.ClosingNodeId == nil && channel.ClosingNodeId != nil ||
				!existingChannelSettings.HasChannelFlags(core.ClosedOn) && channel.HasChannelFlags(core.ClosedOn)) {
			err := updateChannelClosing(db, existingChannelId,
				*channel.ClosingTransactionHash, channel.ClosingBlockHeight, channel.ClosedOn, channel.ClosingNodeId,
				channel.Flags)
			if err != nil {
				return 0, errors.Wrapf(err, "Updating channel status and closing transaction hash %v.", existingChannelId)
			}
		}
	}
	cache.SetChannel(existingChannelId, channel.ShortChannelID, channel.LNDShortChannelID, channel.Status,
		channel.FundingTransactionHash, channel.FundingOutputIndex,
		channel.FundingBlockHeight, channel.FundedOn,
		channel.Capacity, channel.Private, channel.FirstNodeId, channel.SecondNodeId,
		channel.InitiatingNodeId, channel.AcceptingNodeId,
		channel.ClosingTransactionHash, channel.ClosingNodeId, channel.ClosingBlockHeight, channel.ClosedOn,
		channel.Flags)
	if channel.Status >= core.Closing {
		cache.RemoveChannelStateFromCache(channel.ChannelID)
	}
	return existingChannelId, nil
}

func updateChannelClosing(db *sqlx.DB, channelId int,
	closingTransactionHash string, closingBlockHeight *uint32, closedOn *time.Time, closingNodeId *int,
	flags core.ChannelFlags) error {
	_, err := db.Exec(`
		UPDATE channel
		SET closing_transaction_hash=$1, updated_on=$2, closing_node_id=$4, closing_block_height=$5, closed_on=$6, flags=$7
		WHERE channel_id=$3 AND
		    (
		        (closing_transaction_hash IS NULL OR closing_transaction_hash != $1) OR
		        closing_node_id IS NULL OR closing_node_id != $4 OR
		        closing_block_height IS NULL OR closing_block_height != $5 OR
		        closed_on IS NULL OR closed_on != $6 OR
		        flags != $7
			)`,
		closingTransactionHash, time.Now().UTC(), channelId, closingNodeId, closingBlockHeight, closedOn, flags)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func updateChannelFunding(db *sqlx.DB, channelId int,
	fundingBlockHeight *uint32, fundedOn *time.Time, flags core.ChannelFlags) error {
	_, err := db.Exec(`
		UPDATE channel
		SET updated_on=$1, funding_block_height=$2, funded_on=$3, flags=$4
		WHERE channel_id=$5 AND
		    (
		        funding_block_height IS NULL OR funding_block_height!=$2 OR
		        funded_on IS NULL OR funded_on!=$3 OR
		        flags!=$4
			)`,
		time.Now().UTC(), fundingBlockHeight, fundedOn, flags, channelId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func updateChannelStatusAndLndIds(db *sqlx.DB, channelId int, status core.ChannelStatus, shortChannelId *string,
	lndShortChannelId *uint64) error {
	if shortChannelId != nil && (*shortChannelId == "" || *shortChannelId == "0x0x0") {
		shortChannelId = nil
	}
	if lndShortChannelId != nil && *lndShortChannelId == 0 {
		lndShortChannelId = nil
	}
	_, err := db.Exec(`
		UPDATE channel
		SET status_id=$2, short_channel_id=$3, lnd_short_channel_id=$4, updated_on=$5
		WHERE channel_id=$1 AND (
		    status_id!=$2 OR
		    short_channel_id IS NULL OR short_channel_id!=$3 OR
		    lnd_short_channel_id IS NULL OR lnd_short_channel_id!=$4)`,
		channelId, status, shortChannelId, lndShortChannelId, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}
