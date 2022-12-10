import { ViewResponse } from "features/viewManagement/types";
import { channel } from "./channelsTypes";
import { ColumnMetaData } from "features/table/types";
import { FilterInterface } from "features/sidebar/sections/filter/filter";

export const AllChannelsColumns: ColumnMetaData<channel>[] = [
  {
    heading: "Active",
    type: "BooleanCell",
    key: "active",
    valueType: "boolean",
    locked: false,
  },
  {
    heading: "Peer Alias",
    type: "AliasCell",
    key: "peerAlias",
    valueType: "string",
  },
  {
    heading: "Short Channel ID",
    type: "LongTextCell",
    key: "shortChannelId",
    valueType: "string",
  },
  {
    heading: "Remote Balance",
    type: "NumericCell",
    key: "remoteBalance",
    valueType: "number",
  },
  {
    heading: "Balance",
    type: "BalanceCell",
    key: "localBalance",
    key2: "remoteBalance",
    valueType: "number",
    percent: false,
  },
  {
    heading: "Local Balance",
    type: "NumericCell",
    key: "localBalance",
    valueType: "number",
  },
  {
    heading: "Capacity",
    type: "NumericCell",
    key: "capacity",
    valueType: "number",
  },
  {
    heading: "Fee rate (PPM)",
    type: "NumericDoubleCell",
    key: "feeRatePpm",
    key2: "remoteFeeRatePpm",
    suffix: "ppm",
    valueType: "number",
  },
  {
    heading: "Base Fee Msat",
    type: "NumericDoubleCell",
    key: "baseFeeMsat",
    key2: "remoteBaseFeeMsat",
    suffix: "msat",
    valueType: "number",
  },
  {
    heading: "Minimum HTLC",
    type: "NumericDoubleCell",
    key: "minHtlc",
    key2: "remoteMinHtlc",
    suffix: "msat",
    valueType: "number",
  },
  {
    heading: "Maximum HTLC Amount",
    type: "NumericDoubleCell",
    key: "maxHtlcMsat",
    key2: "remoteMaxHtlcMsat",
    suffix: "sat",
    valueType: "number",
  },
  {
    heading: "Time Lock Delta",
    type: "NumericCell",
    key: "timeLockDelta",
    valueType: "number",
  },
  {
    heading: "LND Short Channel ID",
    type: "LongTextCell",
    key: "lndShortChannelId",
    valueType: "string",
  },
  {
    heading: "Funding Transaction",
    type: "LongTextCell",
    key: "fundingTransactionHash",
    valueType: "string",
  },
  {
    heading: "Unsettled Balance",
    type: "NumericCell",
    key: "unsettledBalance",
    valueType: "number",
  },
  {
    heading: "Satoshis Sent",
    type: "NumericCell",
    key: "totalSatoshisSent",
    valueType: "number",
  },
  {
    heading: "Satoshis Received",
    type: "NumericCell",
    key: "totalSatoshisReceived",
    valueType: "number",
  },
  {
    heading: "Pending Forwarding HTLCs count",
    type: "NumericCell",
    key: "pendingForwardingHTLCsCount",
    valueType: "number",
  },
  {
    heading: "Pending Forwarding HTLCs",
    type: "NumericCell",
    key: "pendingForwardingHTLCsAmount",
    valueType: "number",
  },
  {
    heading: "Pending Forwarding HTLCs count",
    type: "NumericCell",
    key: "pendingLocalHTLCsCount",
    valueType: "number",
  },
  {
    heading: "Pending Forwarding HTLCs",
    type: "NumericCell",
    key: "pendingLocalHTLCsAmount",
    valueType: "number",
  },
  {
    heading: "Total Pending Forwarding HTLCs count",
    type: "NumericCell",
    key: "pendingTotalHTLCsCount",
    valueType: "number",
  },
  {
    heading: "Total Pending Forwarding HTLCs",
    type: "NumericCell",
    key: "pendingTotalHTLCsAmount",
    valueType: "number",
  },
  {
    heading: "Commit Fee",
    type: "NumericCell",
    key: "commitFee",
    valueType: "number",
  },
  {
    heading: "Node Name",
    type: "AliasCell",
    key: "nodeName",
    valueType: "string",
  },
  {
    heading: "Mempool",
    type: "LinkCell",
    key: "mempoolSpace",
    valueType: "link",
  },
  {
    heading: "Amboss",
    type: "LinkCell",
    key: "ambossSpace",
    valueType: "link",
  },
  {
    heading: "1ML",
    type: "LinkCell",
    key: "oneMl",
    valueType: "link",
  },
];

const defaultColumns: Array<keyof channel> = [
  "active",
  "peerAlias",
  "shortChannelId",
  "feeRatePpm",
  "remoteBalance",
  "gauge",
  "localBalance",
  "capacity",
  "totalSatoshisSent",
  "totalSatoshisReceived",
  "unsettledBalance",
  "commitFee",
  "baseFeeMsat",
  "minHtlc",
  "maxHtlcMsat",
  "nodeName",
];

export const DefaultChannelsColumns = AllChannelsColumns.filter((c) => defaultColumns.includes(c.key));

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
  view: {
    title: "Untitled View",
    columns: DefaultChannelsColumns,
    sortBy: [ChannelsSortTemplate],
  },
};

const sortableColumns: Array<keyof channel> = [
  "active",
  "peerAlias",
  "shortChannelId",
  "feeRatePpm",
  "remoteBalance",
  "localBalance",
  "capacity",
  "totalSatoshisSent",
  "totalSatoshisReceived",
  "unsettledBalance",
  "commitFee",
  "baseFeeMsat",
  "minHtlc",
  "maxHtlcMsat",
  "nodeName",
];

export const SortableChannelsColumns = AllChannelsColumns.filter((c) => sortableColumns.includes(c.key));



