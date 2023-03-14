import { torqApi } from "apiSlice";
import { CloseChannelRequest, CloseChannelResponse } from "./closeChannelTypes";

export const api = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    closeChannel: builder.mutation<CloseChannelResponse, CloseChannelRequest>({
      query: (body) => ({
        url: `channels/close`,
        method: "POST",
        body: body,
      }),
    }),
  }),
});

export const { useCloseChannelMutation } = api;
