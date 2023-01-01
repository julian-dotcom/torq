import { torqApi } from "apiSlice";
import {
  CreateWorkflowVersionNodeLink,
  FullWorkflow,
  NewWorkflowNodeRequest,
  UpdateWorkflow,
  UpdateWorkflowNodeRequest,
  workflowListItem,
} from "./workflowTypes";
import { createSelector } from "@reduxjs/toolkit";

// Define a service using a base URL and expected endpoints
export const workflowApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getWorkflows: builder.query<Array<workflowListItem>, void>({
      query: (params) => "workflows",
    }),
    getWorkflow: builder.query<FullWorkflow, { version: number; workflowId: number }>({
      query: (params) => `workflows/${params.workflowId}/versions/${params.version}`,
      providesTags: ["workflow"],
    }),
    newWorkflow: builder.mutation<{ workflowId: number; version: number }, void>({
      query: (body: void) => ({
        url: "workflows",
        method: "POST",
        body: {},
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
      invalidatesTags: ["workflows"],
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
  useDeleteNodeMutation,
  useDeleteStageMutation,
  useAddNodeLinkMutation,
} = workflowApi;

type State = {
  fullWorkflow: FullWorkflow;
};

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

// Create a selector that get a specific workflow node from the workflow in the store
export const SelectWorkflowNodes = (props: SelectWorkflowNode) => {
  return createSelector(
    [workflowApi.endpoints.getWorkflow.select({ version: props.version, workflowId: props.workflowId })],
    (workflow) => {
      return workflow?.data?.nodes.filter((node) => props.nodeIds.includes(node.workflowVersionNodeId));
    }
  );
};

export type selectWorkflowNodeLinks = { version: number; workflowId: number; nodeId: number; childLinks: boolean };

// Create a selector that get specific workflow node links from the workflow in the store
export const SelectWorkflowNodeLinks = (props: selectWorkflowNodeLinks) => {
  return createSelector(
    [workflowApi.endpoints.getWorkflow.select({ version: props.version, workflowId: props.workflowId })],
    (workflow) => {
      const linkType = props.childLinks ? "childWorkflowVersionNodeId" : "parentWorkflowVersionNodeId";
      return (workflow?.data?.links || []).filter((link) => link[linkType] === props.nodeId);
    }
  );
};
