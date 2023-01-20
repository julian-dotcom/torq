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
}

export const TriggerNodeTypes = [
  WorkflowNodeType.TimeTrigger,
  WorkflowNodeType.ChannelBalanceEventTrigger,
  WorkflowNodeType.StageTrigger,
];
