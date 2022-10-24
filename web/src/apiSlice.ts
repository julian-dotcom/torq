import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { ViewInterface, viewOrderInterface } from "features/forwards/forwardsSlice";
import { PolicyInterface } from "features/channels/ChannelsSlice"
import { getRestEndpoint, getWsEndpoint } from "utils/apiUrlBuilder";
import { UpdatedChannelResponse } from "features/channels/channelsTypes"
import {
  ChannelOnchainCostResponse,
  ChannelHistoryResponse,
  ChannelRebalancingResponse,
  ChannelBalanceResponse,
  ChannelEventResponse,
} from "features/channel/channelTypes"

import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from "@reduxjs/toolkit/query";
import type {
  GetChannelHistoryQueryParams,
  GetDecodedInvoiceQueryParams,
  GetForwardsQueryParams,
  GetInvoicesQueryParams,
  GetOnChainTransactionsQueryParams,
  GetPaymentsQueryParams,
  SendOnChainRequest,
  SendOnChainResponse,
} from "types/api";
import { queryParamsBuilder } from "utils/queryParamsBuilder";
import type { localNode, settings, timeZone, channel } from "./apiTypes";
import { NewInvoiceRequest, NewInvoiceResponse } from "./features/transact/Invoices/newInvoice/newInvoiceTypes";

const API_URL = getRestEndpoint();
export const WS_URL = getWsEndpoint();

const baseQuery = fetchBaseQuery({
  baseUrl: API_URL,
  // prepareHeaders: (headers, _) => {
  //   if (!headers.get("Content-Type")) {
  //     headers.set("Content-Type", "application/json");
  //   }
  //   return headers;
  // },
  credentials: "include",
  mode: "cors",
});

const baseQueryWithRedirect: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError> = async (
  args,
  api,
  extraOptions
) => {
  const result = await baseQuery(args, api, extraOptions);
  if (result.error && result.error.status === 401) {
    window.location.href = "/login";
  }
  return result;
};

// Define a service using a base URL and expected endpoints
export const torqApi = createApi({
  reducerPath: "api",
  baseQuery: baseQueryWithRedirect,
  tagTypes: ["settings", "tableView", "localNodes", "channels"],
  endpoints: (builder) => ({
    getFlow: builder.query<any, GetChannelHistoryQueryParams>({
      query: (params) => queryParamsBuilder("flow", params),
    }),
    getChannelHistory: builder.query<ChannelHistoryResponse, GetChannelHistoryQueryParams>({
      query: (params) => queryParamsBuilder({ endpoint: "channels", baseParam: "chanIds", suffixEndpoint: "history" }, params),
    }),
    getChannelEvent: builder.query<ChannelEventResponse, GetChannelHistoryQueryParams>({
      query: (params) => queryParamsBuilder({ endpoint: "channels", baseParam: "chanIds", suffixEndpoint: "event" }, params),
    }),
    getChannelBalance: builder.query<ChannelBalanceResponse, GetChannelHistoryQueryParams>({
      query: (params) => queryParamsBuilder({ endpoint: "channels", baseParam: "chanIds", suffixEndpoint: "balance" }, params),
    }),
    getChannelRebalancing: builder.query<ChannelRebalancingResponse, GetChannelHistoryQueryParams>({
      query: (params) => queryParamsBuilder({ endpoint: "channels", baseParam: "chanIds", suffixEndpoint: "rebalancing" }, params),
    }),
    getChannelOnChainCost: builder.query<ChannelOnchainCostResponse, GetChannelHistoryQueryParams>({
      query: (params) => queryParamsBuilder({ endpoint: "channels", baseParam: "chanIds", suffixEndpoint: "onchaincost" }, params),
    }),
    getForwards: builder.query<any, GetForwardsQueryParams>({
      query: (params) => queryParamsBuilder("forwards", params, true),
    }),
    getChannels: builder.query<channel[], void>({
      query: () => ({
        url: `channels`,
        method: "GET",
      }),
      providesTags: ["channels"],
    }),
    updateChannel: builder.mutation<UpdatedChannelResponse, PolicyInterface>({
      query: (data: PolicyInterface) => ({
        url: "channels/update",
        method: "PUT",
        body: data,
      }),
      invalidatesTags: ["channels"],
    }),
    getDecodedInvoice: builder.query<any, GetDecodedInvoiceQueryParams>({
      query: (params) => queryParamsBuilder("invoices/decode/", params),
    }),
    getPayments: builder.query<any, GetPaymentsQueryParams>({
      query: (params) => queryParamsBuilder("payments", params, true),
    }),
    getInvoices: builder.query<any, GetInvoicesQueryParams>({
      query: (params) => queryParamsBuilder("invoices", params, true),
    }),
    getOnChainTx: builder.query<any, GetOnChainTransactionsQueryParams>({
      query: (params) => queryParamsBuilder("on-chain-tx", params, true),
    }),
    sendOnChain: builder.mutation<SendOnChainResponse, SendOnChainRequest>({
      query: (data: SendOnChainRequest) => ({
        url: "on-chain-tx/sendcoins",
        method: "POST",
        body: data,
      }),
    }),
    newInvoice: builder.mutation<NewInvoiceResponse, NewInvoiceRequest>({
      query: (data: NewInvoiceRequest) => ({
        url: "invoices/newinvoice",
        method: "POST",
        body: data,
      }),
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
    getLocalNodes: builder.query<localNode[], void>({
      query: () => `settings/local-nodes`,
      providesTags: ["localNodes"],
    }),
    getLocalNode: builder.query<localNode, number>({
      query: (localNodeId) => `settings/local-nodes/${localNodeId}`,
      providesTags: ["localNodes"],
    }),
    addLocalNode: builder.mutation<any, FormData>({
      query: (localNode) => ({
        url: "settings/local-nodes",
        method: "POST",
        body: localNode,
      }),
      invalidatesTags: ["localNodes","channels"],
    }),
    updateLocalNode: builder.mutation<any, { form: FormData; localNodeId: number }>({
      query: (localNode) => ({
        url: `settings/local-nodes/${localNode.localNodeId}`,
        method: "PUT",
        body: localNode.form,
      }),
      invalidatesTags: ["localNodes","channels"],
    }),
    updateLocalNodeSetDisabled: builder.mutation<any, { localNodeId: number; disabled: boolean }>({
      query: (localNode) => ({
        url: `settings/local-nodes/${localNode.localNodeId}/set-disabled`,
        method: "PUT",
        body: localNode,
      }),
      invalidatesTags: ["localNodes","channels"],
    }),
    updateLocalNodeSetDeleted: builder.mutation<any, { localNodeId: number }>({
      query: (localNode) => ({
        url: `settings/local-nodes/${localNode.localNodeId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["localNodes","channels"],
    }),
  }),
});

// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const {
  useGetFlowQuery,
  useGetChannelHistoryQuery,
  useGetChannelEventQuery,
  useGetChannelBalanceQuery,
  useGetChannelRebalancingQuery,
  useGetChannelOnChainCostQuery,
  useGetForwardsQuery,
  useGetChannelsQuery,
  useGetDecodedInvoiceQuery,
  useGetPaymentsQuery,
  useGetInvoicesQuery,
  useGetOnChainTxQuery,
  useSendOnChainMutation,
  useNewInvoiceMutation,
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
  useGetLocalNodesQuery,
  useUpdateLocalNodeMutation,
  useAddLocalNodeMutation,
  useUpdateLocalNodeSetDeletedMutation,
  useUpdateLocalNodeSetDisabledMutation,
  useUpdateChannelMutation,
} = torqApi;
