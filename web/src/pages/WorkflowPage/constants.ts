export enum WorkflowNodeType {
  TimeTrigger,
  ChannelBalanceEventTrigger,
  ChannelFilter,
  RoutingPolicyParameters,
  RebalanceParameters,
  StageTrigger,
  RebalanceRun,
  RoutingPolicyRun,
  SetVariable,
  FilterOnVariable,
}

export const TriggerNodeTypes = [
  WorkflowNodeType.TimeTrigger,
  WorkflowNodeType.ChannelBalanceEventTrigger,
  WorkflowNodeType.StageTrigger,
];
