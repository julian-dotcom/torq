import { torqApi } from "apiSlice";
import { OpenChannelRequest, OpenChannelResponse } from "./openChannelTypes";

export const openChannelApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    openChannel: builder.mutation<OpenChannelResponse, OpenChannelRequest>({
      query: (body) => ({
        url: `channels/open`,
        method: "POST",
        body: body,
      }),
    }),
  }),
});

export const { useOpenChannelMutation } = openChannelApi;
