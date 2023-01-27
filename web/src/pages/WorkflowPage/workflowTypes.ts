import { Status } from "constants/backend";
import { AnyObject } from "utils/types";

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
  parameters: AnyObject;
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
  parameters?: AnyObject;
};

export type UpdateWorkflowNodeRequest = {
  workflowVersionNodeId: number;
  name?: string;
  status?: number;
  parameters?: AnyObject;
};

export type UpdateWorkflowNodeVisibilitySettingsRequest = {
  workflowVersionNodeId: number;
  visibilitySettings: VisibilitySettings;
};

export type WorkflowVersionNodeLink = {
  stage: number;
  workflowVersionId: number;
  workflowVersionNodeLinkId: number;
  name: string;
  visibilitySettings?: LinkVisibilitySettings;
  parentOutput: string;
  parentWorkflowVersionNodeId: number;
  childInput: string;
  childWorkflowVersionNodeId: number;
  createdOn: Date;
  updatedOn: Date;
};

export type CreateWorkflowVersionNodeLink = {
  workflowVersionId: number;
  parentOutput: string;
  parentWorkflowVersionNodeId: number;
  childInput: string;
  childWorkflowVersionNodeId: number;
};

export type UpdateWorkflowVersionNodeLink = {
  workflowVersionNodeLinkId: number;
  parentOutput: string;
  parentWorkflowVersionNodeId: number;
  childInput: string;
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

export type TriggerWorkflowRequest = {
  type: number;
  workflowVersionId: number;
  workflowId: number;
  workflowVersionNodeId: number;
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
