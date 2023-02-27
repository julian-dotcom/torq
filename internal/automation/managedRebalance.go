package automation

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/commons"
)

var ManagedRebalanceChannel = make(chan ManagedRebalance) //nolint:gochecknoglobals

type ManagedRebalanceCacheOperationType uint

const (
	READ_REBALANCER ManagedRebalanceCacheOperationType = iota
	READ_REBALANCERS
	WRITE_REBALANCER
	DELETE_REBALANCER
	READ_REBALANCE_RESULT
	READ_REBALANCE_RESULT_BY_ORIGIN
	WRITE_REBALANCE_RESULT
)

type ManagedRebalance struct {
	Type               ManagedRebalanceCacheOperationType
	Origin             commons.RebalanceRequestOrigin
	OriginId           int
	OriginReference    string
	IncomingChannelId  int
	IncomingPublicKey  string
	OutgoingChannelId  int
	OutgoingPublicKey  string
	AmountMsat         uint64
	Status             *commons.Status
	Rebalancer         *Rebalancer
	RebalanceResult    RebalanceResult
	Out                chan ManagedRebalance
	BoolOut            chan bool
	RebalanceResultOut chan RebalanceResult
	RebalancersOut     chan []*Rebalancer
}

// originId is workflowVersionNodeId in case of workflow triggered rebalances.
type originIdInt int
type channelIdInt int

func ManagedRebalanceCache(ch chan ManagedRebalance, ctx context.Context) {
	rebalancers := make(map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer)
	outgoingResultsCache := make(map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult)
	incomingResultsCache := make(map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult)

	for {
		select {
		case <-ctx.Done():
			return
		case managedRebalance := <-ch:
			switch managedRebalance.Type {
			case READ_REBALANCER:
				if !isValidRequest(managedRebalance) {
					SendToManagedRebalanceChannel(managedRebalance.Out, managedRebalance)
					continue
				}
				initializeRebalancersCache(managedRebalance, rebalancers)
				managedRebalance.Rebalancer = getRebalancersCache(managedRebalance, rebalancers)
				SendToManagedRebalanceChannel(managedRebalance.Out, managedRebalance)
			case READ_REBALANCERS:
				initializeRebalancersCache(managedRebalance, rebalancers)
				var rebalancersArray []*Rebalancer
				for _, originIdRebalancers := range rebalancers {
					for _, focusChannelIdRebalancers := range originIdRebalancers {
						for _, rebalancer := range focusChannelIdRebalancers {
							if managedRebalance.Status != nil && *managedRebalance.Status != rebalancer.Status {
								continue
							}
							rebalancersArray = append(rebalancersArray, rebalancer)
						}
					}
				}
				SendToRebalancersChannel(managedRebalance.RebalancersOut, rebalancersArray)
			case READ_REBALANCE_RESULT:
				if !isValidRequest(managedRebalance) {
					SendToManagedRebalanceResultChannel(managedRebalance.RebalanceResultOut, RebalanceResult{})
					continue
				}
				initializeResultsCache(managedRebalance, incomingResultsCache, outgoingResultsCache)
				SendToManagedRebalanceResultChannel(managedRebalance.RebalanceResultOut,
					getRebalanceLatestResultCache(managedRebalance, incomingResultsCache, outgoingResultsCache))
			case READ_REBALANCE_RESULT_BY_ORIGIN:
				if !isValidRequest(managedRebalance) {
					SendToManagedRebalanceResultChannel(managedRebalance.RebalanceResultOut, RebalanceResult{})
					continue
				}
				initializeResultsCache(managedRebalance, incomingResultsCache, outgoingResultsCache)
				SendToManagedRebalanceResultChannel(managedRebalance.RebalanceResultOut,
					getRebalanceLatestResultByOriginCache(managedRebalance, incomingResultsCache, outgoingResultsCache))
			case WRITE_REBALANCER:
				if !isValidRequest(managedRebalance) {
					commons.SendToManagedBoolChannel(managedRebalance.BoolOut, false)
					continue
				}
				initializeRebalancersCache(managedRebalance, rebalancers)
				if getRebalancersCache(managedRebalance, rebalancers) != nil {
					commons.SendToManagedBoolChannel(managedRebalance.BoolOut, false)
					continue
				}
				setRebalancersCache(managedRebalance, rebalancers)
				commons.SendToManagedBoolChannel(managedRebalance.BoolOut, true)
			case WRITE_REBALANCE_RESULT:
				if !isValidRequest(managedRebalance) {
					continue
				}
				initializeResultsCache(managedRebalance, incomingResultsCache, outgoingResultsCache)
				appendRebalanceResult(managedRebalance, incomingResultsCache, outgoingResultsCache)
			case DELETE_REBALANCER:
				if !isValidRequest(managedRebalance) {
					continue
				}
				initializeRebalancersCache(managedRebalance, rebalancers)
				if getRebalancersCache(managedRebalance, rebalancers) == nil {
					continue
				}
				removeRebalancersCache(managedRebalance, rebalancers)
			}
		}
	}
}

