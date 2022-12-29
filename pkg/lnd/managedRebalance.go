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
	// rebalancers = map["workflow"]map[workflowVersionNodeId]Rebalancer
	rebalancers := make(map[commons.RebalanceRequestOrigin]map[int]*Rebalancer)
	// outgoingResultsCache = map["workflow"]map[workflowVersionNodeId][outgoingChannelId][]RebalanceResult
	outgoingResultsCache := make(map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult)
	// outgoingResultsCache = map["workflow"]map[workflowVersionNodeId][incomingChannelId][]RebalanceResult
	incomingResultsCache := make(map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult)

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
			case READ_REBALANCE_RESULT:
				if !isValidRequest(managedRebalance) {
					SendToManagedRebalanceResultChannel(managedRebalance.RebalanceResultOut, RebalanceResult{})
					continue
				}
				initializeResultsCache(managedRebalance, incomingResultsCache, outgoingResultsCache)
				rebalanceResults := getRebalanceResultsCache(managedRebalance, incomingResultsCache, outgoingResultsCache)
				SendToManagedRebalanceResultChannel(managedRebalance.RebalanceResultOut, getLatestResultWithStatus(rebalanceResults, managedRebalance))
			case WRITE_REBALANCER:
				managedRebalance = copyFromRebalancer(managedRebalance)
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
				managedRebalance = copyFromRebalancer(managedRebalance)
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
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult) {

	var results []RebalanceResult
	if managedRebalance.IncomingChannelId != 0 {
		results = incomingResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.IncomingChannelId]
	} else {
		results = outgoingResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.OutgoingChannelId]
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
		incomingResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.IncomingChannelId] =
			append(results, managedRebalance.RebalanceResult)
	} else {
		outgoingResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.OutgoingChannelId] =
			append(results, managedRebalance.RebalanceResult)
	}
}

func getRebalanceResultsCache(managedRebalance ManagedRebalance,
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult) []RebalanceResult {

	if managedRebalance.IncomingChannelId != 0 {
		return incomingResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.IncomingChannelId]
	}
	return outgoingResultsCache[managedRebalance.Origin][managedRebalance.OriginId][managedRebalance.OutgoingChannelId]
}

func removeRebalancersCache(managedRebalance ManagedRebalance,
	rebalancers map[commons.RebalanceRequestOrigin]map[int]*Rebalancer) {

	delete(rebalancers[managedRebalance.Origin], managedRebalance.OriginId)
}

func setRebalancersCache(managedRebalance ManagedRebalance,
	rebalancers map[commons.RebalanceRequestOrigin]map[int]*Rebalancer) {

	rebalancers[managedRebalance.Origin][managedRebalance.OriginId] = managedRebalance.Rebalancer
}

func getRebalancersCache(managedRebalance ManagedRebalance,
	rebalancers map[commons.RebalanceRequestOrigin]map[int]*Rebalancer) *Rebalancer {

	return rebalancers[managedRebalance.Origin][managedRebalance.OriginId]
}

func initializeResultsCache(managedRebalance ManagedRebalance,
	incomingResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult,
	outgoingResultsCache map[commons.RebalanceRequestOrigin]map[int]map[int][]RebalanceResult) {

	if managedRebalance.IncomingChannelId != 0 {
		if incomingResultsCache[managedRebalance.Origin] == nil {
			incomingResultsCache[managedRebalance.Origin] = make(map[int]map[int][]RebalanceResult)
		}
		if incomingResultsCache[managedRebalance.Origin][managedRebalance.OriginId] == nil {
			incomingResultsCache[managedRebalance.Origin][managedRebalance.OriginId] = make(map[int][]RebalanceResult)
		}
		return
	}
	if outgoingResultsCache[managedRebalance.Origin] == nil {
		outgoingResultsCache[managedRebalance.Origin] = make(map[int]map[int][]RebalanceResult)
	}
	if outgoingResultsCache[managedRebalance.Origin][managedRebalance.OriginId] == nil {
		outgoingResultsCache[managedRebalance.Origin][managedRebalance.OriginId] = make(map[int][]RebalanceResult)
	}
}

func initializeRebalancersCache(managedRebalance ManagedRebalance,
	rebalancers map[commons.RebalanceRequestOrigin]map[int]*Rebalancer) {

	if rebalancers[managedRebalance.Origin] == nil {
		rebalancers[managedRebalance.Origin] = make(map[int]*Rebalancer)
	}
}

func isValidRequest(managedRebalance ManagedRebalance) bool {
	if managedRebalance.IncomingChannelId != 0 && managedRebalance.OutgoingChannelId != 0 {
		log.Error().Msgf("IncomingChannelId (%v) and OutgoingChannelId (%v) cannot both be populated at the same time",
			managedRebalance.IncomingChannelId, managedRebalance.OutgoingChannelId)
		return false
	}
	if managedRebalance.OriginId == 0 {
		log.Error().Msgf("No empty OriginId (%v) allowed", managedRebalance.OriginId)
		return false
	}
	switch managedRebalance.Type {
	case WRITE_REBALANCE_RESULT:
		fallthrough
	case READ_REBALANCE_RESULT:
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
}

func SendToManagedRebalanceResultChannel(ch chan RebalanceResult, rebalanceResult RebalanceResult) {
	ch <- rebalanceResult
}

func getRebalancer(origin commons.RebalanceRequestOrigin, originId int) *Rebalancer {
	responseChannel := make(chan ManagedRebalance)
	managedRebalance := ManagedRebalance{
		Origin:   origin,
		OriginId: originId,
		Type:     READ_REBALANCER,
		Out:      responseChannel,
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
