import { ChannelStatus } from "../channelsTypes";

export type CloseChannelRequest = {
  nodeId: number;
  channelId: number;
  force?: boolean;
  targetConf?: number;
  deliveryAddress?: string;
  satPerVbyte?: number;
};

export type CloseChannelResponse = {
  request: CloseChannelRequest;
  status: ChannelStatus;
  closingTransactionHash: string;
};
