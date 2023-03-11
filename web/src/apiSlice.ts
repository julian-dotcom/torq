import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { channel, PolicyInterface } from "features/channels/channelsTypes";
import { getRestEndpoint, getWsEndpoint } from "utils/apiUrlBuilder";
import { UpdateChannelResponse } from "features/channels/channelsTypes";
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
  GetForwardsQueryParams,
  GetFlowQueryParams,
  GetChannelHistoryData,
  LoginResponse,
  ActiveNetwork,
} from "types/api";
import { queryParamsBuilder } from "utils/queryParamsBuilder";
import { Forward } from "features/forwards/forwardsTypes";
import type { nodeConfiguration, settings, timeZone, services, updateSettingsRequest } from "apiTypes";
import { createSelector } from "@reduxjs/toolkit";
import { Network } from "features/network/networkSlice";
import { lndServices, nodeWalletBalances } from "apiTypes";

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
  tagTypes: [
    "settings",
    "tableView",
    "nodeConfigurations",
    "forwards",
    "channels",
    "channelsPending",
    "channelsClosed",
    "channelHistory",
    "services",
    "tags",
    "tag",
    "categories",
    "corridors",
    "workflows",
    "workflow",
    "nodeWalletBalance",
    // "tagsForChannel",
    // "tagsForNodes",
  ],
  endpoints: (builder) => ({
    getFlow: builder.query<FlowData[], GetFlowQueryParams>({
      query: (params) => "flow" + queryParamsBuilder(params),
    }),
    getChannelHistory: builder.query<ChannelHistoryResponse, GetChannelHistoryData>({
      query: (data) => `channels/${data.params.chanId}/history` + queryParamsBuilder(data.queryParams),
      providesTags: ["channelHistory"],
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
    getForwards: builder.query<Array<Forward>, GetForwardsQueryParams>({
      query: (params) => "forwards" + queryParamsBuilder(params, false),
      providesTags: ["forwards"],
    }),
    getChannels: builder.query<channel[], ActiveNetwork>({
      query: (params) => ({
        url: `channels` + queryParamsBuilder(params),
        method: "GET",
      }),
      providesTags: ["channels"],
    }),
    getChannelsClosed: builder.query<channel[], ActiveNetwork>({
      query: (params) => ({
        url: `channels/closed` + queryParamsBuilder(params),
        method: "GET",
      }),
      providesTags: ["channelsClosed"],
    }),
    getChannelsPending: builder.query<channel[], ActiveNetwork>({
      query: (params) => ({
        url: `channels/pending` + queryParamsBuilder(params),
        method: "GET",
      }),
      providesTags: ["channelsPending"],
    }),
    updateChannel: builder.mutation<UpdateChannelResponse, PolicyInterface>({
      query: (data: PolicyInterface) => ({
        url: "channels/update",
        method: "PUT",
        body: data,
      }),
      invalidatesTags: ["channels"],
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    updateSettings: builder.mutation<any, updateSettingsRequest>({
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
    getNodesWalletBalances: builder.query<nodeWalletBalances[], void>({
      query: () => "nodes/walletBalances",
      providesTags: ["nodeWalletBalance"],
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    addNodeConfiguration: builder.mutation<any, FormData>({
      query: (nodeConfiguration) => ({
        url: "settings/nodeConnectionDetails",
        method: "POST",
        body: nodeConfiguration,
      }),
      invalidatesTags: ["nodeConfigurations", "channels", "channelsClosed", "channelsPending"],
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    updateNodeConfiguration: builder.mutation<any, FormData>({
      query: (nodeConfiguration) => ({
        url: `settings/nodeConnectionDetails`,
        method: "PUT",
        body: nodeConfiguration,
      }),
      invalidatesTags: ["nodeConfigurations", "channels"],
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    updateNodeConfigurationStatus: builder.mutation<any, { nodeId: number; status: number }>({
      query: (nodeConfiguration) => ({
        url: `settings/nodeConnectionDetails/${nodeConfiguration.nodeId}/${nodeConfiguration.status}`,
        method: "PUT",
      }),
      invalidatesTags: ["nodeConfigurations", "channels", "channelsClosed", "channelsPending"],
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
    getLndServices: builder.query<lndServices, number>({
      query: (nodeId) => `services/${nodeId}/lndStatus`,
    }),
    getAutoLoginSetting: builder.query<boolean, void>({
      query: () => `auto-login-setting`,
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
  useLoginMutation,
  useCookieLoginMutation,
  useLogoutMutation,
  useGetSettingsQuery,
  useUpdateSettingsMutation,
  useGetTimeZonesQuery,
  useGetNodeConfigurationsQuery,
  useGetNodeConfigurationQuery,
  useGetNodesWalletBalancesQuery,
  useUpdateNodeConfigurationMutation,
  useAddNodeConfigurationMutation,
  useUpdateNodeConfigurationStatusMutation,
  useUpdateNodePingSystemStatusMutation,
  useUpdateChannelMutation,
  useGetServicesQuery,
  useGetLndServicesQuery,
  useGetAutoLoginSettingQuery,
  useGetChannelsClosedQuery,
  useGetChannelsPendingQuery,
} = torqApi;

export const SelectChannel = (props: { network: Network; channelId: number }) => {
  return createSelector([torqApi.endpoints.getChannels.select({ network: props.network })], (channels) => {
    return channels?.data?.find((channel) => channel.channelId === props.channelId);
  });
};

// Select channel and node tags for a given channel
export const SelectChannelTags = (props: { network: Network; channelId: number }) => {
  return createSelector(
    [
      SelectChannel({ network: props.network, channelId: props.channelId }),
      torqApi.endpoints.getChannels.select({ network: props.network }),
    ],
    (channel) => {
      return {
        tags: channel?.tags,
      };
    }
  );
};

export const SelectChannels = (props: { network: Network; channelIds: Array<number> }) => {
  return createSelector([torqApi.endpoints.getChannels.select({ network: props.network })], (channels) => {
    return channels?.data?.filter((channel) => {
      return props.channelIds.includes(channel.channelId);
    });
  });
};
