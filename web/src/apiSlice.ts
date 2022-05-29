import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { ViewInterface, viewOrderInterface } from "features/table/tableSlice";
import { settings, timeZone, localNode } from "./apiTypes";

const API_URL =
  window.location.port === "3000"
    ? "//" + window.location.hostname + ":8080/api"
    : "//" + window.location.host + "/api";

// Define a service using a base URL and expected endpoints
export const torqApi = createApi({
  reducerPath: "api",
  baseQuery: fetchBaseQuery({
    baseUrl: API_URL,
    // prepareHeaders: (headers, _) => {
    //   if (!headers.get("Content-Type")) {
    //     headers.set("Content-Type", "application/json");
    //   }
    //   return headers;
    // },
    credentials: "include",
    mode: "cors",
  }),
  tagTypes: ["settings", "tableView", "localNode"],
  endpoints: (builder) => ({
    getFlow: builder.query<any, { from: string; to: string; chanId: string }>({
      query: ({ from, to, chanId }) => `flow?from=${from}&to=${to}&chan_id=${chanId}`,
    }),
    getChannelHistory: builder.query<any, { from: string; to: string; chanIds: string }>({
      query: ({ from, to, chanIds }) => `channels/${chanIds}?from=${from}&to=${to}`,
    }),
    getChannels: builder.query<any, { from: string; to: string }>({
      query: ({ from, to }) => `channels?from=${from}&to=${to}`,
    }),
    getTableViews: builder.query<any, void>({
      query: () => `table-views`,
      providesTags: ["tableView"],
    }),
    createTableView: builder.mutation<any, { view: ViewInterface; index: number }>({
      query: (data) => ({
        url: "table-views",
        method: "POST",
        body: { id: null, view: data.view },
      }),
      transformResponse: (response: { view: ViewInterface }, _, arg) => ({
        view: response,
        index: arg.index,
      }),
    }),
    updateTableView: builder.mutation<any, ViewInterface>({
      query: (view: ViewInterface) => ({
        url: "table-views",
        method: "PUT",
        body: { id: view.id, view: view },
      }),
    }),
    deleteTableView: builder.mutation<any, { view: ViewInterface; index: number }>({
      query: (data) => ({
        url: `table-views/${data.view.id}`,
        method: "DELETE",
      }),
      transformResponse: (_, __, arg) => ({ index: arg.index }),
    }),
    updateTableViewsOrder: builder.mutation<any, viewOrderInterface[]>({
      query: (order: viewOrderInterface[]) => ({
        url: "table-views/order",
        method: "PATCH",
        body: order,
      }),
    }),
    logout: builder.mutation<any, void>({
      query: () => ({
        url: "logout",
        method: "POST",
      }),
      invalidatesTags: ["tableView"],
    }),
    login: builder.mutation<any, FormData>({
      query: (form) => ({
        url: "login",
        method: "POST",
        body: new URLSearchParams(form as any),
        // headers: { "Content-Type": "application/x-www-form-urlencoded" },
      }),
    }),
    getSettings: builder.query<settings, void>({
      query: () => `settings`,
      providesTags: ["settings"],
    }),
    updateSettings: builder.mutation<any, settings>({
      query: (settings) => ({
        url: "settings",
        method: "PUT",
        body: settings,
      }),
      invalidatesTags: ["settings"],
    }),
    getTimeZones: builder.query<timeZone[], void>({
      query: () => `settings/timezones`,
    }),
    getLocalNode: builder.query<localNode, void>({
      query: () => `settings/local-node`,
      providesTags: ["localNode"],
    }),
    updateLocalNode: builder.mutation<any, FormData>({
      query: (localNode) => ({
        url: "settings/local-node",
        method: "PUT",
        body: localNode,
      }),
      invalidatesTags: ["localNode"],
    }),
  }),
});

// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const {
  useGetFlowQuery,
  useGetChannelHistoryQuery,
  useGetChannelsQuery,
  useGetTableViewsQuery,
  useUpdateTableViewMutation,
  useCreateTableViewMutation,
  useDeleteTableViewMutation,
  useUpdateTableViewsOrderMutation,
  useLoginMutation,
  useLogoutMutation,
  useGetSettingsQuery,
  useUpdateSettingsMutation,
  useGetTimeZonesQuery,
  useGetLocalNodeQuery,
  useUpdateLocalNodeMutation,
} = torqApi;
