import { ColumnMetaData } from "features/table/types";
import { ExpandedTag } from "pages/tags/tagsTypes";
import { channel } from "features/channels/channelsTypes";
import { ChannelClosed } from "features/channelsClosed/channelsClosedTypes";
import { Invoice } from "features/transact/Invoices/invoiceTypes";
import { OnChainTx } from "features/transact/OnChain/types";
import { Payment } from "features/transact/Payments/types";
import { OrderBy } from "features/sidebar/sections/sort/SortSection";
import { Forward } from "features/forwards/forwardsTypes";
import { workflowListItem } from "pages/WorkflowPage/workflowTypes";
import { SerialisableFilterQuery } from "features/sidebar/sections/filter/filter";

export type ViewResponse<T> = {
  view: ViewInterface<T>;
  page: keyof AllViewsResponse;
  id?: number;
  dirty?: boolean;
};

export type TableResponses = Forward | OnChainTx | Payment | Invoice | ExpandedTag | channel | workflowListItem;

export type AllViewsResponse = {
  forwards: Array<ViewResponse<Forward>>;
  onChain: Array<ViewResponse<OnChainTx>>;
  payments: Array<ViewResponse<Payment>>;
  invoices: Array<ViewResponse<Invoice>>;
  tags: Array<ViewResponse<ExpandedTag>>;
  channel: Array<ViewResponse<channel>>;
  channelsClosed: Array<ViewResponse<ChannelClosed>>;
};

export type CreateViewRequest = {
  index: number;
  page: keyof AllViewsResponse;
  view: ViewInterface<TableResponses>;
};
export type UpdateViewRequest = { id: number; view: ViewInterface<TableResponses> };

export type ViewInterface<T> = {
  title: string;
  filters?: SerialisableFilterQuery;
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
