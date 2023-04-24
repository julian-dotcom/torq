import { torqApi } from "apiSlice";
import { OpenChannelRequest, OpenChannelResponse } from "./openChannelTypes";

export const openChannelApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    openChannel: builder.mutation<OpenChannelResponse, OpenChannelRequest>({
      query: (body) => ({
        url: `lightning/open`,
        method: "POST",
        body: body,
      }),
      invalidatesTags: ["channelsPending"],
    }),
  }),
});

export const { useOpenChannelMutation } = openChannelApi;
