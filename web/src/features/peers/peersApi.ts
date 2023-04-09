import { torqApi } from "apiSlice";
import {
  ConnectPeerRequest,
  ConnectPeerResponse,
  DisconnectPeerRequest,
  DisconnectPeerResponse,
  Peer,
  UpdatePeerRequest,
  UpdatePeerResponse,
} from "./peersTypes";
import { ActiveNetwork } from "types/api";
import { queryParamsBuilder } from "utils/queryParamsBuilder";

export const peersApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    connectPeer: builder.mutation<ConnectPeerResponse, ConnectPeerRequest>({
      query: (body) => ({
        url: `peers/connect`,
        method: "POST",
        body: body,
      }),
      invalidatesTags: ["peers"],
    }),
    disconnectPeer: builder.mutation<DisconnectPeerResponse, DisconnectPeerRequest>({
      query: (body) => ({
        url: `peers/disconnect`,
        method: "PATCH",
        body: body,
      }),
      invalidatesTags: ["peers"],
    }),
    reconnectPeer: builder.mutation<DisconnectPeerResponse, DisconnectPeerRequest>({
      query: (body) => ({
        url: `peers/reconnect`,
        method: "PATCH",
        body: body,
      }),
      invalidatesTags: ["peers"],
    }),
    updatePeer: builder.mutation<UpdatePeerResponse, UpdatePeerRequest>({
      query: (body) => ({
        url: `peers/update`,
        method: "PATCH",
        body: body,
      }),
      invalidatesTags: ["peers"],
    }),
    getPeers: builder.query<Peer[], ActiveNetwork>({
      query: (params) => ({
        url: `peers` + queryParamsBuilder(params),
        method: "GET",
      }),
      providesTags: ["peers"],
    }),
  }),
});

// Select a single peer from the list

export const {
  useConnectPeerMutation,
  useDisconnectPeerMutation,
  useReconnectPeerMutation,
  useUpdatePeerMutation,
  useGetPeersQuery,
} = peersApi;
