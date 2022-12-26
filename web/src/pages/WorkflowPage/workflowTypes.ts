export type workflowListItem = {
  workflowId: number;
  workflowName: string;
  workflowStatus: number;
  latestVersionName: string;
  latestVersion: number;
  latestWorkflowVersionId: number;
  latestVersionStatus: number;
  activeVersionName: string;
  activeVersion: number;
  activeWorkflowVersionId: number;
  activeVersionStatus: number;
};

export type VisibilitySettings = {
  collapsed: boolean;
  xPosition: number;
  yPosition: number;
};

export type WorkflowNode = {
  LinkDetails: {};
  childNodes: {};
  name: string;
  parameters: {};
  parentNodes: {};
  status: number;
  type: number;
  updatedOn: string;
  visibilitySettings: VisibilitySettings;
  workflowVersionId: number;
  workflowVersionNodeId: number;
};

export type NewWorkflowNodeRequest = {
  type: number;
  visibilitySettings: VisibilitySettings;
  workflowVersionId: number;
  stage: number;
};

export type WorkflowVersion = {
  workflowVersionId: number;
  name: string;
  version: number;
  status: number;
  workflowId: number;
  createdOn: Date;
  updatedOn: Date;
};

export type Workflow = {
  workflowId: number;
  name: string;
  status: number;
  createdOn: Date;
  updatedOn: Date;
};

export type WorkflowStages = {
  [key: number]: Array<WorkflowNode>;
};

export type WorkflowForest = {
  sortedStageTrees: WorkflowStages;
};

export type FullWorkflow = {
  workflow: Workflow;
  version: WorkflowVersion;
  workflowForest: WorkflowForest;
};
