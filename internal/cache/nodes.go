package cache

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/core"
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
	readActiveChannelPeerNode
	readChannelPeerNode
	writeActiveChannelPeerNode
	writeInactiveChannelPeerNode
	readAllChannelPeerNodeIds
	readAllChannelPeerPublicKeys
	readActiveChannelPeerPublicKeys
	readNodeSetting
	removeNodeFromCached
	readConnectedPeerNode
	readAllConnectedPeerNodeIds
	writeConnectedPeerNode
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
	// only populated when it's a node managed by Torq
	Name *string
	// only populated when there is a channel
	ChannelStatus *core.Status
}

// NodesCacheHandler parameter Context is for test cases...
func NodesCacheHandler(ch <-chan NodeCache, ctx context.Context) {
	allTorqNodeIdCache := make(map[core.Chain]map[core.Network]map[publicKey]nodeId)
	activeTorqNodeIdCache := make(map[core.Chain]map[core.Network]map[publicKey]nodeId)
	channelPeerNodeIdCache := make(map[core.Chain]map[core.Network]map[publicKey]nodeId)
	allChannelPeerNodeIdCache := make(map[core.Chain]map[core.Network]map[publicKey]nodeId)
	connectedPeerNodeIdCache := make(map[core.Chain]map[core.Network]map[publicKey]nodeId)
	nodeSettingsByNodeIdCache := make(map[nodeId]NodeSettingsCache)
	torqNodeNameByNodeIdCache := make(map[nodeId]string)
	for {
		select {
		case <-ctx.Done():
			return
		case nodeCache := <-ch:
			handleNodeOperation(nodeCache, allTorqNodeIdCache, activeTorqNodeIdCache,
				channelPeerNodeIdCache, allChannelPeerNodeIdCache, connectedPeerNodeIdCache, nodeSettingsByNodeIdCache,
				torqNodeNameByNodeIdCache)
		}
	}
}

