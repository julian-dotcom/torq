import { torqApi } from "apiSlice";
// import { GetOnChainTransactionsQueryParams } from "types/api";
// import { queryParamsBuilder } from "utils/queryParamsBuilder";
import { FullWorkflow, workflowListItem } from "./workflowTypes";

// Define a service using a base URL and expected endpoints
export const workflowApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getWorkflows: builder.query<Array<workflowListItem>, void>({
      query: (params) => "workflows",
      providesTags: ["workflows"],
    }),
    getWorkflow: builder.query<FullWorkflow, { version: number; workflowId: number }>({
      query: (params) => `workflows/${params.workflowId}/versions/${params.version}`,
    }),
    newWorkflow: builder.mutation<{ workflowId: number; version: number }, void>({
      query: (body: void) => ({
        url: "workflows",
        method: "POST",
        body: {},
      }),
      invalidatesTags: ["workflows"],
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const { useGetWorkflowsQuery, useGetWorkflowQuery, useNewWorkflowMutation } = workflowApi;
