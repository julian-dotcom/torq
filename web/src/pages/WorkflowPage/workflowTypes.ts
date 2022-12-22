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

export type workflow = {
  LinkDetails: {};
  childNodes: {};
  name: string;
  parameters: {};
  parentNodes: {};
  status: number;
  type: number;
  updatedOn: string;
  visibilitySettings: number;
  workflowVersionId: number;
  workflowVersionNodeId: number;
};
