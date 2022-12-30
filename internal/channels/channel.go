package channels

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/commons"
)

// GetClosureStatus returns Closing when our API is outdated and a new lnrpc.ChannelCloseSummary_ClosureType is added
func GetClosureStatus(lndClosureType lnrpc.ChannelCloseSummary_ClosureType) commons.ChannelStatus {
	switch lndClosureType {
	case lnrpc.ChannelCloseSummary_COOPERATIVE_CLOSE:
		return commons.CooperativeClosed
	case lnrpc.ChannelCloseSummary_LOCAL_FORCE_CLOSE:
		return commons.LocalForceClosed
	case lnrpc.ChannelCloseSummary_REMOTE_FORCE_CLOSE:
		return commons.RemoteForceClosed
	case lnrpc.ChannelCloseSummary_BREACH_CLOSE:
		return commons.BreachClosed
	case lnrpc.ChannelCloseSummary_FUNDING_CANCELED:
		return commons.FundingCancelledClosed
	case lnrpc.ChannelCloseSummary_ABANDONED:
		return commons.AbandonedClosed
	}
	return commons.Closing
}

type Channel struct {
	// ChannelID A database primary key. NOT a channel_id as specified in BOLT 2
	ChannelID int `json:"channelId" db:"channel_id"`
	// ShortChannelID In the c-lighting and BOLT format e.g. 505580:1917:1
	ShortChannelID         *string               `json:"shortChannelId" db:"short_channel_id"`
	FundingTransactionHash string                `json:"fundingTransactionHash" db:"funding_transaction_hash"`
	FundingOutputIndex     int                   `json:"fundingOutputIndex" db:"funding_output_index"`
	ClosingTransactionHash *string               `json:"closingTransactionHash" db:"closing_transaction_hash"`
	LNDShortChannelID      *uint64               `json:"lndShortChannelId" db:"lnd_short_channel_id"`
	Capacity               int64                 `json:"capacity" db:"capacity"`
	Private                bool                  `json:"private" db:"private"`
	FirstNodeId            int                   `json:"firstNodeId" db:"first_node_id"`
	SecondNodeId           int                   `json:"secondNodeId" db:"second_node_id"`
	InitiatingNodeId       *int                  `json:"initiatingNodeId" db:"initiating_node_id"`
	AcceptingNodeId        *int                  `json:"acceptingNodeId" db:"accepting_node_id"`
	ClosingNodeId          *int                  `json:"closingNodeId" db:"closing_node_id"`
	Status                 commons.ChannelStatus `json:"status" db:"status_id"`
	CreatedOn              time.Time             `json:"createdOn" db:"created_on"`
	UpdateOn               *time.Time            `json:"updatedOn" db:"updated_on"`
}

func AddChannelOrUpdateChannelStatus(db *sqlx.DB, channel Channel) (int, error) {
	var err error
	var existingChannelId int
	if channel.FundingTransactionHash == "" {
		if channel.ShortChannelID == nil || *channel.ShortChannelID == "" || *channel.ShortChannelID == "0x0x0" {
			existingChannelId = 0
		} else {
			existingChannelId = commons.GetChannelIdByShortChannelId(*channel.ShortChannelID)
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
		existingChannelId = commons.GetChannelIdByFundingTransaction(channel.FundingTransactionHash, channel.FundingOutputIndex)
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
	existingChannelSettings := commons.GetChannelSettingByChannelId(existingChannelId)
	if existingChannelSettings.ChannelId == 0 {
		existingChannel, err := GetChannel(db, existingChannelId)
		if err != nil {
			return 0, errors.Wrapf(err, "Obtaining existing channel for channelId: %v.", existingChannelId)
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
		if existingChannel.ClosingTransactionHash == nil && channel.Status >= commons.CooperativeClosed &&
			channel.ClosingTransactionHash != nil {
			err := updateChannelClosingTransactionHash(db, existingChannelId, *channel.ClosingTransactionHash, channel.ClosingNodeId)
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
		if channel.Status >= commons.CooperativeClosed && channel.ClosingTransactionHash != nil {
			err := updateChannelClosingTransactionHash(db, existingChannelId, *channel.ClosingTransactionHash, channel.ClosingNodeId)
			if err != nil {
				return 0, errors.Wrapf(err, "Updating channel status and closing transaction hash %v.", existingChannelId)
			}
		}
	}
	commons.SetChannel(existingChannelId, channel.ShortChannelID, channel.LNDShortChannelID,
		channel.Status, channel.FundingTransactionHash, channel.FundingOutputIndex, channel.Capacity, channel.Private,
		channel.FirstNodeId, channel.SecondNodeId, channel.InitiatingNodeId, channel.AcceptingNodeId)
	return existingChannelId, nil
}

func UpdateChannelStatus(db *sqlx.DB, channelId int, status commons.ChannelStatus) error {
	_, err := db.Exec(`
		UPDATE channel SET status_id=$1, updated_on=$2 WHERE channel_id=$3 AND status_id!=$1`,
		status, time.Now().UTC(), channelId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	commons.SetChannelStatus(channelId, status)
	return nil
}

func updateChannelClosingTransactionHash(db *sqlx.DB, channelId int, closingTransactionHash string, closingNodeId *int) error {
	_, err := db.Exec(`
		UPDATE channel SET closing_transaction_hash=$1, closing_node_id=$4, updated_on=$2 WHERE channel_id=$3 AND
		    ( closing_transaction_hash IS NULL OR closing_transaction_hash != $1 )`,
		closingTransactionHash, time.Now().UTC(), channelId, closingNodeId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func updateChannelStatusAndLndIds(db *sqlx.DB, channelId int, status commons.ChannelStatus, shortChannelId *string,
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
