package commons

import "github.com/rs/zerolog/log"

var ManagedNodeChannel = make(chan ManagedNode)

type ManagedNodeCacheOperationType uint

const (
	READ_ALL_TORQ_NODE ManagedNodeCacheOperationType = iota
	WRITE_INACTIVE_TORQ_NODE
	READ_ALL_TORQ_NODEIDS
	READ_ALL_TORQ_PUBLICKEYS
	READ_ACTIVE_TORQ_NODE
	WRITE_ACTIVE_TORQ_NODE
	READ_ACTIVE_TORQ_NODEIDS
	READ_ACTIVE_TORQ_PUBLICKEYS
	READ_CHANNEL_NODE
	WRITE_CHANNEL_NODE
	DELETE_CHANNEL_NODE
	READ_CHANNEL_NODEIDS
	READ_CHANNEL_PUBLICKEYS
	READ_NODE_SETTINGS
)

type ManagedNode struct {
	Type              ManagedNodeCacheOperationType
	NodeId            int
	Chain             *Chain
	Network           *Network
	PublicKey         string
	Out               chan ManagedNode
	NodeIdsOut        chan []int
	NodeIdSettingsOut chan ManagedNodeSettings
	PublicKeysOut     chan []string
}

type ManagedNodeSettings struct {
	NodeId    int
	Chain     Chain
	Network   Network
	PublicKey string
}

