import { torqApi } from "apiSlice";
import { channelTag, Tag, ChannelNode, ChannelGroup, Corridor } from "./tagsTypes";
import { stringMap } from "apiTypes";

// Define a service using a base URL and expected endpoints
export const onChainApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getTags: builder.query<Tag[], void>({
      query: () => `tags`,
      providesTags: ["tags"],
    }),
    getTag: builder.query<Tag, number>({
      query: (tagId) => `tags/${tagId}`,
      providesTags: ["tag"],
    }),
    addTag: builder.mutation<Tag, Tag>({
      query: (tag) => ({
        url: `tags`,
        method: "POST",
        body: tag,
      }),
      invalidatesTags: ["tags", "tag"],
    }),
    setTag: builder.mutation<Tag, Tag>({
      query: (tag) => ({
        url: `tags`,
        method: "PUT",
        body: tag,
      }),
      invalidatesTags: ["tags", "tag"],
    }),
    deleteTag: builder.mutation<number, number>({
      query: (tagId) => ({
        url: `channelGroups/tag/${tagId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["tags", "tag"],
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
    addChannelsGroups: builder.mutation<ChannelGroup, ChannelGroup>({
      query: (channelGroups) => ({
        url: `channelGroups`,
        method: "POST",
        body: channelGroups,
      }),
      invalidatesTags: ["corridors"],
    }),
    getCorridorByReference: builder.query<Corridor, number>({
      query: (tagId) => ({
        url: `corridors/${tagId}`,
        method: "GET",
      }),
      providesTags: ["corridors"],
    }),
    deleteChannelGroupByTag: builder.mutation<number, number>({
      query: (corridorId) => ({
        url: `channelGroups/corridor/${corridorId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["corridors"],
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const {
  useGetTagsQuery,
  useGetTagQuery,
  useGetCorridorByReferenceQuery,
  useAddTagMutation,
  useAddChannelsGroupsMutation,
  useDeleteChannelGroupByTagMutation,
  useDeleteTagMutation,
  useSetTagMutation,
  useGetNodesChannelsQuery,
} = onChainApi;