func handleNodeOperation(nodeCache NodeCache,
	allTorqNodeIdCache map[core.Chain]map[core.Network]map[publicKey]nodeId,
	activeTorqNodeIdCache map[core.Chain]map[core.Network]map[publicKey]nodeId,
	channelPeerNodeIdCache map[core.Chain]map[core.Network]map[publicKey]nodeId,
	allChannelPeerNodeIdCache map[core.Chain]map[core.Network]map[publicKey]nodeId,
	connectedPeerNodeIdCache map[core.Chain]map[core.Network]map[publicKey]nodeId,
	nodeSettingsByNodeIdCache map[nodeId]NodeSettingsCache,
	torqNodeNameByNodeIdCache map[nodeId]string) {
	switch nodeCache.Type {
	case readAllTorqNode:
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(allTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			nodeCache.NodeId = int(allTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)])
			setNameInNodeCache(nodeCache, torqNodeNameByNodeIdCache)
		}
		nodeCache.Out <- nodeCache
	case readActiveTorqNode:
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			nodeCache.NodeId = int(activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)])
			setNameInNodeCache(nodeCache, torqNodeNameByNodeIdCache)
		}
		nodeCache.Out <- nodeCache
	case readActiveChannelPeerNode:
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(channelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			nodeCache.NodeId = int(channelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)])
			setNameInNodeCache(nodeCache, torqNodeNameByNodeIdCache)
		}
		nodeCache.Out <- nodeCache
	case readChannelPeerNode:
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(allChannelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			nodeCache.NodeId = int(allChannelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)])
			setNameInNodeCache(nodeCache, torqNodeNameByNodeIdCache)
		}
		nodeCache.Out <- nodeCache
	case readConnectedPeerNode:
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(connectedPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			nodeCache.NodeId = int(connectedPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)])
			setNameInNodeCache(nodeCache, torqNodeNameByNodeIdCache)
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
	case readAllChannelPeerNodeIds:
		var allNodeIds []int
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(allChannelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for _, value := range allChannelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				allNodeIds = append(allNodeIds, int(value))
			}
		}
		nodeCache.NodeIdsOut <- allNodeIds
	case readAllConnectedPeerNodeIds:
		var allNodeIds []int
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(connectedPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for _, value := range connectedPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
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
								nodes = append(nodes,
									setNameInNodeSettings(nodeSettingsByNodeIdCache, nId, torqNodeNameByNodeIdCache))
							}
						}
					}
				}
			}
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for _, nId := range activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				nodes = append(nodes, setNameInNodeSettings(nodeSettingsByNodeIdCache, nId, torqNodeNameByNodeIdCache))
			}
		}
		nodeCache.NodeSettingsOut <- nodes
	case readAllChannelPeerPublicKeys:
		var channelPublicKeys []string
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(allChannelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for key := range allChannelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				channelPublicKeys = append(channelPublicKeys, string(key))
			}
		}
		nodeCache.PublicKeysOut <- channelPublicKeys
	case readActiveChannelPeerPublicKeys:
		var channelPublicKeys []string
		if nodeCache.Chain == nil || nodeCache.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(channelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			for key := range channelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network] {
				channelPublicKeys = append(channelPublicKeys, string(key))
			}
		}
		nodeCache.PublicKeysOut <- channelPublicKeys
	case readNodeSetting:
		if nodeCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", nodeCache.NodeId)
			nodeCache.NodeSettingOut <- NodeSettingsCache{}
		} else {
			nodeCache.NodeSettingOut <-
				setNameInNodeSettings(nodeSettingsByNodeIdCache, nodeId(nodeCache.NodeId), torqNodeNameByNodeIdCache)
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
			initializeNodeIdCache(channelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			delete(channelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
			initializeNodeIdCache(allChannelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			allChannelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			nodeSettingsByNodeIdCache[nodeId(nodeCache.NodeId)] = NodeSettingsCache{
				NodeId:    nodeCache.NodeId,
				Network:   *nodeCache.Network,
				Chain:     *nodeCache.Chain,
				PublicKey: nodeCache.PublicKey,
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
			initializeNodeIdCache(channelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			channelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			initializeNodeIdCache(allChannelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			allChannelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			nodeSettingsByNodeIdCache[nodeId(nodeCache.NodeId)] = NodeSettingsCache{
				NodeId:    nodeCache.NodeId,
				Network:   *nodeCache.Network,
				Chain:     *nodeCache.Chain,
				PublicKey: nodeCache.PublicKey,
			}
		}
	case writeActiveChannelPeerNode:
		if nodeCache.PublicKey == "" || nodeCache.NodeId == 0 || nodeCache.Chain == nil ||
			nodeCache.Network == nil {
			log.Error().Msgf("No empty publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				nodeCache.PublicKey, nodeCache.NodeId, nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(channelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			channelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			initializeNodeIdCache(allChannelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			allChannelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			active := core.Active
			nodeSettingsByNodeIdCache[nodeId(nodeCache.NodeId)] = NodeSettingsCache{
				NodeId:        nodeCache.NodeId,
				Network:       *nodeCache.Network,
				Chain:         *nodeCache.Chain,
				PublicKey:     nodeCache.PublicKey,
				ChannelStatus: &active,
			}
		}
	case writeInactiveChannelPeerNode:
		if nodeCache.PublicKey == "" || nodeCache.NodeId == 0 || nodeCache.Chain == nil ||
			nodeCache.Network == nil {
			log.Error().Msgf("No empty publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				nodeCache.PublicKey, nodeCache.NodeId, nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(channelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			delete(channelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
			initializeNodeIdCache(allChannelPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			allChannelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			inactive := core.Inactive
			nodeSettingsByNodeIdCache[nodeId(nodeCache.NodeId)] = NodeSettingsCache{
				NodeId:        nodeCache.NodeId,
				Network:       *nodeCache.Network,
				Chain:         *nodeCache.Chain,
				PublicKey:     nodeCache.PublicKey,
				ChannelStatus: &inactive,
			}
		}
	case writeConnectedPeerNode:
		if nodeCache.PublicKey == "" || nodeCache.NodeId == 0 || nodeCache.Chain == nil ||
			nodeCache.Network == nil {
			log.Error().Msgf("No empty publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				nodeCache.PublicKey, nodeCache.NodeId, nodeCache.Chain, nodeCache.Network)
		} else {
			initializeNodeIdCache(connectedPeerNodeIdCache, *nodeCache.Chain, *nodeCache.Network)
			connectedPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network][publicKey(nodeCache.PublicKey)] = nodeId(nodeCache.NodeId)
			nodeSettingsByNodeIdCache[nodeId(nodeCache.NodeId)] = NodeSettingsCache{
				NodeId:    nodeCache.NodeId,
				Network:   *nodeCache.Network,
				Chain:     *nodeCache.Chain,
				PublicKey: nodeCache.PublicKey,
			}
		}
	case removeNodeFromCached:
		delete(channelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
		delete(allTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
		delete(nodeSettingsByNodeIdCache, nodeId(nodeCache.NodeId))
		delete(activeTorqNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
		delete(allChannelPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
		delete(connectedPeerNodeIdCache[*nodeCache.Chain][*nodeCache.Network], publicKey(nodeCache.PublicKey))
		delete(torqNodeNameByNodeIdCache, nodeId(nodeCache.NodeId))
	}
}

func setNameInNodeCache(nodeCache NodeCache, torqNodeNameByNodeIdCache map[nodeId]string) {
	nodeName, exists := torqNodeNameByNodeIdCache[nodeId(nodeCache.NodeId)]
	if exists {
		nodeCache.Name = &nodeName
	}
}

func setNameInNodeSettings(nodeSettingsByNodeIdCache map[nodeId]NodeSettingsCache,
	nId nodeId,
	torqNodeNameByNodeIdCache map[nodeId]string) NodeSettingsCache {

	nodeSettings := nodeSettingsByNodeIdCache[nId]
	nodeName, exists := torqNodeNameByNodeIdCache[nId]
	if exists {
		nodeSettings.Name = &nodeName
	}
	return nodeSettings
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

func GetActiveChannelPeerPublicKeys(chain core.Chain, network core.Network) []string {
	publicKeysResponseChannel := make(chan []string)
	nodeCache := NodeCache{
		Chain:         &chain,
		Network:       &network,
		Type:          readActiveChannelPeerPublicKeys,
		PublicKeysOut: publicKeysResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-publicKeysResponseChannel
}

func GetAllChannelPeerPublicKeys(chain core.Chain, network core.Network) []string {
	publicKeysResponseChannel := make(chan []string)
	nodeCache := NodeCache{
		Chain:         &chain,
		Network:       &network,
		Type:          readAllChannelPeerPublicKeys,
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

func GetChannelPeerNodeIds(chain core.Chain, network core.Network) []int {
	nodeIdsResponseChannel := make(chan []int)
	nodeCache := NodeCache{
		Chain:      &chain,
		Network:    &network,
		Type:       readAllChannelPeerNodeIds,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	return <-nodeIdsResponseChannel
}

func GetConnectedPeerNodeIds(chain core.Chain, network core.Network) []int {
	nodeIdsResponseChannel := make(chan []int)
	nodeCache := NodeCache{
		Chain:      &chain,
		Network:    &network,
		Type:       readAllConnectedPeerNodeIds,
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

func GetChannelPeerNodeIdByPublicKey(publicKey string, chain core.Chain, network core.Network) int {
	nodeResponseChannel := make(chan NodeCache)
	nodeCache := NodeCache{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      readChannelPeerNode,
		Out:       nodeResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	nodeResponse := <-nodeResponseChannel
	return nodeResponse.NodeId
}

func GetActiveChannelPeerNodeIdByPublicKey(publicKey string, chain core.Chain, network core.Network) int {
	nodeResponseChannel := make(chan NodeCache)
	nodeCache := NodeCache{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      readActiveChannelPeerNode,
		Out:       nodeResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	nodeResponse := <-nodeResponseChannel
	return nodeResponse.NodeId
}

func GetConnectedPeerNodeIdByPublicKey(publicKey string, chain core.Chain, network core.Network) int {
	nodeResponseChannel := make(chan NodeCache)
	nodeCache := NodeCache{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      readConnectedPeerNode,
		Out:       nodeResponseChannel,
	}
	NodesCacheChannel <- nodeCache
	nodeResponse := <-nodeResponseChannel
	return nodeResponse.NodeId
}

func SetChannelPeerNode(nodeId int, publicKey string, chain core.Chain, network core.Network, status core.ChannelStatus) {
	if status < core.CooperativeClosed {
		NodesCacheChannel <- NodeCache{
			NodeId:    nodeId,
			PublicKey: publicKey,
			Chain:     &chain,
			Network:   &network,
			Type:      writeActiveChannelPeerNode,
		}
	} else {
		SetInactiveChannelPeerNode(nodeId, publicKey, chain, network)
	}
}

func SetInactiveChannelPeerNode(nodeId int, publicKey string, chain core.Chain, network core.Network) {
	NodesCacheChannel <- NodeCache{
		NodeId:    nodeId,
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      writeInactiveChannelPeerNode,
	}
}

func SetConnectedPeerNode(nodeId int, publicKey string, chain core.Chain, network core.Network) {
	NodesCacheChannel <- NodeCache{
		NodeId:    nodeId,
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      writeConnectedPeerNode,
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
