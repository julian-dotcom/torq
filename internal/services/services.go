package services

import (
	"time"

	"github.com/lncapital/torq/pkg/core"
)

type CommonService struct {
	ServiceType       core.ServiceType   `json:"type"`
	ServiceTypeString string             `json:"typeString"`
	Status            core.ServiceStatus `json:"status"`
	StatusString      string             `json:"statusString"`
	BootTime          *time.Time         `json:"bootTime,omitempty"`
}
type CoreService struct {
	CommonService
}

type LndService struct {
	CommonService
	NodeId         int          `json:"nodeId"`
	BitcoinNetwork core.Network `json:"bitcoinNetwork"`
}

type ServiceMismatch struct {
	ServiceType         core.ServiceType   `json:"type"`
	ServiceTypeString   string             `json:"typeString"`
	Status              core.ServiceStatus `json:"status"`
	StatusString        string             `json:"statusString"`
	DesiredStatus       core.ServiceStatus `json:"desiredStatus"`
	DesiredStatusString string             `json:"desiredStatusString"`
	NodeId              *int               `json:"nodeId,omitempty"`
	BitcoinNetwork      *core.Network      `json:"bitcoinNetwork,omitempty"`
	FailureTime         *time.Time         `json:"failureTime,omitempty"`
}

type Services struct {
	Version           string            `json:"version"`
	BitcoinNetworks   []core.Network    `json:"bitcoinNetworks"`
	MainService       CoreService       `json:"mainService"`
	TorqServices      []CoreService     `json:"torqServices"`
	LndServices       []LndService      `json:"lndServices,omitempty"`
	ServiceMismatches []ServiceMismatch `json:"serviceMismatches,omitempty"`
}
