package lightning_helpers

import (
	"time"

	"github.com/lncapital/torq/internal/core"
)

type Status int

const (
	Inactive = Status(core.Inactive)
	Active   = Status(core.Active)
)

type CommunicationResponse struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type FailedRequest struct {
	Reason string `json:"reason"`
	Error  string `json:"error"`
}

type RebalanceResponse struct {
	Request RebalanceRequest `json:"request"`
	CommunicationResponse
}

type InformationResponse struct {
	Request InformationRequest `json:"request"`
	CommunicationResponse
	Implementation          core.Implementation `json:"implementation"`
	Version                 string              `json:"version"`
	PublicKey               string              `json:"publicKey"`
	Alias                   string              `json:"alias"`
	Color                   string              `json:"color"`
	PendingChannelCount     int                 `json:"pendingChannelCount"`
	ActiveChannelCount      int                 `json:"activeChannelCount"`
	InactiveChannelCount    int                 `json:"inactiveChannelCount"`
	PeerCount               int                 `json:"peerCount"`
	BlockHeight             uint32              `json:"blockHeight"`
	BlockHash               string              `json:"blockHash"`
	BestHeaderTimestamp     time.Time           `json:"bestHeaderTimestamp"`
	ChainSynced             bool                `json:"chainSynced"`
	GraphSynced             bool                `json:"graphSynced"`
	Addresses               []string            `json:"addresses"`
	HtlcInterceptorRequired bool                `json:"htlcInterceptorRequired"`
}

type SignMessageResponse struct {
	Request SignMessageRequest `json:"request"`
	CommunicationResponse
	Signature string `json:"signature"`
}

type SignatureVerificationResponse struct {
	Request SignatureVerificationRequest `json:"request"`
	CommunicationResponse
	PublicKey string `json:"publicKey"`
	Valid     bool   `json:"valid"`
}

type RoutingPolicyUpdateResponse struct {
	Request RoutingPolicyUpdateRequest `json:"request"`
	CommunicationResponse
	FailedUpdates []FailedRequest `json:"failedUpdates"`
}

type ConnectPeerResponse struct {
	Request ConnectPeerRequest `json:"request"`
	CommunicationResponse
	RequestFailCurrentlyConnected bool `json:"requestFailCurrentlyConnected"`
}

type DisconnectPeerResponse struct {
	Request DisconnectPeerRequest `json:"request"`
	CommunicationResponse
	RequestFailedCurrentlyDisconnected bool `json:"requestFailedCurrentlyDisconnected"`
}

type WalletBalanceResponse struct {
	Request WalletBalanceRequest `json:"request"`
	CommunicationResponse
	TotalBalance              int64 `json:"totalBalance"`
	ConfirmedBalance          int64 `json:"confirmedBalance"`
	UnconfirmedBalance        int64 `json:"unconfirmedBalance"`
	LockedBalance             int64 `json:"lockedBalance"`
	ReservedBalanceAnchorChan int64 `json:"reservedBalanceAnchorChan"`
}

type ListPeersResponse struct {
	Request ListPeersRequest `json:"request"`
	CommunicationResponse
	Peers map[string]Peer `json:"peers"`
}

type NewAddressResponse struct {
	Request NewAddressRequest `json:"request"`
	CommunicationResponse
	Address string `json:"address"`
}

type OpenChannelResponse struct {
	Request OpenChannelRequest `json:"request"`
	CommunicationResponse
	ChannelStatus          core.ChannelStatus `json:"channelStatus"`
	ChannelPoint           string             `json:"channelPoint"`
	FundingTransactionHash string             `json:"fundingTransactionHash,omitempty"`
	FundingOutputIndex     uint32             `json:"fundingOutputIndex,omitempty"`
}

type BatchOpenChannelResponse struct {
	Request BatchOpenChannelRequest `json:"request"`
	CommunicationResponse
	PendingChannelPoints []string `json:"pendingChannelPoints"`
}

type CloseChannelResponse struct {
	Request CloseChannelRequest `json:"request"`
	CommunicationResponse
	ChannelStatus          core.ChannelStatus `json:"channelStatus"`
	ClosingTransactionHash string             `json:"closingTransactionHash"`
}

type NewInvoiceResponse struct {
	Request NewInvoiceRequest `json:"request"`
	CommunicationResponse
	PaymentRequest string `json:"paymentRequest"`
	AddIndex       uint64 `json:"addIndex"`
	PaymentAddress string `json:"paymentAddress"`
}

type ChannelStatusUpdateResponse struct {
	Request ChannelStatusUpdateRequest `json:"request"`
	CommunicationResponse
}

type OnChainPaymentResponse struct {
	Request OnChainPaymentRequest `json:"request"`
	CommunicationResponse
	TxId string `json:"txId"`
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
	Request NewPaymentRequest `json:"request"`
	CommunicationResponse
	PaymentStatus  string    `json:"paymentStatus"`
	FailureReason  string    `json:"failureReason"`
	Hash           string    `json:"hash"`
	Preimage       string    `json:"preimage"`
	PaymentRequest string    `json:"paymentRequest"`
	AmountMsat     int64     `json:"amountMsat"`
	FeeLimitMsat   int64     `json:"feeLimitMsat"`
	FeePaidMsat    int64     `json:"feePaidMsat"`
	CreationDate   time.Time `json:"creationDate"`
	Attempt        Attempt   `json:"path"`
}

type FeatureMap map[uint32]Feature

type HopHint struct {
	LNDShortChannelId uint64 `json:"lndShortChannelId"`
	ShortChannelId    string `json:"shortChannelId"`
	NodeId            string `json:"localNodeId"`
	FeeBase           uint32 `json:"feeBase"`
	CltvExpiryDelta   uint32 `json:"cltvExpiryDelta"`
	FeeProportional   uint32 `json:"feeProportionalMillionths"`
}

type RouteHint struct {
	HopHints []HopHint `json:"hopHints"`
}

type DecodeInvoiceResponse struct {
	Request DecodeInvoiceRequest `json:"request"`
	CommunicationResponse
	NodeAlias         string      `json:"nodeAlias"`
	PaymentRequest    string      `json:"paymentRequest"`
	DestinationPubKey string      `json:"destinationPubKey"`
	RHash             string      `json:"rHash"`
	Memo              string      `json:"memo"`
	ValueMsat         int64       `json:"valueMsat"`
	PaymentAddr       string      `json:"paymentAddr"`
	FallbackAddr      string      `json:"fallbackAddr"`
	Expiry            int64       `json:"expiry"`
	CreatedAt         int64       `json:"createdAt"`
	CltvExpiry        int64       `json:"cltvExpiry"`
	Private           bool        `json:"private"`
	Features          FeatureMap  `json:"features"`
	RouteHints        []RouteHint `json:"routeHints"`
}
