import { ViewResponse } from "features/viewManagement/types";
import { channel } from "features/channels/channelsTypes";
import { AllChannelsColumns, ChannelsSortableColumns, ChannelsFilterableColumns } from "features/channels/channelsColumns.generated";
import { FilterInterface } from "features/sidebar/sections/filter/filter";
import { ColumnMetaData } from "features/table/types";

const defaultColumns: Array<keyof channel> = [
  "peerAlias",
  "active",
  "balance",
  "feeRateMilliMsat",
  "feeBase",
  "minHtlc",
  "maxHtlc",
  "shortChannelId",
  "nodeName",
];

export const DefaultChannelsColumns = AllChannelsColumns.filter((c) => defaultColumns.includes(c.key));

export const FilterableChannelsColumns = AllChannelsColumns.filter((column: ColumnMetaData<channel>) =>
  ChannelsFilterableColumns.includes(column.key)
);

export const ChannelsFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number",
  parameter: 0,
  key: "capacity",
};

export const ChannelsSortTemplate: { key: keyof channel; direction: "desc" | "asc" } = {
  key: "peerAlias",
  direction: "asc",
};

export const DefaultChannelsView: ViewResponse<channel> = {
  page: "channel",
  dirty: true,
  view: {
    title: "Draft View",
    columns: DefaultChannelsColumns,
    sortBy: [ChannelsSortTemplate],
  },
};

export const SortableChannelsColumns = AllChannelsColumns.filter((c) => ChannelsSortableColumns.includes(c.key));
