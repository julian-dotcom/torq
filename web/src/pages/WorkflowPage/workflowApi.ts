import { torqApi } from "apiSlice";
import { FullWorkflow, NewWorkflowNodeRequest, UpdateWorkflowNodeRequest, workflowListItem } from "./workflowTypes";

// Define a service using a base URL and expected endpoints
export const workflowApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getWorkflows: builder.query<Array<workflowListItem>, void>({
      query: (params) => "workflows",
      providesTags: ["workflows"],
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
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const {
  useGetWorkflowsQuery,
  useGetWorkflowQuery,
  useNewWorkflowMutation,
  useAddNodeMutation,
  useUpdateNodeMutation,
  useDeleteNodeMutation,
} = workflowApi;
