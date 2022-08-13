import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { ColumnMetaData } from "../table/Table";
import { SortByOptionType } from "../sidebar/sections/sort/SortSectionOld";
import { RootState } from "../../store/store";

export interface ViewInterface {
  title: string;
  id?: number;
  saved: boolean;
  filters?: any;
  columns?: ColumnMetaData[];
  order?: SortByOptionType[];
}

export interface initialStateProps {
  paymentViews: ViewInterface[];
}

const initialState: initialStateProps = {
  paymentViews: [
    {
      title: "Default View",
      saved: true,
    },
  ],
};

export const paymentsSlice = createSlice({
  name: "payments",
  initialState,
  reducers: {
    updatePaymentsFilters: (state, actions: PayloadAction<{ filters: any }>) => {
      state.paymentViews[0].filters = actions.payload.filters;
    },
  },
});

export const { updatePaymentsFilters } = paymentsSlice.actions;

export const selectPaymentsFilters = (state: RootState) => {
  return state.payments.paymentViews[0].filters;
};

export default paymentsSlice.reducer;
