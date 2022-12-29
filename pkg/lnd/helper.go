package lnd

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

func convertMicro(ns int64) time.Time {
	return time.Unix(0, ns).Round(time.Microsecond).UTC()
}

func getDeltaPerMille(base uint64, amt uint64) int {
	if base > amt {
		return int((base - amt) / base * 1_000)
	} else if base == amt {
		return 0
	} else {
		return int((amt - base) / amt * 1_000)
	}
}

func SendStreamEvent(eventChannel chan interface{},
	nodeId int, subscriptionStream commons.SubscriptionStream, status commons.Status, previousStatus commons.Status) commons.Status {
	if eventChannel != nil && (status == commons.Initializing || previousStatus != status) {
		commons.RunningServices[commons.LndService].SetStreamStatus(nodeId, subscriptionStream, status, eventChannel)
	}
	return status
}
