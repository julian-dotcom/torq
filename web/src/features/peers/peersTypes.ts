import { Tag } from "pages/tags/tagsTypes";

export type Peer = {
  nodeId: number;
  pubKey: string;
  torqNodeId: number;
  torqNodeAlias: string;
  connectionStatus: ConnectionStatus;
  setting: NodeConnectionSetting;
  tags: Tag[];
  peerAlias: string;
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
  torqNodeId: number;
};

export type DisconnectPeerResponse = {
  success: boolean;
};

export type UpdatePeerRequest = {
  nodeId: number;
  torqNodeId: number;
  setting: NodeConnectionSetting;
};

export type UpdatePeerResponse = {
  success: boolean;
};

export enum ConnectionStatus {
  Disconnected = 0,
  Connected = 1,
}

export enum NodeConnectionSetting {
  AlwaysReconnect = 0,
  DisableReconnect = 1,
}
