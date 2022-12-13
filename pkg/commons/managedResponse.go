package commons

func SendToManagedNodeIdsChannel(ch chan []int, nodeIds []int) {
	ch <- nodeIds
}

func SendToManagedChannelIdsChannel(ch chan []int, channelIds []int) {
	ch <- channelIds
}

func SendToManagedPublicKeysChannel(ch chan []string, publicKeys []string) {
	ch <- publicKeys
}

func SendToManagedChannelGroupSettingsChannel(ch chan *ManagedChannelGroupSettings, managedChannelGroupSettings *ManagedChannelGroupSettings) {
	ch <- managedChannelGroupSettings
}

func SendToManagedNodeChannel(ch chan ManagedNode, managedNode ManagedNode) {
	ch <- managedNode
}

func SendToManagedNodeSettingChannel(ch chan ManagedNodeSettings, nodeSetting ManagedNodeSettings) {
	ch <- nodeSetting
}

func SendToManagedNodeSettingsChannel(ch chan []ManagedNodeSettings, nodeSettings []ManagedNodeSettings) {
	ch <- nodeSettings
}

func SendToManagedChannelChannel(ch chan ManagedChannel, managedChannel ManagedChannel) {
	ch <- managedChannel
}

func SendToManagedChannelSettingChannel(ch chan ManagedChannelSettings, channelSettings ManagedChannelSettings) {
	ch <- channelSettings
}

func SendToManagedChannelSettingsChannel(ch chan []ManagedChannelSettings, channelSettings []ManagedChannelSettings) {
	ch <- channelSettings
}

func SendToManagedChannelBalanceStatesSettingsChannel(ch chan []ManagedChannelBalanceStateSettings, managedChannelBalanceStateSettings []ManagedChannelBalanceStateSettings) {
	ch <- managedChannelBalanceStateSettings
}

func SendToManagedChannelBalanceStateSettingsChannel(ch chan *ManagedChannelBalanceStateSettings, managedChannelBalanceStateSettings *ManagedChannelBalanceStateSettings) {
	ch <- managedChannelBalanceStateSettings
}

func SendToManagedChannelStatesSettingsChannel(ch chan []ManagedChannelStateSettings, managedChannelBalanceStateSettings []ManagedChannelStateSettings) {
	ch <- managedChannelBalanceStateSettings
}

func SendToManagedChannelStateSettingsChannel(ch chan *ManagedChannelStateSettings, managedChannelStateSettings *ManagedChannelStateSettings) {
	ch <- managedChannelStateSettings
}

func SendToManagedTriggerChannel(ch chan ManagedTrigger, managedTrigger ManagedTrigger) {
	ch <- managedTrigger
}

func SendToManagedTriggerSettingsChannel(ch chan ManagedTriggerSettings, managedTriggerSettings ManagedTriggerSettings) {
	ch <- managedTriggerSettings
}
