package lnd

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

func convertMicro(ns int64) time.Time {
	return time.Unix(0, ns).Round(time.Microsecond).UTC()
}

func SendStreamEvent(serviceEventChannel chan commons.ServiceEvent,
	nodeId int, subscriptionStream commons.SubscriptionStream, status commons.Status, previousStatus commons.Status) commons.Status {
	if serviceEventChannel != nil && (status == commons.Initializing || previousStatus != status) {
		previousStatusFromRunningService := commons.RunningServices[commons.LndService].SetStreamStatus(nodeId, subscriptionStream, status)
		if previousStatusFromRunningService != status {
			commons.SendServiceEvent(nodeId, serviceEventChannel, previousStatusFromRunningService, status, commons.LndService, &subscriptionStream)
		}
	}
	return status
}
