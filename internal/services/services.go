package services

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

type Service struct {
	Status   commons.Status
	BootTime *time.Time
}
type TorqService struct {
	Service
	Version string
}

type LndService struct {
	Service
	TransactionStreamStatus       commons.Status
	TransactionStreamBootTime     *time.Time
	HtlcEventStreamStatus         commons.Status
	HtlcEventStreamBootTime       *time.Time
	ChannelEventStreamStatus      commons.Status
	ChannelEventStreamBootTime    *time.Time
	GraphEventStreamStatus        commons.Status
	GraphEventStreamBootTime      *time.Time
	ForwardStreamStatus           commons.Status
	ForwardStreamBootTime         *time.Time
	InvoiceStreamStatus           commons.Status
	InvoiceStreamBootTime         *time.Time
	PaymentStreamStatus           commons.Status
	PaymentStreamBootTime         *time.Time
	InFlightPaymentStreamStatus   commons.Status
	InFlightPaymentStreamBootTime *time.Time
	PeerEventStreamStatus         commons.Status
	PeerEventStreamBootTime       *time.Time
}

type VectorService struct {
	Service
}

type AmbossService struct {
	Service
}

type Services struct {
	TorqService    TorqService
	LndServices    []LndService
	VectorServices []VectorService
	AmbossServices []AmbossService
}
