package lnd

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
	WRITE_REBALANCER
	DELETE_REBALANCER
	READ_REBALANCE_RESULT
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
	Status             commons.Status
	Rebalancer         *Rebalancer
	RebalanceResult    RebalanceResult
	Out                chan ManagedRebalance
	BoolOut            chan bool
	RebalanceResultOut chan RebalanceResult
}

func ManagedRebalanceCache(ch chan ManagedRebalance, ctx context.Context) {
	// rebalancers = map["workflow"]map[workflowVersionNodeId]map[ChannelId]Rebalancer
	incomingRebalancers := make(map[commons.RebalanceRequestOrigin]map[int]map[int]*Rebalancer)
	incomingRebalanceResultsCache := make(map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult)

	outgoingRebalancers := make(map[commons.RebalanceRequestOrigin]map[int]map[int]*Rebalancer)
	outgoingRebalanceResultsCache := make(map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult)

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
				initializeRebalancersCache(managedRebalance, incomingRebalancers, outgoingRebalancers)
				managedRebalance.Rebalancer = getRebalancersCache(managedRebalance, incomingRebalancers, outgoingRebalancers)
				SendToManagedRebalanceChannel(managedRebalance.Out, managedRebalance)
			case READ_REBALANCE_RESULT:
				if !isValidRequest(managedRebalance) {
					SendToManagedRebalanceResultChannel(managedRebalance.RebalanceResultOut, RebalanceResult{})
					continue
				}
				initializeResultsCache(managedRebalance, incomingRebalanceResultsCache, outgoingRebalanceResultsCache)
				rebalanceResults := getRebalanceResultsCache(managedRebalance, incomingRebalanceResultsCache, outgoingRebalanceResultsCache)
				SendToManagedRebalanceResultChannel(managedRebalance.RebalanceResultOut, getLatestResultWithStatus(rebalanceResults, managedRebalance))
			case WRITE_REBALANCER:
				managedRebalance = copyFromRebalancer(managedRebalance)
				if !isValidRequest(managedRebalance) {
					commons.SendToManagedBoolChannel(managedRebalance.BoolOut, false)
					continue
				}
				initializeRebalancersCache(managedRebalance, incomingRebalancers, outgoingRebalancers)
				if getRebalancersCache(managedRebalance, incomingRebalancers, outgoingRebalancers) != nil {
					commons.SendToManagedBoolChannel(managedRebalance.BoolOut, false)
					continue
				}
				setRebalancersCache(managedRebalance, incomingRebalancers, outgoingRebalancers)
				commons.SendToManagedBoolChannel(managedRebalance.BoolOut, true)
			case WRITE_REBALANCE_RESULT:
				if !isValidRequest(managedRebalance) {
					continue
				}
				initializeResultsCache(managedRebalance, incomingRebalanceResultsCache, outgoingRebalanceResultsCache)
				appendRebalanceResult(managedRebalance, incomingRebalanceResultsCache, outgoingRebalanceResultsCache)
			case DELETE_REBALANCER:
				managedRebalance = copyFromRebalancer(managedRebalance)
				if !isValidRequest(managedRebalance) {
					continue
				}
				initializeRebalancersCache(managedRebalance, incomingRebalancers, outgoingRebalancers)
				if getRebalancersCache(managedRebalance, incomingRebalancers, outgoingRebalancers) == nil {
					continue
				}
				removeRebalancersCache(managedRebalance, incomingRebalancers, outgoingRebalancers)
			}
		}
	}
}

func copyFromRebalancer(managedRebalance ManagedRebalance) ManagedRebalance {
	switch managedRebalance.Type {
	case DELETE_REBALANCER:
		fallthrough
	case WRITE_REBALANCER:
		managedRebalance.Origin = managedRebalance.Rebalancer.Origin
		managedRebalance.OriginId = managedRebalance.Rebalancer.OriginId
		managedRebalance.OriginReference = managedRebalance.Rebalancer.OriginReference
		managedRebalance.IncomingChannelId = managedRebalance.Rebalancer.IncomingChannelId
		managedRebalance.OutgoingChannelId = managedRebalance.Rebalancer.OutgoingChannelId
	}
	return managedRebalance
}

