package automation

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/commons"
)

const rebalanceResultsTimeoutSeconds = 5 * 60

var ManagedRebalanceResultChannel = make(chan ManagedRebalanceResult) //nolint:gochecknoglobals

type ManagedRebalanceResultCacheOperationType uint

const (
	readRebalanceResult ManagedRebalanceResultCacheOperationType = iota
	readRebalanceResultByOrigin
	writeRebalanceResult
)

type ManagedRebalanceResult struct {
	Type              ManagedRebalanceResultCacheOperationType
	Origin            commons.RebalanceRequestOrigin
	OriginId          int
	OriginReference   string
	IncomingChannelId int
	OutgoingChannelId int
	Status            *commons.Status
	RebalanceResult   RebalanceResult
	Out               chan<- RebalanceResult
}

func ManagedRebalanceResultCache(ch <-chan ManagedRebalanceResult, ctx context.Context) {
	outgoingResultsCache := make(map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult)
	incomingResultsCache := make(map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult)

	for {
		select {
		case <-ctx.Done():
			return
		case managedRebalance := <-ch:
			switch managedRebalance.Type {
			case readRebalanceResult:
				if !isValidResultRequest(managedRebalance) {
					SendToManagedRebalanceResultChannel(managedRebalance.Out, RebalanceResult{})
					continue
				}
				initializeResultsCache(managedRebalance, incomingResultsCache, outgoingResultsCache)
				SendToManagedRebalanceResultChannel(managedRebalance.Out,
					getRebalanceLatestResultCache(managedRebalance, incomingResultsCache, outgoingResultsCache))
			case readRebalanceResultByOrigin:
				if !isValidResultRequest(managedRebalance) {
					SendToManagedRebalanceResultChannel(managedRebalance.Out, RebalanceResult{})
					continue
				}
				initializeResultsCache(managedRebalance, incomingResultsCache, outgoingResultsCache)
				SendToManagedRebalanceResultChannel(managedRebalance.Out,
					getRebalanceLatestResultByOriginCache(managedRebalance, incomingResultsCache, outgoingResultsCache))
			case writeRebalanceResult:
				if !isValidResultRequest(managedRebalance) {
					continue
				}
				initializeResultsCache(managedRebalance, incomingResultsCache, outgoingResultsCache)
				appendRebalanceResult(managedRebalance, incomingResultsCache, outgoingResultsCache)
			}
		}
	}
}

func appendRebalanceResult(managedRebalanceResult ManagedRebalanceResult,
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult) {

	var results []RebalanceResult
	if managedRebalanceResult.IncomingChannelId != 0 {
		results = incomingResultsCache[managedRebalanceResult.Origin][originIdInt(managedRebalanceResult.OriginId)][channelIdInt(managedRebalanceResult.IncomingChannelId)]
	} else {
		results = outgoingResultsCache[managedRebalanceResult.Origin][originIdInt(managedRebalanceResult.OriginId)][channelIdInt(managedRebalanceResult.OutgoingChannelId)]
	}
	if len(results)%100 == 0 {
		for i := 0; i < len(results); i++ {
			result := results[i]
			if time.Since(result.UpdateOn).Seconds() < rebalanceResultsTimeoutSeconds {
				results = results[i:]
				break
			}
		}
	}
	if managedRebalanceResult.IncomingChannelId != 0 {
		incomingResultsCache[managedRebalanceResult.Origin][originIdInt(managedRebalanceResult.OriginId)][channelIdInt(managedRebalanceResult.IncomingChannelId)] =
			append(results, managedRebalanceResult.RebalanceResult)
	} else {
		outgoingResultsCache[managedRebalanceResult.Origin][originIdInt(managedRebalanceResult.OriginId)][channelIdInt(managedRebalanceResult.OutgoingChannelId)] =
			append(results, managedRebalanceResult.RebalanceResult)
	}
}

func getRebalanceLatestResultByOriginCache(managedRebalanceResult ManagedRebalanceResult,
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult) RebalanceResult {

	var rebalanceResults []RebalanceResult
	if managedRebalanceResult.IncomingChannelId != 0 {
		rebalanceResults = incomingResultsCache[managedRebalanceResult.Origin][originIdInt(managedRebalanceResult.OriginId)][channelIdInt(managedRebalanceResult.IncomingChannelId)]
	} else {
		rebalanceResults = outgoingResultsCache[managedRebalanceResult.Origin][originIdInt(managedRebalanceResult.OriginId)][channelIdInt(managedRebalanceResult.OutgoingChannelId)]
	}
	if len(rebalanceResults) != 0 {
		for i := len(rebalanceResults) - 1; i >= 0; i-- {
			if managedRebalanceResult.Status == nil || *managedRebalanceResult.Status == rebalanceResults[i].Status {
				return rebalanceResults[i]
			}
		}
	}
	return RebalanceResult{}
}

func getRebalanceLatestResultCache(managedRebalanceResult ManagedRebalanceResult,
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult) RebalanceResult {

	var latest RebalanceResult
	if managedRebalanceResult.IncomingChannelId != 0 {
		latest = processResultsCache(managedRebalanceResult, incomingResultsCache, latest)
	}
	if managedRebalanceResult.OutgoingChannelId != 0 {
		latest = processResultsCache(managedRebalanceResult, outgoingResultsCache, latest)
	}
	return latest
}

