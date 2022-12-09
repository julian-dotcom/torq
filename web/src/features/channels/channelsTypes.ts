export type UpdatedChannelResponse = {
  status: string;
  failedUpdates: null | FailedUpdates[];
};

type FailedUpdates = {
  reason: string;
  updateError: string;
  OutPoint: FailedUpdatesOutPoint;
};

type FailedUpdatesOutPoint = {
  outputIndex: number;
  Txid: string;
};

interface PendingHTLC {
  forwardingCount: number;
  forwardingAmount: number;
  localCount: number;
  localAmount: number;
  toalCount: number;
  totalAmount: number;
}

export interface channel {
  channelId: number;
  peerAlias: string;
  active: boolean;
  gauge: number;
  remotePubkey: string;
  fundingTransactionHash: string;
  fundingOutputIndex: number;
  lndShortChannelId: number;
  shortChannelId: string;
  capacity: number;
  localBalance: number;
  remoteBalance: number;
  unsettledBalance: number;
  commitFee: number;
  commitWeight: number;
  feePerKw: number;
  baseFeeMsat: number;
  minHtlc: number;
  maxHtlcMsat: number;
  pendingHtlcs: PendingHTLC;
  totalSatoshisSent: number;
  numUpdates: number;
  initiator: boolean;
  chanStatusFlags: string;
  localChanReserveSat: number;
  remoteChanReserveSat: number;
  commitmentType: number;
  lifetime: number;
  totalSatoshisReceived: number;
  timeLockDelta: number;
  feeRatePpm: number;
  firstNodeId: number;
  secondNodeId: number;
  pendingLocalHTLCsAmount: number;
  pendingLocalHTLCsCount: number;
  pendingForwardingHTLCsCount: number;
  pendingForwardingHTLCsAmount: number;
  pendingTotalHTLCsCount: number;
  pendingTotalHTLCsAmount: number;
  nodeId: number;
  channelPoint: string;
  nodeName: string;
  mempoolSpace: string;
  ambossSpace: string;
  oneMl: string;
}

export type PolicyInterface = {
  feeRatePpm: number;
  timeLockDelta: number;
  maxHtlcMsat: number;
  minHtlcMsat: number;
  baseFeeMsat: number;
  fundingTransactionHash: string;
  fundingOutputIndex: number;
  nodeId: number;
};
