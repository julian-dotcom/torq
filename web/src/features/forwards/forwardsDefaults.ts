import { AndClause, FilterInterface } from "features/sidebar/sections/filter/filter";
import { ColumnMetaData } from "features/table/types";
import { ViewResponse } from "../viewManagement/types";
import { OrderBy } from "features/sidebar/sections/sort/SortSection";
import { Forward } from "./forwardsTypes";

export const AllForwardsColumns: ColumnMetaData<Forward>[] = [
  {
    heading: "Name",
    type: "AliasCell",
    key: "alias",
    locked: true,
    valueType: "string",
  },
  {
    heading: "Revenue",
    type: "BarCell",
    key: "revenueOut",
    valueType: "number",
  },
  {
    heading: "Total Forwards",
    type: "BarCell",
    key: "countTotal",
    valueType: "number",
  },
  {
    heading: "Outbound Amount",
    type: "BarCell",
    key: "amountOut",
    valueType: "number",
  },
  {
    heading: "Inbound Amount",
    type: "BarCell",
    key: "amountIn",
    valueType: "number",
  },
  {
    heading: "Total Amount",
    type: "BarCell",
    key: "amountTotal",
    valueType: "number",
  },
  {
    heading: "Turnover Outbound",
    type: "BarCell",
    key: "turnoverOut",
    valueType: "number",
  },
  {
    heading: "Turnover Inbound",
    type: "BarCell",
    key: "turnoverIn",
    valueType: "number",
  },
  {
    heading: "Total Turnover",
    type: "BarCell",
    key: "turnoverTotal",
    valueType: "number",
  },
  {
    heading: "Outbound Forwards",
    type: "BarCell",
    key: "countOut",
    valueType: "number",
  },
  {
    heading: "Inbound Forwards",
    type: "BarCell",
    key: "countIn",
    valueType: "number",
  },
  {
    heading: "Revenue inbound",
    type: "BarCell",
    key: "revenueIn",
    valueType: "number",
  },
  {
    heading: "Revenue total",
    type: "BarCell",
    key: "revenueTotal",
    valueType: "number",
  },
  {
    heading: "Capacity",
    type: "BarCell",
    key: "capacity",
    valueType: "number",
  },
  {
    heading: "Public key",
    type: "LongTextCell",
    key: "pubKey",
    valueType: "string",
  },
  {
    heading: "Funding Tx Hash",
    type: "LongTextCell",
    key: "fundingTransactionHash",
    valueType: "string",
  },
  {
    heading: "Funding Tx Output Index",
    type: "TextCell",
    key: "fundingOutputIndex",
    valueType: "string",
  },
  {
    heading: "Channel short ID",
    type: "LongTextCell",
    key: "shortChannelId",
    valueType: "string",
  },
  {
    heading: "LND Channel short ID",
    type: "LongTextCell",
    key: "lndShortChannelId",
    valueType: "string",
  },
  {
    heading: "Channel Point",
    type: "LongTextCell",
    key: "channelPoint",
    valueType: "string",
  },
  {
    heading: "Open",
    type: "BooleanCell",
    key: "open",
    valueType: "boolean",
  },
];

// const ForwardsFilterTemplate = new AndClause();
// ForwardsFilterTemplate.addChildClause(
//   new FilterClause({
//     funcName: "gt",
//     category: "number",
//     key: "amountTotal",
//     parameter: 0,
//   })
// );

export const ForwardsFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number",
  parameter: 0,
  key: "amountTotal",
};

const defaultColumns: Array<keyof Forward> = [
  "alias",
  "revenueOut",
  "countTotal",
  "amountOut",
  "amountIn",
  "amountTotal",
  "turnoverTotal",
  "capacity",
];

export const ForwardsSortByTemplate: OrderBy = { key: "revenueOut", direction: "desc" };

export const DefaultForwardsColumns = AllForwardsColumns.filter((c) => defaultColumns.includes(c.key));

export const DefaultForwardsView: ViewResponse<Forward> = {
  page: "forwards",
  dirty: true,
  view: {
    title: "Draft View",
    filters: new AndClause().toJSON(),
    columns: DefaultForwardsColumns,
    sortBy: [ForwardsSortByTemplate],
    groupBy: "channels",
  },
};
