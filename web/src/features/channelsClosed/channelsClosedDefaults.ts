import { ViewResponse } from "features/viewManagement/types";
import { ChannelClosed } from "features/channelsClosed/channelsClosedTypes";

import { FilterInterface } from "features/sidebar/sections/filter/filter";
import { ColumnMetaData } from "features/table/types";
import {
  AllChannelClosedColumns,
  ChannelsClosedFilterableColumns,
  ChannelsClosedSortableColumns,
} from "./channelsClosedColumns.generated";
import { OrderBy } from "features/sidebar/sections/sort/SortSection";

const defaultColumns: Array<keyof ChannelClosed> = [
  "peerAlias",
  "capacity",
  "closedOn",
  "fundedOn",
  "status",
  "shortChannelId",
  "lndShortChannelId",
  "fundingTransactionHash",
  "closingTransactionHash",
];

//changed to not have to rely on the generated file for default columns ordering
export const DefaultChannelsClosedColumns = defaultColumns.map((col) =>
  AllChannelClosedColumns.find((x) => x.key === col)
) as Array<ColumnMetaData<ChannelClosed>>;

export const FilterableChannelsClosedColumns = AllChannelClosedColumns.filter((column: ColumnMetaData<ChannelClosed>) =>
  ChannelsClosedFilterableColumns.includes(column.key)
);

export const ChannelsClosedFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number",
  parameter: 0,
  key: "capacity",
};

export const ChannelsClosedSortTemplate: Array<OrderBy> = [
  {
    key: "closedOnSecondsDelta",
    direction: "desc",
  },
];

export const DefaultClosedChannelsView: ViewResponse<ChannelClosed> = {
  page: "channelsClosed",
  dirty: true,
  view: {
    title: "Closed Channels",
    columns: DefaultChannelsClosedColumns,
    sortBy: ChannelsClosedSortTemplate,
  },
};

export const SortableChannelsClosedColumns = AllChannelClosedColumns.filter((c) =>
  ChannelsClosedSortableColumns.includes(c.key)
);
