package core

const UnknownEnumString = "Unknown"

type Status int

const (
	Inactive = Status(iota)
	Active
	Pending
	Deleted
	Initializing
	Archived
	TimedOut
)

type Implementation int

const (
	LND = Implementation(iota)
	CLN
)

type Chain int

const (
	Bitcoin = Chain(iota)
	Litecoin
)

type Network int

const (
	MainNet = Network(iota)
	TestNet
	RegTest
	SigNet
	SimNet
)

type ChannelStatus int

const (
	Opening = ChannelStatus(iota)
	Open
	Closing
	CooperativeClosed      = 100
	LocalForceClosed       = 101
	RemoteForceClosed      = 102
	BreachClosed           = 103
	FundingCancelledClosed = 104
	AbandonedClosed        = 105
)

type ChannelFlags uint32

const (
	FundedOn ChannelFlags = 1 << iota
	ClosedOn
)
const ChannelFlagsMax = int(ClosedOn)*2 - 1

type TransactionFlags uint32

const (
	TransactionTime TransactionFlags = 1 << iota
)
const TransactionFlagsMax = int(TransactionTime)*2 - 1

type PingSystem uint32

const (
	Amboss PingSystem = 1 << iota
	Vector
)
const PingSystemMax = int(Vector)*2 - 1

type NodeConnectionDetailCustomSettings uint32

const (
	ImportFailedPayments NodeConnectionDetailCustomSettings = 1 << iota
	ImportHtlcEvents
	ImportPeerEventsDeleted
	ImportTransactions
	ImportPayments
	ImportInvoices
	ImportForwards
	ImportHistoricForwards
)
const NodeConnectionDetailCustomSettingsMax = int(ImportHistoricForwards)*2 - 1

const (
	MEMPOOL string = "https://mempool.space/lightning/channel/"
	AMBOSS  string = "https://amboss.space/edge/"
	ONEML   string = "https://1ml.com/channel/"
)

type BalanceUpdateEventOrigin int

const (
	BalanceUpdateForwardEvent = BalanceUpdateEventOrigin(iota)
	BalanceUpdateInvoiceEvent
	BalanceUpdatePaymentEvent
)

type NotificationType int

const (
	NodeDetails NotificationType = iota
)

type NodeConnectionSetting int

const (
	NodeConnectionSettingAlwaysReconnect = NodeConnectionSetting(iota)
	NodeConnectionSettingDisableReconnect
)

type NodeConnectionStatus int

const (
	NodeConnectionStatusDisconnected = NodeConnectionStatus(iota)
	NodeConnectionStatusConnected
)
