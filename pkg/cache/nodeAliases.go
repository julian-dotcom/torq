package cache

import (
	"context"

	"github.com/rs/zerolog/log"
)

var NodeAliasesCacheChannel = make(chan NodeAliasCache) //nolint:gochecknoglobals

type NodeAliasCacheOperationType uint

const (
	readAlias NodeAliasCacheOperationType = iota
	writeAlias
)

type NodeAliasCache struct {
	Type   NodeAliasCacheOperationType
	NodeId int
	Alias  string
	Out    chan<- string
}

func NodeAliasesCacheHandler(ch <-chan NodeAliasCache, ctx context.Context) {
	nodeAliasesByNodeIdCache := make(map[nodeId]string, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case nodeAliasCache := <-ch:
			handleNodeAliasOperation(nodeAliasCache, nodeAliasesByNodeIdCache)
		}
	}
}

func handleNodeAliasOperation(nodeAliasCache NodeAliasCache, nodeAliasesByNodeIdCache map[nodeId]string) {
	switch nodeAliasCache.Type {
	case readAlias:
		if nodeAliasCache.NodeId == 0 {
			log.Error().Msgf("No empty NodeId allowed")
			nodeAliasCache.Out <- ""
			break
		}
		nodeAlias, exists := nodeAliasesByNodeIdCache[nodeId(nodeAliasCache.NodeId)]
		if exists {
			nodeAliasCache.Out <- nodeAlias
			break
		}
		nodeAliasCache.Out <- ""
	case writeAlias:
		if nodeAliasCache.NodeId == 0 {
			log.Error().Msg("No empty NodeId allowed")
			break
		}
		if nodeAliasCache.Alias == "" {
			log.Debug().Msgf("No empty Alias allowed (nodeId: %v)", nodeAliasCache.NodeId)
			break
		}
		nodeAliasesByNodeIdCache[nodeId(nodeAliasCache.NodeId)] = nodeAliasCache.Alias
	}
}

func GetNodeAlias(nodeId int) string {
	nodeAliasResponseChannel := make(chan string)
	nodeAliasCache := NodeAliasCache{
		NodeId: nodeId,
		Type:   readAlias,
		Out:    nodeAliasResponseChannel,
	}
	NodeAliasesCacheChannel <- nodeAliasCache
	return <-nodeAliasResponseChannel
}

func SetNodeAlias(nodeId int, nodeAlias string) {
	nodeAliasCache := NodeAliasCache{
		Alias:  nodeAlias,
		NodeId: nodeId,
		Type:   writeAlias,
	}
	NodeAliasesCacheChannel <- nodeAliasCache
}
