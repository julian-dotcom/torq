package commons

import (
	"context"

	"github.com/rs/zerolog/log"
)

var ManagedNodeChannel = make(chan ManagedNode) //nolint:gochecknoglobals

type ManagedNodeCacheOperationType uint

const (
	READ_ALL_TORQ_NODE ManagedNodeCacheOperationType = iota
	WRITE_INACTIVE_TORQ_NODE
	READ_ALL_TORQ_NODEIDS
	READ_ALL_TORQ_NODEIDS_ALL_NETWORKS
	READ_ALL_TORQ_PUBLICKEYS
	READ_ACTIVE_TORQ_NODE
	WRITE_ACTIVE_TORQ_NODE
	READ_ACTIVE_TORQ_NODEIDS
	READ_ACTIVE_TORQ_PUBLICKEYS
	READ_ALL_ACTIVE_TORQ_NODEIDS
	READ_ALL_ACTIVE_TORQ_NODESETTINGS
	READ_ACTIVE_CHANNEL_NODE
	READ_CHANNEL_NODE
	WRITE_ACTIVE_CHANNEL_NODE
	WRITE_INACTIVE_CHANNEL_NODE
	INACTIVATE_CHANNEL_NODE
	READ_ALL_CHANNEL_NODEIDS
	READ_ALL_CHANNEL_PUBLICKEYS
	READ_NODE_SETTING
	RESET_MANAGED_NODE_CACHE
)

type ManagedNode struct {
	Type            ManagedNodeCacheOperationType
	NodeId          int
	Chain           *Chain
	Network         *Network
	PublicKey       string
	Name            *string
	Out             chan ManagedNode
	NodeIdsOut      chan []int
	NodeSettingOut  chan ManagedNodeSettings
	NodeSettingsOut chan []ManagedNodeSettings
	PublicKeysOut   chan []string
}

type ManagedNodeSettings struct {
	NodeId    int
	Chain     Chain
	Network   Network
	PublicKey string
	Name      *string
	Status    Status
}

// ManagedNodeCache parameter Context is for test cases...
func ManagedNodeCache(ch chan ManagedNode, ctx context.Context) {
	allTorqNodeIdCache := make(map[Chain]map[Network]map[string]int, 0)
	nodeSettingsByNodeIdCache := make(map[int]ManagedNodeSettings, 0)
	activeTorqNodeIdCache := make(map[Chain]map[Network]map[string]int, 0)
	channelNodeIdCache := make(map[Chain]map[Network]map[string]int, 0)
	allChannelNodeIdCache := make(map[Chain]map[Network]map[string]int, 0)
	torqNodeNameByNodeIdCache := make(map[int]string, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case managedNode := <-ch:
			if managedNode.Type == RESET_MANAGED_NODE_CACHE {
				allTorqNodeIdCache = make(map[Chain]map[Network]map[string]int, 0)
				nodeSettingsByNodeIdCache = make(map[int]ManagedNodeSettings, 0)
				activeTorqNodeIdCache = make(map[Chain]map[Network]map[string]int, 0)
				channelNodeIdCache = make(map[Chain]map[Network]map[string]int, 0)
				allChannelNodeIdCache = make(map[Chain]map[Network]map[string]int, 0)
				torqNodeNameByNodeIdCache = make(map[int]string, 0)
			}
			processManagedNode(managedNode, allTorqNodeIdCache, activeTorqNodeIdCache,
				channelNodeIdCache, allChannelNodeIdCache, nodeSettingsByNodeIdCache, torqNodeNameByNodeIdCache)
		}
	}
}

