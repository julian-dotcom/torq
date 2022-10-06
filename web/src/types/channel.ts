type Channel = {
  alias: string;
  channelDbId: number;
  channelPoint: string;
  pubKey: string;
  shortChannelId: string;
  channelId: string;
  isOpen: boolean;
  capacity: number;
};

type HistoryRecord = {
  alias: string;
  date: string;
  amountOut: number;
  amountIn: number;
  amountTotal: number;
  revenueOut: number;
  revenueIn: number;
  revenueTotal: number;
  countOut: number;
  countIn: number;
  countTotal: number;
};

type Rebalancing = {
  amountMsat: number;
  totalCostMsat: number;
  splitCostMsat: number;
  count: number;
};

type Event = {
  date: string;
  datetime: string;
  lndChannelPoint: string;
  shortChannelId: string;
  type: string;
  isOutbound: boolean;
  announcingPubKey: string;
  value: number;
  previousValue: number;
};

export type ChannelHistory = {
  label: string;
  onChainCost: number;
  rebalancingCost: number;
  amountOut: number;
  amountIn: number;
  amountTotal: number;
  revenueOut: number;
  revenueIn: number;
  revenueTotal: number;
  countOut: number;
  countIn: number;
  countTotal: number;
  rebalancing: Rebalancing;
  channel_balance: null;
  channels: Channel[];
  history: HistoryRecord[];
  events: Event[];
};
