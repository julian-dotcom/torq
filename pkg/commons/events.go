package commons

import (
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
)

type EventData struct {
	EventTime time.Time `json:"eventTime"`
	NodeId    int       `json:"nodeId"`
}

type ManualTriggerEvent struct {
	EventData
	WorkflowVersionNodeId int `json:"workflowVersionNodeId"`
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
	FeeBaseMsat      int64  `json:"feeBaseMsat"`
	FeeRateMilliMsat int64  `json:"feeRateMilliMsat"`
}

type ChannelBalanceEventData struct {
	Capacity                      int64 `json:"capacity"`
	LocalBalance                  int64 `json:"localBalance"`
	LocalBalancePerMilleRatio     int   `json:"localBalancePerMilleRatio"`
	RemoteBalance                 int64 `json:"remoteBalance"`
	PeerChannelCapacity           int64 `json:"peerChannelCapacity"`
	PeerChannelCount              int   `json:"peerChannelCount"`
	PeerLocalBalance              int64 `json:"peerLocalBalance"`
	PeerLocalBalancePerMilleRatio int   `json:"peerLocalBalancePerMilleRatio"`
}

type ServiceStatus int

const (
	ServiceInactive               = ServiceStatus(Inactive)
	ServiceActive                 = ServiceStatus(Active)
	ServicePending                = ServiceStatus(Pending)
	ServiceDeleted                = ServiceStatus(Deleted)
	ServiceInitializing           = ServiceStatus(Initializing)
	ServiceBootRequested          = ServiceStatus(100)
	ServiceBootRequestedWithDelay = ServiceStatus(101)
)

