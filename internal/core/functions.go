package core

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
)

func (s Network) String() string {
	switch s {
	case MainNet:
		return "MainNet"
	case TestNet:
		return "TestNet"
	case RegTest:
		return "RegTest"
	case SigNet:
		return "SigNet"
	case SimNet:
		return "SimNet"
	}
	return UnknownEnumString
}

// GetNetwork defaults to MainNet when no match is found
func GetNetwork(network string) Network {
	switch network {
	case "testnet":
		return TestNet
	case "signet":
		return SigNet
	case "simnet":
		return SimNet
	case "regtest":
		return RegTest
	}
	return MainNet
}

func (s Chain) String() string {
	switch s {
	case Bitcoin:
		return "MainNet"
	case Litecoin:
		return "Litecoin"
	}
	return UnknownEnumString
}

// GetChain defaults to Bitcoin when no match is found
func GetChain(chain string) Chain {
	switch chain {
	case "litecoin":
		return Litecoin
	}
	return Bitcoin
}

// override to get human readable enum
func (s ChannelStatus) String() string {
	switch s {
	case Opening:
		return "Opening"
	case Open:
		return "Open"
	case Closing:
		return "Closing"
	case CooperativeClosed:
		return "Cooperative Closed"
	case LocalForceClosed:
		return "Local Force Closed"
	case RemoteForceClosed:
		return "Remote Force Closed"
	case BreachClosed:
		return "Breach Closed"
	case FundingCancelledClosed:
		return "Funding Cancelled Closed"
	case AbandonedClosed:
		return "Abandoned Closed"
	}
	return UnknownEnumString
}

// NodeConnectionStatus is the status of a node connection.
func (s NodeConnectionSetting) String() string {
	switch s {
	case NodeConnectionSettingAlwaysReconnect:
		return "AlwaysReconnect"
	case NodeConnectionSettingDisableReconnect:
		return "DisableReconnect"
	}
	return UnknownEnumString
}

const mutexLocked = 1

func MutexLocked(m *sync.Mutex) bool {
	state := reflect.ValueOf(m).Elem().FieldByName("state")
	return state.Int()&mutexLocked == mutexLocked
}

func RWMutexWriteLocked(rw *sync.RWMutex) bool {
	// RWMutex has a "w" sync.Mutex field for write lock
	state := reflect.ValueOf(rw).Elem().FieldByName("w").FieldByName("state")
	return state.Int()&mutexLocked == mutexLocked
}

func RWMutexReadLocked(rw *sync.RWMutex) bool {
	return reflect.ValueOf(rw).Elem().FieldByName("readerCount").Int() > 0
}

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
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

func ParseChannelPoint(channelPoint string) (*string, *int) {
	parts := strings.Split(channelPoint, ":")
	if channelPoint != "" && strings.Contains(channelPoint, ":") && len(parts) == 2 {
		outputIndex, err := strconv.Atoi(parts[1])
		if err == nil {
			return &parts[0], &outputIndex
		}
		log.Debug().Err(err).Msgf("Failed to parse channelPoint %v", channelPoint)
	}
	return nil, nil
}

func CreateChannelPoint(fundingTransactionHash string, fundingOutputIndex int) string {
	return fmt.Sprintf("%s:%v", fundingTransactionHash, fundingOutputIndex)
}

func (s *Status) String() string {
	if s == nil {
		return UnknownEnumString
	}
	switch *s {
	case Inactive:
		return "Inactive"
	case Active:
		return "Active"
	case Pending:
		return "Pending"
	case Deleted:
		return "Deleted"
	case Initializing:
		return "Initializing"
	case Archived:
		return "Archived"
	case TimedOut:
		return "TimedOut"
	}
	return UnknownEnumString
}

func (ps PingSystem) AddPingSystem(pingSystem PingSystem) PingSystem {
	return ps | pingSystem
}
func (ps PingSystem) HasPingSystem(pingSystem PingSystem) bool {
	return ps&pingSystem != 0
}
func (ps PingSystem) RemovePingSystem(pingSystem PingSystem) PingSystem {
	return ps & ^pingSystem
}

func (cs NodeConnectionDetailCustomSettings) AddNodeConnectionDetailCustomSettings(
	customSettings NodeConnectionDetailCustomSettings) NodeConnectionDetailCustomSettings {

	return cs | customSettings
}
func (cs NodeConnectionDetailCustomSettings) HasNodeConnectionDetailCustomSettings(
	customSettings NodeConnectionDetailCustomSettings) bool {

	return cs&customSettings != 0
}
func (cs NodeConnectionDetailCustomSettings) RemoveNodeConnectionDetailCustomSettings(
	customSettings NodeConnectionDetailCustomSettings) NodeConnectionDetailCustomSettings {

	return cs & ^customSettings
}

func (cf ChannelFlags) AddChannelFlag(channelFlags ChannelFlags) ChannelFlags {
	return cf | channelFlags
}
func (cf ChannelFlags) HasChannelFlag(channelFlags ChannelFlags) bool {
	return cf&channelFlags != 0
}
func (cf ChannelFlags) RemoveChannelFlag(channelFlags ChannelFlags) ChannelFlags {
	return cf & ^channelFlags
}

func GetImplementations() []Implementation {
	return []Implementation{
		LND,
		CLN,
	}
}

func GetNodeConnectionDetailCustomSettings() []NodeConnectionDetailCustomSettings {
	return []NodeConnectionDetailCustomSettings{
		ImportFailedPayments,
		ImportHtlcEvents,
		ImportTransactions,
		ImportPayments,
		ImportInvoices,
		ImportForwards,
		ImportHistoricForwards,
	}
}

func GetDeltaPerMille(base uint64, amt uint64) int {
	if base > amt {
		return int((base - amt) / base * 1_000)
	} else if base == amt {
		return 0
	} else {
		return int((amt - base) / amt * 1_000)
	}
}

func Sleep(ctx context.Context, d time.Duration) bool {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-ticker.C:
	}
	return true
}
