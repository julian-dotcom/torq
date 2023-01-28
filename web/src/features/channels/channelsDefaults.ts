import { ViewResponse } from "features/viewManagement/types";
import { channel } from "./channelsTypes";
import { ColumnMetaData } from "features/table/types";
import { FilterInterface } from "features/sidebar/sections/filter/filter";

export const AllChannelsColumns: ColumnMetaData<channel>[] = [
  {
    heading: "Peer Alias",
    type: "AliasCell",
    key: "peerAlias",
    locked: true,
    valueType: "string",
  },
  {
    heading: "Active",
    type: "BooleanCell",
    key: "active",
    valueType: "boolean",
    locked: false,
  },
  {
    heading: "Balance",
    type: "BalanceCell",
    key: "balance",
    valueType: "number",
  },
  {
    heading: "Tags",
    type: "TagsCell",
    key: "tags",
    valueType: "tag",
  },
  {
    heading: "Short Channel ID",
    type: "LongTextCell",
    key: "shortChannelId",
    valueType: "string",
  },
  {
    heading: "Channel Balance (%)",
    type: "BarCell",
    key: "gauge",
    valueType: "number",
    suffix: "%",
  },
  {
    heading: "Remote Balance",
    type: "NumericCell",
    key: "remoteBalance",
    valueType: "number",
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
    key: "feeRateMilliMsat",
    key2: "remoteFeeRateMilliMsat",
    suffix: "ppm",
    valueType: "number",
  },
  {
    heading: "Base Fee Msat",
    type: "NumericDoubleCell",
    key: "feeBaseMsat",
    key2: "remoteFeeBaseMsat",
    suffix: "msat",
    valueType: "number",
  },
  {
    heading: "Remote Fee rate (PPM)",
    type: "NumericCell",
    key: "remoteFeeRateMilliMsat",
    suffix: "ppm",
    valueType: "number",
  },
  {
    heading: "Remote Base Fee Msat",
    type: "NumericCell",
    key: "remoteFeeBaseMsat",
    suffix: "msat",
    valueType: "number",
  },
  {
    heading: "Minimum HTLC",
    type: "NumericDoubleCell",
    key: "minHtlcMsat",
    key2: "remoteMinHtlcMsat",
    suffix: "msat",
    valueType: "number",
  },
  {
    heading: "Maximum HTLC",
    type: "NumericDoubleCell",
    key: "maxHtlcMsat",
    key2: "remoteMaxHtlcMsat",
    suffix: "sat",
    valueType: "number",
  },
  {
    heading: "Remote Minimum HTLC",
    type: "NumericCell",
    key: "remoteMinHtlcMsat",
    suffix: "msat",
    valueType: "number",
  },
  {
    heading: "Remote Maximum HTLC",
    type: "NumericCell",
    key: "remoteMaxHtlcMsat",
    suffix: "msat",
    valueType: "number",
  },
  {
    heading: "Time Lock Delta",
    type: "NumericDoubleCell",
    key: "timeLockDelta",
    key2: "remoteTimeLockDelta",
    valueType: "number",
  },
  {
    heading: "Remote Time Lock Delta",
    type: "NumericDoubleCell",
    key: "remoteTimeLockDelta",
    valueType: "number",
  },
  {
    heading: "LND Short Channel ID",
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
    heading: "Current BlockHeight",
    type: "NumericCell",
    key: "currentBlockHeight",
    valueType: "number",
  },
  {
    heading: "Funding Transaction",
    type: "LongTextCell",
    key: "fundingTransactionHash",
    valueType: "string",
  },
  {
    heading: "Funding BlockHeight",
    type: "NumericCell",
    key: "fundingBlockHeight",
    valueType: "number",
  },
  {
    heading: "Funding BlockHeight Delta",
    type: "NumericCell",
    key: "fundingBlockHeightDelta",
    valueType: "number",
  },
  {
    heading: "Funding Date",
    key: "fundedOn",
    type: "DateCell",
    valueType: "date",
  },
  {
    heading: "Funding Date Delta (Seconds)",
    type: "DurationCell",
    key: "fundedOnSecondsDelta",
    valueType: "duration",
  },
  {
    heading: "Closing Transaction",
    type: "LongTextCell",
    key: "closingTransactionHash",
    valueType: "string",
  },
  {
    heading: "Closing BlockHeight",
    type: "NumericCell",
    key: "closingBlockHeight",
    valueType: "number",
  },
  {
    heading: "Closing BlockHeight Delta",
    type: "NumericCell",
    key: "closingBlockHeightDelta",
    valueType: "number",
  },
  {
    heading: "Closing Date",
    key: "closedOn",
    valueType: "date",
    type: "DateCell",
  },
  {
    heading: "Closing Date Delta (Seconds)",
    type: "DurationCell",
    key: "closedOnSecondsDelta",
    valueType: "duration",
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
    heading: "Local Channel Reserve",
    type: "NumericDoubleCell",
    key: "localChanReserveSat",
    key2: "remoteChanReserveSat",
    valueType: "number",
    suffix: "sat",
  },
  {
    heading: "Remote Channel Reserve",
    type: "NumericCell",
    key: "remoteChanReserveSat",
    valueType: "number",
    suffix: "sat",
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
  "peerAlias",
  "active",
  "balance",
  "feeRateMilliMsat",
  "feeBaseMsat",
  "minHtlcMsat",
  "maxHtlcMsat",
  "shortChannelId",
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
  dirty: true,
  view: {
    title: "Draft View",
    columns: DefaultChannelsColumns,
    sortBy: [ChannelsSortTemplate],
  },
};

const sortableColumns: Array<keyof channel> = [
  "active",
  "gauge",
  "peerAlias",
  "shortChannelId",
  "feeRateMilliMsat",
  "remoteBalance",
  "localBalance",
  "capacity",
  "totalSatoshisSent",
  "totalSatoshisReceived",
  "unsettledBalance",
  "commitFee",
  "feeBaseMsat",
  "minHtlcMsat",
  "maxHtlcMsat",
  "remoteChanReserveSat",
  "localChanReserveSat",
  "remoteFeeRateMilliMsat",
  "remoteFeeBaseMsat",
  "remoteMinHtlcMsat",
  "remoteMaxHtlcMsat",
  "nodeName",
];

export const SortableChannelsColumns = AllChannelsColumns.filter((c) => sortableColumns.includes(c.key));
