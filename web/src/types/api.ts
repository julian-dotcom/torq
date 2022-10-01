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
  filter?: Record<string, any> | undefined;
};

export type BaseQuery = Flow | BaseQueryCollectionParams;

export type GetFlowQueryParams = Flow & {
  chanId: string;
};

export type GetChannelHistoryQueryParams = Flow & {
  chanIds: string;
};

export type GetForwardsQueryParams = Flow;

export type GetDecodedInvoiceQueryParams = {
  invoice: string;
};

export type GetPaymentsQueryParams = BaseQueryCollectionParams;

export type GetInvoicesQueryParams = BaseQueryCollectionParams;

export type GetOnChainTransactionsQueryParams = BaseQueryCollectionParams;
