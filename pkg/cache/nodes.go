package cache

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/core"
)

var NodesCacheChannel = make(chan NodeCache) //nolint:gochecknoglobals

type NodeCacheOperationType uint
type publicKey string

const (
	readAllTorqNode NodeCacheOperationType = iota
	writeInactiveTorqNode
	readAllTorqNodeIds
	readAllTorqNodeIdsAllNetworks
	readAllTorqPublicKeys
	readActiveTorqNode
	writeActiveTorqNode
	readActiveTorqNodeIds
	readActiveTorqPublicKeys
	readAllActiveTorqNodeIds
	readAllActiveTorqNodeSettings
	readActiveChannelNode
	readChannelNode
	writeActiveChannelNode
	writeInactiveChannelNode
	readAllChannelNodeIds
	readAllChannelPublicKeys
	readActiveChannelPublicKeys
	readNodeSetting
	removeNodeFromCached
)

type NodeCache struct {
	Type            NodeCacheOperationType
	NodeId          int
	Chain           *core.Chain
	Network         *core.Network
	PublicKey       string
	Name            *string
	Out             chan<- NodeCache
	NodeIdsOut      chan<- []int
	NodeSettingOut  chan<- NodeSettingsCache
	NodeSettingsOut chan<- []NodeSettingsCache
	PublicKeysOut   chan<- []string
}

type NodeSettingsCache struct {
	NodeId    int
	Chain     core.Chain
	Network   core.Network
	PublicKey string
	Name      *string
	Status    core.Status
}

// NodesCacheHandler parameter Context is for test cases...
func NodesCacheHandler(ch <-chan NodeCache, ctx context.Context) {
	allTorqNodeIdCache := make(map[core.Chain]map[core.Network]map[publicKey]nodeId)
	activeTorqNodeIdCache := make(map[core.Chain]map[core.Network]map[publicKey]nodeId)
	channelNodeIdCache := make(map[core.Chain]map[core.Network]map[publicKey]nodeId)
	allChannelNodeIdCache := make(map[core.Chain]map[core.Network]map[publicKey]nodeId)
	nodeSettingsByNodeIdCache := make(map[nodeId]NodeSettingsCache)
	torqNodeNameByNodeIdCache := make(map[nodeId]string)
	for {
		select {
		case <-ctx.Done():
			return
		case nodeCache := <-ch:
			handleNodeOperation(nodeCache, allTorqNodeIdCache, activeTorqNodeIdCache,
				channelNodeIdCache, allChannelNodeIdCache, nodeSettingsByNodeIdCache,
				torqNodeNameByNodeIdCache)
		}
	}
}