func getLatestResultWithStatus(rebalanceResults []RebalanceResult, managedRebalance ManagedRebalance) RebalanceResult {
	if len(rebalanceResults) != 0 {
		for i := len(rebalanceResults) - 1; i < 0; i-- {
			if rebalanceResults[i].Status == managedRebalance.Status {
				return rebalanceResults[i]
			}
		}
	}
	return RebalanceResult{}
}

func appendRebalanceResult(managedRebalance ManagedRebalance,
	incomingRebalanceResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult,
	outgoingRebalanceResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult) {

	var results []RebalanceResult
	if managedRebalance.IncomingChannelId != 0 {
		results = incomingRebalanceResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.IncomingChannelId]
	} else {
		results = outgoingRebalanceResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.OutgoingChannelId]
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
		incomingRebalanceResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.IncomingChannelId] =
			append(results, managedRebalance.RebalanceResult)
	} else {
		outgoingRebalanceResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.OutgoingChannelId] =
			append(results, managedRebalance.RebalanceResult)
	}
}

func getRebalanceResultsCache(managedRebalance ManagedRebalance,
	incomingRebalanceResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult,
	outgoingRebalanceResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult) []RebalanceResult {

	if managedRebalance.IncomingChannelId != 0 {
		return incomingRebalanceResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.IncomingChannelId]
	}
	return outgoingRebalanceResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.OutgoingChannelId]
}

func removeRebalancersCache(managedRebalance ManagedRebalance,
	incomingRebalancers map[commons.RebalanceRequestOrigin]map[int]map[int]*Rebalancer,
	outgoingRebalancers map[commons.RebalanceRequestOrigin]map[int]map[int]*Rebalancer) {

	if managedRebalance.IncomingChannelId != 0 {
		delete(incomingRebalancers[managedRebalance.Origin][managedRebalance.OriginId], managedRebalance.IncomingChannelId)
		return
	}
	delete(outgoingRebalancers[managedRebalance.Origin][managedRebalance.OriginId], managedRebalance.OutgoingChannelId)
}

func setRebalancersCache(managedRebalance ManagedRebalance,
	incomingRebalancers map[commons.RebalanceRequestOrigin]map[int]map[int]*Rebalancer,
	outgoingRebalancers map[commons.RebalanceRequestOrigin]map[int]map[int]*Rebalancer) {

	if managedRebalance.IncomingChannelId != 0 {
		incomingRebalancers[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.IncomingChannelId] = managedRebalance.Rebalancer
		return
	}
	outgoingRebalancers[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.OutgoingChannelId] = managedRebalance.Rebalancer
}

func getRebalancersCache(managedRebalance ManagedRebalance,
	incomingRebalancers map[commons.RebalanceRequestOrigin]map[int]map[int]*Rebalancer,
	outgoingRebalancers map[commons.RebalanceRequestOrigin]map[int]map[int]*Rebalancer) *Rebalancer {

	if managedRebalance.IncomingChannelId != 0 {
		return incomingRebalancers[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.IncomingChannelId]
	}
	return outgoingRebalancers[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.OutgoingChannelId]
}

func initializeResultsCache(managedRebalance ManagedRebalance,
	incomingRebalanceResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult,
	outgoingRebalanceResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult,
) {

	if managedRebalance.IncomingChannelId != 0 {
		if incomingRebalanceResultsCache[managedRebalance.Origin] == nil {
			incomingRebalanceResultsCache[managedRebalance.Origin] = make(map[int]map[int][]RebalanceResult)
		}
		if incomingRebalanceResultsCache[managedRebalance.Origin][managedRebalance.OriginId] == nil {
			incomingRebalanceResultsCache[managedRebalance.Origin][managedRebalance.OriginId] = make(map[int][]RebalanceResult)
		}
		return
	}
	if outgoingRebalanceResultsCache[managedRebalance.Origin] == nil {
		outgoingRebalanceResultsCache[managedRebalance.Origin] = make(map[int]map[int][]RebalanceResult)
	}
	if outgoingRebalanceResultsCache[managedRebalance.Origin][managedRebalance.OriginId] == nil {
		outgoingRebalanceResultsCache[managedRebalance.Origin][managedRebalance.OriginId] = make(map[int][]RebalanceResult)
	}
}

