package channels

import (
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

type Status int

const (
	Opening = Status(iota)
	Open
	Closing
	CooperativeClosed      = 100
	LocalForceClosed       = 101
	RemoteForceClosed      = 102
	BreachClosed           = 103
	FundingCancelledClosed = 104
	AbandonedClosed        = 105
)

// GetClosureStatus returns Closing when our API is outdated and a new lnrpc.ChannelCloseSummary_ClosureType is added
func GetClosureStatus(lndClosureType lnrpc.ChannelCloseSummary_ClosureType) Status {
	switch lndClosureType {
	case lnrpc.ChannelCloseSummary_COOPERATIVE_CLOSE:
		return CooperativeClosed
	case lnrpc.ChannelCloseSummary_LOCAL_FORCE_CLOSE:
		return LocalForceClosed
	case lnrpc.ChannelCloseSummary_REMOTE_FORCE_CLOSE:
		return RemoteForceClosed
	case lnrpc.ChannelCloseSummary_BREACH_CLOSE:
		return BreachClosed
	case lnrpc.ChannelCloseSummary_FUNDING_CANCELED:
		return FundingCancelledClosed
	case lnrpc.ChannelCloseSummary_ABANDONED:
		return AbandonedClosed
	}
	return Closing
}

type Channel struct {
	// ChannelDBID A database primary key. NOT a channel_id as specified in BOLT 2
	ChannelDBID int `json:"channelDBId" db:"channel_id"`
	// ShortChannelID In the c-lighting and BOLT format e.g. 505580:1917:1
	ShortChannelID string `json:"shortChannelId" db:"short_channel_id"`
	// LNDChannelPoint At the moment only used by LND. Format is "funding tx id : output id"
	LNDChannelPoint   null.String `json:"lndChannelPoint" db:"lnd_channel_point"`
	FirstNodeId       int         `json:"firstNodeId" db:"first_node_id"`
	SecondNodeId      int         `json:"secondNodeId" db:"second_node_id"`
	CreatedOn         time.Time   `json:"createdOn" db:"created_on"`
	UpdateOn          null.Time   `json:"updatedOn" db:"updated_on"`
	LNDShortChannelID uint64      `json:"lndShortChannelId" db:"lnd_short_channel_id"`
	Status            Status      `json:"status" db:"status_id"`
}

func AddChannelOrUpdateChannelStatus(db *sqlx.DB, channel Channel) (int, error) {
	if channel.ShortChannelID == "" {
		// This is a new channel that is in a pending state
		channel.Status = Opening
		// The channelPoint should be available (for LND)
		existingChannelId := commons.GetChannelIdFromChannelPoint(channel.LNDChannelPoint.String)
		var err error
		if existingChannelId == 0 {
			existingChannelId, err = getChannelIdByLndChannelPoint(db, channel.LNDChannelPoint.String)
			if err != nil {
				return 0, errors.Wrapf(err, "Getting channelId by LNDChannelPoint %v", channel.LNDChannelPoint.String)
			}
		}
		if existingChannelId == 0 {
			storedChannel, err := addChannel(db, channel)
			if err != nil {
				return 0, errors.Wrapf(err, "Adding channel channelPoint: %v", channel.LNDChannelPoint.String)
			}
			commons.SetOpeningChannel(storedChannel.ChannelDBID, storedChannel.LNDChannelPoint.String, int(Opening))
			return storedChannel.ChannelDBID, nil
		} else {
			log.Error().Msgf("Impossible cache miss (except for torq bootstap)!!! channelPoint: %v", channel.LNDChannelPoint.String)
			err = UpdateChannelStatus(db, existingChannelId, channel.Status)
			if err != nil {
				return 0, errors.Wrapf(err, "Updating channel status %v", existingChannelId)
			}
			return existingChannelId, nil
		}
	} else {
		var err error
		existingChannelId := commons.GetChannelIdFromShortChannelId(channel.ShortChannelID)
		if existingChannelId == 0 {
			existingChannelId, err = getChannelIdByShortChannelId(db, channel.ShortChannelID)
			if err != nil {
				return 0, errors.Wrapf(err, "Getting channelId by ShortChannelID %v", channel.ShortChannelID)
			}
		}
		if existingChannelId == 0 {
			// The channelPoint should be available (for LND)
			if channel.LNDChannelPoint.Valid {
				existingChannelId = commons.GetChannelIdFromChannelPoint(channel.LNDChannelPoint.String)
				if existingChannelId == 0 {
					existingChannelId, err = getChannelIdByLndChannelPoint(db, channel.LNDChannelPoint.String)
					if err != nil {
						return 0, errors.Wrapf(err, "Getting channelId by LNDChannelPoint %v",
							channel.LNDChannelPoint.String)
					}
				}
				if existingChannelId == 0 {
					storedChannel, err := addChannel(db, channel)
					if err != nil {
						return 0, errors.Wrapf(err, "Adding channel channelPoint: %v",
							channel.LNDChannelPoint.String)
					}
					if channel.Status == Opening || channel.Status == Open || channel.Status == Closing {
						commons.SetChannel(storedChannel.ChannelDBID, storedChannel.ShortChannelID,
							int(storedChannel.Status), storedChannel.LNDChannelPoint.String)
					}
					return storedChannel.ChannelDBID, nil
				} else {
					err = updateChannelStatusAndLndIds(db, existingChannelId, channel.Status, channel.ShortChannelID,
						channel.LNDShortChannelID)
					if err != nil {
						return 0, errors.Wrapf(err, "Updating existing channel with channelPoint %v",
							channel.LNDChannelPoint.String)
					}
					return existingChannelId, nil
				}
			} else {
				return 0, errors.Wrapf(err, "No valid channelPoint (%v) found", channel.LNDChannelPoint.String)
			}
		} else {
			statusId := commons.GetChannelStatusIdFromChannelId(existingChannelId)
			if Status(statusId) != channel.Status {
				err := UpdateChannelStatus(db, existingChannelId, channel.Status)
				if err != nil {
					return 0, errors.Wrap(err, "Updating channel status.")
				}
			}
			return existingChannelId, nil
		}
	}
}

func UpdateChannelStatus(db *sqlx.DB, channelId int, status Status) error {
	_, err := db.Exec(`
		UPDATE channel SET status_id=$1, updated_on=$2 WHERE channel_id=$3 AND status_id!=$1`,
		status, time.Now().UTC(), channelId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	if status == Opening || status == Open || status == Closing {
		commons.SetChannelStatus(channelId, int(status))
	} else {
		commons.RemoveChannel(channelId)
	}
	return nil
}

func updateChannelStatusAndLndIds(db *sqlx.DB, channelId int, status Status, shortChannelId string,
	lndShortChannelId uint64) error {
	_, err := db.Exec(`
		UPDATE channel
		SET status_id=$2, short_channel_id=$3, lnd_short_channel_id=$4, updated_on=$5
		WHERE channel_id=$1`,
		channelId, status, shortChannelId, lndShortChannelId, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	if status == Opening || status == Open || status == Closing {
		commons.SetChannelStatus(channelId, int(status))
	} else {
		commons.RemoveChannel(channelId)
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
