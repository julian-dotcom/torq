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

// interface PendingHTLC {
//   forwardingCount: number;
//   forwardingAmount: number;
//   localCount: number;
//   localAmount: number;
//   toalCount: number;
//   totalAmount: number;
// }

export interface channel {
  active: boolean;
  ambossSpace: string;
  baseFeeMsat: number;
  capacity: number;
  channelPoint: string;
  chanStatusFlags: string;
  commitFee: number;
  commitmentType: number;
  commitWeight: number;
  feePerKw: number;
  feeRatePpm: number;
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
  minHtlc: number;
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
  remoteBaseFeeMsat: number;
  remoteChanReserveSat: number;
  remoteFeeRatePpm: number;
  remoteMaxHtlcMsat: number;
  remoteMinHtlc: number;
  remotePubkey: number;
  remoteTimeLockDelta: number;
  shortChannelId: string;
  timeLockDelta: number;
  totalSatoshisReceived: number;
  totalSatoshisSent: number;
  unsettledBalance: number;
}

export type PolicyInterface = {
  feeRatePpm?: number;
  timeLockDelta?: number;
  maxHtlcMsat?: number;
  minHtlcMsat?: number;
  baseFeeMsat?: number;
  channelId?: number;
  nodeId: number;
};
