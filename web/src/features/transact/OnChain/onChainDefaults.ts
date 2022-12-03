import { uuid } from "uuidv4";
import { ColumnMetaData } from "features/table/types";
import { OnChainTx } from "./types";
import { ViewResponse } from "features/viewManagement/types";
import { FilterInterface } from "features/sidebar/sections/filter/filter";

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

export const DefaultOnChainColumns = AllOnChainColumns.filter((c) => defaultColumns.includes(c.key));

const sortableColumnsKeys: Array<keyof OnChainTx> = [
  "date",
  "destAddresses",
  "destAddressesCount",
  "amount",
  "totalFees",
  "label",
  "lndTxTypeLabel",
  "lndShortChanId",
];

export const SortableOnChainColumns = AllOnChainColumns.filter((column: ColumnMetaData<OnChainTx>) =>
  sortableColumnsKeys.includes(column.key)
);

export const OnChainSortTemplate: { key: keyof OnChainTx; direction: "desc" | "asc" } = {
  key: "date",
  direction: "desc",
};

export const OnChainFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number",
  parameter: 0,
  key: "amount",
};

export const DefaultOnChainView: ViewResponse<OnChainTx> = {
  page: "onChain",
  uuid: uuid(),
  dirty: true,
  view: {
    title: "Untitled View",
    columns: DefaultOnChainColumns,
    sortBy: [OnChainSortTemplate],
  },
};
