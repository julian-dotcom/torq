import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { ColumnMetaData } from "features/table/Table";
import { SortByOptionType } from "features/sidebar/sections/sort/SortSectionOld";
import { RootState } from "features/../store/store";

export interface ViewInterface {
  title: string;
  id?: number;
  saved: boolean;
  filters?: any;
  columns?: ColumnMetaData[];
  order?: SortByOptionType[];
}

export interface initialStateProps {
  onChainViews: ViewInterface[];
}

export const AllOnChainColumns: Array<ColumnMetaData> = [
  { key: "date", heading: "Date", type: "DateCell", valueType: "date" },
  { key: "amount", heading: "Amount", type: "NumericCell", valueType: "number" },
  { key: "total_fees", heading: "Fees", type: "NumericCell", valueType: "number" },
  { key: "tx_hash", heading: "Tx Hash", type: "TextCell", valueType: "string" },
  { key: "lnd_short_chan_id", heading: "LND Short Channel ID", type: "TextCell", valueType: "string" },
  { key: "lnd_tx_type_label", heading: "LND Tx type label", type: "TextCell", valueType: "string" },
  { key: "dest_addresses_count", heading: "Destination Addresses Count", type: "NumericCell", valueType: "number" },
  { key: "label", heading: "Label", type: "TextCell", valueType: "string" },
];

export const ActiveOnChainColumns = AllOnChainColumns.filter(({ key }) =>
  ["date", "amount", "total_fees", "lnd_short_chan_id", "lnd_tx_type_label", "tx_hash", "label"].includes(key)
);

const initialState: initialStateProps = {
  onChainViews: [
    {
      title: "Default View",
      saved: true,
      columns: ActiveOnChainColumns,
    },
  ],
};

export const onChainSlice = createSlice({
  name: "onChain",
  initialState,
  reducers: {
    updateOnChainFilters: (state, actions: PayloadAction<{ filters: any }>) => {
      state.onChainViews[0].filters = actions.payload.filters;
    },
    updateColumns: (state, actions: PayloadAction<{ columns: ColumnMetaData[] }>) => {
      state.onChainViews[0].columns = actions.payload.columns;
    },
  },
});

export const { updateOnChainFilters, updateColumns } = onChainSlice.actions;

export const selectOnChainFilters = (state: RootState) => {
  return state.onChain.onChainViews[0].filters;
};

export const selectActiveColumns = (state: RootState) => {
  return state.onChain.onChainViews[0].columns || [];
};

export const selectAllColumns = (_: RootState) => AllOnChainColumns;

export default onChainSlice.reducer;
