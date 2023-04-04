package workflows

import (
	"context"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/pkg/core"
)

var RebalancesCacheChannel = make(chan RebalanceCache) //nolint:gochecknoglobals

type RebalanceCacheOperationType uint
type originIdInt int
type channelIdInt int

const (
	readRebalancerOperation RebalanceCacheOperationType = iota
	readRebalancersOperation
	writeRebalancerOperation
	deleteRebalancerOperation
	cancelRebalancerOperation
	cancelRebalancersOperation
	cancelRebalancersByOriginIdOperation
)

type RebalanceCache struct {
	Type              RebalanceCacheOperationType
	Origin            core.RebalanceRequestOrigin
	OriginId          int
	OriginReference   string
	IncomingChannelId int
	OutgoingChannelId int
	AmountMsat        uint64
	ChannelIds        []int
	Status            *core.Status
	Rebalancer        *Rebalancer
	Out               chan<- RebalanceCache
	BoolOut           chan<- bool
	RebalancersOut    chan<- []*Rebalancer
}

func RebalanceCacheHandler(ch <-chan RebalanceCache, ctx context.Context) {
	rebalancers := make(map[core.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer)

	for {
		select {
		case <-ctx.Done():
			return
		case rebalanceCache := <-ch:
			switch rebalanceCache.Type {
			case readRebalancerOperation:
				if !isValidRequest(rebalanceCache) {
					rebalanceCache.Out <- rebalanceCache
					close(rebalanceCache.Out)
					continue
				}
				initializeRebalancersCache(rebalanceCache, rebalancers)
				rebalanceCache.Rebalancer = getRebalancerCache(rebalanceCache, rebalancers)
				rebalanceCache.Out <- rebalanceCache
				close(rebalanceCache.Out)
			case readRebalancersOperation:
				initializeRebalancersCache(rebalanceCache, rebalancers)
				var rebalancersArray []*Rebalancer
				for _, originIdRebalancers := range rebalancers {
					for _, focusChannelIdRebalancers := range originIdRebalancers {
						for _, rebalancer := range focusChannelIdRebalancers {
							if rebalanceCache.Status != nil && *rebalanceCache.Status != rebalancer.Status {
								continue
							}
							rebalancersArray = append(rebalancersArray, rebalancer)
						}
					}
				}
				rebalanceCache.RebalancersOut <- rebalancersArray
				close(rebalanceCache.RebalancersOut)
			case writeRebalancerOperation:
				if !isValidRequest(rebalanceCache) {
					rebalanceCache.BoolOut <- false
					close(rebalanceCache.BoolOut)
					continue
				}
				initializeRebalancersCache(rebalanceCache, rebalancers)
				if getRebalancerCache(rebalanceCache, rebalancers) != nil {
					rebalanceCache.BoolOut <- false
					close(rebalanceCache.BoolOut)
					continue
				}
				setRebalancersCache(rebalanceCache, rebalancers)
				rebalanceCache.BoolOut <- true
				close(rebalanceCache.BoolOut)
			case deleteRebalancerOperation:
				if !isValidRequest(rebalanceCache) {
					continue
				}
				initializeRebalancersCache(rebalanceCache, rebalancers)
				if getRebalancerCache(rebalanceCache, rebalancers) == nil {
					continue
				}
				removeRebalancersCache(rebalanceCache, rebalancers)
			case cancelRebalancerOperation:
				if rebalanceCache.OriginId == 0 || len(rebalanceCache.ChannelIds) != 1 {
					continue
				}
				initializeRebalancersCache(rebalanceCache, rebalancers)
				rebalanceCache.IncomingChannelId = rebalanceCache.ChannelIds[0]
				rebalancer := getRebalancerCache(rebalanceCache, rebalancers)
				if rebalancer != nil {
					log.Debug().Msgf("Cancelling rebalancer for channelId: %v, origin: %v, originId: %v",
						rebalanceCache.ChannelIds[0], rebalanceCache.Origin, rebalanceCache.OriginId)
					rebalancer.RebalanceCancel()
					delete(rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)], channelIdInt(rebalanceCache.ChannelIds[0]))
				}
			case cancelRebalancersOperation:
				if rebalanceCache.OriginId == 0 {
					continue
				}
				initializeRebalancersCache(rebalanceCache, rebalancers)
				for channelId, rebalancer := range getRebalancersCache(rebalanceCache, rebalancers) {
					if slices.Contains(rebalanceCache.ChannelIds, int(channelId)) {
						continue
					}
					log.Debug().Msgf("Cancelling rebalancer for channelId: %v, origin: %v, originId: %v",
						channelId, rebalanceCache.Origin, rebalanceCache.OriginId)
					rebalancer.RebalanceCancel()
					delete(rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)], channelId)
				}
			case cancelRebalancersByOriginIdOperation:
				if rebalanceCache.OriginId == 0 {
					continue
				}
				initializeRebalancersCache(rebalanceCache, rebalancers)
				_, exists := rebalancers[rebalanceCache.Origin]
				if !exists {
					continue
				}
				rebalancersForOriginId, exists := rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)]
				if !exists {
					continue
				}
				for channelId, rebalancer := range rebalancersForOriginId {
					log.Debug().Msgf("Cancelling rebalancer for channelId: %v, origin: %v, originId: %v",
						channelId, rebalanceCache.Origin, rebalanceCache.OriginId)
					rebalancer.RebalanceCancel()
					delete(rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)], channelId)
				}
			}
		}
	}
}

