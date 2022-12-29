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

const triggerNodeTypes = [
  WorkflowNodeType.TimeTrigger,
  WorkflowNodeType.ChannelBalanceEventTrigger,
  WorkflowNodeType.StageTrigger,
];
