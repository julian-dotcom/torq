import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { ColumnMetaData } from "features/table/Table";
//import { SortByOptionType } from "features/sidebar/sections/sort/SortSectionOld";
import { RootState } from "store/store";
import { ViewInterface } from "features/table/Table";
import { torqApi } from "apiSlice";

// export interface ViewInterface {
//   title: string;
//   id?: number;
//   saved: boolean;
//   filters?: any;
//   columns?: ColumnMetaData[];
//   order?: SortByOptionType[];
// }


export interface TablePaymentsState {
  paymentViews: ViewInterface[];
  selectedViewIndex: number;
}

export const AllPaymentsColumns: Array<ColumnMetaData> = [
  { key: "date", heading: "Date", type: "DateCell", valueType: "date" },
  { key: "status", heading: "Status", type: "TextCell", valueType: "array" },
  { key: "value", heading: "Value", type: "NumericCell", valueType: "number" },
  { key: "fee", heading: "Fee", type: "NumericCell", valueType: "number" },
  { key: "ppm", heading: "PPM", type: "NumericCell", valueType: "number" },
  { key: "is_rebalance", heading: "Rebalance", type: "BooleanCell", valueType: "boolean" },
  { key: "seconds_in_flight", heading: "Seconds In Flight", type: "DurationCell", valueType: "duration" },
  { key: "failure_reason", heading: "Failure Reason", type: "TextCell", valueType: "array" },
  { key: "is_mpp", heading: "MPP", type: "BooleanCell", valueType: "boolean" },
  { key: "count_failed_attempts", heading: "Failed Attempts", type: "NumericCell", valueType: "number" },
  { key: "count_successful_attempts", heading: "Successful Attempts", type: "NumericCell", valueType: "number" },
  { key: "destination_pub_key", heading: "Destination", type: "TextCell", valueType: "string" },
  { key: "payment_hash", heading: "Payment Hash", type: "TextCell", valueType: "string" },
  { key: "payment_preimage", heading: "Payment Preimage", type: "TextCell", valueType: "string" },
];

export const ActivePaymentsColumns: Array<ColumnMetaData> = [
  { key: "date", heading: "Date", type: "DateCell", valueType: "date" },
  { key: "status", heading: "Status", type: "TextCell", valueType: "array" },
  { key: "value", heading: "Value", type: "NumericCell", valueType: "number" },
  { key: "fee", heading: "Fee", type: "NumericCell", valueType: "number" },
  { key: "ppm", heading: "PPM", type: "NumericCell", valueType: "number" },
  { key: "is_rebalance", heading: "Rebalance", type: "BooleanCell", valueType: "boolean" },
  { key: "seconds_in_flight", heading: "Seconds In Flight", type: "DurationCell", valueType: "duration" },
  { key: "failure_reason", heading: "Failure Reason", type: "TextCell", valueType: "array" },
  { key: "count_failed_attempts", heading: "Failed Attempts", type: "NumericCell", valueType: "number" },
];

const initialState: TablePaymentsState = {
  selectedViewIndex: 0,
  paymentViews: [
    {
      title: "Default View",
      saved: true,
      columns: ActivePaymentsColumns,
      sortBy: [],
      page: 'payments'
    },
  ],
  // newPayments: new Map<string, Array<NewPaymentResponse>>(),
};

export const DefaultView: ViewInterface = {
  title: "Untitled View",
  saved: true,
  columns: ActivePaymentsColumns,
  page: 'payments',
  sortBy: []
};

export const paymentsSlice = createSlice({
  name: "payments",
  initialState,
  reducers: {
    updatePaymentsFilters: (state, actions: PayloadAction<{ filters: any }>) => {
      state.paymentViews[state.selectedViewIndex].filters = actions.payload.filters;
    },
    updateColumns: (state, actions: PayloadAction<{ columns: ColumnMetaData[] }>) => {
      state.paymentViews[state.selectedViewIndex].columns = actions.payload.columns;
    },
    updateViews: (state, actions: PayloadAction<{ views: ViewInterface[]; index: number }>) => {
      state.paymentViews = actions.payload.views;
      state.selectedViewIndex = actions.payload.index;
    },
    updateViewsOrder: (state, actions: PayloadAction<{ views: ViewInterface[]; index: number }>) => {
      state.paymentViews = actions.payload.views;
      state.selectedViewIndex = actions.payload.index;
    },
    deleteView: (state, actions: PayloadAction<{ view: ViewInterface; index: number }>) => {
      state.paymentViews = [
        ...state.paymentViews.slice(0, actions.payload.index),
        ...state.paymentViews.slice(actions.payload.index + 1, state.paymentViews.length),
      ];
      state.selectedViewIndex = 0;
    },
    updateSelectedView: (state, actions: PayloadAction<{ index: number }>) => {
      state.selectedViewIndex = actions.payload.index;
    },
    // updateNewPaymentResponse: (state, actions: PayloadAction<{ reqId: string; payment: NewPaymentResponse }>) => {
    //   const payment = actions.payload.payment;
    //   const reqId = actions.payload.reqId;
    //   const paymentArray = state.newPayments.get(reqId);
    //   if (paymentArray) {
    //     paymentArray.push(payment);
    //   } else {
    //     state.newPayments.set(reqId, [payment]);
    //   }
    // },
  },
  // The `extraReducers` field lets the slice handle actions defined elsewhere,
  // including actions generated by createAsyncThunk or in other slices.
  extraReducers: (builder) => {
    builder.addMatcher(
      (action) => {
        return (
          ["payments/updateOnChainFilters", "payments/updateColumns"].findIndex(
            (item) => action.type === item
          ) !== -1
        );
      },
      (state, _) => {
        state.paymentViews[state.selectedViewIndex].saved = false;
      }
    );

    builder.addMatcher(torqApi.endpoints.createTableView.matchFulfilled, (state, { payload }) => {
      state.paymentViews[payload.index] = {
        ...payload.view.view,
        id: payload.view.id,
      };
      state.selectedViewIndex = payload.index;
    });

    builder.addMatcher(torqApi.endpoints.deleteTableView.matchFulfilled, (state, { payload }) => {
      state.paymentViews = [
        ...state.paymentViews.slice(0, payload.index),
        ...state.paymentViews.slice(payload.index + 1, state.paymentViews.length),
      ];
      state.selectedViewIndex = 0;
    });

    builder.addMatcher(torqApi.endpoints.getTableViews.matchFulfilled, (state, { payload }) => {
      if (payload !== null) {
        state.paymentViews = payload.map((view: { id: number; view: ViewInterface }) => {
          return { ...view.view, id: view.id };
        });
      }
    });

    builder.addMatcher(torqApi.endpoints.updateTableView.matchFulfilled, (state, { payload }) => {
      const view = state.paymentViews.find((v) => v.id === payload.id);
      if (view) {
        view.saved = true;
      }
    });
  },
});

export const {
  updatePaymentsFilters,
  updateColumns,
  updateViews,
  updateViewsOrder,
  deleteView,
  updateSelectedView,
} = paymentsSlice.actions;

export const selectPaymentsFilters = (state: RootState) => {
  return state.payments.paymentViews[0].filters;
};

export const selectActiveColumns = (state: RootState) => {
  return state.payments.paymentViews[0].columns || [];
};

export const selectAllColumns = (_: RootState) => AllPaymentsColumns;
export const selectViews = (state: RootState) => state.payments.paymentViews;
export const selectCurrentView = (state: RootState) => state.payments.paymentViews[state.payments.selectedViewIndex];
export const selectedViewIndex = (state: RootState) => state.payments.selectedViewIndex;

export default paymentsSlice.reducer;
