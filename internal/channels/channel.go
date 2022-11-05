package channels

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"gopkg.in/guregu/null.v4"

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

func ParseChannelPoint(channelPoint string) (string, int) {
	parts := strings.Split(channelPoint, ":")
	if channelPoint != "" && strings.Contains(channelPoint, ":") && len(parts) == 2 {
		outputIndex, err := strconv.Atoi(parts[1])
		if err == nil {
			return parts[0], outputIndex
		} else {
			log.Debug().Err(err).Msgf("Failed to parse channelPoint %v", channelPoint)
		}
	}
	return "", 0
}

func CreateChannelPoint(fundingTransactionHash string, fundingOutputIndex int) string {
	return fmt.Sprintf("%s:%v", fundingTransactionHash, fundingOutputIndex)
}

type Channel struct {
	// ChannelID A database primary key. NOT a channel_id as specified in BOLT 2
	ChannelID int `json:"channelId" db:"channel_id"`
	// ShortChannelID In the c-lighting and BOLT format e.g. 505580:1917:1
	ShortChannelID         *string               `json:"shortChannelId" db:"short_channel_id"`
	FundingTransactionHash string                `json:"fundingTransactionHash" db:"funding_transaction_hash"`
	FundingOutputIndex     int                   `json:"fundingOutputIndex" db:"funding_output_index"`
	ClosingTransactionHash *string               `json:"closingTransactionHash" db:"closing_transaction_hash"`
	FirstNodeId            int                   `json:"firstNodeId" db:"first_node_id"`
	SecondNodeId           int                   `json:"secondNodeId" db:"second_node_id"`
	CreatedOn              time.Time             `json:"createdOn" db:"created_on"`
	UpdateOn               null.Time             `json:"updatedOn" db:"updated_on"`
	LNDShortChannelID      *uint64               `json:"lndShortChannelId" db:"lnd_short_channel_id"`
	Status                 commons.ChannelStatus `json:"status" db:"status_id"`
}