func ManagedNodeCache(ch chan ManagedNode) {
	allTorqNodeIdCache := make(map[Chain]map[Network]map[string]int, 0)
	nodeSettingsByNodeIdCache := make(map[int]ManagedNodeSettings, 0)
	activeTorqNodeIdCache := make(map[Chain]map[Network]map[string]int, 0)
	channelNodeIdCache := make(map[Chain]map[Network]map[string]int, 0)
	for {
		managedNode := <-ch
		switch managedNode.Type {
		case READ_ALL_TORQ_NODE:
			if managedNode.Chain == nil || managedNode.Network == nil {
				log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(allTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
				managedNode.NodeId = allTorqNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey]
			}
			go SendToManagedNodeChannel(managedNode.Out, managedNode)
		case READ_ACTIVE_TORQ_NODE:
			if managedNode.Chain == nil || managedNode.Network == nil {
				log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(activeTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
				managedNode.NodeId = activeTorqNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey]
			}
			go SendToManagedNodeChannel(managedNode.Out, managedNode)
		case READ_CHANNEL_NODE:
			if managedNode.Chain == nil || managedNode.Network == nil {
				log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(channelNodeIdCache, *managedNode.Chain, *managedNode.Network)
				managedNode.NodeId = channelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey]
			}
			go SendToManagedNodeChannel(managedNode.Out, managedNode)
		case READ_ALL_TORQ_NODEIDS:
			allNodeIds := make([]int, len(allTorqNodeIdCache))
			if managedNode.Chain == nil || managedNode.Network == nil {
				log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(allTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
				for _, value := range allTorqNodeIdCache[*managedNode.Chain][*managedNode.Network] {
					allNodeIds = append(allNodeIds, value)
				}
			}
			go SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, allNodeIds)
		case READ_ACTIVE_TORQ_NODEIDS:
			allNodeIds := make([]int, len(activeTorqNodeIdCache))
			if managedNode.Chain == nil || managedNode.Network == nil {
				log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(activeTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
				for _, value := range activeTorqNodeIdCache[*managedNode.Chain][*managedNode.Network] {
					allNodeIds = append(allNodeIds, value)
				}
			}
			go SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, allNodeIds)
		case READ_CHANNEL_NODEIDS:
			allNodeIds := make([]int, len(channelNodeIdCache))
			if managedNode.Chain == nil || managedNode.Network == nil {
				log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(channelNodeIdCache, *managedNode.Chain, *managedNode.Network)
				for _, value := range channelNodeIdCache[*managedNode.Chain][*managedNode.Network] {
					allNodeIds = append(allNodeIds, value)
				}
			}
			go SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, allNodeIds)
		case READ_ALL_TORQ_PUBLICKEYS:
			allPublicKeys := make([]string, len(allTorqNodeIdCache))
			if managedNode.Chain == nil || managedNode.Network == nil {
				log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(allTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
				for key := range allTorqNodeIdCache[*managedNode.Chain][*managedNode.Network] {
					allPublicKeys = append(allPublicKeys, key)
				}
			}
			go SendToManagedPublicKeysChannel(managedNode.PublicKeysOut, allPublicKeys)
		case READ_ACTIVE_TORQ_PUBLICKEYS:
			activePublicKeys := make([]string, len(activeTorqNodeIdCache))
			if managedNode.Chain == nil || managedNode.Network == nil {
				log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(activeTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
				for key := range activeTorqNodeIdCache[*managedNode.Chain][*managedNode.Network] {
					activePublicKeys = append(activePublicKeys, key)
				}
			}
			go SendToManagedPublicKeysChannel(managedNode.PublicKeysOut, activePublicKeys)
		case READ_CHANNEL_PUBLICKEYS:
			channelPublicKeys := make([]string, len(channelNodeIdCache))
			if managedNode.Chain == nil || managedNode.Network == nil {
				log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(channelNodeIdCache, *managedNode.Chain, *managedNode.Network)
				for key := range channelNodeIdCache[*managedNode.Chain][*managedNode.Network] {
					channelPublicKeys = append(channelPublicKeys, key)
				}
			}
			go SendToManagedPublicKeysChannel(managedNode.PublicKeysOut, channelPublicKeys)
		case READ_NODE_SETTINGS:
			go SendToManagedNodeSettingsChannel(managedNode.NodeIdSettingsOut, nodeSettingsByNodeIdCache[managedNode.NodeId])
		case WRITE_INACTIVE_TORQ_NODE:
			if managedNode.PublicKey == "" || managedNode.NodeId == 0 || managedNode.Chain == nil ||
				managedNode.Network == nil {
				log.Error().Msgf("No empty publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
					managedNode.PublicKey, managedNode.NodeId, managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(allTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
				allTorqNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
				nodeSettingsByNodeIdCache[managedNode.NodeId] = ManagedNodeSettings{
					NodeId:    managedNode.NodeId,
					Network:   *managedNode.Network,
					Chain:     *managedNode.Chain,
					PublicKey: managedNode.PublicKey,
				}
			}
		case WRITE_ACTIVE_TORQ_NODE:
			if managedNode.PublicKey == "" || managedNode.NodeId == 0 || managedNode.Chain == nil ||
				managedNode.Network == nil {
				log.Error().Msgf("No empty publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
					managedNode.PublicKey, managedNode.NodeId, managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(activeTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
				activeTorqNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
				initializeIdCache(allTorqNodeIdCache, *managedNode.Chain, *managedNode.Network)
				allTorqNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
				initializeIdCache(channelNodeIdCache, *managedNode.Chain, *managedNode.Network)
				channelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
				nodeSettingsByNodeIdCache[managedNode.NodeId] = ManagedNodeSettings{
					NodeId:    managedNode.NodeId,
					Network:   *managedNode.Network,
					Chain:     *managedNode.Chain,
					PublicKey: managedNode.PublicKey,
				}
			}
		case WRITE_CHANNEL_NODE:
			if managedNode.PublicKey == "" || managedNode.NodeId == 0 || managedNode.Chain == nil ||
				managedNode.Network == nil {
				log.Error().Msgf("No empty publicKey (%v), chain (%v), network (%v) or nodeId (%v) allowed",
					managedNode.PublicKey, managedNode.NodeId, managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(channelNodeIdCache, *managedNode.Chain, *managedNode.Network)
				channelNodeIdCache[*managedNode.Chain][*managedNode.Network][managedNode.PublicKey] = managedNode.NodeId
				nodeSettingsByNodeIdCache[managedNode.NodeId] = ManagedNodeSettings{
					NodeId:    managedNode.NodeId,
					Network:   *managedNode.Network,
					Chain:     *managedNode.Chain,
					PublicKey: managedNode.PublicKey,
				}
			}
		case DELETE_CHANNEL_NODE:
			if managedNode.Chain == nil || managedNode.Network == nil {
				log.Error().Msgf("No empty Chain (%v) or Network (%v) allowed", managedNode.Chain, managedNode.Network)
			} else {
				initializeIdCache(channelNodeIdCache, *managedNode.Chain, *managedNode.Network)
				delete(channelNodeIdCache[*managedNode.Chain][*managedNode.Network], managedNode.PublicKey)
			}
		}
	}
}

func initializeIdCache(nodeIdCache map[Chain]map[Network]map[string]int, chain Chain, network Network) {
	if nodeIdCache[chain] == nil {
		nodeIdCache[chain] = make(map[Network]map[string]int, 0)
	}
	if nodeIdCache[chain][network] == nil {
		nodeIdCache[chain][network] = make(map[string]int, 0)
	}
}

func SendToManagedNodeChannel(ch chan ManagedNode, managedNode ManagedNode) {
	ch <- managedNode
}

func SendToManagedNodeIdsChannel(ch chan []int, channelIds []int) {
	ch <- channelIds
}

func SendToManagedPublicKeysChannel(ch chan []string, publicKeys []string) {
	ch <- publicKeys
}

func SendToManagedNodeSettingsChannel(ch chan ManagedNodeSettings, nodeSettings ManagedNodeSettings) {
	ch <- nodeSettings
}

func GetAllTorqPublicKeys(chain Chain, network Network) []string {
	publicKeysResponseChannel := make(chan []string)
	managedNode := ManagedNode{
		Chain:         &chain,
		Network:       &network,
		Type:          READ_ALL_TORQ_PUBLICKEYS,
		PublicKeysOut: publicKeysResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-publicKeysResponseChannel
}

func GetAllTorqNodeIds(chain Chain, network Network) []int {
	nodeIdsResponseChannel := make(chan []int)
	managedNode := ManagedNode{
		Chain:      &chain,
		Network:    &network,
		Type:       READ_ALL_TORQ_NODEIDS,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-nodeIdsResponseChannel
}

func GetChannelPublicKeys(chain Chain, network Network) []string {
	publicKeysResponseChannel := make(chan []string)
	managedNode := ManagedNode{
		Chain:         &chain,
		Network:       &network,
		Type:          READ_CHANNEL_PUBLICKEYS,
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
		Type:       READ_CHANNEL_NODEIDS,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-nodeIdsResponseChannel
}

func GetActiveTorqNodeIdFromPublicKey(publicKey string, chain Chain, network Network) int {
	nodeResponseChannel := make(chan ManagedNode)
	managedNode := ManagedNode{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      READ_ACTIVE_TORQ_NODE,
		Out:       nodeResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	nodeResponse := <-nodeResponseChannel
	return nodeResponse.NodeId
}

// SetActiveTorqNode also adds to channelNodes
func SetActiveTorqNode(nodeId int, publicKey string, chain Chain, network Network) {
	managedNode := ManagedNode{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		NodeId:    nodeId,
		Type:      WRITE_ACTIVE_TORQ_NODE,
	}
	ManagedNodeChannel <- managedNode
}

func SetInactiveTorqNode(nodeId int, publicKey string, chain Chain, network Network) {
	managedNode := ManagedNode{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		NodeId:    nodeId,
		Type:      WRITE_INACTIVE_TORQ_NODE,
	}
	ManagedNodeChannel <- managedNode
}

func GetNodeIdFromPublicKey(publicKey string, chain Chain, network Network) int {
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

func SetChannelNode(nodeId int, publicKey string, chain Chain, network Network) {
	ManagedNodeChannel <- ManagedNode{
		NodeId:    nodeId,
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      WRITE_CHANNEL_NODE,
	}
}
func RemoveChannelNode(publicKey string, chain Chain, network Network) {
	managedNode := ManagedNode{
		PublicKey: publicKey,
		Chain:     &chain,
		Network:   &network,
		Type:      DELETE_CHANNEL_NODE,
	}
	ManagedNodeChannel <- managedNode
}

func GetNodeSettingsByNodeId(nodeId int) ManagedNodeSettings {
	nodeResponseChannel := make(chan ManagedNodeSettings)
	managedNode := ManagedNode{
		NodeId:            nodeId,
		Type:              READ_NODE_SETTINGS,
		NodeIdSettingsOut: nodeResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-nodeResponseChannel
}
