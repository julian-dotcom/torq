import { Tag } from "pages/tags/tagsTypes";

export type Peer = ExpandedPeer & {
  nodeId: number;
  pubKey: string;
  chain: string;
  network: string;
  nodeConnectionDetailsNodeId: number;
  connectionStatus: PeerStatus;
  tags: Tag[];
  peerAlias: string;
};

export type ExpandedPeer = {
  actions: boolean;
};

export type ConnectPeerRequest = {
  nodeId: number;
  connectionString: string;
  network: number;
};

export type ConnectPeerResponse = {
  success: boolean;
};

export type DisconnectPeerRequest = {
  nodeId: number;
  nodeConnectionDetailsNodeId: number;
};

export type DisconnectPeerResponse = {
  success: boolean;
};

export enum PeerStatus {
  Inactive = 0,
  Active = 1,
}
