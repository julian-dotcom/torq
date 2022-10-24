export interface settings {
  defaultDateRange: string;
  defaultLanguage: "en" | "nl";
  preferredTimezone: string;
  weekStartsOn: "saturday" | "sunday" | "monday";
}

export interface timeZone {
  name: string;
}

export interface localNode {
  localNodeId: number;
  name?: string;
  implementation: string;
  grpcAddress?: string;
  tlsFileName?: string;
  tlsFile: File | null;
  macaroonFileName?: string;
  macaroonFile: File | null;
  createdOn?: Date;
  updatedOn?: Date;
  disabled: boolean;
  deleted: boolean;
}

export interface channel {
  active: boolean;
  gauge: number;
  remotePubkey: string;
  lndChannelPoint: string;
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
  localNodeId: number;
  timeLockDelta: number;
  feeRatePpm: number;
}

interface PendingHTLC {
  forwardingCount: number;
  forwardingAmount: number;
  localCount: number;
  localAmount: number;
  toalCount: number;
  totalAmount: number;
}