func initializeRebalancersCache(managedRebalance ManagedRebalance,
	incomingRebalancers map[commons.RebalanceRequestOrigin]map[int]map[int]*Rebalancer,
	outgoingRebalancers map[commons.RebalanceRequestOrigin]map[int]map[int]*Rebalancer) {

	if managedRebalance.IncomingChannelId != 0 {
		if incomingRebalancers[managedRebalance.Origin] == nil {
			incomingRebalancers[managedRebalance.Origin] = make(map[int]map[int]*Rebalancer)
		}
		if incomingRebalancers[managedRebalance.Origin][managedRebalance.OriginId] == nil {
			incomingRebalancers[managedRebalance.Origin][managedRebalance.OriginId] = make(map[int]*Rebalancer)
		}
		return
	}

	if outgoingRebalancers[managedRebalance.Origin] == nil {
		outgoingRebalancers[managedRebalance.Origin] = make(map[int]map[int]*Rebalancer)
	}
	if outgoingRebalancers[managedRebalance.Origin][managedRebalance.OriginId] == nil {
		outgoingRebalancers[managedRebalance.Origin][managedRebalance.OriginId] = make(map[int]*Rebalancer)
	}
}

func isValidRequest(managedRebalance ManagedRebalance) bool {
	if managedRebalance.IncomingChannelId != 0 && managedRebalance.OutgoingChannelId != 0 {
		log.Error().Msgf("IncomingChannelId (%v) and OutgoingChannelId (%v) cannot both be populated at the same time",
			managedRebalance.IncomingChannelId, managedRebalance.OutgoingChannelId)
		return false
	}
	if managedRebalance.OriginId == 0 ||
		(managedRebalance.IncomingChannelId == 0 && managedRebalance.OutgoingChannelId == 0) {
		log.Error().Msgf("No empty OriginId (%v) or IncomingChannelId (%v) and OutgoingChannelId (%v) allowed",
			managedRebalance.OriginId, managedRebalance.IncomingChannelId, managedRebalance.OutgoingChannelId)
		return false
	}
	return true
}

func SendToManagedRebalanceChannel(ch chan ManagedRebalance, managedRebalance ManagedRebalance) {
	ch <- managedRebalance
}

func SendToManagedRebalanceResultChannel(ch chan RebalanceResult, rebalanceResult RebalanceResult) {
	ch <- rebalanceResult
}

func getRebalancer(origin commons.RebalanceRequestOrigin, originId int, incomingChannelId int, outgoingChannelId int) *Rebalancer {
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

func getLatestResult(origin commons.RebalanceRequestOrigin, originId int,
	incomingChannelId int, outgoingChannelId int, status commons.Status) RebalanceResult {
	responseChannel := make(chan RebalanceResult)
	managedRebalance := ManagedRebalance{
		Origin:             origin,
		OriginId:           originId,
		IncomingChannelId:  incomingChannelId,
		OutgoingChannelId:  outgoingChannelId,
		Status:             status,
		Type:               READ_REBALANCE_RESULT,
		RebalanceResultOut: responseChannel,
	}
	ManagedRebalanceChannel <- managedRebalance
	return <-responseChannel
}

func addRebalancer(rebalancer *Rebalancer) bool {
	responseChannel := make(chan bool)
	managedRebalance := ManagedRebalance{
		Rebalancer: rebalancer,
		Type:       WRITE_REBALANCER,
		BoolOut:    responseChannel,
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

func removeRebalancer(rebalancer *Rebalancer) {
	managedRebalance := ManagedRebalance{
		Rebalancer: rebalancer,
		Type:       DELETE_REBALANCER,
	}
	ManagedRebalanceChannel <- managedRebalance
}
