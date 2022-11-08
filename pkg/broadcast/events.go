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
	Timestamp time.Time `json:"timestamp"`
	Amount    int64     `json:"amount"`
	TotalFees int64     `json:"totalFees"`
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