func copyFromRebalancer(rebalanceCache RebalanceCache) RebalanceCache {
	rebalanceCache.Origin = rebalanceCache.Rebalancer.Request.Origin
	rebalanceCache.OriginId = rebalanceCache.Rebalancer.Request.OriginId
	rebalanceCache.OriginReference = rebalanceCache.Rebalancer.Request.OriginReference
	rebalanceCache.IncomingChannelId = rebalanceCache.Rebalancer.Request.IncomingChannelId
	rebalanceCache.OutgoingChannelId = rebalanceCache.Rebalancer.Request.OutgoingChannelId
	return rebalanceCache
}

func removeRebalancersCache(rebalanceCache RebalanceCache,
	rebalancers map[core.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) {

	_, exists := rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)]
	if exists {
		if rebalanceCache.Rebalancer.Request.IncomingChannelId != 0 {
			delete(rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)], channelIdInt(rebalanceCache.Rebalancer.Request.IncomingChannelId))
		}
		if rebalanceCache.Rebalancer.Request.OutgoingChannelId != 0 {
			delete(rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)], channelIdInt(rebalanceCache.Rebalancer.Request.OutgoingChannelId))
		}
		if len(rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)]) == 0 {
			delete(rebalancers[rebalanceCache.Origin], originIdInt(rebalanceCache.OriginId))
		}
	}
}

func setRebalancersCache(rebalanceCache RebalanceCache,
	rebalancers map[core.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) {

	_, exists := rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)]
	if !exists {
		rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)] = make(map[channelIdInt]*Rebalancer)
	}
	if rebalanceCache.Rebalancer.Request.IncomingChannelId != 0 {
		rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)][channelIdInt(rebalanceCache.Rebalancer.Request.IncomingChannelId)] = rebalanceCache.Rebalancer
	}
	if rebalanceCache.Rebalancer.Request.OutgoingChannelId != 0 {
		rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)][channelIdInt(rebalanceCache.Rebalancer.Request.OutgoingChannelId)] = rebalanceCache.Rebalancer
	}
}

func getRebalancersCache(rebalanceCache RebalanceCache,
	rebalancers map[core.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) map[channelIdInt]*Rebalancer {

	_, exists := rebalancers[rebalanceCache.Origin]
	if !exists {
		return nil
	}
	_, exists = rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)]
	if !exists {
		return nil
	}
	results := make(map[channelIdInt]*Rebalancer)
	for channelId, rebalancer := range rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)] {
		results[channelId] = rebalancer
	}
	return results

}

