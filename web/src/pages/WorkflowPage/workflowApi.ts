import { torqApi } from "apiSlice";
import {
  CreateWorkflowVersionNodeLink,
  FullWorkflow,
  NewWorkflowNodeRequest,
  UpdateWorkflow,
  UpdateWorkflowNodeRequest,
  workflowListItem,
  UpdateWorkflowNodeVisibilitySettingsRequest,
  TriggerWorkflowRequest,
} from "./workflowTypes";
import { createSelector } from "@reduxjs/toolkit";
import { TriggerNodeTypes, WorkflowNodeType } from "./constants";

// Define a service using a base URL and expected endpoints
export const workflowApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getWorkflows: builder.query<Array<workflowListItem>, void>({
      query: () => "workflows",
    }),
    getWorkflow: builder.query<FullWorkflow, { version: number; workflowId: number }>({
      query: (params) => `workflows/${params.workflowId}/versions/${params.version}`,
      providesTags: ["workflow"],
    }),
    newWorkflow: builder.mutation<{ workflowId: number; version: number }, void>({
      query: () => ({
        url: "workflows",
        method: "POST",
        body: {},
      }),
      invalidatesTags: ["workflows"],
    }),
    addManualWorkflowTrigger: builder.mutation<void, TriggerWorkflowRequest>({
      query: (body: TriggerWorkflowRequest) => ({
        url: "workflows/trigger",
        method: "POST",
        body,
      }),
      invalidatesTags: ["workflows"],
    }),
    updateWorkflow: builder.mutation<void, UpdateWorkflow>({
      query: (body: UpdateWorkflow) => ({
        url: `workflows`,
        method: "PUT",
        body: body,
      }),
      invalidatesTags: ["workflow", "workflows"],
    }),
    // deleteWorflow is currently not used
    deleteWorkflow: builder.mutation<void, { workflowId: number }>({
      query: (params) => ({
        url: `workflows/${params.workflowId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["workflows", "workflow"],
    }),
    addNode: builder.mutation<void, NewWorkflowNodeRequest>({
      query: (body: NewWorkflowNodeRequest) => ({
        url: `workflows/nodes`,
        method: "POST",
        body: body,
      }),
      invalidatesTags: ["workflow"],
    }),
    updateNode: builder.mutation<void, UpdateWorkflowNodeRequest>({
      query: (body: UpdateWorkflowNodeRequest) => ({
        url: `workflows/nodes`,
        method: "PUT",
        body: body,
      }),
      invalidatesTags: ["workflow"],
    }),
    updateNodeVisibilitySettings: builder.mutation<void, UpdateWorkflowNodeVisibilitySettingsRequest>({
      query: (body: UpdateWorkflowNodeVisibilitySettingsRequest) => ({
        url: `workflows/nodes/${body.workflowVersionNodeId}/visibility-settings`,
        method: "PUT",
        body: body.visibilitySettings,
      }),
    }),
    deleteNode: builder.mutation<void, { nodeId: number }>({
      query: (body: { nodeId: number }) => ({
        url: `workflows/nodes/${body.nodeId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["workflow"],
    }),
    deleteStage: builder.mutation<void, { workflowId: number; version: number; stage: number }>({
      query: (params) => ({
        url: `workflows/${params.workflowId}/versions/${params.version}/stage/${params.stage}`,
        method: "DELETE",
      }),
      invalidatesTags: ["workflow"],
    }),
    addNodeLink: builder.mutation<void, CreateWorkflowVersionNodeLink>({
      query: (body) => ({
        url: `workflows/links`,
        method: "POST",
        body: body,
      }),
      invalidatesTags: ["workflow"],
    }),
    updateNodeLink: builder.mutation<void, CreateWorkflowVersionNodeLink>({
      query: (body) => ({
        url: `workflows/links`,
        method: "PUT",
        body: body,
      }),
      invalidatesTags: ["workflow"],
    }),
    deleteNodeLink: builder.mutation<void, { linkId: number }>({
      query: (body) => ({
        url: `workflows/links/${body.linkId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["workflow"],
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const {
  useGetWorkflowsQuery,
  useGetWorkflowQuery,
  useNewWorkflowMutation,
  useDeleteWorkflowMutation,
  useUpdateWorkflowMutation,
  useAddNodeMutation,
  useUpdateNodeMutation,
  useUpdateNodeVisibilitySettingsMutation,
  useDeleteNodeMutation,
  useDeleteStageMutation,
  useAddNodeLinkMutation,
  useDeleteNodeLinkMutation,
  useAddManualWorkflowTriggerMutation,
} = workflowApi;

// // Select the FullWorkflow object from the state
// const selectFullWorkflow = (state: State) => state.fullWorkflow;
//
// // Create a selector that filters the links array within FullWorkflow by the specific workflow version node
// export const selectLinksForWorkflowVersionNode = createSelector(
//   [selectFullWorkflow],
//   (fullWorkflow: FullWorkflow) => (workflowVersionNodeId: number) =>
//     fullWorkflow.links.filter(
//       (link: WorkflowVersionNodeLink) => link.childWorkflowVersionNodeId === workflowVersionNodeId
//     )
// );

type SelectWorkflowNode = { version: number; workflowId: number; nodeIds: Array<number> };

// Get specific workflow nodes from the workflow in the store
export const SelectWorkflowNodes = (props: SelectWorkflowNode) => {
  return createSelector(
    [workflowApi.endpoints.getWorkflow.select({ version: props.version, workflowId: props.workflowId })],
    (workflow) => {
      return workflow?.data?.nodes.filter((node) => props.nodeIds.includes(node.workflowVersionNodeId));
    }
  );
};

type SelectWorkflowLinksType = { version: number; workflowId: number; stage: number };

// Get all links
export const SelectWorkflowLinks = (props: SelectWorkflowLinksType) => {
  return createSelector(
    [workflowApi.endpoints.getWorkflow.select({ version: props.version, workflowId: props.workflowId })],
    (workflow) => {
      return (workflow?.data?.links || []).filter((link) => link.stage === props.stage) || [];
    }
  );
};

export type selectWorkflowNodeLinks = { version: number; workflowId: number; nodeId: number; stage: number };

// Get specific workflow node links from the workflow in the store
export const SelectWorkflowNodeLinks = (props: selectWorkflowNodeLinks) => {
  return createSelector(
    [workflowApi.endpoints.getWorkflow.select({ version: props.version, workflowId: props.workflowId })],
    (workflow) => {
      const parentLinks = (workflow?.data?.links || []).filter((link) => {
        if (link.parentWorkflowVersionNodeId === props.nodeId && link.stage === props.stage) {
          return true;
        }
      });
      const childLinks = (workflow?.data?.links || []).filter(
        (link) => link.childWorkflowVersionNodeId === props.nodeId && link.stage === props.stage
      );
      return { parentLinks, childLinks };
    }
  );
};

function getStageNodes(workflow: FullWorkflow, stage: number) {
  const stageNodes = workflow.nodes.filter((node) => node.stage === stage);
  const mainTriggerNode = stageNodes.find((node) => node.type === WorkflowNodeType.Trigger);
  const triggers = stageNodes?.filter((node) => TriggerNodeTypes.includes(node.type));
  const actions = stageNodes?.filter(
    (node) => !TriggerNodeTypes.includes(node.type) && node.type !== WorkflowNodeType.Trigger
  );
  return { triggers, actions, mainTriggerNode };
}

// Select WorkflowNodeType.Trigger node from the workflow
export const SelectWorkflowMainTriggerNode = (props: { version: number; workflowId: number }) => {
  return createSelector(
    [workflowApi.endpoints.getWorkflow.select({ version: props.version, workflowId: props.workflowId })],
    (workflow) => {
      return workflow?.data?.nodes.find((node) => node.type === WorkflowNodeType.Trigger);
    }
  );
};

// Get nodes belonging to a stage divided into an array of trigger nodes and an array of action nodes
export const SelectWorkflowStageNodes = (props: { version: number; workflowId: number; stage: number }) => {
  return createSelector(
    [workflowApi.endpoints.getWorkflow.select({ version: props.version, workflowId: props.workflowId })],
    (workflow) => {
      return getStageNodes(workflow?.data || ({} as FullWorkflow), props.stage);
    }
  );
};

// Get trigger and action nodes per stage for a workflow
export const SelectWorkflowStages = (props: { version: number; workflowId: number }) => {
  return createSelector(
    [workflowApi.endpoints.getWorkflow.select({ version: props.version, workflowId: props.workflowId })],
    (workflow) => {
      const stages = workflow?.data?.nodes?.reduce((acc: Array<number>, node) => {
        if (!acc.includes(node.stage)) {
          acc.push(node.stage);
        }
        return acc;
      }, []);
      return (stages || [])?.map((stage) => {
        return { ...getStageNodes(workflow?.data || ({} as FullWorkflow), stage), stage: stage };
      });
    }
  );
};
