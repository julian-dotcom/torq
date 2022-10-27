package commons

var ManagedNodeChannel = make(chan ManagedNode)

type ManagedNodeCacheOperationType uint

const (
	READ_ALL_TORQ_NODE ManagedNodeCacheOperationType = iota
	WRITE_INACTIVE_TORQ_NODE
	DELETE_INACTIVE_TORQ_NODE
	READ_ALL_TORQ_NODEIDS
	READ_ALL_TORQ_PUBLICKEYS
	READ_ACTIVE_TORQ_NODE
	WRITE_ACTIVE_TORQ_NODE
	DELETE_ACTIVE_TORQ_NODE
	READ_ACTIVE_TORQ_NODEIDS
	READ_ACTIVE_TORQ_PUBLICKEYS
	READ_CHANNEL_NODE
	WRITE_CHANNEL_NODE
	DELETE_CHANNEL_NODE
	READ_CHANNEL_NODEIDS
	READ_CHANNEL_PUBLICKEYS
)

type ManagedNode struct {
	Type          ManagedNodeCacheOperationType
	NodeId        int
	PublicKey     string
	Out           chan ManagedNode
	NodeIdsOut    chan []int
	PublicKeysOut chan []string
}

func ManagedNodeCache(ch chan ManagedNode) {
	allTorqNodeCache := make(map[string]int, 0)
	activeTorqNodeCache := make(map[string]int, 0)
	channelNodeCache := make(map[string]int, 0)
	for {
		managedNode := <-ch
		switch managedNode.Type {
		case READ_ALL_TORQ_NODE:
			managedNode.NodeId = allTorqNodeCache[managedNode.PublicKey]
			go SendToManagedNodeChannel(managedNode.Out, managedNode)
		case READ_ACTIVE_TORQ_NODE:
			managedNode.NodeId = activeTorqNodeCache[managedNode.PublicKey]
			go SendToManagedNodeChannel(managedNode.Out, managedNode)
		case READ_CHANNEL_NODE:
			managedNode.NodeId = channelNodeCache[managedNode.PublicKey]
			go SendToManagedNodeChannel(managedNode.Out, managedNode)
		case READ_ALL_TORQ_NODEIDS:
			allNodeIds := make([]int, len(allTorqNodeCache))
			for _, value := range allTorqNodeCache {
				allNodeIds = append(allNodeIds, value)
			}
			go SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, allNodeIds)
		case READ_ACTIVE_TORQ_NODEIDS:
			allNodeIds := make([]int, len(activeTorqNodeCache))
			for _, value := range activeTorqNodeCache {
				allNodeIds = append(allNodeIds, value)
			}
			go SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, allNodeIds)
		case READ_CHANNEL_NODEIDS:
			allNodeIds := make([]int, len(channelNodeCache))
			for _, value := range channelNodeCache {
				allNodeIds = append(allNodeIds, value)
			}
			go SendToManagedNodeIdsChannel(managedNode.NodeIdsOut, allNodeIds)
		case READ_ALL_TORQ_PUBLICKEYS:
			allPublicKeys := make([]string, len(allTorqNodeCache))
			for key := range allTorqNodeCache {
				allPublicKeys = append(allPublicKeys, key)
			}
			go SendToManagedPublicKeysChannel(managedNode.PublicKeysOut, allPublicKeys)
		case READ_ACTIVE_TORQ_PUBLICKEYS:
			activePublicKeys := make([]string, len(activeTorqNodeCache))
			for key := range activeTorqNodeCache {
				activePublicKeys = append(activePublicKeys, key)
			}
			go SendToManagedPublicKeysChannel(managedNode.PublicKeysOut, activePublicKeys)
		case READ_CHANNEL_PUBLICKEYS:
			channelPublicKeys := make([]string, len(channelNodeCache))
			for key := range channelNodeCache {
				channelPublicKeys = append(channelPublicKeys, key)
			}
			go SendToManagedPublicKeysChannel(managedNode.PublicKeysOut, channelPublicKeys)
		case WRITE_INACTIVE_TORQ_NODE:
			allTorqNodeCache[managedNode.PublicKey] = managedNode.NodeId
		case WRITE_ACTIVE_TORQ_NODE:
			activeTorqNodeCache[managedNode.PublicKey] = managedNode.NodeId
			allTorqNodeCache[managedNode.PublicKey] = managedNode.NodeId
		case WRITE_CHANNEL_NODE:
			channelNodeCache[managedNode.PublicKey] = managedNode.NodeId
		case DELETE_INACTIVE_TORQ_NODE:
			delete(allTorqNodeCache, managedNode.PublicKey)
		case DELETE_ACTIVE_TORQ_NODE:
			delete(activeTorqNodeCache, managedNode.PublicKey)
			delete(allTorqNodeCache, managedNode.PublicKey)
		case DELETE_CHANNEL_NODE:
			delete(channelNodeCache, managedNode.PublicKey)
		}
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

func GetAllTorqPublicKeys() []string {
	publicKeysResponseChannel := make(chan []string)
	managedNode := ManagedNode{
		Type:          READ_ALL_TORQ_PUBLICKEYS,
		PublicKeysOut: publicKeysResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-publicKeysResponseChannel
}

func GetActiveTorqNodeIdFromPublicKey(publicKey string) int {
	nodeResponseChannel := make(chan ManagedNode)
	managedNode := ManagedNode{
		PublicKey: publicKey,
		Type:      READ_ACTIVE_TORQ_NODE,
		Out:       nodeResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	nodeResponse := <-nodeResponseChannel
	return nodeResponse.NodeId
}

func GetNodeIdFromPublicKey(publicKey string) int {
	nodeResponseChannel := make(chan ManagedNode)
	managedNode := ManagedNode{
		PublicKey: publicKey,
		Type:      READ_CHANNEL_NODE,
		Out:       nodeResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	nodeResponse := <-nodeResponseChannel
	return nodeResponse.NodeId
}

func GetAllTorqNodeIds() []int {
	nodeIdsResponseChannel := make(chan []int)
	managedNode := ManagedNode{
		Type:       READ_ALL_TORQ_NODEIDS,
		NodeIdsOut: nodeIdsResponseChannel,
	}
	ManagedNodeChannel <- managedNode
	return <-nodeIdsResponseChannel
}