func copyFromRebalancer(managedRebalance ManagedRebalance) ManagedRebalance {
	managedRebalance.Origin = managedRebalance.Rebalancer.Request.Origin
	managedRebalance.OriginId = managedRebalance.Rebalancer.Request.OriginId
	managedRebalance.OriginReference = managedRebalance.Rebalancer.Request.OriginReference
	managedRebalance.IncomingChannelId = managedRebalance.Rebalancer.Request.IncomingChannelId
	managedRebalance.OutgoingChannelId = managedRebalance.Rebalancer.Request.OutgoingChannelId
	return managedRebalance
}

func appendRebalanceResult(managedRebalance ManagedRebalance,
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult) {

	var results []RebalanceResult
	if managedRebalance.IncomingChannelId != 0 {
		results = incomingResultsCache[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)][channelIdInt(managedRebalance.IncomingChannelId)]
	} else {
		results = outgoingResultsCache[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)][channelIdInt(managedRebalance.OutgoingChannelId)]
	}
	if len(results)%100 == 0 {
		for i := 0; i < len(results); i++ {
			result := results[i]
			if time.Since(result.UpdateOn).Seconds() < commons.REBALANCE_RESULTS_TIMEOUT_SECONDS {
				results = results[i:]
				break
			}
		}
	}
	if managedRebalance.IncomingChannelId != 0 {
		incomingResultsCache[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)][channelIdInt(managedRebalance.IncomingChannelId)] =
			append(results, managedRebalance.RebalanceResult)
	} else {
		outgoingResultsCache[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)][channelIdInt(managedRebalance.OutgoingChannelId)] =
			append(results, managedRebalance.RebalanceResult)
	}
}

func getRebalanceLatestResultByOriginCache(managedRebalance ManagedRebalance,
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult) RebalanceResult {

	var rebalanceResults []RebalanceResult
	if managedRebalance.IncomingChannelId != 0 {
		rebalanceResults = incomingResultsCache[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)][channelIdInt(managedRebalance.IncomingChannelId)]
	} else {
		rebalanceResults = outgoingResultsCache[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)][channelIdInt(managedRebalance.OutgoingChannelId)]
	}
	if len(rebalanceResults) != 0 {
		for i := len(rebalanceResults) - 1; i >= 0; i-- {
			if managedRebalance.Status == nil || *managedRebalance.Status == rebalanceResults[i].Status {
				return rebalanceResults[i]
			}
		}
	}
	return RebalanceResult{}
}

func getRebalanceLatestResultCache(managedRebalance ManagedRebalance,
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult) RebalanceResult {

	var latest RebalanceResult
	if managedRebalance.IncomingChannelId != 0 {
		latest = processResultsCache(managedRebalance, incomingResultsCache, latest)
	}
	if managedRebalance.OutgoingChannelId != 0 {
		latest = processResultsCache(managedRebalance, outgoingResultsCache, latest)
	}
	return latest
}

func processResultsCache(
	managedRebalance ManagedRebalance,
	resultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult,
	latest RebalanceResult) RebalanceResult {

	for origin := range resultsCache {
		for originId := range resultsCache[origin] {
			for channelId := range resultsCache[origin][originId] {
				rebalanceResults := resultsCache[origin][originId][channelId]
				// Reverse order, so we process the most recent result first.
				for i := len(rebalanceResults) - 1; i >= 0; i-- {
					if rebalanceResults[i].OutgoingChannelId == managedRebalance.OutgoingChannelId &&
						rebalanceResults[i].IncomingChannelId == managedRebalance.IncomingChannelId &&
						(managedRebalance.Status == nil || *managedRebalance.Status == rebalanceResults[i].Status) {
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

func removeRebalancersCache(managedRebalance ManagedRebalance,
	rebalancers map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) {

	_, exists := rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)]
	if exists {
		if managedRebalance.Rebalancer.Request.IncomingChannelId != 0 {
			delete(rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)], channelIdInt(managedRebalance.Rebalancer.Request.IncomingChannelId))
		}
		if managedRebalance.Rebalancer.Request.OutgoingChannelId != 0 {
			delete(rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)], channelIdInt(managedRebalance.Rebalancer.Request.OutgoingChannelId))
		}
		if len(rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)]) == 0 {
			delete(rebalancers[managedRebalance.Origin], originIdInt(managedRebalance.OriginId))
		}
	}
}

