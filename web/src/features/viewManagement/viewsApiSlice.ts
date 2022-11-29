import { torqApi } from "apiSlice";
import {
  AllViewsResponse,
  CreateViewRequest,
  UpdateViewRequest,
  ViewOrderInterface,
} from "features/viewManagement/types";

// Define a service using a base URL and expected endpoints
export const viewApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getTableViews: builder.query<AllViewsResponse, void>({
      query: () => "table-views",
      providesTags: ["tableView"],
    }),
    createTableView: builder.mutation<AllViewsResponse, CreateViewRequest>({
      query: (body: CreateViewRequest) => ({
        url: "table-views",
        method: "POST",
        body: body,
      }),
      // invalidatesTags: ["tableView"],
    }),
    updateTableView: builder.mutation<AllViewsResponse, UpdateViewRequest>({
      query: (body: UpdateViewRequest) => ({
        url: "table-views",
        method: "PUT",
        body: body,
      }),
      // invalidatesTags: ["tableView"],
    }),
    deleteTableView: builder.mutation<AllViewsResponse, { id: number }>({
      query: (data) => ({
        url: `table-views/${data.id}`,
        method: "DELETE",
      }),
      // invalidatesTags: ["tableView"],
    }),
    updateTableViewsOrder: builder.mutation<AllViewsResponse, Array<ViewOrderInterface>>({
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
