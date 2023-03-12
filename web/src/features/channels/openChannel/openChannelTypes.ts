enum ChannelStatus {
  Opening = 1,
  Open = 2,
  Closing = 3,
  CooperativeClosed = 100,
  LocalForceClosed = 101,
  RemoteForceClosed = 102,
  BreachClosed = 103,
  FundingCancelledClosed = 104,
  AbandonedClosed = 105,
}

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