func AddChannelOrUpdateChannelStatus(db *sqlx.DB, channel Channel) (int, error) {
	if channel.ShortChannelID == nil || *channel.ShortChannelID == "" || *channel.ShortChannelID == "0x0x0" {
		// This is a new channel that is in a pending state
		channel.Status = commons.Opening
		// The channelPoint should be available (for LND)
		existingChannelId := commons.GetChannelIdFromFundingTransaction(channel.FundingTransactionHash, channel.FundingOutputIndex)
		var err error
		if existingChannelId == 0 {
			existingChannelId, err = getChannelIdByFundingTransaction(db, channel.FundingTransactionHash, channel.FundingOutputIndex)
			if err != nil {
				return 0, errors.Wrapf(err,
					"Getting channelId by FundingTransactionHash %v, FundingOutputIndex %v",
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
		} else {
			log.Error().Msgf("Impossible cache miss (except for torq bootstap)!!! FundingTransactionHash %v, FundingOutputIndex %v",
				channel.FundingTransactionHash, channel.FundingOutputIndex)
			if channel.Status >= commons.CooperativeClosed && channel.ClosingTransactionHash != nil {
				err := updateChannelStatusAndClosingTransactionHash(db, existingChannelId, channel.Status, *channel.ClosingTransactionHash)
				if err != nil {
					return 0, errors.Wrapf(err, "Updating channel status and closing transaction hash %v.", existingChannelId)
				}
			} else {
				err = UpdateChannelStatus(db, existingChannelId, channel.Status)
				if err != nil {
					return 0, errors.Wrapf(err, "Updating channel status %v", existingChannelId)
				}
			}
			return existingChannelId, nil
		}
	} else {
		var err error
		existingChannelId := commons.GetChannelIdFromShortChannelId(*channel.ShortChannelID)
		if existingChannelId == 0 {
			existingChannelId, err = getChannelIdByShortChannelId(db, channel.ShortChannelID)
			if err != nil {
				return 0, errors.Wrapf(err, "Getting channelId by ShortChannelID %v", channel.ShortChannelID)
			}
		}
		if existingChannelId == 0 {
			if channel.FundingTransactionHash != "" {
				existingChannelId = commons.GetChannelIdFromFundingTransaction(channel.FundingTransactionHash, channel.FundingOutputIndex)
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
				} else {
					err = updateChannelStatusAndLndIds(db, existingChannelId, channel.Status, channel.ShortChannelID,
						channel.LNDShortChannelID)
					if err != nil {
						return 0, errors.Wrapf(err, "Updating existing channel with FundingTransactionHash %v, FundingOutputIndex %v",
							channel.FundingTransactionHash, channel.FundingOutputIndex)
					}
					commons.SetChannel(existingChannelId, channel.ShortChannelID,
						channel.Status, channel.FundingTransactionHash, channel.FundingOutputIndex)
					if channel.Status >= commons.CooperativeClosed && channel.ClosingTransactionHash != nil {
						err := updateChannelStatusAndClosingTransactionHash(db, existingChannelId, channel.Status, *channel.ClosingTransactionHash)
						if err != nil {
							return 0, errors.Wrapf(err, "Updating channel status and closing transaction hash %v.", existingChannelId)
						}
					}
					return existingChannelId, nil
				}
			} else {
				return 0, errors.Wrapf(err, "No valid FundingTransactionHash %v, FundingOutputIndex %v",
					channel.FundingTransactionHash, channel.FundingOutputIndex)
			}
		} else {
			status := commons.GetChannelStatusFromChannelId(existingChannelId)
			if status != channel.Status {
				if channel.Status >= commons.CooperativeClosed && channel.ClosingTransactionHash != nil {
					err := updateChannelStatusAndClosingTransactionHash(db, existingChannelId, channel.Status, *channel.ClosingTransactionHash)
					if err != nil {
						return 0, errors.Wrapf(err, "Updating channel status and closing transaction hash %v.", existingChannelId)
					}
				} else {
					err := UpdateChannelStatus(db, existingChannelId, channel.Status)
					if err != nil {
						return 0, errors.Wrapf(err, "Updating channel status %v.", existingChannelId)
					}
				}
			}
			return existingChannelId, nil
		}
	}
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

func updateChannelStatusAndClosingTransactionHash(db *sqlx.DB, channelId int, status commons.ChannelStatus, closingTransactionHash string) error {
	_, err := db.Exec(`
		UPDATE channel SET status_id=$1, closing_transaction_hash=$2, updated_on=$3 WHERE channel_id=$4 AND (status_id!=$1 OR closing_transaction_hash!=$2)`,
		status, closingTransactionHash, time.Now().UTC(), channelId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	commons.SetChannelStatus(channelId, status)
	return nil
}

func updateChannelStatusAndLndIds(db *sqlx.DB, channelId int, status commons.ChannelStatus, shortChannelId *string,
	lndShortChannelId *uint64) error {
	if shortChannelId != nil && *shortChannelId == "" || *shortChannelId == "0x0x0" {
		shortChannelId = nil
	}
	if lndShortChannelId != nil && *lndShortChannelId == 0 {
		lndShortChannelId = nil
	}
	_, err := db.Exec(`
		UPDATE channel
		SET status_id=$2, short_channel_id=$3, lnd_short_channel_id=$4, updated_on=$5
		WHERE channel_id=$1`,
		channelId, status, shortChannelId, lndShortChannelId, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func ConvertLNDShortChannelID(LNDShortChannelID uint64) string {
	blockHeight := uint32(LNDShortChannelID >> 40)
	txIndex := uint32(LNDShortChannelID>>16) & 0xFFFFFF
	outputIndex := uint16(LNDShortChannelID)
	return strconv.FormatUint(uint64(blockHeight), 10) +
		"x" + strconv.FormatUint(uint64(txIndex), 10) +
		"x" + strconv.FormatUint(uint64(outputIndex), 10)
}

func ConvertShortChannelIDToLND(ShortChannelID string) (uint64, error) {
	parts := strings.Split(ShortChannelID, "x")
	blockHeight, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, errors.Wrap(err, "Converting block height from string to int")
	}
	txIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, errors.Wrap(err, "Converting tx index from string to int")
	}
	txPosition, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, errors.Wrap(err, "Converting tx position from string to int")
	}

	return (uint64(blockHeight) << 40) |
		(uint64(txIndex) << 16) |
		(uint64(txPosition)), nil
}
