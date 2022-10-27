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
