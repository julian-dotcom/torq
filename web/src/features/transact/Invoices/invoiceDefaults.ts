// import { uuid } from "uuidv4";
import { ViewResponse } from "features/viewManagement/types";
import { Invoice } from "features/transact/Invoices/invoiceTypes";
import {
  AllInvoicesColumns,
  InvoicesSortableColumns,
  InvoicesFilterableColumns,
} from "features/transact/Invoices/invoicesColumns.generated";
import { ColumnMetaData } from "features/table/types";
import { FilterInterface } from "features/sidebar/sections/filter/filter";

const defaultKeys: Array<keyof Invoice> = [
  "creationDate",
  "settleDate",
  "invoiceState",
  "amtPaid",
  "memo",
  "value",
  "isRebalance",
  "isKeysend",
  "destinationPubKey",
];

export const DefaultInvoicesColumns = AllInvoicesColumns.filter(({ key }) => defaultKeys.includes(key));

export const SortableInvoiceColumns = AllInvoicesColumns.filter((column: ColumnMetaData<Invoice>) => {
  return InvoicesSortableColumns.includes(column.key);
});

export const FilterableInvoiceColumns = AllInvoicesColumns.filter((column: ColumnMetaData<Invoice>) => {
  return InvoicesFilterableColumns.includes(column.key);
});

export const InvoiceSortTemplate: { key: keyof Invoice; direction: "desc" | "asc" } = {
  key: "creationDate",
  direction: "desc",
};

export const InvoiceFilterTemplate: FilterInterface = {
  key: "value",
  funcName: "gte",
  parameter: 0,
  category: "number",
};

export const DefaultInvoiceView: ViewResponse<Invoice> = {
  page: "invoices",
  dirty: true,
  view: {
    title: "Draft View",
    columns: DefaultInvoicesColumns,
    sortBy: [InvoiceSortTemplate],
  },
};