func setRebalancersCache(managedRebalance ManagedRebalance,
	rebalancers map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) {

	_, exists := rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)]
	if !exists {
		rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)] = make(map[channelIdInt]*Rebalancer)
	}
	if managedRebalance.Rebalancer.Request.IncomingChannelId != 0 {
		rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)][channelIdInt(managedRebalance.Rebalancer.Request.IncomingChannelId)] = managedRebalance.Rebalancer
	}
	if managedRebalance.Rebalancer.Request.OutgoingChannelId != 0 {
		rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)][channelIdInt(managedRebalance.Rebalancer.Request.OutgoingChannelId)] = managedRebalance.Rebalancer
	}
}

func getRebalancersCache(managedRebalance ManagedRebalance,
	rebalancers map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) *Rebalancer {

	_, exists := rebalancers[managedRebalance.Origin]
	if !exists {
		return nil
	}
	_, exists = rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)]
	if !exists {
		return nil
	}
	var rebalancer *Rebalancer
	if managedRebalance.IncomingChannelId != 0 {
		rebalancer, exists = rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)][channelIdInt(managedRebalance.IncomingChannelId)]
		if !exists {
			return nil
		}
	}
	if managedRebalance.OutgoingChannelId != 0 {
		rebalancer, exists = rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)][channelIdInt(managedRebalance.OutgoingChannelId)]
		if !exists {
			return nil
		}
	}
	return rebalancer

}

func initializeResultsCache(managedRebalance ManagedRebalance,
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt][]RebalanceResult) {

	if managedRebalance.IncomingChannelId != 0 {
		if incomingResultsCache[managedRebalance.Origin] == nil {
			incomingResultsCache[managedRebalance.Origin] = make(map[originIdInt]map[channelIdInt][]RebalanceResult)
		}
		if incomingResultsCache[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)] == nil {
			incomingResultsCache[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)] = make(map[channelIdInt][]RebalanceResult)
		}
		return
	}
	if outgoingResultsCache[managedRebalance.Origin] == nil {
		outgoingResultsCache[managedRebalance.Origin] = make(map[originIdInt]map[channelIdInt][]RebalanceResult)
	}
	if outgoingResultsCache[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)] == nil {
		outgoingResultsCache[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)] = make(map[channelIdInt][]RebalanceResult)
	}
}

func initializeRebalancersCache(managedRebalance ManagedRebalance,
	rebalancers map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) {

	if rebalancers[managedRebalance.Origin] == nil {
		rebalancers[managedRebalance.Origin] = make(map[originIdInt]map[channelIdInt]*Rebalancer)
	}
}

func isValidRequest(managedRebalance ManagedRebalance) bool {
	if managedRebalance.Type != READ_REBALANCE_RESULT && managedRebalance.IncomingChannelId != 0 && managedRebalance.OutgoingChannelId != 0 {
		log.Error().Msgf("IncomingChannelId (%v) and OutgoingChannelId (%v) cannot both be populated at the same time",
			managedRebalance.IncomingChannelId, managedRebalance.OutgoingChannelId)
		return false
	}
	if managedRebalance.Type != READ_REBALANCE_RESULT && managedRebalance.OriginId == 0 {
		log.Error().Msgf("No empty OriginId (%v) allowed", managedRebalance.OriginId)
		return false
	}
	switch managedRebalance.Type {
	case WRITE_REBALANCER:
		fallthrough
	case DELETE_REBALANCER:
		fallthrough
	case READ_REBALANCER:
		fallthrough
	case WRITE_REBALANCE_RESULT:
		fallthrough
	case READ_REBALANCE_RESULT:
		fallthrough
	case READ_REBALANCE_RESULT_BY_ORIGIN:
		if managedRebalance.IncomingChannelId == 0 && managedRebalance.OutgoingChannelId == 0 {
			log.Error().Msgf("IncomingChannelId (%v) and OutgoingChannelId (%v) cannot both be 0",
				managedRebalance.IncomingChannelId, managedRebalance.OutgoingChannelId)
			return false
		}
	}
	return true
}

