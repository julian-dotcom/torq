package commons

import (
	"context"

	"github.com/rs/zerolog/log"
)

var ManagedNodeAliasChannel = make(chan ManagedNodeAlias) //nolint:gochecknoglobals

type ManagedNodeAliasCacheOperationType uint

const (
	readAlias ManagedNodeAliasCacheOperationType = iota
	writeAlias
)

type ManagedNodeAlias struct {
	Type   ManagedNodeAliasCacheOperationType
	NodeId int
	Alias  string
	Out    chan<- string
}

func ManagedNodeAliasCache(ch <-chan ManagedNodeAlias, ctx context.Context) {
	nodeAliasesByNodeIdCache := make(map[int]string, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case managedNodeAlias := <-ch:
			processManagedNodeAlias(managedNodeAlias, nodeAliasesByNodeIdCache)
		}
	}
}

func processManagedNodeAlias(managedNodeAlias ManagedNodeAlias, nodeAliasesByNodeIdCache map[int]string) {
	switch managedNodeAlias.Type {
	case readAlias:
		if managedNodeAlias.NodeId == 0 {
			log.Error().Msgf("No empty NodeId allowed")
			SendToManagedNodeAliasChannel(managedNodeAlias.Out, "")
			break
		}
		nodeAlias, exists := nodeAliasesByNodeIdCache[managedNodeAlias.NodeId]
		if exists {
			SendToManagedNodeAliasChannel(managedNodeAlias.Out, nodeAlias)
			break
		}
		SendToManagedNodeAliasChannel(managedNodeAlias.Out, "")
	case writeAlias:
		if managedNodeAlias.NodeId == 0 {
			log.Error().Msg("No empty NodeId allowed")
			break
		}
		if managedNodeAlias.Alias == "" {
			log.Debug().Msgf("No empty Alias allowed (nodeId: %v)", managedNodeAlias.NodeId)
			break
		}
		nodeAliasesByNodeIdCache[managedNodeAlias.NodeId] = managedNodeAlias.Alias
	}
}

func GetNodeAlias(nodeId int) string {
	nodeAliasResponseChannel := make(chan string)
	managedNodeAlias := ManagedNodeAlias{
		NodeId: nodeId,
		Type:   readAlias,
		Out:    nodeAliasResponseChannel,
	}
	ManagedNodeAliasChannel <- managedNodeAlias
	return <-nodeAliasResponseChannel
}

func SetNodeAlias(nodeId int, nodeAlias string) {
	managedNodeAlias := ManagedNodeAlias{
		Alias:  nodeAlias,
		NodeId: nodeId,
		Type:   writeAlias,
	}
	ManagedNodeAliasChannel <- managedNodeAlias
}
