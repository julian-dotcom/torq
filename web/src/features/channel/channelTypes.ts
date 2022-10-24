export type Sections = {
  filter: boolean;
  sort: boolean;
  group: boolean;
  columns: boolean;
};

export type ChannelHistoryResponse = {
  label: string
  amountOut: number
  amountIn: number
  amountTotal: number
  revenueOut: number
  revenueIn: number
  revenueTotal: number
  countOut: number
  countIn: number
  countTotal: number
  channels: Channel[]
  history: History[]
}

type History = {
  alias: string;
  amountOut: number
  amountIn: number
  amountTotal: number
  revenueOut: number
  revenueIn: number
  revenueTotal: number
  countOut: number
  countIn: number
  countTotal: number
}

export type Channel = {
  alias: string;
  channelDbId: string;
  channelPoint: string;
  pubKey: string;
  shortChannelId: string;
  chanId: string;
  open: boolean;
  capacity: number;
}

export type ChannelRebalancingResponse = {
  rebalancingCost: number;
  rebalancing: Rebalancing;
}

type Rebalancing = {
  amountMsat: number;
  totalCostMsat: number;
  splitCostMsat: number;
  count: number;
}

export type ChannelOnchainCostResponse = {
  onChainCost: number;
}

export type ChannelBalanceResponse = {
  channelBalance: ChannelBalance[] | null;
}

type ChannelBalance = {
  LndShortChannelId: string;
  balances: Balance[];
}

type Balance = {
	date: Date
	inboundCapacity:  number;
	outboundCapacity: number;
	capacityDiff:     number;
}

export type ChannelEventResponse = {
	events: Event[] | null;
}

export type Event = {
	date: string;
	datetime: Date;
	lndChannelPoint: string;
	shortChannelId: string
	type: string;
	outbound: boolean;
	announcingPubKey: string;
	value: number;
	previousValue: number;
}
