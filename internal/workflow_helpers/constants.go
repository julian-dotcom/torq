package workflow_helpers

type WorkflowNodeType int

const (
	WorkflowNodeIntervalTrigger = WorkflowNodeType(iota)
	WorkflowNodeChannelBalanceEventTrigger
	WorkflowNodeChannelFilter
	WorkflowNodeChannelPolicyConfigurator
	WorkflowNodeRebalanceConfigurator
	WorkflowNodeStageTrigger
	WorkflowNodeRebalanceRun
	WorkflowNodeChannelPolicyRun
	WorkflowNodeSetVariable
	WorkflowNodeFilterOnVariable
	WorkflowTrigger
	WorkflowNodeAddTag
	WorkflowNodeRemoveTag
	WorkflowNodeChannelOpenEventTrigger
	WorkflowNodeChannelCloseEventTrigger
	WorkflowNodeCronTrigger
	WorkflowNodeManualTrigger
	WorkflowNodeChannelPolicyAutoRun
	WorkflowNodeRebalanceAutoRun
	WorkflowNodeDataSourceTorqChannels
	WorkflowNodeChannelBalanceEventFilter
)

type WorkflowParameterType string

const (
	WorkflowParameterTypeChannelIds            = WorkflowParameterType("channelIds")
	WorkflowParameterTypeRoutingPolicySettings = WorkflowParameterType("routingPolicySettings")
	WorkflowParameterTypeRebalanceSettings     = WorkflowParameterType("rebalanceSettings")
	WorkflowParameterTypeTagSettings           = WorkflowParameterType("tagSettings")
	WorkflowParameterTypeStatus                = WorkflowParameterType("status")
)

type WorkflowParameterLabel string

const (
	WorkflowParameterLabelChannels              = WorkflowParameterLabel("channels")
	WorkflowParameterLabelRoutingPolicySettings = WorkflowParameterLabel("routingPolicySettings")
	WorkflowParameterLabelRebalanceSettings     = WorkflowParameterLabel("rebalanceSettings")
	WorkflowParameterLabelTagSettings           = WorkflowParameterLabel("tagSettings")
	WorkflowParameterLabelIncomingChannels      = WorkflowParameterLabel("incomingChannels")
	WorkflowParameterLabelOutgoingChannels      = WorkflowParameterLabel("outgoingChannels")
	WorkflowParameterLabelStatus                = WorkflowParameterLabel("status")
	WorkflowParameterLabelAllChannels           = WorkflowParameterLabel("allChannels")
	WorkflowParameterLabelEventChannels         = WorkflowParameterLabel("eventChannels")
	WorkflowParameterLabelEvents                = WorkflowParameterLabel("events")
)

type WorkflowNodeTypeParameters struct {
	WorkflowNodeType WorkflowNodeType
	RequiredInputs   map[WorkflowParameterLabel]WorkflowParameterType
	OptionalInputs   map[WorkflowParameterLabel]WorkflowParameterType
	RequiredOutputs  map[WorkflowParameterLabel]WorkflowParameterType
	OptionalOutputs  map[WorkflowParameterLabel]WorkflowParameterType
}
