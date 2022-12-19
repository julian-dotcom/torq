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
	MinHtlcMsat      uint64 `json:"minHtlcMsat"`
	MaxHtlcMsat      uint64 `json:"maxHtlcMsat"`
	FeeBaseMsat      uint64 `json:"feeBaseMsat"`
	FeeRateMilliMsat uint64 `json:"feeRateMilliMsat"`
}

type ChannelBalanceEventData struct {
	Capacity                            int64 `json:"capacity"`
	LocalBalance                        int64 `json:"localBalance"`
	LocalBalancePerMilleRatio           int   `json:"localBalancePerMilleRatio"`
	RemoteBalance                       int64 `json:"remoteBalance"`
	AggregatedLocalBalance              int64 `json:"aggregatedLocalBalance"`
	AggregatedLocalBalancePerMilleRatio int   `json:"aggregatedLocalBalancePerMilleRatio"`
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
	PreviousEventTime *time.Time          `json:"previousEventTime"`
	PreviousEventData *NodeGraphEventData `json:"previous"`
}

type ChannelGraphEvent struct {
	GraphEventData
	ChannelGraphEventData
	PreviousEventTime *time.Time             `json:"previousEventTime"`
	PreviousEventData *ChannelGraphEventData `json:"previous"`
}

type ChannelBalanceEvent struct {
	EventData
	ChannelId int `json:"channelId"`
	ChannelBalanceEventData
	PreviousEventData *ChannelBalanceEventData `json:"previous"`
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
	ChannelId         int                        `json:"channelId"`
	AddIndex          uint64                     `json:"addIndex"`
	ValueMSat         uint64                     `json:"valueMSat"`
	State             lnrpc.Invoice_InvoiceState `json:"state"`
	AmountPaidMsat    uint64                     `json:"amountPaidMsat"`
	SettledDate       time.Time                  `json:"settledDate"`
	DestinationNodeId *int                       `json:"destinationNodeId"`
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

// GENERIC REQUEST/RESPONSE STRUCTS
type FailedRequest struct {
	Reason string `json:"reason"`
	Error  string `json:"error"`
}

type ChannelPoint struct {
	TxId        []byte `json:"txId"`
	OutputIndex uint32 `json:"outputIndex"`
}

// OPEN CHANNEL
type OpenChannelRequest struct {
	NodeId             int     `json:"nodeId"`
	SatPerVbyte        *uint64 `json:"satPerVbyte"`
	NodePubKey         string  `json:"nodePubKey"`
	Host               *string `json:"host"`
	LocalFundingAmount int64   `json:"localFundingAmount"`
	PushSat            *int64  `json:"pushSat"`
	TargetConf         *int32  `json:"targetConf"`
	Private            *bool   `json:"private"`
	MinHtlcMsat        *uint64 `json:"minHtlcMsat"`
	RemoteCsvDelay     *uint32 `json:"remoteCsvDelay"`
	MinConfs           *int32  `json:"minConfs"`
	SpendUnconfirmed   *bool   `json:"spendUnconfirmed"`
	CloseAddress       *string `json:"closeAddress"`
}

type OpenChannelResponse struct {
	ReqId               string             `json:"reqId"`
	Request             OpenChannelRequest `json:"request"`
	Status              ChannelStatus      `json:"status"`
	ChannelPoint        string             `json:"channelPoint,omitempty"`
	PendingChannelPoint string             `json:"pendingChannelPoint,omitempty"`
}

// CLOSE CHANNEL
type CloseChannelRequest struct {
	NodeId          int     `json:"nodeId"`
	ChannelId       int     `json:"channelId"`
	Force           *bool   `json:"force"`
	TargetConf      *int32  `json:"targetConf"`
	DeliveryAddress *string `json:"deliveryAddress"`
	SatPerVbyte     *uint64 `json:"satPerVbyte"`
}

type CloseChannelResponse struct {
	ReqId                    string              `json:"reqId"`
	Request                  CloseChannelRequest `json:"request"`
	Status                   ChannelStatus       `json:"status"`
	ClosePendingChannelPoint ChannelPoint        `json:"closePendingChannelPoint"`
	CloseChannelStatus       CloseChannelStatus  `json:"closeChannelStatus"`
}

type CloseChannelStatus struct {
	ClosingTxId []byte `json:"closingTxId"`
	Success     bool   `json:"success"`
}

// NEW ADDRESS
type NewAddressRequest struct {
	NodeId int   `json:"nodeId"`
	Type   int32 `json:"type"`
	//The name of the account to generate a new address for. If empty, the default wallet account is used.
	Account string `json:"account"`
}

type NewAddressResponse struct {
	ReqId   string            `json:"reqId"`
	Request NewAddressRequest `json:"request"`
	Address string            `json:"address"`
}

// NEW PAYMENT
type NewPaymentRequest struct {
	NodeId           int     `json:"nodeId"`
	Invoice          *string `json:"invoice"`
	TimeOutSecs      int32   `json:"timeoutSecs"`
	Dest             *string `json:"dest"`
	AmtMSat          *int64  `json:"amtMSat"`
	FeeLimitMsat     *int64  `json:"feeLimitMsat"`
	AllowSelfPayment *bool   `json:"allowSelfPayment"`
}

type MppRecord struct {
	PaymentAddr  string
	TotalAmtMsat int64
}

type Hops struct {
	ChanId           string    `json:"chanId"`
	Expiry           uint32    `json:"expiry"`
	AmtToForwardMsat int64     `json:"amtToForwardMsat"`
	PubKey           string    `json:"pubKey"`
	MppRecord        MppRecord `json:"mppRecord"`
	// TODO: Imolement AMP record here when needed
}

type Route struct {
	TotalTimeLock uint32 `json:"totalTimeLock"`
	Hops          []Hops `json:"hops"`
	TotalAmtMsat  int64  `json:"totalAmtMsat"`
}

type FailureDetails struct {
	Reason             string `json:"reason"`
	FailureSourceIndex uint32 `json:"failureSourceIndex"`
	Height             uint32 `json:"height"`
}

type Attempt struct {
	AttemptId     uint64         `json:"attemptId"`
	Status        string         `json:"status"`
	Route         Route          `json:"route"`
	AttemptTimeNs time.Time      `json:"attemptTimeNs"`
	ResolveTimeNs time.Time      `json:"resolveTimeNs"`
	Preimage      string         `json:"preimage"`
	Failure       FailureDetails `json:"failure"`
}
type NewPaymentResponse struct {
	ReqId          string            `json:"reqId"`
	Request        NewPaymentRequest `json:"request"`
	Status         string            `json:"status"`
	FailureReason  string            `json:"failureReason"`
	Hash           string            `json:"hash"`
	Preimage       string            `json:"preimage"`
	PaymentRequest string            `json:"paymentRequest"`
	AmountMsat     int64             `json:"amountMsat"`
	FeeLimitMsat   int64             `json:"feeLimitMsat"`
	FeePaidMsat    int64             `json:"feePaidMsat"`
	CreationDate   time.Time         `json:"creationDate"`
	Attempt        Attempt           `json:"path"`
}

// PAY ONCHAIN
type PayOnChainRequest struct {
	NodeId           int     `json:"nodeId"`
	Address          string  `json:"address"`
	AmountSat        int64   `json:"amountSat"`
	TargetConf       *int32  `json:"targetConf"`
	SatPerVbyte      *uint64 `json:"satPerVbyte"`
	SendAll          *bool   `json:"sendAll"`
	Label            *string `json:"label"`
	MinConfs         *int32  `json:"minConfs"`
	SpendUnconfirmed *bool   `json:"spendUnconfirmed"`
}

type PayOnChainResponse struct {
	Request PayOnChainRequest `json:"request"`
	TxId    string            `json:"txId"`
}

// UPDATE CHANNEL
type UpdateChannelRequest struct {
	NodeId           int     `json:"nodeId"`
	ChannelId        *int    `json:"channelId"`
	FeeRateMilliMsat *uint64 `json:"feeRateMilliMsat"`
	FeeBaseMsat      *uint64 `json:"feeBaseMsat"`
	MaxHtlcMsat      *uint64 `json:"maxHtlcMsat"`
	MinHtlcMsat      *uint64 `json:"minHtlcMsat"`
	TimeLockDelta    *uint32 `json:"timeLockDelta"`
}

type UpdateChannelResponse struct {
	Request       UpdateChannelRequest `json:"request"`
	Status        Status               `json:"status"`
	FailedUpdates []FailedRequest      `json:"failedUpdates"`
}

// BATCH OPEN CHANNELS
type BatchOpenChannel struct {
	NodePubkey         string  `json:"nodePubkey"`
	LocalFundingAmount int64   `json:"localFundingAmount"`
	PushSat            *int64  `json:"pushSat"`
	Private            *bool   `json:"private"`
	MinHtlcMsat        *uint64 `json:"minHtlcMsat"`
}

type BatchOpenRequest struct {
	NodeId      int                `json:"nodeId"`
	Channels    []BatchOpenChannel `json:"channels"`
	TargetConf  *int32             `json:"targetConf"`
	SatPerVbyte *int64             `json:"satPerVbyte"`
}

type BatchOpenResponse struct {
	PendingChannels []PendingChannel `json:"pendingChannels"`
}

type PendingChannel struct {
	PendingChannelPoint string `json:"pendingChannelPoint"`
}
