package commons

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
	FeeBaseMsat      uint64 `json:"feeBaseMsat"`
	FeeRateMilliMsat uint64 `json:"feeRateMilliMsat"`
}

type ServiceEvent struct {
	EventData
	Type               ServiceType
	SubscriptionStream *SubscriptionStream
	Status             Status
	PreviousStatus     Status
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
}

type ChannelEvent struct {
	EventData
	Type      lnrpc.ChannelEventUpdate_UpdateType `json:"type"`
	ChannelId int                                 `json:"channelId"`
}

type InvoiceEvent struct {
	EventData
	ChannelId      int                        `json:"channelId"`
	AddIndex       uint64                     `json:"addIndex"`
	ValueMSat      uint64                     `json:"valueMSat"`
	State          lnrpc.Invoice_InvoiceState `json:"state"`
	AmountPaidMsat uint64                     `json:"amountPaidMsat"`
	SettledDate    time.Time                  `json:"settledDate"`
}

type PeerEvent struct {
	EventData
	Type        lnrpc.PeerEvent_EventType `json:"type"`
	EventNodeId int                       `json:"eventNodeId"`
}

type PaymentEvent struct {
	EventData
	AmountPaid           int64                       `json:"amountPaid"`
	FeeMsat              uint64                      `json:"feeMsat"`
	PaymentStatus        lnrpc.Payment_PaymentStatus `json:"paymentStatus"`
	PaymentFailureReason lnrpc.PaymentFailureReason  `json:"paymentFailureReason"`
	OutgoingChannelId    *int                        `json:"outgoingChannelId"`
	RebalanceAmountMsat  *uint64                     `json:"rebalanceAmount"`
	IncomingChannelId    *int                        `json:"incomingChannelId"`
}

type BlockEvent struct {
	EventData
	Height uint32 `json:"height"`
	Hash   []byte `json:"hash"`
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

type ForwardEvent struct {
	EventData
	Timestamp         time.Time `json:"timestamp"`
	FeeMsat           uint64    `json:"feeMsat"`
	AmountInMsat      uint64    `json:"amountInMsat"`
	AmountOutMsat     uint64    `json:"amountOutMsat"`
	OutgoingChannelId *int      `json:"outgoingChannelId"`
	IncomingChannelId *int      `json:"incomingChannelId"`
}
