import { ViewResponse } from "features/viewManagement/types";

import { FilterInterface } from "features/sidebar/sections/filter/filter";
import { ColumnMetaData } from "features/table/types";
import {
  AllChannelPendingColumns,
  ChannelsPendingFilterableColumns,
  ChannelsPendingSortableColumns,
} from "features/channelsPending/channelsPendingColumns.generated";
import { ChannelPending } from "./channelsPendingTypes";

const defaultColumns: Array<keyof ChannelPending> = [
  "peerAlias",
  "capacity",
  "status",
  "fundingTransactionHash",
  "closingTransactionHash",
  "lndShortChannelId",
  "fundedOn",
  "shortChannelId",
  "nodeName",
  "fundingBlockHeight",
  "fundingBlockHeightDelta",
  "fundedOnSecondsDelta",
  "closingBlockHeight",
  "closingBlockHeightDelta",
];

//changed to not have to rely on the generated file for default columns ordering
export const DefaultChannelsPendingColumns = defaultColumns.map((col) =>
  AllChannelPendingColumns.find((x) => x.key === col)
) as Array<ColumnMetaData<ChannelPending>>;

export const FilterableChannelsPendingColumns = AllChannelPendingColumns.filter(
  (column: ColumnMetaData<ChannelPending>) => ChannelsPendingFilterableColumns.includes(column.key)
);

export const ChannelsPendingFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number",
  parameter: 0,
  key: "capacity",
};

export const ChannelsPendingSortTemplate: { key: keyof ChannelPending; direction: "desc" | "asc" } = {
  key: "peerAlias",
  direction: "asc",
};

export const DefaultPendingChannelsView: ViewResponse<ChannelPending> = {
  page: "channelsPending",
  dirty: true,
  view: {
    title: "Pending Channels",
    columns: DefaultChannelsPendingColumns,
    sortBy: [ChannelsPendingSortTemplate],
  },
};

export const SortableChannelsPendingColumns = AllChannelPendingColumns.filter((c) =>
  ChannelsPendingSortableColumns.includes(c.key)
);
