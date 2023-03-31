package services

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

type CommonService struct {
	ServiceType       commons.ServiceType   `json:"type"`
	ServiceTypeString string                `json:"typeString"`
	Status            commons.ServiceStatus `json:"status"`
	StatusString      string                `json:"statusString"`
	BootTime          *time.Time            `json:"bootTime,omitempty"`
}
type TorqService struct {
	CommonService
}

type LndService struct {
	CommonService
	NodeId         int             `json:"nodeId"`
	BitcoinNetwork commons.Network `json:"bitcoinNetwork"`
}

type ServiceMismatch struct {
	ServiceType         commons.ServiceType   `json:"type"`
	ServiceTypeString   string                `json:"typeString"`
	Status              commons.ServiceStatus `json:"status"`
	StatusString        string                `json:"statusString"`
	DesiredStatus       commons.ServiceStatus `json:"desiredStatus"`
	DesiredStatusString string                `json:"desiredStatusString"`
	NodeId              *int                  `json:"nodeId,omitempty"`
	BitcoinNetwork      *commons.Network      `json:"bitcoinNetwork,omitempty"`
	FailureTime         *time.Time            `json:"failureTime,omitempty"`
}

type Services struct {
	Version           string            `json:"version"`
	BitcoinNetworks   []commons.Network `json:"bitcoinNetworks"`
	MainService       TorqService       `json:"mainService"`
	TorqServices      []TorqService     `json:"torqServices"`
	LndServices       []LndService      `json:"lndServices,omitempty"`
	ServiceMismatches []ServiceMismatch `json:"serviceMismatches,omitempty"`
}
