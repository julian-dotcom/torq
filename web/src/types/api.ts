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

export type ForwardResponse = {
  alias: string;
  channelDbId: number;
  lndChannelPoint: string;
  pubKey: string;
  shortChannelId: string;
  lndShortChannelId: string;
  color: string;
  open: number;
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

export type OnchainResponse = {
  data: OnchainData[];
  pagination: Pagination;
}

type Pagination = {
  limit: number;
  offset: number;
  total: number;
}

type OnchainData = {
  date: string;
  amount: number;
  destAddresses: string[];
  destAddressesCount: string;
  label: string;
  lndShortChanId: string;
  lndTxTypeLabel: string;
  totalFees: number;
  txHash: number;
}

export type PaymentsResponse = {
  data: PaymentData[];
  pagination: Pagination;
}

type PaymentData = {
  paymentIndex: number;
  date: string;
  destinationPubKey: string;
  status: string;
  value: number;
  fee: number;
  ppm: number;
  failureReason: string;
  txHash: number;
}

export type InvoicesResponse = {
  data : InvoiceData[];
  pagination: Pagination;
}

type InvoiceData = {
  creationDate: string;
  settleDate: string;
  addIndex: number;
  settleIndex: number;
  paymentRequest: string;
  destinationPubKey: string;
  rHash: string;
  rPreimage: string;
  memo: string;
  value: number;
  amtPaid: number;
  invoiceState: string;
  isRebalance: boolean;
  isKeysend: boolean;
  isAmp: boolean;
  paymentAddr: string;
  fallbackAddr: string;
  updatedOn: string;
  expiry: number;
  cltvExpiry: number;
  private: boolean;
}
