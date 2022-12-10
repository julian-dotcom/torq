package services

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

type Service struct {
	Status   commons.Status `json:"status"`
	BootTime *time.Time     `json:"bootTime,omitempty"`
}
type TorqService struct {
	Service
	Version string `json:"version"`
}

type LndService struct {
	Service
	NodeId                        int            `json:"nodeId"`
	TransactionStreamStatus       commons.Status `json:"transactionStreamStatus"`
	TransactionStreamBootTime     *time.Time     `json:"transactionStreamBootTime,omitempty"`
	HtlcEventStreamStatus         commons.Status `json:"htlcEventStreamStatus"`
	HtlcEventStreamBootTime       *time.Time     `json:"htlcEventStreamBootTime,omitempty"`
	ChannelEventStreamStatus      commons.Status `json:"channelEventStreamStatus"`
	ChannelEventStreamBootTime    *time.Time     `json:"channelEventStreamBootTime,omitempty"`
	GraphEventStreamStatus        commons.Status `json:"graphEventStreamStatus"`
	GraphEventStreamBootTime      *time.Time     `json:"graphEventStreamBootTime,omitempty"`
	ForwardStreamStatus           commons.Status `json:"forwardStreamStatus"`
	ForwardStreamBootTime         *time.Time     `json:"forwardStreamBootTime,omitempty"`
	InvoiceStreamStatus           commons.Status `json:"invoiceStreamStatus"`
	InvoiceStreamBootTime         *time.Time     `json:"invoiceStreamBootTime,omitempty"`
	PaymentStreamStatus           commons.Status `json:"paymentStreamStatus"`
	PaymentStreamBootTime         *time.Time     `json:"paymentStreamBootTime,omitempty"`
	InFlightPaymentStreamStatus   commons.Status `json:"inFlightPaymentStreamStatus"`
	InFlightPaymentStreamBootTime *time.Time     `json:"inFlightPaymentStreamBootTime,omitempty"`
	PeerEventStreamStatus         commons.Status `json:"peerEventStreamStatus"`
	PeerEventStreamBootTime       *time.Time     `json:"peerEventStreamBootTime,omitempty"`
}

type VectorService struct {
	Service
	NodeId int `json:"nodeId"`
}

type AmbossService struct {
	Service
	NodeId int `json:"nodeId"`
}

type Services struct {
	TorqService    TorqService     `json:"torqService"`
	LndServices    []LndService    `json:"lndServices,omitempty"`
	VectorServices []VectorService `json:"vectorServices,omitempty"`
	AmbossServices []AmbossService `json:"ambossServices,omitempty"`
}
