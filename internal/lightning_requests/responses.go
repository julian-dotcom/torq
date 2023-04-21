package lightning_requests

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
