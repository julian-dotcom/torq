package lnd

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

func convertMicro(ns int64) time.Time {
	return time.Unix(0, ns).Round(time.Microsecond).UTC()
}

func SendStreamEvent(eventChannel chan interface{},
	nodeId int, subscriptionStream commons.SubscriptionStream, status commons.Status, previousStatus commons.Status) commons.Status {
	if eventChannel != nil && (status == commons.Initializing || previousStatus != status) {
		previousStatusFromRunningService := commons.RunningServices[commons.LndService].SetStreamStatus(nodeId, subscriptionStream, status)
		if previousStatusFromRunningService != status {
			commons.SendServiceEvent(nodeId, eventChannel, previousStatusFromRunningService, status, commons.LndService, &subscriptionStream)
		}
	}
	return status
}
