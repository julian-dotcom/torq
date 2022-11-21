type Order = {
  key: string;
  direction: "asc" | "desc";
};

type Paginable = {
  limit: number;
  offset: number;
};

export type FromAndTo = {
  from: string;
  to: string;
};

export type BaseQueryCollectionParams = Paginable & {
  order?: Order;
  filter?: Record<string, unknown>;
};

export type GetFlowQueryParams = FromAndTo & {
  chanIds: string;
};

export type GetChannelHistoryParams = {
  chanId: string;
};
export type GetChannelHistoryQueryParams = FromAndTo;
export type GetChannelHistoryData = {
  params: GetChannelHistoryParams;
  queryParams: GetChannelHistoryQueryParams;
};

export type GetForwardsQueryParams = FromAndTo;

export type GetDecodedInvoiceQueryParams = {
  invoice: string;
  nodeId: number;
};

export type SendOnChainRequest = {
  nodeId: number;
  address: string;
  amountSat: number;
  targetConf?: number;
  satPerVbyte?: number;
  sendAll?: boolean;
  label?: string;
  minConfs?: number;
  spendUnconfirmed?: boolean;
};

export type SendOnChainResponse = {
  txId: string;
};

export type LoginResponse = {
  error?: LoginFail;
  data?: LoginSuccess;
};

type LoginFail = {
  data?: { error: string };
  error?: string;
  status: number;
};
type LoginSuccess = {
  message: string;
};

export type GetPaymentsQueryParams = BaseQueryCollectionParams;

export type GetInvoicesQueryParams = BaseQueryCollectionParams;

export type GetOnChainTransactionsQueryParams = BaseQueryCollectionParams;

type InvoiceFeature = {
  Name: string;
  IsKnown: boolean;
  IsRequired: boolean;
};

type FeatureMap = Map<number, InvoiceFeature>;

type HopHint = {
  lNDShortChannelId: number;
  shortChannelId: string;
  nodeId: string;
  feeBase: number;
  cltvExpiryDelta: number;
  feeProportional: number;
};

type RouteHint = {
  hopHints: Array<HopHint>;
};

export type DecodedInvoice = {
  nodeAlias: string;
  paymentRequest: string;
  destinationPubKey: string;
  rHash: string;
  memo: string;
  valueMsat: number;
  paymentAddr: string;
  fallbackAddr: string;
  expiry: number;
  createdAt: number;
  cltvExpiry: number;
  private: boolean;
  features: FeatureMap;
  routeHints: Array<RouteHint>;
};

export type Forward = {
  alias: string;
  channelDbId: number;
  lndChannelPoint: string;
  pubKey: string;
  shortChannelId: string;
  lndShortChannelId: string;
  fundingOutputIndex: number;
  fundingTransactionHash: string;
  color: string;
  open: boolean;
  capacity: number;
  amountOut: number;
  amountIn: number;
  amountTotal: number;
  revenueOut: number;
  revenueIn: number;
  revenueTotal: number;
  countOut: number;
  countIn: number;
  countTotal: number;
  turnoverOut: number;
  turnoverIn: number;
  turnoverTotal: number;
};

export type Pagination = {
  limit: number;
  offset: number;
  total: number;
};
