package workflow_helpers

func CloneParameters(parameters map[WorkflowParameterLabel]string) map[WorkflowParameterLabel]string {
	parametersCopy := make(map[WorkflowParameterLabel]string)
	for k, v := range parameters {
		parametersCopy[k] = v
	}
	return parametersCopy
}

func CopyParameters(destination map[WorkflowParameterLabel]string, source map[WorkflowParameterLabel]string) {
	for k, v := range source {
		destination[k] = v
	}
}

func IsWorkflowNodeTypeGrouped(workflowNodeType WorkflowNodeType) bool {
	switch workflowNodeType {
	case WorkflowNodeIntervalTrigger:
		return true
	case WorkflowNodeCronTrigger:
		return true
	case WorkflowNodeChannelBalanceEventTrigger:
		return true
	case WorkflowNodeChannelOpenEventTrigger:
		return true
	case WorkflowNodeChannelCloseEventTrigger:
		return true
	}
	return false
}

func GetWorkflowParameterLabelsEnforced() []WorkflowParameterLabel {
	return []WorkflowParameterLabel{
		WorkflowParameterLabelRoutingPolicySettings,
		WorkflowParameterLabelRebalanceSettings,
		WorkflowParameterLabelStatus,
	}
}

func GetWorkflowNodes() map[WorkflowNodeType]WorkflowNodeTypeParameters {
	all := make(map[WorkflowParameterLabel]WorkflowParameterType)
	all[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	all[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings
	all[WorkflowParameterLabelRebalanceSettings] = WorkflowParameterTypeRebalanceSettings
	all[WorkflowParameterLabelTagSettings] = WorkflowParameterTypeTagSettings
	all[WorkflowParameterLabelIncomingChannels] = WorkflowParameterTypeChannelIds
	all[WorkflowParameterLabelOutgoingChannels] = WorkflowParameterTypeChannelIds
	all[WorkflowParameterLabelStatus] = WorkflowParameterTypeStatus

	channelsOnly := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelsOnly[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds

	channelFilterRequiredInputs := channelsOnly
	channelFilterRequiredOutputs := channelsOnly

	channelBalanceEventFilterRequiredInputs := channelsOnly
	channelBalanceEventFilterRequiredOutputs := channelsOnly

	channelPolicyConfiguratorOptionalInputs := all
	channelPolicyConfiguratorOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyConfiguratorOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	channelPolicyConfiguratorOptionalOutputs[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings

	channelPolicyAutoRunRequiredInputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyAutoRunRequiredInputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	channelPolicyAutoRunOptionalInputs := all
	channelPolicyAutoRunRequiredOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyAutoRunOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyAutoRunOptionalOutputs[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings
	channelPolicyAutoRunOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	channelPolicyAutoRunOptionalOutputs[WorkflowParameterLabelStatus] = WorkflowParameterTypeStatus

	channelPolicyRunRequiredInputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyRunRequiredInputs[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings
	channelPolicyRunOptionalInputs := channelPolicyAutoRunOptionalInputs
	channelPolicyRunRequiredOutputs := channelPolicyAutoRunRequiredOutputs
	channelPolicyRunOptionalOutputs := channelPolicyAutoRunOptionalOutputs

	rebalanceConfiguratorOptionalInputs := all
	rebalanceConfiguratorOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceConfiguratorOptionalOutputs[WorkflowParameterLabelRebalanceSettings] = WorkflowParameterTypeRebalanceSettings

	rebalanceAutoRunRequiredInputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceAutoRunRequiredInputs[WorkflowParameterLabelIncomingChannels] = WorkflowParameterTypeChannelIds
	rebalanceAutoRunRequiredInputs[WorkflowParameterLabelOutgoingChannels] = WorkflowParameterTypeChannelIds
	rebalanceAutoRunOptionalInputs := all
	rebalanceAutoRunRequiredOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceAutoRunOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceAutoRunOptionalOutputs[WorkflowParameterLabelRebalanceSettings] = WorkflowParameterTypeRebalanceSettings
	rebalanceAutoRunOptionalOutputs[WorkflowParameterLabelIncomingChannels] = WorkflowParameterTypeChannelIds
	rebalanceAutoRunOptionalOutputs[WorkflowParameterLabelOutgoingChannels] = WorkflowParameterTypeChannelIds
	rebalanceAutoRunOptionalOutputs[WorkflowParameterLabelStatus] = WorkflowParameterTypeStatus

	rebalanceRunRequiredInputs := rebalanceAutoRunRequiredInputs
	rebalanceRunOptionalInputs := rebalanceAutoRunOptionalInputs
	rebalanceRunRequiredOutputs := rebalanceAutoRunRequiredOutputs
	rebalanceRunOptionalOutputs := rebalanceAutoRunOptionalOutputs

	addTagOptionalInputs := channelsOnly
	addTagOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	addTagOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	addTagOptionalOutputs[WorkflowParameterLabelTagSettings] = WorkflowParameterTypeTagSettings

	removeTagOptionalInputs := channelsOnly
	removeTagOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	removeTagOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	removeTagOptionalOutputs[WorkflowParameterLabelTagSettings] = WorkflowParameterTypeTagSettings

	return map[WorkflowNodeType]WorkflowNodeTypeParameters{
		WorkflowTrigger: {
			WorkflowNodeType: WorkflowTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeManualTrigger: {
			WorkflowNodeType: WorkflowNodeManualTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeIntervalTrigger: {
			WorkflowNodeType: WorkflowNodeIntervalTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeCronTrigger: {
			WorkflowNodeType: WorkflowNodeCronTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelBalanceEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelOpenEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelCloseEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeStageTrigger: {
			WorkflowNodeType: WorkflowNodeStageTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeDataSourceTorqChannels: {
			WorkflowNodeType: WorkflowNodeDataSourceTorqChannels,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  channelsOnly,
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelFilter: {
			WorkflowNodeType: WorkflowNodeChannelFilter,
			RequiredInputs:   channelFilterRequiredInputs,
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  channelFilterRequiredOutputs,
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelBalanceEventFilter: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventFilter,
			RequiredInputs:   channelBalanceEventFilterRequiredInputs,
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  channelBalanceEventFilterRequiredOutputs,
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelPolicyConfigurator: {
			WorkflowNodeType: WorkflowNodeChannelPolicyConfigurator,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   channelPolicyConfiguratorOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  channelPolicyConfiguratorOptionalOutputs,
		},
		WorkflowNodeChannelPolicyAutoRun: {
			WorkflowNodeType: WorkflowNodeChannelPolicyAutoRun,
			RequiredInputs:   channelPolicyAutoRunRequiredInputs,
			OptionalInputs:   channelPolicyAutoRunOptionalInputs,
			RequiredOutputs:  channelPolicyAutoRunRequiredOutputs,
			OptionalOutputs:  channelPolicyAutoRunOptionalOutputs,
		},
		WorkflowNodeChannelPolicyRun: {
			WorkflowNodeType: WorkflowNodeChannelPolicyRun,
			RequiredInputs:   channelPolicyRunRequiredInputs,
			OptionalInputs:   channelPolicyRunOptionalInputs,
			RequiredOutputs:  channelPolicyRunRequiredOutputs,
			OptionalOutputs:  channelPolicyRunOptionalOutputs,
		},
		WorkflowNodeRebalanceConfigurator: {
			WorkflowNodeType: WorkflowNodeRebalanceConfigurator,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   rebalanceConfiguratorOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  rebalanceConfiguratorOptionalOutputs,
		},
		WorkflowNodeRebalanceAutoRun: {
			WorkflowNodeType: WorkflowNodeRebalanceAutoRun,
			RequiredInputs:   rebalanceAutoRunRequiredInputs,
			OptionalInputs:   rebalanceAutoRunOptionalInputs,
			RequiredOutputs:  rebalanceAutoRunRequiredOutputs,
			OptionalOutputs:  rebalanceAutoRunOptionalOutputs,
		},
		WorkflowNodeRebalanceRun: {
			WorkflowNodeType: WorkflowNodeRebalanceRun,
			RequiredInputs:   rebalanceRunRequiredInputs,
			OptionalInputs:   rebalanceRunOptionalInputs,
			RequiredOutputs:  rebalanceRunRequiredOutputs,
			OptionalOutputs:  rebalanceRunOptionalOutputs,
		},
		WorkflowNodeAddTag: {
			WorkflowNodeType: WorkflowNodeAddTag,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   addTagOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  addTagOptionalOutputs,
		},
		WorkflowNodeRemoveTag: {
			WorkflowNodeType: WorkflowNodeRemoveTag,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   removeTagOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  removeTagOptionalOutputs,
		},
		WorkflowNodeSetVariable: {
			WorkflowNodeType: WorkflowNodeSetVariable,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   all,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeFilterOnVariable: {
			WorkflowNodeType: WorkflowNodeFilterOnVariable,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   all,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
	}
}
