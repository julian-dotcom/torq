import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { PolicyInterface } from "features/channels/ChannelsSlice";
import { getRestEndpoint, getWsEndpoint } from "utils/apiUrlBuilder";
import { UpdatedChannelResponse } from "features/channels/channelsTypes";
import { ViewInterface, viewOrderInterface } from "features/table/Table";
import {
  ChannelOnchainCostResponse,
  ChannelHistoryResponse,
  ChannelRebalancingResponse,
  ChannelBalanceResponse,
  ChannelEventResponse,
  FlowData,
} from "features/channel/channelTypes";

import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from "@reduxjs/toolkit/query";
import type {
  GetDecodedInvoiceQueryParams,
  GetForwardsQueryParams,
  GetInvoicesQueryParams,
  GetOnChainTransactionsQueryParams,
  GetPaymentsQueryParams,
  SendOnChainRequest,
  SendOnChainResponse,
  GetTableViewQueryParams,
  GetFlowQueryParams,
  GetChannelHistoryData,
  LoginResponse,
  ForwardResponse,
  OnchainResponse,
  PaymentsResponse,
  InvoicesResponse,
  DecodedInvoice,
} from "types/api";
import { queryParamsBuilder } from "utils/queryParamsBuilder";
import type { nodeConfiguration, settings, timeZone, channel, services } from "apiTypes";
import { NewInvoiceRequest, NewInvoiceResponse } from "./features/transact/Invoices/newInvoice/newInvoiceTypes";

const API_URL = getRestEndpoint();
export const WS_URL = getWsEndpoint();

const baseQuery = fetchBaseQuery({
  baseUrl: API_URL,
  credentials: "include",
  mode: "cors",
});

const baseQueryWithRedirect: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError> = async (
  args,
  api,
  extraOptions
) => {
  const result = await baseQuery(args, api, extraOptions);
  if (result.error && result.error.status === 401 && window.location.pathname !== "/login") {
    window.location.href = "/login";
  }
  if (result.error && result.error.status === 424 && window.location.pathname !== "/services") {
    window.location.href = "/services";
  }
  return result;
};