func processResultsCache(
	managedRebalanceResult ManagedRebalanceResult,
	resultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult,
	latest RebalanceResult) RebalanceResult {

	for origin := range resultsCache {
		for originId := range resultsCache[origin] {
			for channelId := range resultsCache[origin][originId] {
				rebalanceResults := resultsCache[origin][originId][channelId]
				// Reverse order, so we process the most recent result first.
				for i := len(rebalanceResults) - 1; i >= 0; i-- {
					if rebalanceResults[i].OutgoingChannelId == managedRebalanceResult.OutgoingChannelId &&
						rebalanceResults[i].IncomingChannelId == managedRebalanceResult.IncomingChannelId &&
						(managedRebalanceResult.Status == nil || *managedRebalanceResult.Status == rebalanceResults[i].Status) {
						if latest.RebalanceId == 0 {
							latest = rebalanceResults[i]
							break
						}
						if latest.UpdateOn.Before(rebalanceResults[i].UpdateOn) {
							latest = rebalanceResults[i]
							break
						}
					}
				}
			}
		}
	}
	return latest
}

func initializeResultsCache(managedRebalanceResult ManagedRebalanceResult,
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult) {

	if managedRebalanceResult.IncomingChannelId != 0 {
		if incomingResultsCache[managedRebalanceResult.Origin] == nil {
			incomingResultsCache[managedRebalanceResult.Origin] = make(map[originIdInt]map[channelIdInt][]RebalanceResult)
		}
		if incomingResultsCache[managedRebalanceResult.Origin][originIdInt(managedRebalanceResult.OriginId)] == nil {
			incomingResultsCache[managedRebalanceResult.Origin][originIdInt(managedRebalanceResult.OriginId)] = make(map[channelIdInt][]RebalanceResult)
		}
		return
	}
	if outgoingResultsCache[managedRebalanceResult.Origin] == nil {
		outgoingResultsCache[managedRebalanceResult.Origin] = make(map[originIdInt]map[channelIdInt][]RebalanceResult)
	}
	if outgoingResultsCache[managedRebalanceResult.Origin][originIdInt(managedRebalanceResult.OriginId)] == nil {
		outgoingResultsCache[managedRebalanceResult.Origin][originIdInt(managedRebalanceResult.OriginId)] = make(map[channelIdInt][]RebalanceResult)
	}
}

func isValidResultRequest(managedRebalanceResult ManagedRebalanceResult) bool {
	if managedRebalanceResult.Type == readRebalanceResultByOrigin {
		if managedRebalanceResult.IncomingChannelId != 0 && managedRebalanceResult.OutgoingChannelId != 0 {
			log.Error().Msgf("IncomingChannelId (%v) and OutgoingChannelId (%v) cannot both be populated at the same time",
				managedRebalanceResult.IncomingChannelId, managedRebalanceResult.OutgoingChannelId)
			return false
		}
		if managedRebalanceResult.OriginId == 0 {
			log.Error().Msgf("No empty OriginId (%v) allowed", managedRebalanceResult.OriginId)
			return false
		}
	}
	if managedRebalanceResult.IncomingChannelId == 0 && managedRebalanceResult.OutgoingChannelId == 0 {
		log.Error().Msgf("IncomingChannelId (%v) and OutgoingChannelId (%v) cannot both be 0",
			managedRebalanceResult.IncomingChannelId, managedRebalanceResult.OutgoingChannelId)
		return false
	}
	return true
}

func SendToManagedRebalanceResultChannel(ch chan<- RebalanceResult, rebalanceResult RebalanceResult) {
	ch <- rebalanceResult
	close(ch)
}

func getLatestResult(incomingChannelId int, outgoingChannelId int, status *commons.Status) RebalanceResult {
	responseChannel := make(chan RebalanceResult)
	managedRebalance := ManagedRebalanceResult{
		IncomingChannelId: incomingChannelId,
		OutgoingChannelId: outgoingChannelId,
		Status:            status,
		Type:              readRebalanceResult,
		Out:               responseChannel,
	}
	ManagedRebalanceResultChannel <- managedRebalance
	return <-responseChannel
}

func getLatestResultByOrigin(origin commons.RebalanceRequestOrigin, originId int,
	incomingChannelId int, outgoingChannelId int, status *commons.Status) RebalanceResult {
	responseChannel := make(chan RebalanceResult)
	managedRebalance := ManagedRebalanceResult{
		Origin:            origin,
		OriginId:          originId,
		IncomingChannelId: incomingChannelId,
		OutgoingChannelId: outgoingChannelId,
		Status:            status,
		Type:              readRebalanceResultByOrigin,
		Out:               responseChannel,
	}
	ManagedRebalanceResultChannel <- managedRebalance
	return <-responseChannel
}

func addRebalanceResult(origin commons.RebalanceRequestOrigin, originId int, incomingChannelId int, outgoingChannelId int,
	rebalanceResult RebalanceResult) {

	managedRebalance := ManagedRebalanceResult{
		Origin:            origin,
		OriginId:          originId,
		IncomingChannelId: incomingChannelId,
		OutgoingChannelId: outgoingChannelId,
		RebalanceResult:   rebalanceResult,
		Type:              writeRebalanceResult,
	}
	ManagedRebalanceResultChannel <- managedRebalance
}
