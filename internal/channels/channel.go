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

	"github.com/lncapital/torq/pkg/commons"
)

//type pendingHtlc struct {
//	Incoming          bool   `json:"incoming"`
//	Amount            int64  `json:"amount"`
//	HashLock          []byte `json:"hashLock"`
//	ExpirationHeight  uint32 `json:"expirationHeight"`
//	ForwardingChannel uint64 `json:"forwardingChannel"`
//}
//
//type channelConstraints struct {
//	CsvDelay          uint32 `json:"csvDelay"`
//	ChanReserveMsat   uint64 `json:"chanReserveMsat"`
//	DustLimitMsat     uint64 `json:"dustLimitMsat"`
//	MaxPendingAmtMsat uint64 `json:"maxPendingAmtMsat"`
//	MinHtlcMsat       uint64 `json:"minHtlcMsat"`
//	MaxAcceptedHtlcs  uint32 `json:"maxAcceptedHtlcs"`
//}
//
//type ChannelState struct {
//	Active                bool               `json:"active"`
//	RemotePubKey          string             `json:"remotePubkey"`
//	ChannelPoint          string             `json:"channelPoint"`
//	ChanId                uint64             `json:"chanId"`
//	Capacity              int64              `json:"capacity"`
//	LocalBalance          int64              `json:"localBalance"`
//	RemoteBalance         int64              `json:"remoteBalance"`
//	CommitFee             int64              `json:"commitFee"`
//	CommitWeight          int64              `json:"commitWeight"`
//	FeePerKw              int64              `json:"feePerKw"`
//	UnsettledBalance      int64              `json:"unsettledBalance"`
//	TotalSatoshisSent     int64              `json:"totalSatoshisSent"`
//	TotalSatoshisReceived int64              `json:"totalSatoshisReceived"`
//	NumUpdates            uint64             `json:"numUpdates"`
//	PendingHtlcs          []*pendingHtlc     `json:"pendingHtlcs"`
//	CsvDelay              uint32             `json:"csvDelay"`
//	Private               bool               `json:"private"`
//	Initiator             bool               `json:"initiator"`
//	ChanStatusFlags       string             `json:"chanStatusFlags"`
//	LocalChanReserveSat   int64              `json:"localChanReserveSat"`
//	RemoteChanReserveSat  int64              `json:"remoteChanReserveSat"`
//	StaticRemoteKey       bool               `json:"staticRemoteKey"`
//	CommitmentType        string             `json:"commitmentType"`
//	Lifetime              int64              `json:"lifetime"`
//	Uptime                int64              `json:"uptime"`
//	CloseAddress          string             `json:"closeAddress"`
//	PushAmountSat         uint64             `json:"pushAmountSat"`
//	ThawHeight            uint32             `json:"thawHeight"`
//	LocalConstraints      channelConstraints `json:"localConstraints"`
//	RemoteConstraints     channelConstraints `json:"remoteConstraints"`
//}
//

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
	ChannelDBID int `json:"channelDBId" db:"channel_db_id"`
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
	existingChannelId := commons.GetChannelIdFromShortChannelId(channel.ShortChannelID)
	if existingChannelId == 0 {
		channelId, err := getChannelIdByShortChannelId(db, channel.ShortChannelID)
		if err != nil {
			return 0, errors.Wrap(err, "Getting channelId by ShortChannelID.")
		}
		if channelId == 0 {
			storedChannel, err := addChannel(db, channel)
			if err != nil {
				return 0, errors.Wrap(err, "Adding channel.")
			}
			channelId = storedChannel.ChannelDBID
		} else {
			log.Error().Msgf("Impossible cache miss (except for torq bootstap)!!! shortChannelId: %v", channel.ShortChannelID)
			err = updateChannelStatus(db, channelId, channel.Status)
			if err != nil {
				return 0, errors.Wrap(err, "Updating channel status.")
			}
		}
		return channelId, nil
	} else {
		statusId := commons.GetChannelStatusIdFromChannelId(existingChannelId)
		if Status(statusId) != Open {
			err := UpdateChannelStatus(db, existingChannelId, Open)
			if err != nil {
				return 0, errors.Wrap(err, "Updating channel status.")
			}
		}
		return existingChannelId, nil
	}
}

func UpdateChannelStatus(db *sqlx.DB, channelId int, status Status) error {
	err := updateChannelStatus(db, channelId, status)
	if err != nil {
		return errors.Wrap(err, "Updating channel status.")
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
