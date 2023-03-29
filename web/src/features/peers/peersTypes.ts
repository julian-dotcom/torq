import { Tag } from "pages/tags/tagsTypes";

export type Peer = ExpandedPeer & {
  nodeId: number;
  pubKey: string;
  chain: string;
  network: string;
  nodeConnectionDetailsNodeId: number;
  status: number;
  tags: Tag[];
};

export type ExpandedPeer = {
  actions: boolean;
};

export type ConnectPeerRequest = {
  nodeId: number;
  connectionString: string;
  network: number;
};

export type ConnectPeerResponse = {};

export type DisconnectPeerRequest = {
  nodeId: number;
  nodeConnectionDetailsNodeId: number;
};

export type DisconnectPeerResponse = {};
