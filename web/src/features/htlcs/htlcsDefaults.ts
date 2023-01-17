import { AndClause, FilterInterface } from "features/sidebar/sections/filter/filter";
import { ColumnMetaData } from "features/table/types";
import { ViewResponse } from "features/viewManagement/types";
import { OrderBy } from "features/sidebar/sections/sort/SortSection";
import { Htlc } from "features/htlcs/htlcsTypes";

export const AllHtlcColumns: ColumnMetaData<Htlc>[] = [
  {
    heading: "Event Time",
    type: "TextCell",
    key: "eventTime",
    locked: true,
    valueType: "date",
  },
  {
    heading: "Event Type",
    type: "TextCell",
    key: "eventType",
    valueType: "string",
  },
  {
    heading: "Event Origin",
    type: "TextCell",
    key: "eventOrigin",
    valueType: "string",
  },
  {
    heading: "Incoming Name",
    type: "LongTextCell",
    key: "incomingAlias",
    valueType: "string",
  },
  {
    heading: "Incoming Amount (msat)",
    type: "BarCell",
    key: "incomingAmountMsat",
    valueType: "number",
  },
  {
    heading: "Incoming Tags",
    type: "TagsCell",
    key: "incomingChannelTags",
    valueType: "tag",
  },
  {
    heading: "Incoming Capacity",
    type: "BarCell",
    key: "incomingChannelCapacity",
    valueType: "number",
  },
  {
    heading: "Incoming Public key",
    type: "LongTextCell",
    key: "incomingPublicKey",
    valueType: "string",
  },
  {
    heading: "Incoming Channel short ID",
    type: "LongTextCell",
    key: "incomingShortChannelId",
    valueType: "string",
  },
  {
    heading: "Incoming LND Channel short ID",
    type: "LongTextCell",
    key: "incomingLndShortChannelId",
    valueType: "string",
  },
  {
    heading: "Outgoing Name",
    type: "AliasCell",
    key: "outgoingAlias",
    valueType: "string",
  },
  {
    heading: "Outgoing Amount (msat)",
    type: "BarCell",
    key: "outgoingAmountMsat",
    valueType: "number",
  },
  {
    heading: "Outgoing Tags",
    type: "TagsCell",
    key: "outgoingChannelTags",
    valueType: "tag",
  },
  {
    heading: "Outgoing Capacity",
    type: "BarCell",
    key: "outgoingChannelCapacity",
    valueType: "number",
  },
  {
    heading: "Outgoing Public key",
    type: "LongTextCell",
    key: "outgoingPublicKey",
    valueType: "string",
  },
  {
    heading: "Outgoing Channel short ID",
    type: "LongTextCell",
    key: "outgoingShortChannelId",
    valueType: "string",
  },
  {
    heading: "Outgoing LND Channel short ID",
    type: "LongTextCell",
    key: "outgoingLndShortChannelId",
    valueType: "string",
  },
  {
    heading: "Failure Code",
    type: "TextCell",
    key: "boltFailureCode",
    valueType: "string",
  },
  {
    heading: "Failure String",
    type: "TextCell",
    key: "boltFailureString",
    valueType: "string",
  },
  {
    heading: "Failure Detail",
    type: "TextCell",
    key: "lndFailureDetail",
    valueType: "string",
  },
];

export const HtlcsFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number",
  parameter: 0,
  key: "incomingAmountMsat",
};

const defaultColumns: Array<keyof Htlc> = [
  "eventTime",
  "eventType",
  "eventOrigin",
  "incomingAlias",
  "incomingAmountMsat",
  "incomingChannelCapacity",
  "outgoingAlias",
  "outgoingAmountMsat",
  "outgoingChannelCapacity",
  "boltFailureCode",
  "boltFailureString",
  "lndFailureDetail",
];

export const HtlcsSortByTemplate: OrderBy = { key: "incomingAlias", direction: "asc" };

export const DefaultHtlcsColumns = AllHtlcColumns.filter((c) => defaultColumns.includes(c.key));

export const DefaultHtlcsView: ViewResponse<Htlc> = {
  page: "htlcs",
  dirty: true,
  view: {
    title: "Draft View",
    filters: new AndClause().toJSON(),
    columns: DefaultHtlcsColumns,
    sortBy: [HtlcsSortByTemplate],
  },
};
