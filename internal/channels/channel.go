package channels

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type pendingHtlc struct {
	Incoming          bool   `json:"incoming"`
	Amount            int64  `json:"amount"`
	HashLock          []byte `json:"hashLock"`
	ExpirationHeight  uint32 `json:"expirationHeight"`
	ForwardingChannel uint64 `json:"forwardingChannel"`
}

type channelConstraints struct {
	CsvDelay          uint32 `json:"csvDelay"`
	ChanReserveMsat   uint64 `json:"chanReserveMsat"`
	DustLimitMsat     uint64 `json:"dustLimitMsat"`
	MaxPendingAmtMsat uint64 `json:"maxPendingAmtMsat"`
	MinHtlcMsat       uint64 `json:"minHtlcMsat"`
	MaxAcceptedHtlcs  uint32 `json:"maxAcceptedHtlcs"`
}

type ChannelState struct {
	Active                bool               `json:"active"`
	RemotePubKey          string             `json:"remotePubkey"`
	ChannelPoint          string             `json:"channelPoint"`
	ChanId                uint64             `json:"chanId"`
	Capacity              int64              `json:"capacity"`
	LocalBalance          int64              `json:"localBalance"`
	RemoteBalance         int64              `json:"remoteBalance"`
	CommitFee             int64              `json:"commitFee"`
	CommitWeight          int64              `json:"commitWeight"`
	FeePerKw              int64              `json:"feePerKw"`
	UnsettledBalance      int64              `json:"unsettledBalance"`
	TotalSatoshisSent     int64              `json:"totalSatoshisSent"`
	TotalSatoshisReceived int64              `json:"totalSatoshisReceived"`
	NumUpdates            uint64             `json:"numUpdates"`
	PendingHtlcs          []*pendingHtlc     `json:"pendingHtlcs"`
	CsvDelay              uint32             `json:"csvDelay"`
	Private               bool               `json:"private"`
	Initiator             bool               `json:"initiator"`
	ChanStatusFlags       string             `json:"chanStatusFlags"`
	LocalChanReserveSat   int64              `json:"localChanReserveSat"`
	RemoteChanReserveSat  int64              `json:"remoteChanReserveSat"`
	StaticRemoteKey       bool               `json:"staticRemoteKey"`
	CommitmentType        string             `json:"commitmentType"`
	Lifetime              int64              `json:"lifetime"`
	Uptime                int64              `json:"uptime"`
	CloseAddress          string             `json:"closeAddress"`
	PushAmountSat         uint64             `json:"pushAmountSat"`
	ThawHeight            uint32             `json:"thawHeight"`
	LocalConstraints      channelConstraints `json:"localConstraints"`
	RemoteConstraints     channelConstraints `json:"remoteConstraints"`
}

func getChannelList(db *sqlx.DB) ([]Channel, error) {

	// TODO: change to select which local node
	connectionDetails, err := settings.GetConnectionDetails(db)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting node connection details from the db: %s", err.Error())
		return nil, errors.New("Error getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails[0].GRPCAddress,
		connectionDetails[0].TLSFileBytes,
		connectionDetails[0].MacaroonFileBytes)
	if err != nil {
		log.Error().Err(err).Msgf("can't connect to LND: %s", err.Error())
		return nil, errors.Newf("can't connect to LND %s", err.Error())
	}
	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)

	ctx := context.Background()
	req := lnrpc.ListChannelsRequest{}
	r, err := client.ListChannels(ctx, &req)
	if err != nil {
		return nil, err
	}

}

func getChannelsFromSourceHandler(c *gin.Context, db *sqlx.DB) {

	c.JSON(http.StatusOK, r)
}

type Channel struct {
	// A database primary key. NOT a channel_id as specified in BOLT 2
	ChannelDBID int `json:"channelDBId" db:"channel_db_id"`
	// In the c-lighting and BOLT format e.g. 505580:1917:1
	ShortChannelID string `json:"shortChannelId" db:"short_channel_id"`
	// At the moment only used by LND. Format is "funding tx id : output id"
	LNDChannelPoint   null.String `json:"lndChannelPoint" db:"lnd_channel_point"`
	Alias             null.String `json:"alias" db:"alias"`
	DestinationPubKey null.String `json:"destinationPubKey" db:"destination_pub_key"`
	LocalNodeId       int         `json:"localNodeId" db:"local_node_id"`
	CreatedOn         time.Time   `json:"createdOn" db:"created_on"`
	UpdateOn          null.Time   `json:"updatedOn" db:"updated_on"`
	LNDShortChannelID uint64      `json:"lndShortChannelId" db:"lnd_short_channel_id"`
}

func AddChannelRecordIfDoesntExist(db *sqlx.DB, channel Channel) error {
	dbChannel, err := getChannel(db, channel.ShortChannelID)
	if err != nil {
		return err
	}
	if dbChannel != nil {
		return nil
	}
	err = insertChannel(db, channel)
	if err != nil {
		return err
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
