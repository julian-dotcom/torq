type Flow = {
  from: string;
  to: string;
};

type Order = {
  key: string;
  direction: "asc" | "desc";
};

type BaseQueryParams = {
  limit: number;
  offset: number;
  order: Order;
  filter?: Record<string, any>;
};

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

export type GetPaymentsQueryParams = BaseQueryParams;

export type GetInvoicesQueryParams = BaseQueryParams;

export type GetOnChainTransactionsQueryParams = BaseQueryParams;