func handleNodeOperation(nodeCache NodeCache,
	allTorqNodeIdCache map[core.Chain]map[core.Network]map[publicKey]nodeId,
	activeTorqNodeIdCache map[core.Chain]map[core.Network]map[publicKey]nodeId,
	channelNodeIdCache map[core.Chain]map[core.Network]map[publicKey]nodeId,
	allChannelNodeIdCache map[core.Chain]map[core.Network]map[publicKey]nodeId,
	nodeSettingsByNodeIdCache map[nodeId]NodeSettingsCache,
	torqNodeNameByNodeIdCache map[nodeId]string) {
	switch nodeCache.Type {
	case readAllTorqNode:
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(allTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			nodeCache.NodeId = int(allTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)])
			nodeName, exists := torqNodeNameByNodeIdCache[nodeId(nodeCache.NodeId)]
			if exists {
				nodeCache.Name = &nodeName
			}
		}
		nodeCache.Out <- nodeCache
	case readActiveTorqNode:
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			nodeCache.NodeId = int(activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)])
			nodeName, exists := torqNodeNameByNodeIdCache[nodeId(nodeCache.NodeId)]
			if exists {
				nodeCache.Name = &nodeName
			}
		}
		nodeCache.Out <- nodeCache
	case readActiveChannelNode:
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(channelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			nodeCache.NodeId = int(channelNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)])
		}
		nodeCache.Out <- nodeCache
	case readChannelNode:
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(allChannelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			nodeCache.NodeId = int(allChannelNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)])
		}
		nodeCache.Out <- nodeCache
	case readAllTorqNodeIds:
		var allNodeIds []int
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(allTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for _, value := range allTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				allNodeIds = append(allNodeIds, int(value))
			}
		}
		nodeCache.NodeIdsOut <- allNodeIds
	case readAllTorqNodeIdsAllNetworks:
		var allNodeIds []int
		for chainIndex := range allTorqNodeIdCache {
			for networkIndex := range allTorqNodeIdCache[chainIndex] {
				for nodeIndex := range allTorqNodeIdCache[chainIndex][networkIndex] {
					allNodeIds = append(allNodeIds, int(allTorqNodeIdCache[chainIndex][networkIndex][nodeIndex]))
				}
			}
		}
		nodeCache.NodeIdsOut <- allNodeIds
	case readActiveTorqNodeIds:
		var allNodeIds []int
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for _, value := range activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				allNodeIds = append(allNodeIds, int(value))
			}
		}
		nodeCache.NodeIdsOut <- allNodeIds
	case readAllChannelNodeIds:
		var allNodeIds []int
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(allChannelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for _, value := range allChannelNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				allNodeIds = append(allNodeIds, int(value))
			}
		}
		nodeCache.NodeIdsOut <- allNodeIds
	case readAllTorqPublicKeys:
		var allPublicKeys []string
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(allTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for key := range allTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				allPublicKeys = append(allPublicKeys, string(key))
			}
		}
		nodeCache.PublicKeysOut <- allPublicKeys
	case readActiveTorqPublicKeys:
		var activePublicKeys []string
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for key := range activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				activePublicKeys = append(activePublicKeys, string(key))
			}
		}
		nodeCache.PublicKeysOut <- activePublicKeys
	case readAllActiveTorqNodeIds:
		var nodeIds []int
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			for chain, networkMap := range activeTorqNodeIdCache {
				if nodeCache.Chain == nil || *nodeCache.Chain == chain {
					for network, publicKeyMap := range networkMap {
						if nodeCache.Network == nil || *nodeCache.Network == network {
							for _, nId := range publicKeyMap {
								nodeIds = append(nodeIds, int(nId))
							}
						}
					}
				}
			}
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for _, nId := range activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				nodeIds = append(nodeIds, int(nId))
			}
		}
		nodeCache.NodeIdsOut <- nodeIds
	case readAllActiveTorqNodeSettings:
		var nodes []NodeSettingsCache
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			for chain, networkMap := range activeTorqNodeIdCache {
				if nodeCache.Chain == nil || *nodeCache.Chain == chain {
					for network, publicKeyMap := range networkMap {
						if nodeCache.Network == nil || *nodeCache.Network == network {
							for _, nId := range publicKeyMap {
								nodes = append(nodes, nodeSettingsByNodeIdCache[nodeId(nId)])
							}
						}
					}
				}
			}
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for _, nId := range activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				nodes = append(nodes, nodeSettingsByNodeIdCache[nodeId(nId)])
			}
		}
		nodeCache.NodeSettingsOut <- nodes
	case readAllChannelPublicKeys:
		var channelPublicKeys []string
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(allChannelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for key := range allChannelNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				channelPublicKeys = append(channelPublicKeys, string(key))
			}
		}
		nodeCache.PublicKeysOut <- channelPublicKeys
	case readActiveChannelPublicKeys:
		var channelPublicKeys []string
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(channelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for key := range channelNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				channelPublicKeys = append(channelPublicKeys, string(key))
			}
		}
		nodeCache.PublicKeysOut <- channelPublicKeys
	case readNodeSetting:
		if nodeCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", nodeCache.NodeId)
			nodeCache.NodeSettingOut <- NodeSettingsCache{}
		} else {
			nodeName, exists := torqNodeNameByNodeIdCache[nodeId(nodeCache.NodeId)]
			nodeSettings := nodeSettingsByNodeIdCache[nodeId(nodeCache.NodeId)]
			if exists {
				nodeSettings.Name = &nodeName
			}
			nodeCache.NodeSettingOut <- nodeSettings
		}
	case writeInactiveTorqNode:
		if nodeCache.Name == nil || *nodeCache.Name == "" || nodeCache.PublicKey == "" || nodeCache.NodeId == 0 ||
			nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty name (%v), publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				nodeCache.Name, nodeCache.PublicKey, nodeCache.NodeId, nodeCache.Chain, nodeCache.Network)
		} else {
			torqNodeNameByNodeIdCache[nodeId(nodeCache.NodeId)] = *nodeCache.Name
			initializeNodeIdCache(activeTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			delete(activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
			initializeNodeIdCache(allTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			allTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			initializeNodeIdCache(channelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			delete(channelNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
			initializeNodeIdCache(allChannelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			allChannelNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			nodeSettingsByNodeIdCache[nodeId(nodeCache.NodeId)] = NodeSettingsCache{
				NodeId:    nodeCache.NodeId,
				Network:   *nodeCache.Network,
				Chain:     *nodeCache.Chain,
				PublicKey: nodeCache.PublicKey,
				Name:      nodeCache.Name,
			}
		}
	case writeActiveTorqNode:
		if nodeCache.Name == nil || *nodeCache.Name == "" || nodeCache.PublicKey == "" || nodeCache.NodeId == 0 ||
			nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty name (%v), publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				nodeCache.Name, nodeCache.PublicKey, nodeCache.NodeId, nodeCache.Chain, nodeCache.Network)
		} else {
			torqNodeNameByNodeIdCache[nodeId(nodeCache.NodeId)] = *nodeCache.Name
			initializeNodeIdCache(activeTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			initializeNodeIdCache(allTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			allTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			initializeNodeIdCache(channelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			channelNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			initializeNodeIdCache(allChannelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			allChannelNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			nodeSettingsByNodeIdCache[nodeId(nodeCache.NodeId)] = NodeSettingsCache{
				NodeId:    nodeCache.NodeId,
				Network:   *nodeCache.Network,
				Chain:     *nodeCache.Chain,
				PublicKey: nodeCache.PublicKey,
				Name:      nodeCache.Name,
			}
		}
	case writeActiveChannelNode:
		if nodeCache.PublicKey == "" || nodeCache.NodeId == 0 || nodeCache.Chain == nil ||
			nodeCache.Network == nil {
			log.Error().Msgf("No empty publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				nodeCache.PublicKey, nodeCache.NodeId, nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(channelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			channelNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			initializeNodeIdCache(allChannelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			allChannelNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			nodeSettingsByNodeIdCache[nodeId(nodeCache.NodeId)] = NodeSettingsCache{
				NodeId:    nodeCache.NodeId,
				Network:   *nodeCache.Network,
				Chain:     *nodeCache.Chain,
				PublicKey: nodeCache.PublicKey,
				Status:    core.Active,
			}
		}
	case writeInactiveChannelNode:
		if nodeCache.PublicKey == "" || nodeCache.NodeId == 0 || nodeCache.Chain == nil ||
			nodeCache.Network == nil {
			log.Error().Msgf("No empty publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				nodeCache.PublicKey, nodeCache.NodeId, nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(channelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			delete(channelNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
			initializeNodeIdCache(allChannelNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			allChannelNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			nodeSettingsByNodeIdCache[nodeId(nodeCache.NodeId)] = NodeSettingsCache{
				NodeId:    nodeCache.NodeId,
				Network:   *nodeCache.Network,
				Chain:     *nodeCache.Chain,
				PublicKey: nodeCache.PublicKey,
				Status:    core.Inactive,
			}
		}
	case removeNodeFromCached:
		delete(channelNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
		delete(allTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
		delete(nodeSettingsByNodeIdCache, nodeId(nodeCache.NodeId))
		delete(activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
		delete(allChannelNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
		delete(torqNodeNameByNodeIdCache, nodeId(nodeCache.NodeId))
	}
}

func initializeNodeIdCache(nodeIdCache map[core.Chain]map[core.Network]map[publicKey]nodeId,
	chain core.Chain,
	network core.Network) {

	if nodeIdCache[chain] == nil {
		nodeIdCache[chain] = make(map[core.Network]map[publicKey]nodeId, 0)
	}
	if nodeIdCache[chain][network] == nil {
		nodeIdCache[chain][network] = make(map[publicKey]nodeId, 0)
	}
}

func GetActiveChannelPublicKeys(chain core.Chain, network core.Network) []string {
	publicKeysResponseChannel := make(chan []string)
	nodeCache := NodeCache{
		Chain:         &chain,
		Network:       &network,
		Type:          readActiveChannelPublicKeys,
		PublicKeysOut: publicKeysResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-publicKeysResponseChannel
}

func GetAllChannelPublicKeys(chain core.Chain, network core.Network) []string {
	publicKeysResponseChannel := make(chan []string)
	nodeCache := NodeCache{
		Chain:         &chain,
		Network:       &network,
		Type:          readAllChannelPublicKeys,
		PublicKeysOut: publicKeysResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-publicKeysResponseChannel
}

func GetAllTorqPublicKeys(chain core.Chain, network core.Network) []string {
	publicKeysResponseChannel := make(chan []string)
	nodeCache := NodeCache{
		Chain:         &chain,
		Network:       &network,
		Type:          readAllTorqPublicKeys,
		PublicKeysOut: publicKeysResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-publicKeysResponseChannel
}

func GetAllTorqNodeIdsByNetwork(chain core.Chain, network core.Network) []int {
	nodeIdsResponseChannel := make(chan []int)
	nodeCache := NodeCache{
		Chain:      &chain,
		Network:    &network,
		Type:       readAllTorqNodeIds,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-nodeIdsResponseChannel
}

func GetAllTorqNodeIds() []int {
	nodeIdsResponseChannel := make(chan []int)
	nodeCache := NodeCache{
		Type:       readAllTorqNodeIdsAllNetworks,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-nodeIdsResponseChannel
}

func GetChannelNodeIds(chain core.Chain, network core.Network) []int {
	nodeIdsResponseChannel := make(chan []int)
	nodeCache := NodeCache{
		Chain:      &chain,
		Network:    &network,
		Type:       readAllChannelNodeIds,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-nodeIdsResponseChannel
}

func GetAllActiveTorqNodeIds(chain *core.Chain, network *core.Network) []int {
	nodeIdsResponseChannel := make(chan []int)
	nodeCache := NodeCache{
		Chain:      chain,
		Network:    network,
		Type:       readAllActiveTorqNodeIds,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-nodeIdsResponseChannel
}

func GetActiveTorqNodeSettings() []NodeSettingsCache {
	nodesResponseChannel := make(chan []NodeSettingsCache)
	nodeCache := NodeCache{
		Type:            readAllActiveTorqNodeSettings,
		NodeSettingsOut: nodesResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-nodesResponseChannel
}

// SetTorqNode When active then also adds to channelNodes
func SetTorqNode(nodeId int, name string, status core.Status, publicKey string, chain core.Chain, network core.Network) {
	if status == core.Active {
		nodeCache := NodeCache{
			PublicKey: publicKey,
			Chain:     &chain,
			Network:   &network,
			NodeId:    nodeId,
			Name:      &name,
			Type:      writeActiveTorqNode,
		}
		NodesCacheChannel <- nodeCache
	} else {
		nodeCache := NodeCache{
			PublicKey: publicKey,
			Chain:     &chain,
			Network:   &network,
			NodeId:    nodeId,
			Name:      &name,
			Type:      writeInactiveTorqNode,
		}
		NodesCacheChannel <- nodeCache
	}
}

func GetNodeIdByPublicKey(publicKey string, chain core.Chain, network core.Network) int {
	nodeResponseChannel := make(chan NodeCache)
	nodeCache := NodeCache{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      readChannelNode,
		Out:       nodeResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	nodeResponse := <-nodeResponseChannel
	return nodeResponse.NodeId
}

func GetActiveNodeIdByPublicKey(publicKey string, chain core.Chain, network core.Network) int {
	nodeResponseChannel := make(chan NodeCache)
	nodeCache := NodeCache{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      readActiveChannelNode,
		Out:       nodeResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	nodeResponse := <-nodeResponseChannel
	return nodeResponse.NodeId
}

func SetChannelNode(nodeId int, publicKey string, chain core.Chain, network core.Network, status core.ChannelStatus) {
	if status < core.CooperativeClosed {
		NodesCacheChannel <- NodeCache{
			NodeId:    nodeId,
			PublicKey: publicKey,
			Chain:     &chain,
			Network:   &network,
			Type:      writeActiveChannelNode,
		}
	} else {
		SetInactiveChannelNode(nodeId, publicKey, chain, network)
	}
}

func SetInactiveChannelNode(nodeId int, publicKey string, chain core.Chain, network core.Network) {
	NodesCacheChannel <- NodeCache{
		NodeId:    nodeId,
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      writeInactiveChannelNode,
	}
}

func GetNodeSettingsByNodeId(nodeId int) NodeSettingsCache {
	nodeResponseChannel := make(chan NodeSettingsCache)
	nodeCache := NodeCache{
		NodeId:         nodeId,
		Type:           readNodeSetting,
		NodeSettingOut: nodeResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-nodeResponseChannel
}

func RemoveNodeFromCache(node NodeSettingsCache) {
	NodesCacheChannel <- NodeCache{
		Type:      removeNodeFromCached,
		NodeId:    node.NodeId,
		Network:   &node.Network,
		Chain:     &node.Chain,
		PublicKey: node.PublicKey,
	}
}
