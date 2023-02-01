import { ColumnMetaData } from "features/table/types";
import { OnChainTx } from "features/transact/OnChain/types";
import { ViewResponse } from "features/viewManagement/types";
import { FilterInterface } from "features/sidebar/sections/filter/filter";
import {
  AllOnChainTransactionsColumns,
  OnChainTransactionsSortableColumns,
  OnChainTransactionsFilterableColumns
} from "features/transact/OnChain/onChainColumns";

const defaultColumns: Array<keyof OnChainTx> = [
  "date",
  "amount",
  "totalFees",
  "lndShortChanId",
  "lndTxTypeLabel",
  "txHash",
  "label",
];

export const DefaultOnChainColumns = AllOnChainTransactionsColumns.filter((c) => defaultColumns.includes(c.key));

export const SortableOnChainColumns = AllOnChainTransactionsColumns.filter((column: ColumnMetaData<OnChainTx>) =>
  OnChainTransactionsSortableColumns.includes(column.key)
);

export const FilterableOnChainColumns = AllOnChainTransactionsColumns.filter((column: ColumnMetaData<OnChainTx>) =>
  OnChainTransactionsFilterableColumns.includes(column.key)
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
  dirty: true,
  view: {
    title: "Draft View",
    columns: DefaultOnChainColumns,
    sortBy: [OnChainSortTemplate],
  },
};
