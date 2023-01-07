export type Tag = {
  tagId?: number;
  name: string;
  style: string;
  categoryId?: number;
  createdOn?: Date;
  updateOn?: Date;
};

export type ExpandedTag = Tag & {
  delete?: boolean;
};

export interface channelTag {
  channelTagId: number;
  tagOriginId: number;
  fromNodeId: number;
  toNodeId: number;
  channelId: number;
  tagId: number;
}

export type ChannelNode = {
  channels: ChannelForTag[];
  nodes: NodeForTag[];
};

export type ChannelForTag = {
  shortChannelId: string;
  channelId: number;
  nodeId: number;
  alias: string;
  type: string;
};

export type NodeForTag = {
  nodeId: number;
  alias: string;
  type: string;
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

export type TagNodeChannel = {
  tagId: number;
  channelId?: number;
  nodeId: number;
};
