import { TagColor } from "components/tags/Tag";

export type Tag = {
  tagId?: number;
  name: string;
  style: TagColor;
  categoryId?: number;
  categoryName?: string;
  categoryStyle?: string;
  createdOn?: Date;
  updateOn?: Date;
};

export type TaggedNodes = {
  name: string;
  nodeId: number;
  openChannelCount: number;
  closedChannelCount: number;
};

export type TaggedChannels = {
  name: string;
  shortChannelId: string;
  channelId: number;
};

export type TagResponse = {
  tagId?: number;
  name: string;
  style: string;
  categoryId?: number;
  categoryName?: string;
  categoryStyle?: string;
  createdOn?: Date;
  updateOn?: Date;
  channels: Array<TaggedChannels>;
  nodes: Array<TaggedNodes>;
};

export type ExpandedTag = TagResponse & {
  delete?: boolean;
  edit?: boolean;
};

export type TagChannelRequest = {
  tagId: number;
  channelId: number;
};
export type TagNodeRequest = {
  tagId: number;
  nodeId: number;
};

export type ChannelNode = {
  channels: ChannelForTag[];
  nodes: NodeForTag[];
};

export type ChannelForTag = {
  shortChannelId?: string;
  channelId: number;
  nodeId: number;
  alias?: string;
  type: string;
};

export type NodeForTag = {
  nodeId: number;
  alias: string;
  type: string;
};
