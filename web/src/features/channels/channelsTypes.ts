export type UpdateChannelResponse = {
  status: number;
  failedUpdates: null | FailedRequest[];
};

type FailedRequest = {
  reason: string;
  error: string;
};

export interface channel {
  active: boolean;
  ambossSpace: string;
  feeBaseMsat: number;
  capacity: number;
  channelPoint: string;
  chanStatusFlags: string;
  commitFee: number;
  commitmentType: number;
  commitWeight: number;
  feePerKw: number;
  feeRateMilliMsat: number;
  fundingOutputIndex: number;
  fundingTransactionHash: string;
  gauge: number;
  initiator: boolean;
  lifetime: number;
  lndShortChannelId: number;
  balance: number; // NB! This column only exists in the frontend!
  localBalance: number;
  localChanReserveSat: number;
  maxHtlcMsat: number;
  mempoolSpace: string;
  minHtlcMsat: number;
  nodeId: number;
  channelId: number;
  nodeName: string;
  numUpdates: number;
  oneMl: string;
  peerAlias: string;
  pendingForwardingHTLCsAmount: number;
  pendingForwardingHTLCsCount: number;
  pendingLocalHTLCsAmount: number;
  pendingLocalHTLCsCount: number;
  pendingTotalHTLCsAmount: number;
  pendingTotalHTLCsCount: number;
  remoteBalance: number;
  remoteFeeBaseMsat: number;
  remoteChanReserveSat: number;
  remoteFeeRateMilliMsat: number;
  remoteMaxHtlcMsat: number;
  remoteMinHtlcMsat: number;
  remotePubkey: number;
  remoteTimeLockDelta: number;
  shortChannelId: string;
  timeLockDelta: number;
  totalSatoshisReceived: number;
  totalSatoshisSent: number;
  unsettledBalance: number;
}

export type PolicyInterface = {
  feeRateMilliMsat?: number;
  timeLockDelta?: number;
  maxHtlcMsat?: number;
  minHtlcMsat?: number;
  feeBaseMsat?: number;
  channelId?: number;
  nodeId: number;
};
