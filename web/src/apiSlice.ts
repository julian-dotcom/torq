import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import {channel, PolicyInterface} from "features/channels/channelsTypes";
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
import type { GetForwardsQueryParams, GetFlowQueryParams, GetChannelHistoryData, LoginResponse } from "types/api";
import { queryParamsBuilder } from "utils/queryParamsBuilder";
import { tag, channelTag } from "pages/tagsPage/tagsTypes";
import { Forward } from "./features/forwards/forwardsTypes";
import type { nodeConfiguration, settings, timeZone, stringMap, services } from "apiTypes";

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
  tagTypes: ["settings", "tableView", "nodeConfigurations", "channels", "services", "tags"],
  endpoints: (builder) => ({
    getTags: builder.query<tag[], void>({
      query: () => `tags/all`,
      providesTags: ["tags"],
    }),
    getTag: builder.query<tag, number>({
      query: (tagId) => `tags/get/${tagId}`,
      providesTags: ["tags"],
    }),
    addTag: builder.mutation<tag, tag>({
      query: (tag) => ({
        url: `tags/add`,
        method: "POST",
        body: tag,
      }),
      invalidatesTags: ["tags"],
    }),
    setTag: builder.mutation<tag, tag>({
      query: (tag) => ({
        url: `tags/set`,
        method: "PUT",
        body: tag,
      }),
      invalidatesTags: ["tags"],
    }),
    removeTag: builder.mutation<number, number>({
      query: (tagId) => ({
        url: `tags/${tagId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["tags"],
    }),
    addChannelTag: builder.mutation<stringMap<string>, channelTag>({
      query: (channelTag) => ({
        url: `channelTags/add`,
        method: "POST",
        body: channelTag,
      }),
    }),
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
    getForwards: builder.query<Array<Forward>, GetForwardsQueryParams>({
      query: (params) => "forwards" + queryParamsBuilder(params, false),
    }),
    getChannels: builder.query<channel[], void>({
      query: () => ({
        url: `channels`,
        method: "GET",
      }),
      providesTags: ["channels"],
    }),
    updateChannel: builder.mutation<UpdateChannelResponse, PolicyInterface>({
      query: (data: PolicyInterface) => ({
        url: "channels/update",
        method: "PUT",
        body: data,
      }),
      invalidatesTags: ["channels"],
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
  useGetTagsQuery,
  useGetServicesQuery,
} = torqApi;
