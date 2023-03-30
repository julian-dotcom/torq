package services

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

type CommonService struct {
	Status       commons.ServiceStatus `json:"status"`
	StatusString string                `json:"statusString"`
	BootTime     *time.Time            `json:"bootTime,omitempty"`
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
	TorqService     []TorqService     `json:"torqService"`
	LndServices     []LndService      `json:"lndServices,omitempty"`
}
