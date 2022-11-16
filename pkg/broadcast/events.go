package broadcast

import (
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
)

type EventData struct {
	EventTime time.Time `json:"eventTime"`
	NodeId    int       `json:"nodeId"`
}

type GraphEventData struct {
	EventData
	EventNodeId      *int `json:"eventNodeId"`
	AnnouncingNodeId *int `json:"announcingNodeId"`
	ConnectingNodeId *int `json:"connectingNodeId"`
	ChannelId        *int `json:"channelId"`
}

type NodeGraphEventData struct {
	Alias     string `json:"alias"`
	Color     string `json:"color"`
	Addresses string `json:"addresses"`
	Features  string `json:"features"`
}

type ChannelGraphEventData struct {
	Disabled         bool   `json:"disabled"`
	TimeLockDelta    uint32 `json:"timeLockDelta"`
	MinHtlc          int64  `json:"minHtlc"`
	MaxHtlcMsat      uint64 `json:"maxHtlcMsat"`
	FeeBaseMsat      int64  `json:"feeBaseMsat"`
	FeeRateMilliMsat int64  `json:"feeRateMilliMsat"`
}

type NodeGraphEvent struct {
	GraphEventData
	NodeGraphEventData
	PreviousEventTime time.Time          `json:"previousEventTime"`
	PreviousEventData NodeGraphEventData `json:"previous"`
}

type ChannelGraphEvent struct {
	GraphEventData
	ChannelGraphEventData
	PreviousEventTime time.Time             `json:"previousEventTime"`
	PreviousEventData ChannelGraphEventData `json:"previous"`
}

type TransactionEvent struct {
	EventData
	Timestamp             time.Time `json:"timestamp"`
	TransactionHash       *string   `json:"transactionHash"`
	Amount                *int64    `json:"amount"`
	NumberOfConfirmations *int32    `json:"numberOfConfirmations"`
	BlockHash             *string   `json:"blockHash"`
	BlockHeight           *int32    `json:"blockHeight"`
	TotalFees             *int64    `json:"totalFees"`
	DestinationAddresses  *[]string `json:"destinationAddresses"`
	RawTransactionHex     *string   `json:"rawTransactionHex"`
	Label                 *string   `json:"label"`
	NodeId                int       `json:"nodeId"`
}

type ChannelEvent struct {
	EventData
	Type      lnrpc.ChannelEventUpdate_UpdateType `json:"type"`
	ChannelId int                                 `json:"channelId"`
}

type InvoiceEvent struct {
	EventData
	AddIndex    uint64                     `json:"addIndex"`
	ValueMSat   int64                      `json:"valueMSat"`
	State       lnrpc.Invoice_InvoiceState `json:"state"`
	AmountPaid  int64                      `json:"amountPaid"`
	SettledDate time.Time                  `json:"settledDate"`
}

type PeerEvent struct {
	EventData
	Type           lnrpc.PeerEvent_EventType `json:"type"`
	EventPublicKey string                    `json:"eventPublicKey"`
}

type HtlcEvent struct {
	EventData
	Timestamp         time.Time `json:"timestamp"`
	Data              string    `json:"data"`
	EventOrigin       *string   `json:"eventOrigin"`
	EventType         *string   `json:"eventType"`
	OutgoingHtlcId    *uint64   `json:"outgoingHtlcId"`
	IncomingHtlcId    *uint64   `json:"incomingHtlcId"`
	TimestampNs       *uint64   `json:"timestampNs"`
	IncomingAmtMsat   *uint64   `json:"incomingAmtMsat"`
	OutgoingAmtMsat   *uint64   `json:"outgoingAmtMsat"`
	IncomingTimelock  *uint32   `json:"incomingTimelock"`
	OutgoingTimelock  *uint32   `json:"outgoingTimelock"`
	BoltFailureCode   *string   `json:"boltFailureCode"`
	BoltFailureString *string   `json:"boltFailureString"`
	LndFailureDetail  *string   `json:"lndFailureDetail"`
	OutgoingChannelId *int      `json:"outgoingChannelId"`
	IncomingChannelId *int      `json:"incomingChannelId"`
}
