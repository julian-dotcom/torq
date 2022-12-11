import { ColumnMetaData } from "features/table/types";
import { tag } from "pages/tagsPage/tagsTypes";
import { channel } from "features/channels/channelsTypes";
import { Invoice } from "features/transact/Invoices/invoiceTypes";
import { OnChainTx } from "features/transact/OnChain/types";
import { Payment } from "features/transact/Payments/types";
import { OrderBy } from "../sidebar/sections/sort/SortSection";
import { Forward } from "../forwards/forwardsTypes";

export type ViewResponse<T> = {
  view: ViewInterface<T>;
  page: keyof AllViewsResponse;
  id?: number;
  dirty?: boolean;
};

export type TableResponses = Forward | OnChainTx | Payment | Invoice | tag | channel;

export type AllViewsResponse = {
  forwards: Array<ViewResponse<Forward>>;
  onChain: Array<ViewResponse<OnChainTx>>;
  payments: Array<ViewResponse<Payment>>;
  invoices: Array<ViewResponse<Invoice>>;
  tags: Array<ViewResponse<tag>>;
  channel: Array<ViewResponse<channel>>;
};

export type CreateViewRequest = {
  index: number;
  page: keyof AllViewsResponse;
  view: ViewInterface<TableResponses>;
};
export type UpdateViewRequest = { id: number; view: ViewInterface<TableResponses> };

export type ViewInterface<T> = {
  title: string;
  filters?: any;
  columns: Array<ColumnMetaData<T>>;
  sortBy?: Array<OrderBy>;
  groupBy?: "channels" | "peers" | undefined;
};

export interface ViewOrderInterface {
  id: number | undefined;
  viewOrder: number;
}

export type GetTableViewQueryParams = {
  page: string;
};

export type ViewInterfaceResponse = ViewInterface<TableResponses>;
