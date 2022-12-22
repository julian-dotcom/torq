import { torqApi } from "apiSlice";
import { channelTag, Tag, ChannelNode } from "./tagsTypes";
import { stringMap } from "apiTypes";

// Define a service using a base URL and expected endpoints
export const onChainApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getTags: builder.query<Tag[], void>({
      query: () => `tags/all`,
      providesTags: ["tags"],
    }),
    getTag: builder.query<Tag, number>({
      query: (tagId) => `tags/get/${tagId}`,
      providesTags: ["tags"],
    }),
    addTag: builder.mutation<Tag, Tag>({
      query: (tag) => ({
        url: `tags/add`,
        method: "POST",
        body: tag,
      }),
      invalidatesTags: ["tags"],
    }),
    setTag: builder.mutation<Tag, Tag>({
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
    getNodesChannels: builder.query<ChannelNode, void>({
      query: () => ({
        url: `channels/nodes`,
        method: "GET",
      }),
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const {
  useGetTagsQuery,
  useGetNodesChannelsQuery,
} = onChainApi;
