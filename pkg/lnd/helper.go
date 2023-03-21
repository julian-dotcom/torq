package lnd

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

func convertMicro(ns int64) time.Time {
	return time.Unix(0, ns).Round(time.Microsecond).UTC()
}

func SetStreamStatus(nodeId int,
	subscriptionStream commons.SubscriptionStream,
	previousStatus commons.ServiceStatus,
	status commons.ServiceStatus) commons.ServiceStatus {

	if status == commons.ServiceInitializing || previousStatus != status {
		commons.RunningServices[commons.LndService].SetStreamStatus(nodeId, subscriptionStream, status)
	}
	return status
}
