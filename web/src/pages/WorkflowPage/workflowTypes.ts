import { Status } from "constants/backend";

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

export type LinkVisibilitySettings = {
  child: { xPosition: number; yPosition: number };
  parent: { xPosition: number; yPosition: number };
};

export type WorkflowNode = {
  LinkDetails: { [k: string | number | symbol]: unknown };
  childNodes: { [k: string | number | symbol]: unknown };
  name: string;
  parameters: { [k: string | number | symbol]: unknown };
  parentNodes: { [k: string | number | symbol]: unknown };
  status: number;
  type: number;
  updatedOn: string;
  visibilitySettings: VisibilitySettings;
  workflowVersionId: number;
  workflowVersionNodeId: number;
};

export type WorkflowVersionNode = {
  name: string;
  stage: number;
  parameters: { [k: string | number | symbol]: unknown };
  status: number;
  type: number;
  updatedOn: string;
  visibilitySettings: VisibilitySettings;
  workflowId: number;
  version: number;
  workflowVersionId: number;
  workflowVersionNodeId: number;
};

export type NewWorkflowNodeRequest = {
  type: number;
  name: string;
  visibilitySettings: VisibilitySettings;
  workflowVersionId: number;
  stage: number;
};

export type UpdateWorkflowNodeRequest = Partial<{
  workflowVersionNodeId: number;
  name: string;
  status: number;
  visibilitySettings?: VisibilitySettings;
  parameters?: { [k: string | number | symbol]: unknown };
}>;

export type WorkflowVersionNodeLink = {
  stage: number;
  workflowVersionId: number;
  workflowVersionNodeLinkId: number;
  name: string;
  visibilitySettings: LinkVisibilitySettings;
  parentOutputIndex: number;
  parentWorkflowVersionNodeId: number;
  childInputIndex: number;
  childWorkflowVersionNodeId: number;
  createdOn: Date;
  updatedOn: Date;
};

export type CreateWorkflowVersionNodeLink = {
  workflowVersionId: number;
  parentOutputIndex: number;
  parentWorkflowVersionNodeId: number;
  childInputIndex: number;
  childWorkflowVersionNodeId: number;
  visibilitySettings: LinkVisibilitySettings;
};

export type UpdateWorkflowVersionNodeLink = {
  workflowVersionNodeLinkId: number;
  parentOutputIndex: number;
  parentWorkflowVersionNodeId: number;
  childInputIndex: number;
  childWorkflowVersionNodeId: number;
  visibilitySettings: LinkVisibilitySettings;
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
  [key: number]: Array<WorkflowVersionNode>;
};

export type WorkflowVersionNodeLinks = {
  [key: number]: WorkflowVersionNodeLink;
};

export type WorkflowForest = {
  sortedStageTrees: WorkflowStages;
};

export type FullWorkflow = {
  workflow: Workflow;
  version: WorkflowVersion;
  nodes: Array<WorkflowVersionNode>;
  links: Array<WorkflowVersionNodeLink>;
};

export type UpdateWorkflow = {
  workflowId: number;
  name?: string;
  status?: Status;
};