// Define a service using a base URL and expected endpoints
export const torqApi = createApi({
  reducerPath: "api",
  baseQuery: baseQueryWithRedirect,
  tagTypes: ["settings", "tableView", "nodeConfigurations", "channels", "services"],
  endpoints: (builder) => ({
    getFlow: builder.query<FlowData[], GetFlowQueryParams>({
      query: (params) => "flow" + queryParamsBuilder(params),
    }),
    getChannelHistory: builder.query<ChannelHistoryResponse, GetChannelHistoryData>({
      query: (data) => `channels/${data.params.chanId}/history` + queryParamsBuilder(data.queryParams),
    }),
    getChannelEvent: builder.query<ChannelEventResponse, GetChannelHistoryData>({
      query: (data) => `channels/${data.params.chanId}/event` + queryParamsBuilder(data.queryParams),
    }),
    getChannelBalance: builder.query<ChannelBalanceResponse, GetChannelHistoryData>({
      query: (data) => `channels/${data.params.chanId}/balance` + queryParamsBuilder(data.queryParams),
    }),
    getChannelRebalancing: builder.query<ChannelRebalancingResponse, GetChannelHistoryData>({
      query: (data) => `channels/${data.params.chanId}/rebalancing` + queryParamsBuilder(data.queryParams),
    }),
    getChannelOnChainCost: builder.query<ChannelOnchainCostResponse, GetChannelHistoryData>({
      query: (data) => `channels/${data.params.chanId}/onchaincost` + queryParamsBuilder(data.queryParams),
    }),
    getForwards: builder.query<ForwardResponse[], GetForwardsQueryParams>({
      query: (params) => "forwards" + queryParamsBuilder(params, true),
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
    getDecodedInvoice: builder.query<DecodedInvoice, GetDecodedInvoiceQueryParams>({
      query: (params) => "invoices/decode" + queryParamsBuilder(params),
    }),
    getPayments: builder.query<PaymentsResponse, GetPaymentsQueryParams>({
      query: (params) => "payments" + queryParamsBuilder(params, true),
    }),
    getInvoices: builder.query<InvoicesResponse, GetInvoicesQueryParams>({
      query: (params) => "invoices" + queryParamsBuilder(params, true),
    }),
    getOnChainTx: builder.query<OnchainResponse, GetOnChainTransactionsQueryParams>({
      query: (params) => "on-chain-tx" + queryParamsBuilder(params, true),
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
    getTableViews: builder.query<any, GetTableViewQueryParams>({
      query: (params) => "table-views" + queryParamsBuilder(params),
      providesTags: ["tableView"],
    }),
    createTableView: builder.mutation<any, { view: ViewInterface; index: number; page: string }>({
      query: (data) => ({
        url: "table-views",
        method: "POST",
        body: { id: null, view: data.view, page: data.page },
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
    }),
    login: builder.mutation<LoginResponse, FormData>({
      query: (form) => ({
        url: "login",
        method: "POST",
        body: new URLSearchParams(form as any),
      }),
    }),
    cookieLogin: builder.mutation<LoginResponse, string>({
      query: (key) => ({
        url: `cookie-login`,
        method: "POST",
        body: { accessKey: key },
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
    getNodeConfigurations: builder.query<nodeConfiguration[], void>({
      query: () => `settings/nodeConnectionDetails`,
      providesTags: ["nodeConfigurations"],
    }),
    getNodeConfiguration: builder.query<nodeConfiguration, number>({
      query: (nodeId) => `settings/nodeConnectionDetails/${nodeId}`,
      providesTags: ["nodeConfigurations"],
    }),
    addNodeConfiguration: builder.mutation<any, FormData>({
      query: (nodeConfiguration) => ({
        url: "settings/nodeConnectionDetails",
        method: "POST",
        body: nodeConfiguration,
      }),
      invalidatesTags: ["nodeConfigurations", "channels"],
    }),
    updateNodeConfiguration: builder.mutation<any, FormData>({
      query: (nodeConfiguration) => ({
        url: `settings/nodeConnectionDetails`,
        method: "PUT",
        body: nodeConfiguration,
      }),
      invalidatesTags: ["nodeConfigurations", "channels"],
    }),
    updateNodeConfigurationStatus: builder.mutation<any, { nodeId: number; status: number }>({
      query: (nodeConfiguration) => ({
        url: `settings/nodeConnectionDetails/${nodeConfiguration.nodeId}/${nodeConfiguration.status}`,
        method: "PUT",
      }),
      invalidatesTags: ["nodeConfigurations", "channels"],
    }),
    updateNodePingSystemStatus: builder.mutation<any, { nodeId: number; pingSystem: number; statusId: number }>({
      query: (nodeConfiguration) => ({
        url: `settings/nodePingSystem/${nodeConfiguration.nodeId}/${nodeConfiguration.pingSystem}/${nodeConfiguration.statusId}`,
        method: "PUT",
      }),
      invalidatesTags: ["nodeConfigurations", "channels"],
    }),
    getServices: builder.query<services, void>({
      query: () => "services/status",
      providesTags: ["services"],
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
  useCookieLoginMutation,
  useLogoutMutation,
  useGetSettingsQuery,
  useUpdateSettingsMutation,
  useGetTimeZonesQuery,
  useGetNodeConfigurationsQuery,
  useGetNodeConfigurationQuery,
  useUpdateNodeConfigurationMutation,
  useAddNodeConfigurationMutation,
  useUpdateNodeConfigurationStatusMutation,
  useUpdateNodePingSystemStatusMutation,
  useUpdateChannelMutation,
  useGetServicesQuery,
} = torqApi;