func SendToManagedRebalanceChannel(ch chan ManagedRebalance, managedRebalance ManagedRebalance) {
	ch <- managedRebalance
	close(ch)
}

func SendToRebalancersChannel(ch chan []*Rebalancer, rebalancers []*Rebalancer) {
	ch <- rebalancers
	close(ch)
}

func SendToManagedRebalanceResultChannel(ch chan RebalanceResult, rebalanceResult RebalanceResult) {
	ch <- rebalanceResult
	close(ch)
}

func getRebalancers(status *commons.Status) []*Rebalancer {
	responseChannel := make(chan []*Rebalancer)
	managedRebalance := ManagedRebalance{
		Status:         status,
		Type:           READ_REBALANCERS,
		RebalancersOut: responseChannel,
	}
	ManagedRebalanceChannel <- managedRebalance
	return <-responseChannel
}

func getRebalancer(origin commons.RebalanceRequestOrigin, originId int,
	incomingChannelId int,
	outgoingChannelId int) *Rebalancer {

	responseChannel := make(chan ManagedRebalance)
	managedRebalance := ManagedRebalance{
		Origin:            origin,
		OriginId:          originId,
		IncomingChannelId: incomingChannelId,
		OutgoingChannelId: outgoingChannelId,
		Type:              READ_REBALANCER,
		Out:               responseChannel,
	}
	ManagedRebalanceChannel <- managedRebalance
	response := <-responseChannel
	return response.Rebalancer
}

func addRebalancer(rebalancer *Rebalancer) bool {
	responseChannel := make(chan bool)
	managedRebalance := ManagedRebalance{
		Rebalancer: rebalancer,
		Type:       WRITE_REBALANCER,
		BoolOut:    responseChannel,
	}
	managedRebalance = copyFromRebalancer(managedRebalance)
	ManagedRebalanceChannel <- managedRebalance
	return <-responseChannel
}

func removeRebalancer(rebalancer *Rebalancer) {
	managedRebalance := ManagedRebalance{
		Rebalancer: rebalancer,
		Type:       DELETE_REBALANCER,
	}
	managedRebalance = copyFromRebalancer(managedRebalance)
	ManagedRebalanceChannel <- managedRebalance
}

func getLatestResult(incomingChannelId int, outgoingChannelId int, status *commons.Status) RebalanceResult {
	responseChannel := make(chan RebalanceResult)
	managedRebalance := ManagedRebalance{
		IncomingChannelId:  incomingChannelId,
		OutgoingChannelId:  outgoingChannelId,
		Status:             status,
		Type:               READ_REBALANCE_RESULT,
		RebalanceResultOut: responseChannel,
	}
	ManagedRebalanceChannel <- managedRebalance
	return <-responseChannel
}

func getLatestResultByOrigin(origin commons.RebalanceRequestOrigin, originId int,
	incomingChannelId int, outgoingChannelId int, status *commons.Status) RebalanceResult {
	responseChannel := make(chan RebalanceResult)
	managedRebalance := ManagedRebalance{
		Origin:             origin,
		OriginId:           originId,
		IncomingChannelId:  incomingChannelId,
		OutgoingChannelId:  outgoingChannelId,
		Status:             status,
		Type:               READ_REBALANCE_RESULT_BY_ORIGIN,
		RebalanceResultOut: responseChannel,
	}
	ManagedRebalanceChannel <- managedRebalance
	return <-responseChannel
}

func addRebalanceResult(origin commons.RebalanceRequestOrigin, originId int, incomingChannelId int, outgoingChannelId int,
	rebalanceResult RebalanceResult) {

	managedRebalance := ManagedRebalance{
		Origin:            origin,
		OriginId:          originId,
		IncomingChannelId: incomingChannelId,
		OutgoingChannelId: outgoingChannelId,
		RebalanceResult:   rebalanceResult,
		Type:              WRITE_REBALANCE_RESULT,
	}
	ManagedRebalanceChannel <- managedRebalance
}