type ServiceEvent struct {
	EventData
	Type               ServiceType
	SubscriptionStream *SubscriptionStream
	Status             ServiceStatus
	PreviousStatus     ServiceStatus
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
	ChannelId                int                      `json:"channelId"`
	BalanceDelta             int64                    `json:"balanceDelta"`
	BalanceDeltaAbsolute     int64                    `json:"balanceDeltaAbsolute"`
	BalanceUpdateEventOrigin BalanceUpdateEventOrigin `json:"balanceUpdateEventOrigin"`
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
	BlockHeight           *uint32   `json:"blockHeight"`
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

type ChannelPoint struct {
	TxId        []byte `json:"txId"`
	OutputIndex uint32 `json:"outputIndex"`
}

type CloseChannelStatus struct {
	ClosingTxId []byte `json:"closingTxId"`
	Success     bool   `json:"success"`
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
	RequestId      string            `json:"requestId"`
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

// GENERIC REQUEST/RESPONSE STRUCTS
type FailedRequest struct {
	Reason string `json:"reason"`
	Error  string `json:"error"`
}

// Request/Response for Vector
type ShortChannelIdRequest struct {
	CommunicationRequest
	ResponseChannel chan<- ShortChannelIdResponse `json:"-"`
	TransactionHash string                        `json:"transactionHash"`
	OutputIndex     int                           `json:"outputIndex"`
}

type ShortChannelIdResponse struct {
	Request ShortChannelIdRequest `json:"request"`
	CommunicationResponse
	ShortChannelId string `json:"shortChannelId"`
}

// Request/Response for lightningCommunication
type CommunicationRequest struct {
	RequestId   string     `json:"requestId"`
	RequestTime *time.Time `json:"requestTime"`
	NodeId      int        `json:"nodeId"`
}

type CommunicationResponse struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type ChannelStatusUpdateRequest struct {
	ResponseChannel chan<- ChannelStatusUpdateResponse `json:"-"`
	CommunicationRequest
	ChannelId     int    `json:"channelId"`
	ChannelStatus Status `json:"channelStatus"`
}

type ChannelStatusUpdateResponse struct {
	Request ChannelStatusUpdateRequest `json:"request"`
	CommunicationResponse
}

type RoutingPolicyUpdateRequest struct {
	CommunicationRequest
	ResponseChannel  chan<- RoutingPolicyUpdateResponse `json:"-"`
	RateLimitSeconds int                                `json:"rateLimitSeconds"`
	RateLimitCount   int                                `json:"rateLimitCount"`
	ChannelId        int                                `json:"channelId"`
	FeeRateMilliMsat *int64                             `json:"feeRateMilliMsat"`
	FeeBaseMsat      *int64                             `json:"feeBaseMsat"`
	MaxHtlcMsat      *uint64                            `json:"maxHtlcMsat"`
	MinHtlcMsat      *uint64                            `json:"minHtlcMsat"`
	TimeLockDelta    *uint32                            `json:"timeLockDelta"`
}

type RoutingPolicyUpdateResponse struct {
	Request RoutingPolicyUpdateRequest `json:"request"`
	CommunicationResponse
	FailedUpdates []FailedRequest `json:"failedUpdates"`
}

type SignatureVerificationRequest struct {
	CommunicationRequest
	ResponseChannel chan<- SignatureVerificationResponse `json:"-"`
	Message         string                               `json:"message"`
	Signature       string                               `json:"signature"`
}

type SignatureVerificationResponse struct {
	Request SignatureVerificationRequest `json:"request"`
	CommunicationResponse
	PublicKey string `json:"publicKey"`
	Valid     bool   `json:"valid"`
}

type SignMessageRequest struct {
	CommunicationRequest
	ResponseChannel chan<- SignMessageResponse `json:"-"`
	Message         string                     `json:"message"`
	SingleHash      *bool                      `json:"singleHash"`
}

type SignMessageResponse struct {
	Request SignMessageRequest `json:"request"`
	CommunicationResponse
	Signature string `json:"signature"`
}

type RebalanceRequests struct {
	CommunicationRequest
	Requests        []RebalanceRequest
	ResponseChannel chan<- []RebalanceResponse `json:"-"`
}

type RebalanceRequest struct {
	Origin RebalanceRequestOrigin `json:"origin"`
	// Either manually generated number for manual rebalance or
	// WorkflowVersionNodeId for rebalance originating from workflows
	OriginId        int    `json:"originId"`
	OriginReference string `json:"originReference"`
	// Either IncomingChannelId is populated or OutgoingChannelId is.
	IncomingChannelId int `json:"incomingChannelId"`
	// Either OutgoingChannelId is populated or IncomingChannelId is.
	OutgoingChannelId int `json:"outgoingChannelIds"`
	// ONLY used for previous success rerun validation!
	ChannelIds            []int       `json:"channelIds"`
	AmountMsat            uint64      `json:"amountMsat"`
	MaximumCostMsat       uint64      `json:"maximumCostMsat"`
	MaximumConcurrency    int         `json:"maximumConcurrency"`
	WorkflowUnfocusedPath interface{} `json:"-"`
}

type RebalanceResponse struct {
	Request RebalanceRequest `json:"request"`
	CommunicationResponse
}

type NodeWalletBalanceRequest struct {
	CommunicationRequest
	ResponseChannel chan<- NodeWalletBalanceResponse `json:"-"`
}

type NodeWalletBalanceResponse struct {
	TotalBalance              int64
	ConfirmedBalance          int64
	UnconfirmedBalance        int64
	LockedBalance             int64
	ReservedBalanceAnchorChan int64
	CommunicationResponse
	ResponseChannel chan<- SignatureVerificationResponse `json:"-"`
}
type ImportType int

const (
	ImportChannelRoutingPolicies = ImportType(iota)
	ImportNodeInformation
	ImportAllChannels
	ImportPendingChannelsOnly
)

type ImportRequest struct {
	CommunicationRequest
	ResponseChannel chan<- ImportResponse `json:"-"`
	Force           bool
	ImportType      ImportType
}

type ImportResponse struct {
	Request ImportRequest `json:"request"`
	CommunicationResponse
	Error error `json:"error"`
}

type ConnectPeerRequest struct {
	CommunicationRequest
	PubKey          string                     `json:"pubKey"`
	Host            string                     `json:"host"`
	ResponseChannel chan<- ConnectPeerResponse `json:"-"`
}

type ConnectPeerResponse struct {
	CommunicationResponse
}

type DisconnectPeerRequest struct {
	CommunicationRequest
	PubKey          string                        `json:"pubKey"`
	Host            string                        `json:"host"`
	ResponseChannel chan<- DisconnectPeerResponse `json:"-"`
}

type DisconnectPeerResponse struct {
	CommunicationResponse
}

type ListPeersRequest struct {
	CommunicationRequest
	ResponseChannel chan<- ListPeersResponse `json:"-"`
}

type Feature struct {
	Name       string `json:"name"`
	IsRequired bool   `json:"is_required"`
	IsKnown    bool   `json:"is_known"`
}

type PeerSyncType int32

const (
	// PeerUnknownSync Denotes that we cannot determine the peer's current sync type.
	PeerUnknownSync PeerSyncType = 0
	// PeerActiveSync Denotes that we are actively receiving new graph updates from the peer.
	PeerActiveSync PeerSyncType = 1
	// PeerPassiveSync Denotes that we are not receiving new graph updates from the peer.
	PeerPassiveSync PeerSyncType = 2
	// PeerPinnedSync Denotes that this peer is pinned into an active sync.
	PeerPinnedSync PeerSyncType = 3
)

type Peer struct {
	PubKey          string       `json:"pub_key"`
	Address         string       `json:"address"`
	BytesSent       uint64       `json:"bytes_sent"`
	BytesRecv       uint64       `json:"bytes_recv"`
	SatSent         int64        `json:"sat_sent"`
	SatRecv         int64        `json:"sat_recv"`
	Inbound         bool         `json:"inbound"`
	PingTime        int64        `json:"ping_time"`
	SyncType        PeerSyncType `json:"sync_type"`
	Features        []Feature    `json:"features"`
	Errors          []string     `json:"errors"`
	FlapCount       int          `json:"flap_count"`
	LastFlapNS      string       `json:"last_flap_ns"`
	LastPingPayload string       `json:"last_ping_payload"`
}
type ListPeersResponse struct {
	Peers map[string]Peer `json:"peers"`
	CommunicationResponse
}
