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

type Services struct {
	Version         string            `json:"version"`
	BitcoinNetworks []commons.Network `json:"bitcoinNetworks"`
	MainService     TorqService       `json:"mainService"`
	TorqServices    []TorqService     `json:"torqServices"`
	LndServices     []LndService      `json:"lndServices,omitempty"`
}
