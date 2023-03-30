package lnd

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

func convertMicro(ns int64) time.Time {
	return time.Unix(0, ns).Round(time.Microsecond).UTC()
}

func SetStreamStatus(nodeId int,
	serviceType commons.ServiceType,
	previousStatus commons.ServiceStatus,
	status commons.ServiceStatus) commons.ServiceStatus {

	if status == commons.ServiceInitializing || previousStatus != status {
		switch status {
		case commons.ServiceActive:
			commons.SetActiveLndServiceState(serviceType, nodeId)
		case commons.ServiceInactive:
			commons.SetInactiveLndServiceState(serviceType, nodeId)
		case commons.ServiceInitializing:
			commons.SetInitializingLndServiceState(serviceType, nodeId)
		case commons.ServicePending:
			commons.SetPendingLndServiceState(serviceType, nodeId)
		}
	}
	return status
}
