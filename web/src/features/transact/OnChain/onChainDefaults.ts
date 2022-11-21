import { ColumnMetaData } from "../../table/types";
import { OnChainTx } from "./types";
import { ViewInterface } from "../../viewManagement/types";

export const AllOnChainColumns: Array<ColumnMetaData<OnChainTx>> = [
  { key: "date", heading: "Date", type: "DateCell", valueType: "date" },
  { key: "amount", heading: "Amount", type: "NumericCell", valueType: "number" },
  { key: "totalFees", heading: "Fees", type: "NumericCell", valueType: "number" },
  { key: "txHash", heading: "Tx Hash", type: "TextCell", valueType: "string" },
  { key: "lndShortChanId", heading: "LND Short Channel ID", type: "TextCell", valueType: "string" },
  { key: "lndTxTypeLabel", heading: "LND Tx type label", type: "TextCell", valueType: "string" },
  { key: "destAddressesCount", heading: "Destination Addresses Count", type: "NumericCell", valueType: "number" },
  { key: "label", heading: "Label", type: "TextCell", valueType: "string" },
];

const defaultColumns: Array<keyof OnChainTx> = [
  "date",
  "amount",
  "totalFees",
  "lndShortChanId",
  "lndTxTypeLabel",
  "txHash",
  "label",
];

export const ActiveOnChainColumns = AllOnChainColumns.filter((c) => defaultColumns.includes(c.key));

export const DefaultOnChainView: ViewInterface<OnChainTx> = {
  view_order: 0,
  title: "Untitled View",
  saved: true,
  columns: ActiveOnChainColumns,
  page: "onChain",
  sortBy: [],
};
