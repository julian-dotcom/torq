type Order = {
  key: string;
  direction: "asc" | "desc";
};

type Paginable = {
  limit: number;
  offset: number;
};

export type Flow = {
  from: string;
  to: string;
};

export type BaseQueryCollectionParams = Paginable & {
  order?: Order;
  filter?: Record<string, unknown>;
};

export type GetFlowQueryParams = Flow & {
  chan_id: string;
};

export type GetChannelHistoryQueryParams = Flow & {
  chanIds: string;
};

export type GetForwardsQueryParams = Flow;

export type GetDecodedInvoiceQueryParams = {
  invoice: string;
  nodeId: number;
};

export type SendOnChainRequest = {
  nodeId: number;
  addr: string;
  amountSat: number;
  targetConf?: number;
  satPerVbyte?: number;
  sendAll?: boolean;
  label?: string;
  minConfs?: number;
  spendUnconfirmed?: boolean;
};

export type GetPaymentsQueryParams = BaseQueryCollectionParams;

export type GetInvoicesQueryParams = BaseQueryCollectionParams;

export type GetOnChainTransactionsQueryParams = BaseQueryCollectionParams;