func getRebalancerCache(rebalanceCache RebalanceCache,
	rebalancers map[core.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) *Rebalancer {

	_, exists := rebalancers[rebalanceCache.Origin]
	if !exists {
		return nil
	}
	_, exists = rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)]
	if !exists {
		return nil
	}
	var rebalancer *Rebalancer
	if rebalanceCache.IncomingChannelId != 0 {
		rebalancer, exists = rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)][channelIdInt(rebalanceCache.IncomingChannelId)]
		if !exists {
			return nil
		}
	}
	if rebalanceCache.OutgoingChannelId != 0 {
		rebalancer, exists = rebalancers[rebalanceCache.Origin][originIdInt(rebalanceCache.OriginId)][channelIdInt(rebalanceCache.OutgoingChannelId)]
		if !exists {
			return nil
		}
	}
	return rebalancer

}

func initializeRebalancersCache(rebalanceCache RebalanceCache,
	rebalancers map[core.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) {

	if rebalancers[rebalanceCache.Origin] == nil {
		rebalancers[rebalanceCache.Origin] = make(map[originIdInt]map[channelIdInt]*Rebalancer)
	}
}

func isValidRequest(rebalanceCache RebalanceCache) bool {
	if rebalanceCache.Type != readRebalancersOperation && rebalanceCache.IncomingChannelId == 0 && rebalanceCache.OutgoingChannelId == 0 {
		log.Error().Msgf("IncomingChannelId (%v) and OutgoingChannelId (%v) cannot both be 0",
			rebalanceCache.IncomingChannelId, rebalanceCache.OutgoingChannelId)
		return false
	}
	return true
}

func CancelRebalancersExcept(origin core.RebalanceRequestOrigin, originId int, activeChannelIds []int) {
	rebalanceCache := RebalanceCache{
		Origin:     origin,
		OriginId:   originId,
		ChannelIds: activeChannelIds,
		Type:       cancelRebalancersOperation,
	}
	RebalancesCacheChannel <- rebalanceCache
}

func CancelRebalancer(origin core.RebalanceRequestOrigin, originId int, channelId int) {
	rebalanceCache := RebalanceCache{
		Origin:     origin,
		OriginId:   originId,
		ChannelIds: []int{channelId},
		Type:       cancelRebalancerOperation,
	}
	RebalancesCacheChannel <- rebalanceCache
}

func cancelRebalancersByOriginIds(origin core.RebalanceRequestOrigin, originIds []int) {
	for _, originId := range originIds {
		rebalanceCache := RebalanceCache{
			Origin:   origin,
			OriginId: originId,
			Type:     cancelRebalancersByOriginIdOperation,
		}
		RebalancesCacheChannel <- rebalanceCache
	}
}

func getRebalancers(status *core.Status) []*Rebalancer {
	responseChannel := make(chan []*Rebalancer)
	rebalanceCache := RebalanceCache{
		Status:         status,
		Type:           readRebalancersOperation,
		RebalancersOut: responseChannel,
	}
	RebalancesCacheChannel <- rebalanceCache
	return <-responseChannel
}

func getRebalancer(origin core.RebalanceRequestOrigin, originId int,
	incomingChannelId int,
	outgoingChannelId int) *Rebalancer {

	responseChannel := make(chan RebalanceCache)
	rebalanceCache := RebalanceCache{
		Origin:            origin,
		OriginId:          originId,
		IncomingChannelId: incomingChannelId,
		OutgoingChannelId: outgoingChannelId,
		Type:              readRebalancerOperation,
		Out:               responseChannel,
	}
	RebalancesCacheChannel <- rebalanceCache
	response := <-responseChannel
	return response.Rebalancer
}

func addRebalancer(rebalancer *Rebalancer) bool {
	responseChannel := make(chan bool)
	rebalanceCache := RebalanceCache{
		Rebalancer: rebalancer,
		Type:       writeRebalancerOperation,
		BoolOut:    responseChannel,
	}
	rebalanceCache = copyFromRebalancer(rebalanceCache)
	RebalancesCacheChannel <- rebalanceCache
	return <-responseChannel
}

func removeRebalancer(rebalancer *Rebalancer) {
	rebalanceCache := RebalanceCache{
		Rebalancer: rebalancer,
		Type:       deleteRebalancerOperation,
	}
	rebalanceCache = copyFromRebalancer(rebalanceCache)
	RebalancesCacheChannel <- rebalanceCache
}
