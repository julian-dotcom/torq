import { Tag } from "pages/tags/tagsTypes";

export type Peer = {
  nodeId: number;
  torqNodeId: number;
  peerAlias: string;
  torqNodeAlias: string;
  pubKey: string;
  address: string;
  connectionStatus: ConnectionStatus;
  setting: NodeConnectionSetting;
  tags: Tag[]; // TODO: Add tags to the peers api endpoint.
  daysSinceLastConnected: number;
  dateLastConnected?: Date;
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
};

export type UpdatePeerResponse = {
  success: boolean;
};

export enum ConnectionStatus {
  Disconnected = "Disconnected",
  Connected = "Connected",
}

export enum NodeConnectionSetting {
  AlwaysReconnect = "AlwaysReconnect",
  DisableReconnect = "DisableReconnect",
}

// Add NodeConnectionSetting to integer
export const NodeConnectionSettingInt = {
  [NodeConnectionSetting.AlwaysReconnect]: 0,
  [NodeConnectionSetting.DisableReconnect]: 1,
};
