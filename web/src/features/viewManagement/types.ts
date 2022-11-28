import { ColumnMetaData } from "features/table/types";
import { Clause } from "features/sidebar/sections/filter/filter";
import { Forward } from "types/api";
import { tag } from "pages/tagsPage/tagsTypes";
import { channel } from "features/channels/channelsTypes";
import { Invoice } from "features/transact/Invoices/invoiceTypes";
import { OnChainTx } from "features/transact/OnChain/types";
import { Payment } from "features/transact/Payments/types";
import { OrderBy } from "../sidebar/sections/sort/SortSection";

export type PageViewType<T extends {}> = {
  selectedViewIndex: number;
  views: Array<ViewInterface<T>>;
};

export type ViewInterface<T> = {
  id?: number; // Missing ID means the view is not in the database
  saved: boolean; // False, means the view is different from the version in the database.
  view_order: number; // This is the order of the view set by the user.
  title: string;
  filters?: Clause;
  columns: Array<ColumnMetaData<T>>;
  sortBy?: Array<OrderBy>;
  groupBy?: "channels" | "peers" | undefined;
  page: string;
};

export interface ViewOrderInterface {
  id: number | undefined;
  view_order: number;
}

export type ViewRow<ViewInterfaceResponse> = {
  view: ViewInterfaceResponse;
  index: number;
  onSelectView: (index: number) => void;
  singleView: boolean;
};

export type GetTableViewQueryParams = {
  page: string;
};

export type TableResponses = Forward | OnChainTx | Payment | Invoice | tag | channel;
export type AllViewsResponse = {
  forwards: Array<ViewInterface<Forward>>;
  onChain: Array<ViewInterface<OnChainTx>>;
  payments: Array<ViewInterface<Payment>>;
  invoices: Array<ViewInterface<Invoice>>;
  tags: Array<ViewInterface<tag>>;
  channel: Array<ViewInterface<channel>>;
};

export type ViewInterfaceResponse = ViewInterface<TableResponses>;
