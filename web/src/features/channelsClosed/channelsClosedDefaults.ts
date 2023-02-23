import { ViewResponse } from "features/viewManagement/types";
import { ChannelClosed } from "features/channelsClosed/channelsClosedTypes";

import { FilterInterface } from "features/sidebar/sections/filter/filter";
import { ColumnMetaData } from "features/table/types";
import {
  AllChannelClosedColumns,
  ChannelsClosedFilterableColumns,
  ChannelsClosedSortableColumns,
} from "./channelsClosedColumns.generated";

const defaultColumns: Array<keyof ChannelClosed> = [
  "peerAlias",
  "status",
  "nodeName",
  "pubKey",
  "shortChannelId",
  "lndShortChannelId",
  "closingNodeName",
];

export const DefaultChannelsClosedColumns = AllChannelClosedColumns.filter((c) => defaultColumns.includes(c.key));

export const FilterableChannelsClosedColumns = AllChannelClosedColumns.filter((column: ColumnMetaData<ChannelClosed>) =>
  ChannelsClosedFilterableColumns.includes(column.key)
);

export const ChannelsClosedFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number",
  parameter: 0,
  key: "capacity",
};

export const ChannelsClosedSortTemplate: { key: keyof ChannelClosed; direction: "desc" | "asc" } = {
  key: "peerAlias",
  direction: "asc",
};

export const DefaultClosedChannelsView: ViewResponse<ChannelClosed> = {
  page: "channelsClosed",
  dirty: true,
  view: {
    title: "Closed Channels",
    columns: DefaultChannelsClosedColumns,
    sortBy: [ChannelsClosedSortTemplate],
  },
};

export const SortableChannelsClosedColumns = AllChannelClosedColumns.filter((c) =>
  ChannelsClosedSortableColumns.includes(c.key)
);
