import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { ColumnMetaData } from "features/table/Table";
import { SortByOptionType } from "features/sidebar/sections/sort/SortSectionOld";
import { RootState } from "store/store";

export interface ViewInterface {
  title: string;
  id?: number;
  saved: boolean;
  filters?: any;
  columns?: ColumnMetaData[];
  order?: SortByOptionType[];
}

export interface initialStateProps {
  invoiceViews: ViewInterface[];
}

export const AllInvoicesColumns: Array<ColumnMetaData> = [
  { key: "creation_date", heading: "Creation Date", type: "DateCell", valueType: "date" },
  { key: "settle_date", heading: "Settle Date", type: "DateCell", valueType: "date" },
  { key: "invoice_state", heading: "State", type: "TextCell", valueType: "array" },
  { key: "amt_paid", heading: "Paid Amount", type: "NumericCell", valueType: "number" },
  { key: "memo", heading: "memo", type: "TextCell", valueType: "string" },
  { key: "value", heading: "Invoice Amount", type: "NumericCell", valueType: "number" },
  { key: "is_rebalance", heading: "Rebalance", type: "BooleanCell", valueType: "boolean" },
  { key: "is_keysend", heading: "Keysend", type: "BooleanCell", valueType: "boolean" },
  { key: "destination_pub_key", heading: "Destination", type: "TextCell", valueType: "string" },
  { key: "is_amp", heading: "AMP", type: "BooleanCell", valueType: "boolean" },
  { key: "fallback_addr", heading: "Fallback Address", type: "TextCell", valueType: "string" },
  { key: "payment_addr", heading: "Payment Address", type: "TextCell", valueType: "string" },
  { key: "payment_request", heading: "Payment Request", type: "TextCell", valueType: "string" },
  { key: "private", heading: "Private", type: "BooleanCell", valueType: "boolean" },
  { key: "r_hash", heading: "Hash", type: "TextCell", valueType: "string" },
  { key: "r_preimage", heading: "Preimage", type: "TextCell", valueType: "string" },
  { key: "expiry", heading: "Expiry", type: "NumericCell", valueType: "number" },
  { key: "cltv_expiry", heading: "CLTV Expiry", type: "NumericCell", valueType: "number" },
  { key: "updated_on", heading: "Updated On", type: "DateCell", valueType: "date" },
];

export const ActiveInvoicesColumns = AllInvoicesColumns.filter(({ key }) =>
  [
    "creation_date",
    "settle_date",
    "invoice_state",
    "amt_paid",
    "memo",
    "value",
    "is_rebalance",
    "is_keysend",
    "destination_pub_key",
  ].includes(key)
);

const initialState: initialStateProps = {
  invoiceViews: [
    {
      title: "Default View",
      saved: true,
      columns: ActiveInvoicesColumns,
    },
  ],
};

export const invoicesSlice = createSlice({
  name: "invoices",
  initialState,
  reducers: {
    updateInvoicesFilters: (state, actions: PayloadAction<{ filters: any }>) => {
      state.invoiceViews[0].filters = actions.payload.filters;
    },
    updateColumns: (state, actions: PayloadAction<{ columns: ColumnMetaData[] }>) => {
      state.invoiceViews[0].columns = actions.payload.columns;
    },
  },
});

export const { updateInvoicesFilters, updateColumns } = invoicesSlice.actions;

export const selectInvoicesFilters = (state: RootState) => {
  return state.invoices.invoiceViews[0].filters;
};

export const selectActiveColumns = (state: RootState) => {
  return state.invoices.invoiceViews[0].columns || [];
};

export const selectAllColumns = (_: RootState) => AllInvoicesColumns;

export default invoicesSlice.reducer;
