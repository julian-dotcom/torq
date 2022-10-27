package commons

var ManagedChannelChannel = make(chan ManagedChannel)

type ManagedChannelCacheOperationType uint

const (
	READ_SHORTCHANNELID ManagedChannelCacheOperationType = iota
	READ_CHANNELPOINT
	READ_STATUS
	READ_ALL_CHANNELPOINTS
	WRITE_CHANNEL
	DELETE_CHANNEL
)

type ManagedChannel struct {
	Type             ManagedChannelCacheOperationType
	ChannelId        int
	ShortChannelId   string
	LndChannelPoint  string
	StatusId         int
	Out              chan ManagedChannel
	ChannelPointsOut chan []string
}

func ManagedChannelCache(ch chan ManagedChannel) {
	shortChannelIdCache := make(map[string]int, 0)
	channelPointCache := make(map[string]int, 0)
	channelStatusIdCache := make(map[int]int, 0)
	for {
		managedChannel := <-ch
		switch managedChannel.Type {
		case READ_SHORTCHANNELID:
			managedChannel.ChannelId = shortChannelIdCache[managedChannel.ShortChannelId]
			go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
		case READ_CHANNELPOINT:
			managedChannel.ChannelId = channelPointCache[managedChannel.LndChannelPoint]
			go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
		case READ_STATUS:
			managedChannel.StatusId = shortChannelIdCache[managedChannel.ShortChannelId]
			go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
		case READ_ALL_CHANNELPOINTS:
			allChannelPoints := make([]string, 0, len(channelPointCache))
			for channelPoint := range channelPointCache {
				allChannelPoints = append(allChannelPoints, channelPoint)
			}
			go SendToManagedChannelIdsChannel(managedChannel.ChannelPointsOut, allChannelPoints)
		case WRITE_CHANNEL:
			shortChannelIdCache[managedChannel.ShortChannelId] = managedChannel.ChannelId
			channelPointCache[managedChannel.LndChannelPoint] = managedChannel.ChannelId
			channelStatusIdCache[managedChannel.ChannelId] = managedChannel.StatusId
		case DELETE_CHANNEL:
			delete(shortChannelIdCache, managedChannel.ShortChannelId)
			delete(channelPointCache, managedChannel.LndChannelPoint)
			delete(channelStatusIdCache, managedChannel.ChannelId)
		}
	}
}

func SendToManagedChannelChannel(ch chan ManagedChannel, managedChannel ManagedChannel) {
	ch <- managedChannel
}

func SendToManagedChannelIdsChannel(ch chan []string, channelPoints []string) {
	ch <- channelPoints
}

func GetChannelIdFromChannelPoint(channelPoint string) int {
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		LndChannelPoint: channelPoint,
		Type:            READ_CHANNELPOINT,
		Out:             channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdFromShortChannelId(shortChannelId string) int {
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		ShortChannelId: shortChannelId,
		Type:           READ_SHORTCHANNELID,
		Out:            channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelStatusIdFromChannelId(channelId int) int {
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		ChannelId: channelId,
		Type:      READ_STATUS,
		Out:       channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.StatusId
}
