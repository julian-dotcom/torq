package commons

type Status int

const (
	Active = Status(iota)
	Inactive
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
