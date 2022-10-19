export type Sections = {
  filter: boolean;
  sort: boolean;
  group: boolean;
  columns: boolean;
};

export type UpdatedChannelResponse = {
  status: string
  failedUpdates: null | FailedUpdates[]
}

type FailedUpdates = {
  reason: string;
  updateError: string;
  OutPoint: FailedUpdatesOutPoint;
}

type FailedUpdatesOutPoint = {
  outputIndex: number;
  Txid: string;
}