func processManagedNode(managedNode ManagedNode, allTorqNodeIdCache map[Chain]map[Network]map[string]int,
	activeTorqNodeIdCache map[Chain]map[Network]map[string]int,
	channelNodeIdCache map[Chain]map[Network]map[string]int, allChannelNodeIdCache map[Chain]map[Network]map[string]int,
	nodeSettingsByNodeIdCache map[int]ManagedNodeSettings, torqNodeNameByNodeIdCache map[int]string) {
	switch managedNode.Type {
	case READ_ALL_TORQ_NODE:
		if managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(allTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			managedNode.NodeId = allTorqNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey]
			nodeName, exists := torqNodeNameByNodeIdCache[managedNode.NodeId]
			if exists {
				managedNode.Name = &nodeName
			}
		}
		SendToManagedNodeChannel(managedNode.Out, managedNode)
	case READ_ACTIVE_TORQ_NODE:
		if managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			managedNode.NodeId = activeTorqNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey]
			nodeName, exists := torqNodeNameByNodeIdCache[managedNode.NodeId]
			if exists {
				managedNode.Name = &nodeName
			}
		}
		SendToManagedNodeChannel(managedNode.Out, managedNode)
	case READ_ACTIVE_CHANNEL_NODE:
		if managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(channelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			managedNode.NodeId = channelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey]
		}
		SendToManagedNodeChannel(managedNode.Out, managedNode)
	case READ_CHANNEL_NODE:
		if managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(allChannelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			managedNode.NodeId = allChannelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey]
		}
		SendToManagedNodeChannel(managedNode.Out, managedNode)
	case READ_ALL_TORQ_NODEIDS:
		var allNodeIds []int
		if managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(allTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			for _, value := range allTorqNodeIdCache[*managedNode.Chain][*managedNode.Network] {
				allNodeIds = append(allNodeIds, value)
			}
		}
		SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, allNodeIds)
	case READ_ALL_TORQ_NODEIDS_ALL_NETWORKS:
		var allNodeIds []int
		for chainIndex := range allTorqNodeIdCache {
			for networkIndex := range allTorqNodeIdCache[chainIndex] {
				for nodeIndex := range allTorqNodeIdCache[chainIndex][networkIndex] {
					allNodeIds = append(allNodeIds, allTorqNodeIdCache[chainIndex][networkIndex][nodeIndex])
				}
			}
		}
		SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, allNodeIds)
	case READ_ACTIVE_TORQ_NODEIDS:
		var allNodeIds []int
		if managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			for _, value := range activeTorqNodeIdCache[*managedNode.Chain][*managedNode.Network] {
				allNodeIds = append(allNodeIds, value)
			}
		}
		SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, allNodeIds)
	case READ_ALL_CHANNEL_NODEIDS:
		var allNodeIds []int
		if managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(allChannelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			for _, value := range allChannelNodeIdCache[*managedNode.Chain][*managedNode.Network] {
				allNodeIds = append(allNodeIds, value)
			}
		}
		SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, allNodeIds)
	case READ_ALL_TORQ_PUBLICKEYS:
		var allPublicKeys []string
		if managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(allTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			for key := range allTorqNodeIdCache[*managedNode.Chain][*managedNode.Network] {
				allPublicKeys = append(allPublicKeys, key)
			}
		}
		SendToManagedPublicKeysChannel(managedNode.PublicKeysOut, allPublicKeys)
	case READ_ACTIVE_TORQ_PUBLICKEYS:
		var activePublicKeys []string
		if managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			for key := range activeTorqNodeIdCache[*managedNode.Chain][*managedNode.Network] {
				activePublicKeys = append(activePublicKeys, key)
			}
		}
		SendToManagedPublicKeysChannel(managedNode.PublicKeysOut, activePublicKeys)
	case READ_ALL_ACTIVE_TORQ_NODEIDS:
		var nodeIds []int
		if managedNode.Chain == nil || managedNode.Network == nil {
			for chain, networkMap := range activeTorqNodeIdCache {
				if managedNode.Chain == nil || *managedNode.Chain == chain {
					for network, publicKeyMap := range networkMap {
						if managedNode.Network == nil || *managedNode.Network == network {
							for _, nodeId := range publicKeyMap {
								nodeIds = append(nodeIds, nodeId)
							}
						}
					}
				}
			}
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			for _, nodeId := range activeTorqNodeIdCache[*managedNode.Chain][*managedNode.Network] {
				nodeIds = append(nodeIds, nodeId)
			}
		}
		SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, nodeIds)
	case READ_ALL_ACTIVE_TORQ_NODESETTINGS:
		var nodes []ManagedNodeSettings
		if managedNode.Chain == nil || managedNode.Network == nil {
			for chain, networkMap := range activeTorqNodeIdCache {
				if managedNode.Chain == nil || *managedNode.Chain == chain {
					for network, publicKeyMap := range networkMap {
						if managedNode.Network == nil || *managedNode.Network == network {
							for _, nodeId := range publicKeyMap {
								nodes = append(nodes, nodeSettingsByNodeIdCache[nodeId])
							}
						}
					}
				}
			}
		} else {
			initializeNodeIdCache(activeTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			for _, nodeId := range activeTorqNodeIdCache[*managedNode.Chain][*managedNode.Network] {
				nodes = append(nodes, nodeSettingsByNodeIdCache[nodeId])
			}
		}
		SendToManagedNodeSettingsChannel(managedNode.NodeSettingsOut, nodes)
	case READ_ALL_CHANNEL_PUBLICKEYS:
		var channelPublicKeys []string
		if managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(allChannelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			for key := range allChannelNodeIdCache[*managedNode.Chain][*managedNode.Network] {
				channelPublicKeys = append(channelPublicKeys, key)
			}
		}
		SendToManagedPublicKeysChannel(managedNode.PublicKeysOut, channelPublicKeys)
	case READ_NODE_SETTING:
		if managedNode.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", managedNode.NodeId)
			SendToManagedNodeSettingChannel(managedNode.NodeSettingOut, ManagedNodeSettings{})
		} else {
			nodeName, exists := torqNodeNameByNodeIdCache[managedNode.NodeId]
			nodeSettings := nodeSettingsByNodeIdCache[managedNode.NodeId]
			if exists {
				nodeSettings.Name = &nodeName
			}
			SendToManagedNodeSettingChannel(managedNode.NodeSettingOut, nodeSettings)
		}
	case WRITE_INACTIVE_TORQ_NODE:
		if managedNode.Name == nil || *managedNode.Name == "" || managedNode.PublicKey == "" || managedNode.NodeId == 0 ||
			managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty name (%v), publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				managedNode.Name, managedNode.PublicKey, managedNode.NodeId, managedNode.Chain, managedNode.Network)
		} else {
			torqNodeNameByNodeIdCache[managedNode.NodeId] = *managedNode.Name
			initializeNodeIdCache(allTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			allTorqNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
			initializeNodeIdCache(allChannelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			allChannelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
			nodeSettingsByNodeIdCache[managedNode.NodeId] = ManagedNodeSettings{
				NodeId:    managedNode.NodeId,
				Network:   *managedNode.Network,
				Chain:     *managedNode.Chain,
				PublicKey: managedNode.PublicKey,
				Name:      managedNode.Name,
			}
		}
	case WRITE_ACTIVE_TORQ_NODE:
		if managedNode.Name == nil || *managedNode.Name == "" || managedNode.PublicKey == "" || managedNode.NodeId == 0 ||
			managedNode.Chain == nil || managedNode.Network == nil {
			log.Error().Msgf("No empty name (%v), publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				managedNode.Name, managedNode.PublicKey, managedNode.NodeId, managedNode.Chain, managedNode.Network)
		} else {
			torqNodeNameByNodeIdCache[managedNode.NodeId] = *managedNode.Name
			initializeNodeIdCache(activeTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			activeTorqNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
			initializeNodeIdCache(allTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
			allTorqNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
			initializeNodeIdCache(channelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			channelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
			initializeNodeIdCache(allChannelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			allChannelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
			nodeSettingsByNodeIdCache[managedNode.NodeId] = ManagedNodeSettings{
				NodeId:    managedNode.NodeId,
				Network:   *managedNode.Network,
				Chain:     *managedNode.Chain,
				PublicKey: managedNode.PublicKey,
				Name:      managedNode.Name,
			}
		}
	case WRITE_ACTIVE_CHANNEL_NODE:
		if managedNode.PublicKey == "" || managedNode.NodeId == 0 || managedNode.Chain == nil ||
			managedNode.Network == nil {
			log.Error().Msgf("No empty publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				managedNode.PublicKey, managedNode.NodeId, managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(channelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			channelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
			initializeNodeIdCache(allChannelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			allChannelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
			nodeSettingsByNodeIdCache[managedNode.NodeId] = ManagedNodeSettings{
				NodeId:    managedNode.NodeId,
				Network:   *managedNode.Network,
				Chain:     *managedNode.Chain,
				PublicKey: managedNode.PublicKey,
				Status:    Active,
			}
		}
	case WRITE_INACTIVE_CHANNEL_NODE:
		if managedNode.PublicKey == "" || managedNode.NodeId == 0 || managedNode.Chain == nil ||
			managedNode.Network == nil {
			log.Error().Msgf("No empty publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
				managedNode.PublicKey, managedNode.NodeId, managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(allChannelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			allChannelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
			nodeSettingsByNodeIdCache[managedNode.NodeId] = ManagedNodeSettings{
				NodeId:    managedNode.NodeId,
				Network:   *managedNode.Network,
				Chain:     *managedNode.Chain,
				PublicKey: managedNode.PublicKey,
				Status:    Inactive,
			}
		}
	case INACTIVATE_CHANNEL_NODE:
		if managedNode.Chain == nil || managedNode.Network == nil || managedNode.PublicKey == "" {
			log.Error().Msgf("No empty publicKey (%v), Chain (%v) or Network (%v) allowed", managedNode.PublicKey, managedNode.Chain, managedNode.Network)
		} else {
			initializeNodeIdCache(channelNodeIdCache, *managedNode.Chain, *managedNode.Network)
			delete(channelNodeIdCache[*managedNode.Chain][*managedNode.Network], managedNode.PublicKey)
			nodeSettingsByNodeIdCache[managedNode.NodeId] = ManagedNodeSettings{
				NodeId:    managedNode.NodeId,
				Network:   *managedNode.Network,
				Chain:     *managedNode.Chain,
				PublicKey: managedNode.PublicKey,
				Status:    Inactive,
			}
		}
	}
}

func initializeNodeIdCache(nodeIdCache map[Chain]map[Network]map[string]int, chain Chain, network Network) {
	if nodeIdCache[chain] == nil {
		nodeIdCache[chain] = make(map[Network]map[string]int, 0)
	}
	if nodeIdCache[chain][network] == nil {
		nodeIdCache[chain][network] = make(map[string]int, 0)
	}
}

func GetAllChannelPublicKeys(chain Chain, network Network) []string {
	publicKeysResponseChannel := make(chan []string, 1)
	managedNode := ManagedNode{
		Chain:         &chain,
		Network:       &network,
		Type:          READ_ALL_CHANNEL_PUBLICKEYS,
		PublicKeysOut: publicKeysResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-publicKeysResponseChannel
}

func GetAllTorqPublicKeys(chain Chain, network Network) []string {
	publicKeysResponseChannel := make(chan []string, 1)
	managedNode := ManagedNode{
		Chain:         &chain,
		Network:       &network,
		Type:          READ_ALL_TORQ_PUBLICKEYS,
		PublicKeysOut: publicKeysResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-publicKeysResponseChannel
}

func GetAllTorqNodeIdsByNetwork(chain Chain, network Network) []int {
	nodeIdsResponseChannel := make(chan []int, 1)
	managedNode := ManagedNode{
		Chain:      &chain,
		Network:    &network,
		Type:       READ_ALL_TORQ_NODEIDS,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-nodeIdsResponseChannel
}

func GetAllTorqNodeIds() []int {
	nodeIdsResponseChannel := make(chan []int, 1)
	managedNode := ManagedNode{
		Type:       READ_ALL_TORQ_NODEIDS_ALL_NETWORKS,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-nodeIdsResponseChannel
}

func GetChannelPublicKeys(chain Chain, network Network) []string {
	publicKeysResponseChannel := make(chan []string, 1)
	managedNode := ManagedNode{
		Chain:         &chain,
		Network:       &network,
		Type:          READ_ALL_CHANNEL_PUBLICKEYS,
		PublicKeysOut: publicKeysResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-publicKeysResponseChannel
}

func GetChannelNodeIds(chain Chain, network Network) []int {
	nodeIdsResponseChannel := make(chan []int)
	managedNode := ManagedNode{
		Chain:      &chain,
		Network:    &network,
		Type:       READ_ALL_CHANNEL_NODEIDS,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-nodeIdsResponseChannel
}

func GetAllActiveTorqNodeIds(chain *Chain, network *Network) []int {
	nodeIdsResponseChannel := make(chan []int)
	managedNode := ManagedNode{
		Chain:      chain,
		Network:    network,
		Type:       READ_ALL_ACTIVE_TORQ_NODEIDS,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-nodeIdsResponseChannel
}

func GetActiveTorqNodeSettings() []ManagedNodeSettings {
	nodesResponseChannel := make(chan []ManagedNodeSettings)
	managedNode := ManagedNode{
		Type:            READ_ALL_ACTIVE_TORQ_NODESETTINGS,
		NodeSettingsOut: nodesResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-nodesResponseChannel
}

// SetTorqNode When active then also adds to channelNodes
func SetTorqNode(nodeId int, name string, status Status, publicKey string, chain Chain, network Network) {
	if status == Active {
		managedNode := ManagedNode{
			PublicKey: publicKey,
			Chain:     &chain,
			Network:   &network,
			NodeId:    nodeId,
			Name:      &name,
			Type:      WRITE_ACTIVE_TORQ_NODE,
		}
		ManagedNodeChannel <- managedNode
	} else {
		managedNode := ManagedNode{
			PublicKey: publicKey,
			Chain:     &chain,
			Network:   &network,
			NodeId:    nodeId,
			Name:      &name,
			Type:      WRITE_INACTIVE_TORQ_NODE,
		}
		ManagedNodeChannel <- managedNode
	}
}

func GetNodeIdByPublicKey(publicKey string, chain Chain, network Network) int {
	nodeResponseChannel := make(chan ManagedNode)
	managedNode := ManagedNode{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      READ_CHANNEL_NODE,
		Out:       nodeResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	nodeResponse := <-nodeResponseChannel
	return nodeResponse.NodeId
}

func GetActiveNodeIdByPublicKey(publicKey string, chain Chain, network Network) int {
	nodeResponseChannel := make(chan ManagedNode)
	managedNode := ManagedNode{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      READ_ACTIVE_CHANNEL_NODE,
		Out:       nodeResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	nodeResponse := <-nodeResponseChannel
	return nodeResponse.NodeId
}

func SetChannelNode(nodeId int, publicKey string, chain Chain, network Network, status ChannelStatus) {
	if status < CooperativeClosed {
		ManagedNodeChannel <- ManagedNode{
			NodeId:    nodeId,
			PublicKey: publicKey,
			Chain:     &chain,
			Network:   &network,
			Type:      WRITE_ACTIVE_CHANNEL_NODE,
		}
	} else {
		SetInactiveChannelNode(nodeId, publicKey, chain, network)
	}
}

func SetInactiveChannelNode(nodeId int, publicKey string, chain Chain, network Network) {
	ManagedNodeChannel <- ManagedNode{
		NodeId:    nodeId,
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      WRITE_INACTIVE_CHANNEL_NODE,
	}
}

func InactivateChannelNode(publicKey string, chain Chain, network Network) {
	managedNode := ManagedNode{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      INACTIVATE_CHANNEL_NODE,
	}
	ManagedNodeChannel <- managedNode
}

func GetNodeSettingsByNodeId(nodeId int) ManagedNodeSettings {
	nodeResponseChannel := make(chan ManagedNodeSettings)
	managedNode := ManagedNode{
		NodeId:         nodeId,
		Type:           READ_NODE_SETTING,
		NodeSettingOut: nodeResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-nodeResponseChannel
}

func ResetManagedNodeCache() {
	ManagedNodeChannel <- ManagedNode{
		Type: RESET_MANAGED_NODE_CACHE,
	}
}
