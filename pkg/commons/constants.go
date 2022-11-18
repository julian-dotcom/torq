package commons

type Status int

const (
	Inactive = Status(iota)
	Active
	Pending
	Deleted
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
	SigNet
	SimNet
	RegTest
	TestNet
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

type PingSystem byte

const (
	Amboss PingSystem = 1 << iota
	Vector
)

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

// GetChain defaults to Bitcoin when no match is found
func GetChain(chain string) Chain {
	switch chain {
	case "litecoin":
		return Litecoin
	}
	return Bitcoin
}
