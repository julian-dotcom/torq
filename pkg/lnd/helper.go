package lnd

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

func convertMicro(ns int64) time.Time {
	return time.Unix(0, ns).Round(time.Microsecond).UTC()
}

func sendStreamEvent(serviceEventChannel chan commons.ServiceEvent,
	nodeId int, subscriptionStream commons.SubscriptionStream, status commons.Status, previousStatus commons.Status) commons.Status {
	if previousStatus != status && serviceEventChannel != nil {
		commons.RunningServices[commons.LndService].SetStreamStatus(nodeId, subscriptionStream, status, serviceEventChannel)
	}
	return status
}
