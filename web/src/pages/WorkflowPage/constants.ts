export enum WorkflowNodeType {
  TimeTrigger,
  ChannelBalanceEventTrigger,
  ChannelFilter,
  ChannelPolicyConfigurator,
  RebalanceParameters,
  StageTrigger,
  RebalanceRun,
  RoutingPolicyRun,
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
}

export const TriggerNodeTypes = [
  WorkflowNodeType.TimeTrigger,
  WorkflowNodeType.ChannelBalanceEventTrigger,
  WorkflowNodeType.StageTrigger,
  WorkflowNodeType.ChannelOpenEventTrigger,
  WorkflowNodeType.ChannelCloseEventTrigger,
  WorkflowNodeType.CronTrigger,
];
