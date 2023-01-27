package services

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

type CommonService struct {
	Status       commons.Status `json:"status"`
	StatusString string         `json:"statusString"`
	BootTime     *time.Time     `json:"bootTime,omitempty"`
}
type TorqService struct {
	CommonService
	Version string `json:"version"`
}

type Service struct {
	CommonService
	NodeId     *int                `json:"nodeId,omitempty"`
	Type       commons.ServiceType `json:"type"`
	TypeString string              `json:"typeString"`
}

type Stream struct {
	CommonService
	NodeId     int                        `json:"nodeId,omitempty"`
	Type       commons.SubscriptionStream `json:"type"`
	TypeString string                     `json:"typeString"`
}

type LndService struct {
	CommonService
	NodeId       int      `json:"nodeId"`
	StreamStatus []Stream `json:"streamStatus,omitempty"`
}

type VectorService struct {
	CommonService
	NodeId int `json:"nodeId"`
}

type AmbossService struct {
	CommonService
	NodeId int `json:"nodeId"`
}

type Services struct {
	TorqService    TorqService     `json:"torqService"`
	LndServices    []LndService    `json:"lndServices,omitempty"`
	Services       []Service       `json:"services,omitempty"`
	VectorServices []VectorService `json:"vectorServices,omitempty"`
	AmbossServices []AmbossService `json:"ambossServices,omitempty"`
}
