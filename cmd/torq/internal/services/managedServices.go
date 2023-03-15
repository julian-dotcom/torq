package services

import (
	"context"
	"sort"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/commons"
)

var ManagedServiceChannel = make(chan ManagedService) //nolint:gochecknoglobals

type ManagedServiceCacheOperationType uint

const (
	popDelayedServiceCommand ManagedServiceCacheOperationType = iota
	addDelayedServiceCommand
	removeDelayedServiceCommand
)

type ManagedService struct {
	Type                  ManagedServiceCacheOperationType
	ServiceChannelMessage commons.ServiceChannelMessage
	DelayedServiceCommand DelayedServiceCommand
	Out                   chan<- DelayedServiceCommand
}

type DelayedServiceCommand struct {
	Name                  string
	ServiceChannelMessage commons.ServiceChannelMessage
	Nodes                 []settings.ConnectionDetails
	StartTime             time.Time
}

func ManagedServiceCache(ch <-chan ManagedService, ctx context.Context) {
	var delayedServiceCommands []DelayedServiceCommand
	for {
		select {
		case <-ctx.Done():
			return
		case managedService := <-ch:
			delayedServiceCommands = processManagedService(managedService, delayedServiceCommands)
		}
	}
}

func processManagedService(
	managedService ManagedService,
	delayedServiceCommands []DelayedServiceCommand) []DelayedServiceCommand {

	switch managedService.Type {
	case addDelayedServiceCommand:
		if managedService.DelayedServiceCommand.Name == "" ||
			len(managedService.DelayedServiceCommand.Nodes) == 0 ||
			len(managedService.DelayedServiceCommand.Nodes) > 1 ||
			managedService.DelayedServiceCommand.Nodes[0].NodeId != managedService.DelayedServiceCommand.ServiceChannelMessage.NodeId ||
			managedService.DelayedServiceCommand.ServiceChannelMessage.NodeId == 0 ||
			managedService.DelayedServiceCommand.ServiceChannelMessage.ServiceCommand != commons.Boot {
			log.Error().Msgf("Invalid request to add a delayed BOOT with name (%v), nodeId (%v) and nodes (%v)",
				managedService.DelayedServiceCommand.Name,
				managedService.DelayedServiceCommand.ServiceChannelMessage.NodeId,
				managedService.DelayedServiceCommand.Nodes)
			return delayedServiceCommands
		}
		return append(delayedServiceCommands, managedService.DelayedServiceCommand)
	case popDelayedServiceCommand:
		if len(delayedServiceCommands) != 0 {
			sort.Slice(delayedServiceCommands, func(i, j int) bool {
				return delayedServiceCommands[i].StartTime.Before(delayedServiceCommands[j].StartTime)
			})
			if time.Now().After(delayedServiceCommands[0].StartTime) {
				managedService.Out <- delayedServiceCommands[0]
				return delayedServiceCommands[1:]
			}
		}
		managedService.Out <- DelayedServiceCommand{}
		return delayedServiceCommands
	case removeDelayedServiceCommand:
		if managedService.ServiceChannelMessage.ServiceCommand != commons.Boot {
			log.Error().Msg("Invalid request to remove a delayed BOOT service command.")
			return delayedServiceCommands
		}
		var newList []DelayedServiceCommand
		for _, delayedServiceCommand := range delayedServiceCommands {
			if delayedServiceCommand.ServiceChannelMessage.ServiceType != managedService.ServiceChannelMessage.ServiceType ||
				delayedServiceCommand.ServiceChannelMessage.NodeId != managedService.ServiceChannelMessage.NodeId {
				newList = append(newList, delayedServiceCommand)
			}
		}
		return newList
	}
	return delayedServiceCommands
}

func PopDelayedServiceCommand() DelayedServiceCommand {
	serviceResponseChannel := make(chan DelayedServiceCommand)
	managedChannel := ManagedService{
		Type: popDelayedServiceCommand,
		Out:  serviceResponseChannel,
	}
	ManagedServiceChannel <- managedChannel
	return <-serviceResponseChannel
}

func SetDelayedServiceCommand(delayedServiceCommand DelayedServiceCommand) {
	managedChannel := ManagedService{
		DelayedServiceCommand: delayedServiceCommand,
		Type:                  addDelayedServiceCommand,
	}
	ManagedServiceChannel <- managedChannel
}

func RemoveDelayedServiceCommand(serviceCmd commons.ServiceChannelMessage) {
	managedChannel := ManagedService{
		ServiceChannelMessage: serviceCmd,
		Type:                  removeDelayedServiceCommand,
	}
	ManagedServiceChannel <- managedChannel
}
