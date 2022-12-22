export type Tag = {
  tagId: number;
  name: string;
  style: string;
  createdOn: Date;
  updateOn: Date;
};

export type TagsSidebarSections = {
  filter: boolean;
  sort: boolean;
  group: boolean;
  columns: boolean;
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
  channels: ChannelForTag[]
  nodes: NodeForTag[]
}

export type ChannelForTag = {
  shortChannelId: string;
  alias: string;
  type: string;
}

export type NodeForTag = {
  nodeId: string;
  alias: string;
  type: string;
}
