import { AndClause, FilterClause } from "features/sidebar/sections/filter/filter";
import { ColumnMetaData } from "features/table/types";
import { Forward } from "types/api";
import { ViewInterface } from "../viewManagement/types";

export const availableColumns: ColumnMetaData<Forward>[] = [
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
    type: "TextCell",
    key: "pubKey",
    valueType: "string",
  },
  {
    heading: "Funding Tx Hash",
    type: "TextCell",
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
    type: "TextCell",
    key: "shortChannelId",
    valueType: "string",
  },
  {
    heading: "LND Channel short ID",
    type: "TextCell",
    key: "lndShortChannelId",
    valueType: "string",
  },
  {
    heading: "Channel Point",
    type: "TextCell",
    key: "lndChannelPoint",
    valueType: "string",
  },
  {
    heading: "Open",
    type: "BooleanCell",
    key: "open",
    valueType: "boolean",
  },
];

const defaultFilter = new AndClause();
defaultFilter.addChildClause(
  new FilterClause({
    funcName: "gt",
    category: "number" as "number" | "string",
    key: "amountTotal",
    parameter: 0,
  })
);

export const activeForwardsColumns = availableColumns.filter((c) =>
  ["alias", "revenueOut", "countTotal", "amountOut", "amountIn", "amountTotal", "turnoverTotal", "capacity"].includes(
    c.key
  )
);

export const DefaultView: ViewInterface<Forward> = {
  title: "Untitled View",
  saved: true,
  columns: activeForwardsColumns,
  filters: defaultFilter,
  sortBy: [{ value: "revenueOut", label: "Revenue", direction: "desc" }],
  groupBy: undefined,
  page: "forwards",
  view_order: 0,
};
