import { Tag } from "pages/tags/tagsTypes";

export enum ChannelStatus {
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

export type UpdateChannelResponse = {
  status: number;
  failedUpdates: null | FailedRequest[];
};

type FailedRequest = {
  reason: string;
  error: string;
};

export type channel = {
  active: boolean;
  tags: Tag[]; // this column only exists on the frontend to render the tags cell
  channelTags: Tag[];
  peerTags: Tag[];
  ambossSpace: string;
  feeBase: number;
  capacity: number;
  channelPoint: string;
  chanStatusFlags: string;
  commitFee: number;
  commitmentType: number;
  commitWeight: number;
  feePerKw: number;
  feeRateMilliMsat: number;
  currentBlockHeight: number;
  fundingOutputIndex: number;
  fundingTransactionHash: string;
  fundingBlockHeight: number;
  fundingBlockHeightDelta: number;
  fundedOn: Date;
  fundedOnSecondsDelta: number;
  closingTransactionHash: string;
  closingBlockHeight: number;
  closingBlockHeightDelta: number;
  closedOn: Date;
  closedOnSecondsDelta: number;
  gauge: number;
  initiator: boolean;
  lifetime: number;
  lndShortChannelId: number;
  balance: number; // NB! This column only exists in the frontend!
  localBalance: number;
  localChanReserveSat: number;
  maxHtlc: number;
  mempoolSpace: string;
  minHtlc: number;
  nodeId: number;
  channelId: number;
  nodeName: string;
  numUpdates: number;
  oneMl: string;
  peerAlias: string;
  peerNodeId: number;
  pendingForwardingHTLCsAmount: number;
  pendingForwardingHTLCsCount: number;
  pendingLocalHTLCsAmount: number;
  pendingLocalHTLCsCount: number;
  pendingTotalHTLCsAmount: number;
  pendingTotalHTLCsCount: number;
  remoteBalance: number;
  remoteFeeBase: number;
  remoteChanReserveSat: number;
  remoteFeeRateMilliMsat: number;
  remoteMaxHtlc: number;
  remoteMinHtlc: number;
  remotePubkey: string;
  remoteTimeLockDelta: number;
  shortChannelId: string;
  timeLockDelta: number;
  totalSatoshisReceived: number;
  totalSatoshisSent: number;
  unsettledBalance: number;
  peerChannelCapacity: number;
  peerChannelCount: number;
  peerLocalBalance: number;
  peerGauge: number;
  private: boolean;
};

export type PolicyInterface = {
  feeRateMilliMsat?: number;
  timeLockDelta?: number;
  maxHtlcMsat?: number;
  minHtlcMsat?: number;
  feeBaseMsat?: number;
  channelId?: number;
  nodeId: number;
};
