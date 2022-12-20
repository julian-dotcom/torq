import { torqApi } from "apiSlice";
// import { GetOnChainTransactionsQueryParams } from "types/api";
// import { queryParamsBuilder } from "utils/queryParamsBuilder";
import { workflowListItem } from "./workflowTypes";

// Define a service using a base URL and expected endpoints
export const workflowApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getWorkflows: builder.query<Array<workflowListItem>, void>({
      query: (params) => "workflows",
    }),
    getWorkflow: builder.query<any, { workflowVersionId: number }>({
      query: (params) => `workflows/${params.workflowVersionId}`,
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const { useGetWorkflowsQuery, useGetWorkflowQuery } = workflowApi;
