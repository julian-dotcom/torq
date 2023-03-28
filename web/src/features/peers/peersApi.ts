import { torqApi } from "apiSlice";
import { ConnectPeerRequest, ConnectPeerResponse } from "./peersTypes";

export const peersApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    connectPeer: builder.mutation<ConnectPeerResponse, ConnectPeerRequest>({
      query: (body) => ({
        url: `nodes/peers/connect`,
        method: "POST",
        body: body,
      }),
      invalidatesTags: ["peers"],
    }),
  }),
});

export const { useConnectPeerMutation } = peersApi;
