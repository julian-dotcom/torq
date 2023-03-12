import { ChannelStatus } from "../channelsTypes";

export type OpenChannelRequest = {
  nodeId: number;
  satPerVbyte?: number;
  nodePubKey: string;
  host?: string;
  localFundingAmount: number;
  pushSat?: number;
  targetConf?: number;
  private?: boolean;
  minHtlcMsat?: number;
  remoteCsvDelay?: number;
  minConfs?: number;
  spendUnconfirmed?: boolean;
  closeAddress?: string;
};

export type OpenChannelResponse = {
  request: OpenChannelRequest;
  status: ChannelStatus;
  channelPoint: string;
  fundingTransactionHash: string;
  fundingOutputIndex: number;
};
