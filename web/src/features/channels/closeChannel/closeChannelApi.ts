import { torqApi } from "apiSlice";
import { CloseChannelRequest, CloseChannelResponse } from "./closeChannelTypes";

export const api = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    closeChannel: builder.mutation<CloseChannelResponse, CloseChannelRequest>({
      query: (body) => ({
        url: `lightning/close`,
        method: "POST",
        body: body,
      }),
      invalidatesTags: ["channelsPending", "channels"],
    }),
  }),
});

export const { useCloseChannelMutation } = api;
