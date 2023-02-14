export enum WorkflowNodeType {
  IntervalTrigger,
  ChannelBalanceEventTrigger,
  ChannelFilter,
  ChannelPolicyConfigurator,
  RebalanceConfigurator,
  StageTrigger,
  RebalanceRun,
  ChannelPolicyRun,
  SetVariable,
  FilterOnVariable,
  // Trigger is the main trigger that is used to start the workflow prevent us
  // from having to add and remove links to every trigger.
  Trigger,
  AddTag,
  RemoveTag,
  ChannelOpenEventTrigger,
  ChannelCloseEventTrigger,
  CronTrigger,
  ManualTrigger,
  ChannelPolicyAutoRun,
  RebalanceAutoRun,
  DataSourceAllChannels,
  DataSourceEventChannels
}

export const TriggerNodeTypes = [
  WorkflowNodeType.IntervalTrigger,
  WorkflowNodeType.ChannelBalanceEventTrigger,
  WorkflowNodeType.StageTrigger,
  WorkflowNodeType.ChannelOpenEventTrigger,
  WorkflowNodeType.ChannelCloseEventTrigger,
  WorkflowNodeType.CronTrigger,
];
