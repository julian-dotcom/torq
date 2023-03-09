package workflows

import (
	"context"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/pkg/commons"
)

var ManagedRebalanceChannel = make(chan ManagedRebalance) //nolint:gochecknoglobals

type ManagedRebalanceCacheOperationType uint

const (
	readRebalancer ManagedRebalanceCacheOperationType = iota
	readRebalancers
	writeRebalancer
	deleteRebalancer
	cancelRebalancer
	cancelRebalancers
)

type ManagedRebalance struct {
	Type              ManagedRebalanceCacheOperationType
	Origin            commons.RebalanceRequestOrigin
	OriginId          int
	OriginReference   string
	IncomingChannelId int
	OutgoingChannelId int
	AmountMsat        uint64
	ChannelId         int
	ChannelIds        []int
	Status            *commons.Status
	Rebalancer        *Rebalancer
	Out               chan<- ManagedRebalance
	BoolOut           chan<- bool
	RebalancersOut    chan<- []*Rebalancer
}

func ManagedRebalanceCache(ch <-chan ManagedRebalance, ctx context.Context) {
	rebalancers := make(map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer)

	for {
		select {
		case <-ctx.Done():
			return
		case managedRebalance := <-ch:
			switch managedRebalance.Type {
			case readRebalancer:
				if !isValidRequest(managedRebalance) {
					SendToManagedRebalanceChannel(managedRebalance.Out, managedRebalance)
					continue
				}
				initializeRebalancersCache(managedRebalance, rebalancers)
				managedRebalance.Rebalancer = getRebalancerCache(managedRebalance, rebalancers)
				SendToManagedRebalanceChannel(managedRebalance.Out, managedRebalance)
			case readRebalancers:
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
			case writeRebalancer:
				if !isValidRequest(managedRebalance) {
					commons.SendToManagedBoolChannel(managedRebalance.BoolOut, false)
					continue
				}
				initializeRebalancersCache(managedRebalance, rebalancers)
				if getRebalancerCache(managedRebalance, rebalancers) != nil {
					commons.SendToManagedBoolChannel(managedRebalance.BoolOut, false)
					continue
				}
				setRebalancersCache(managedRebalance, rebalancers)
				commons.SendToManagedBoolChannel(managedRebalance.BoolOut, true)
			case deleteRebalancer:
				if !isValidRequest(managedRebalance) {
					continue
				}
				initializeRebalancersCache(managedRebalance, rebalancers)
				if getRebalancerCache(managedRebalance, rebalancers) == nil {
					continue
				}
				removeRebalancersCache(managedRebalance, rebalancers)
			case cancelRebalancer:
				if managedRebalance.OriginId == 0 || managedRebalance.ChannelId == 0 {
					continue
				}
				initializeRebalancersCache(managedRebalance, rebalancers)
				managedRebalance.IncomingChannelId = managedRebalance.ChannelId
				rebalancer := getRebalancerCache(managedRebalance, rebalancers)
				if rebalancer != nil {
					log.Debug().Msgf("Cancelling rebalancer for channelId: %v, origin: %v, originId: %v",
						managedRebalance.ChannelId, managedRebalance.Origin, managedRebalance.OriginId)
					rebalancer.RebalanceCancel()
					delete(rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)], channelIdInt(managedRebalance.ChannelId))
				}
			case cancelRebalancers:
				if managedRebalance.OriginId == 0 {
					continue
				}
				initializeRebalancersCache(managedRebalance, rebalancers)
				for channelId, rebalancer := range getRebalancersCache(managedRebalance, rebalancers) {
					if slices.Contains(managedRebalance.ChannelIds, int(channelId)) {
						continue
					}
					log.Debug().Msgf("Cancelling rebalancer for channelId: %v, origin: %v, originId: %v",
						channelId, managedRebalance.Origin, managedRebalance.OriginId)
					rebalancer.RebalanceCancel()
					delete(rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)], channelId)
				}
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
	rebalancers map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) map[channelIdInt]*Rebalancer {

	_, exists := rebalancers[managedRebalance.Origin]
	if !exists {
		return nil
	}
	_, exists = rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)]
	if !exists {
		return nil
	}
	results := make(map[channelIdInt]*Rebalancer)
	for channelId, rebalancer := range rebalancers[managedRebalance.Origin][originIdInt(managedRebalance.OriginId)] {
		results[channelId] = rebalancer
	}
	return results

}

func getRebalancerCache(managedRebalance ManagedRebalance,
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

func initializeRebalancersCache(managedRebalance ManagedRebalance,
	rebalancers map[commons.RebalanceRequestOrigin]map[originIdInt]map[channelIdInt]*Rebalancer) {

	if rebalancers[managedRebalance.Origin] == nil {
		rebalancers[managedRebalance.Origin] = make(map[originIdInt]map[channelIdInt]*Rebalancer)
	}
}

func isValidRequest(managedRebalance ManagedRebalance) bool {
	if managedRebalance.Type != readRebalancers && managedRebalance.IncomingChannelId == 0 && managedRebalance.OutgoingChannelId == 0 {
		log.Error().Msgf("IncomingChannelId (%v) and OutgoingChannelId (%v) cannot both be 0",
			managedRebalance.IncomingChannelId, managedRebalance.OutgoingChannelId)
		return false
	}
	return true
}

func SendToManagedRebalanceChannel(ch chan<- ManagedRebalance, managedRebalance ManagedRebalance) {
	ch <- managedRebalance
	close(ch)
}

func SendToRebalancersChannel(ch chan<- []*Rebalancer, rebalancers []*Rebalancer) {
	ch <- rebalancers
	close(ch)
}

func CancelRebalancersExcept(origin commons.RebalanceRequestOrigin, originId int, activeChannelIds []int) {
	managedRebalance := ManagedRebalance{
		Origin:     origin,
		OriginId:   originId,
		ChannelIds: activeChannelIds,
		Type:       cancelRebalancers,
	}
	ManagedRebalanceChannel <- managedRebalance
}

func CancelRebalancer(origin commons.RebalanceRequestOrigin, originId int, channelId int) {
	managedRebalance := ManagedRebalance{
		Origin:    origin,
		OriginId:  originId,
		ChannelId: channelId,
		Type:      cancelRebalancer,
	}
	ManagedRebalanceChannel <- managedRebalance
}

func getRebalancers(status *commons.Status) []*Rebalancer {
	responseChannel := make(chan []*Rebalancer)
	managedRebalance := ManagedRebalance{
		Status:         status,
		Type:           readRebalancers,
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
		Type:              readRebalancer,
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
		Type:       writeRebalancer,
		BoolOut:    responseChannel,
	}
	managedRebalance = copyFromRebalancer(managedRebalance)
	ManagedRebalanceChannel <- managedRebalance
	return <-responseChannel
}

func removeRebalancer(rebalancer *Rebalancer) {
	managedRebalance := ManagedRebalance{
		Rebalancer: rebalancer,
		Type:       deleteRebalancer,
	}
	managedRebalance = copyFromRebalancer(managedRebalance)
	ManagedRebalanceChannel <- managedRebalance
}
