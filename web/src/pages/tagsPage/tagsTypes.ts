export type Tag = {
  tagId?: number;
  name: string;
  style: string;
  categoryId?: number;
  categoryName?: string;
  createdOn?: Date;
  updateOn?: Date;
};

export type TaggedNodes = {
  name: string;
  nodeId: number;
  channelCount: number;
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

export type ExpandedTag = Tag & {
  delete?: boolean;
  edit?: boolean;
};

export interface channelTag {
  channelTagId: number;
  tagOriginId: number;
  fromNodeId: number;
  toNodeId: number;
  channelId: number;
  tagId: number;
}

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

export type NewChannelGroupRequest = {
  nodeId?: number;
  categoryId?: number;
  tagId: number;
  channelId?: number;
};
export type ChannelGroup = {
  nodeId: number;
  categoryId?: number;
  tagId: number;
  channelId?: number;
};

export type CorridorFields = {
  referenceId: number;
  alias: string;
  shortChannelId: string;
  corridorId: number;
};

export type Corridor = {
  corridors: CorridorFields[];
  totalNodes: number;
  totalChannels: number;
};
