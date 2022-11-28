import { torqApi } from "apiSlice";
import {
  AllViewsResponse,
  TableResponses,
  ViewInterface,
  ViewInterfaceResponse,
  ViewOrderInterface,
} from "features/viewManagement/types";

// Define a service using a base URL and expected endpoints
export const viewApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getTableViews: builder.query<AllViewsResponse, void>({
      query: () => "table-views",
      providesTags: ["tableView"],
    }),
    createTableView: builder.mutation<AllViewsResponse, ViewInterfaceResponse>({
      query: (data: ViewInterfaceResponse) => ({
        url: "table-views",
        method: "POST",
        body: data,
      }),
      invalidatesTags: ["tableView"],
    }),
    updateTableView: builder.mutation<AllViewsResponse, ViewInterface<TableResponses>>({
      query: (view: ViewInterfaceResponse) => ({
        url: "table-views",
        method: "PUT",
        body: { id: view.id, view: view },
      }),
      invalidatesTags: ["tableView"],
    }),
    deleteTableView: builder.mutation<AllViewsResponse, { id: number }>({
      query: (data) => ({
        url: `table-views/${data.id}`,
        method: "DELETE",
      }),
      invalidatesTags: ["tableView"],
    }),
    updateTableViewsOrder: builder.mutation<AllViewsResponse, Array<ViewOrderInterface>>({
      query: (order: ViewOrderInterface[]) => ({
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
