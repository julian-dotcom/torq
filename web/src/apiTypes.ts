export interface settings {
  defaultDateRange: string;
  defaultLanguage: "en" | "nl";
  preferredTimezone: string;
  weekStartsOn: "saturday" | "sunday" | "monday";
}

export interface timeZone {
  name: string;
}

export interface nodeConfiguration {
  nodeId: number;
  name?: string;
  implementation: number;
  grpcAddress?: string;
  tlsFileName?: string;
  tlsFile: File | null;
  macaroonFileName?: string;
  macaroonFile: File | null;
  createdOn?: Date;
  updatedOn?: Date;
  status: number;
  pingSystem: number;
  customSettings: number;
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
  nodeId: number;
  channelPoint: string;
}

interface PendingHTLC {
  forwardingCount: number;
  forwardingAmount: number;
  localCount: number;
  localAmount: number;
  toalCount: number;
  totalAmount: number;
}
