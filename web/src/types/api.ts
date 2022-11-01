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
  localNodeId: number;
};

export type SendOnChainRequest = {
  localNodeId: number;
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
}

type LoginFail = {
  data?: { error: string };
  error?: string;
  status: number;
}
type LoginSuccess = {
  message: string;
}

export type GetPaymentsQueryParams = BaseQueryCollectionParams;

export type GetInvoicesQueryParams = BaseQueryCollectionParams;

export type GetOnChainTransactionsQueryParams = BaseQueryCollectionParams;

export type GetTableViewQueryParams = {
  page: string;
};

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
