// import { uuid } from "uuidv4";
import { ViewInterface, ViewResponse } from "../viewManagement/types";
import { channel } from "./channelsTypes";
import { ColumnMetaData } from "../table/types";
import { FilterInterface } from "../sidebar/sections/filter/filter";

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
    type: "TextCell",
    key: "peerAlias",
    valueType: "string",
  },
  {
    heading: "Short Channel ID",
    type: "TextCell",
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
    heading: "Balance Gauge",
    type: "BarCell",
    key: "gauge",
    valueType: "number",
    percent: true,
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
    type: "NumericCell",
    key: "feeRatePpm",
    valueType: "number",
  },
  {
    heading: "Base Fee Msat",
    type: "NumericCell",
    key: "baseFeeMsat",
    valueType: "number",
  },
  {
    heading: "Minimum HTLC",
    type: "NumericCell",
    key: "minHtlc",
    valueType: "number",
  },
  {
    heading: "Maximum HTLC Msat",
    type: "NumericCell",
    key: "maxHtlcMsat",
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
    type: "TextCell",
    key: "lndShortChannelId",
    valueType: "string",
  },
  {
    heading: "Funding Transaction",
    type: "TextCell",
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
    heading: "Node ID",
    type: "NumericCell",
    key: "nodeId",
    valueType: "number",
  },
  {
    heading: "Node Name",
    type: "TextCell",
    key: "nodeName",
    valueType: "string",
  },
  {
    heading: "Mempool",
    type: "LinkCell",
    key: "mempoolSpace",
    valueType: "string",
  },
  {
    heading: "Amboss",
    type: "LinkCell",
    key: "ambossSpace",
    valueType: "string",
  },
  {
    heading: "1ML",
    type: "LinkCell",
    key: "oneMl",
    valueType: "string",
  },
];

export interface TableChannelsState {
  channels: [];
  selectedViewIndex: number;
  views: ViewInterface<channel>[];
  status: "idle" | "loading" | "failed";
}

// const defaultFilter = new AndClause();
// defaultFilter.addChildClause(
//   new FilterClause({
//     funcName: "eq",
//     category: "boolean",
//     key: "active",
//     parameter: true,
//   })
// );

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
  key: "nodeName",
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
