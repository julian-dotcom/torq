import { torqApi } from "apiSlice";
import { ChannelNode, TagChannelRequest, TagNodeRequest, TagResponse } from "./tagsTypes";

// Define a service using a base URL and expected endpoints
export const onChainApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getTags: builder.query<Array<TagResponse>, void>({
      query: () => `tags`,
      providesTags: ["tags"],
    }),
    getTag: builder.query<TagResponse, number>({
      query: (tagId) => `tags/${tagId}`,
      providesTags: ["tag"],
    }),
    addTag: builder.mutation<TagResponse, TagResponse>({
      query: (tag) => ({
        url: `tags`,
        method: "POST",
        body: tag,
      }),
      invalidatesTags: ["tags", "tag"],
    }),
    setTag: builder.mutation<TagResponse, TagResponse>({
      query: (tag) => ({
        url: `tags`,
        method: "PUT",
        body: tag,
      }),
      invalidatesTags: ["tags", "tag"],
    }),
    deleteTag: builder.mutation<number, number>({
      query: (tagId) => ({
        url: `tags/${tagId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["tags", "tag"],
    }),
    tagChannel: builder.mutation<void, TagChannelRequest>({
      query: (tagChannel) => ({
        url: `tags/tag`,
        method: "POST",
        body: tagChannel,
      }),
      invalidatesTags: ["tags", "tag", "channels", "channelHistory", "forwards"],
    }),
    tagNode: builder.mutation<void, TagNodeRequest>({
      query: (tagNode) => ({
        url: `tags/tag`,
        method: "POST",
        body: tagNode,
      }),
      invalidatesTags: ["tags", "tag", "channels", "channelHistory"],
    }),
    untagNode: builder.mutation<void, TagNodeRequest>({
      query: (tagNode) => ({
        url: `tags/untag`,
        method: "POST",
        body: tagNode,
      }),
      invalidatesTags: ["tags", "tag", "channels", "channelHistory"],
    }),
    untagChannel: builder.mutation<void, TagChannelRequest>({
      query: (tagChannel) => ({
        url: `tags/untag`,
        method: "POST",
        body: tagChannel,
      }),
      invalidatesTags: ["tags", "tag", "channels", "channelHistory"],
    }),
    // getChannelsForTag: builder.query<ChannelNode, number>({
    //   query: (tagId) => `tags/${tagId}/channels`,
    //   providesTags: ["tagsForChannel"],
    // }),
    // getNodesForTag: builder.query<ChannelNode, number>({
    //   query: (tagId) => `tags/${tagId}/nodes`,
    //   providesTags: ["tagsForNodes"],
    // }),
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
  useGetTagQuery,
  useTagChannelMutation,
  useTagNodeMutation,
  useUntagNodeMutation,
  useUntagChannelMutation,
  useAddTagMutation,
  useDeleteTagMutation,
  useSetTagMutation,
  useGetNodesChannelsQuery,
} = onChainApi;
