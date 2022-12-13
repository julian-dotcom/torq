import { torqApi } from "apiSlice";
import {
  AllViewsResponse,
  CreateViewRequest,
  TableResponses,
  UpdateViewRequest,
  ViewOrderInterface,
  ViewResponse,
} from "features/viewManagement/types";

// Define a service using a base URL and expected endpoints
export const viewApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getTableViews: builder.query<AllViewsResponse, void>({
      query: () => "table-views",
      providesTags: ["tableView"],
    }),
    createTableView: builder.mutation<ViewResponse<TableResponses>, CreateViewRequest>({
      query: (body: CreateViewRequest) => ({
        url: "table-views",
        method: "POST",
        body: body,
      }),
    }),
    updateTableView: builder.mutation<ViewResponse<TableResponses>, UpdateViewRequest>({
      query: (body: UpdateViewRequest) => ({
        url: "table-views",
        method: "PUT",
        body: body,
      }),
    }),
    deleteTableView: builder.mutation<AllViewsResponse, { page: keyof AllViewsResponse; id: number }>({
      query: (data) => ({
        url: `table-views/${data.id}`,
        method: "DELETE",
      }),
    }),
    updateTableViewsOrder: builder.mutation<undefined, Array<ViewOrderInterface>>({
      query: (order: Array<ViewOrderInterface>) => ({
        url: "table-views/order",
        method: "PATCH",
        body: order,
      }),
      invalidatesTags: ["tableView"],
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const {
  useGetTableViewsQuery,
  useUpdateTableViewMutation,
  useCreateTableViewMutation,
  useDeleteTableViewMutation,
  useUpdateTableViewsOrderMutation,
} = viewApi;
