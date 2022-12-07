import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { TableResponses, ViewResponse } from "./types";
import { viewApi } from "./viewsApiSlice";
import { DefaultForwardsView } from "../forwards/forwardsDefaults";
import { RootState } from "../../store/store";
import { DefaultOnChainView } from "../transact/OnChain/onChainDefaults";
import { DefaultPaymentView } from "../transact/Payments/paymentDefaults";
import { DefaultInvoiceView } from "../transact/Invoices/invoiceDefaults";
import { DefaultChannelsView } from "../channels/channelsDefaults";
import { DefaultTagsView } from "../../pages/tagsPage/tagsDefaults";
import { ColumnMetaData } from "../table/types";
import { OrderBy } from "../sidebar/sections/sort/SortSection";

const initialState = {
  initiated: false,
  pages: {
    forwards: {
      selected: 0,
      views: [DefaultForwardsView],
      // persistedViews: <Array<ViewResponse<Forward>>>[],
    },
    onChain: {
      selected: 0,
      views: [DefaultOnChainView],
      // persistedViews: <Array<ViewResponse<OnChainTx>>>[],
    },
    payments: {
      selected: 0,
      views: [DefaultPaymentView],
      // persistedViews: <Array<ViewResponse<Payment>>>[],
    },
    invoices: {
      selected: 0,
      views: [DefaultInvoiceView],
      // persistedViews: <Array<ViewResponse<Invoice>>>[],
    },
    channel: {
      selected: 0,
      views: [DefaultChannelsView],
      // persistedViews: <Array<ViewResponse<channel>>>[],
    },
    tags: {
      selected: 0,
      views: [DefaultTagsView],
      // persistedViews: <Array<ViewResponse<tag>>>[],
    },
  },
};

type ViewSliceState = typeof initialState;
export type ViewSliceStatePages = keyof typeof initialState.pages;

export const viewsSlice = createSlice({
  name: "views",
  initialState,
  reducers: {
    // --------------------- Views ---------------------
    addView: (state: ViewSliceState, action: PayloadAction<{ view: ViewResponse<TableResponses> }>) => {
      const { view } = action.payload;
      const views = state.pages[view.page].views;
      state.pages[view.page].views = <Array<ViewResponse<TableResponses>>>[...views, view];
    },
    updateViewTitle: (
      state: ViewSliceState,
      action: PayloadAction<{ page: ViewSliceStatePages; viewIndex: number; title: string }>
    ) => {
      const { page, viewIndex, title } = action.payload;
      state.pages[page].views[viewIndex].view.title = title;
    },
    updateSelectedView: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        viewIndex: number;
      }>
    ) => {
      const { page, viewIndex } = actions.payload;
      state.pages[page].selected = viewIndex;
    },
    updateViewsOrder: (
      state,
      actions: PayloadAction<{ page: keyof typeof initialState.pages; fromIndex: number; toIndex: number }>
    ) => {
      const { page, fromIndex, toIndex } = actions.payload;
      const views = state.pages[page].views.slice();
      const view = state.pages[page].views[fromIndex];
      views.splice(fromIndex, 1);
      views.splice(toIndex, 0, view);
      state.pages[page].views = views;
    },
    deleteView: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        viewIndex: number;
      }>
    ) => {
      const { page, viewIndex } = actions.payload;
      if (state.pages[page].views.length !== 1) {
        state.pages[page].views = state.pages[page].views.splice(viewIndex, 1);
        if (state.pages[page].selected === viewIndex) {
          state.pages[page].selected = 0;
        }
      }
    },
    // --------------------- Columns ---------------------
    addColumn: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        viewIndex: number;
        newColumn: ColumnMetaData<TableResponses>;
      }>
    ) => {
      const { page, viewIndex, newColumn } = actions.payload;
      const columns = state.pages[page].views[viewIndex].view.columns;
      if (!columns) {
        state.pages[page].views[viewIndex].view.columns = [newColumn];
      } else {
        state.pages[page].views[viewIndex].view.columns = [...(columns as Array<typeof newColumn>), newColumn];
      }
    },
    updateColumn: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        viewIndex: number;
        columnIndex: number;
        columnUpdate: Partial<ColumnMetaData<TableResponses>>;
      }>
    ) => {
      const { page, viewIndex, columnIndex, columnUpdate } = actions.payload;
      const columns = state.pages[page].views[viewIndex].view.columns;
      const column = columns[columnIndex];
      if (columns) {
        state.pages[page].views[viewIndex].view.columns[columnIndex] = { ...column, ...columnUpdate };
      }
    },
    updateColumnsOrder: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        viewIndex: number;
        fromIndex: number;
        toIndex: number;
      }>
    ) => {
      const { page, viewIndex, fromIndex, toIndex } = actions.payload;
      const columns = state.pages[page].views[viewIndex].view.columns.slice();
      const column = columns[fromIndex];
      columns.splice(fromIndex, 1);
      columns.splice(toIndex, 0, column as ColumnMetaData<TableResponses>);
      state.pages[page].views[viewIndex].view.columns = columns;
    },
    deleteColumn: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        viewIndex: number;
        columnIndex: number;
      }>
    ) => {
      const { page, viewIndex, columnIndex } = actions.payload;
      const columns = state.pages[page].views[viewIndex].view.columns;
      if (columns) {
        state.pages[page].views[viewIndex].view.columns = columns.splice(columnIndex, 1);
      }
    },
    // --------------------- Filters ---------------------
    updateFilters: (
      state: ViewSliceState,
      actions: PayloadAction<{ page: ViewSliceStatePages; viewIndex: number; filterUpdate: any }>
    ) => {
      const { page, viewIndex, filterUpdate } = actions.payload;
      state.pages[page].views[viewIndex].view.filters = filterUpdate;
    },
    // --------------------- Sort ---------------------
    addSortBy: (
      state: ViewSliceState,
      actions: PayloadAction<{ page: ViewSliceStatePages; viewIndex: number; sortBy: OrderBy }>
    ) => {
      const { page, viewIndex, sortBy } = actions.payload;
      const currentSortBy = state.pages[page].views[viewIndex].view.sortBy;
      if (currentSortBy) {
        state.pages[page].views[viewIndex].view.sortBy = [...currentSortBy, sortBy];
      } else {
        state.pages[page].views[viewIndex].view.sortBy = <Array<OrderBy>>[sortBy];
      }
    },
    updateSortBy: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        viewIndex: number;
        sortByUpdate: OrderBy;
        sortByIndex: number;
      }>
    ) => {
      const { page, viewIndex, sortByUpdate, sortByIndex } = actions.payload;
      // Find the current array of sort by
      const currentSortBy = state.pages[page].views[viewIndex].view.sortBy;
      // If there is a current array of sort by update it
      if (currentSortBy) {
        currentSortBy[sortByIndex] = sortByUpdate;
        state.pages[page].views[viewIndex].view.sortBy = currentSortBy;
      }
      // TOOD: Add message
    },
    deleteSortBy: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        viewIndex: number;
        sortByIndex: number;
      }>
    ) => {
      const { page, viewIndex, sortByIndex } = actions.payload;
      // Find the current array of sort by
      const currentSortBy = state.pages[page].views[viewIndex].view.sortBy;
      // Delete the sort by if it exists
      if (currentSortBy) {
        state.pages[page].views[viewIndex].view.sortBy = currentSortBy.splice(sortByIndex, 1);
      }
    },
    updateSortByOrder: (
      state,
      actions: PayloadAction<{ fromIndex: number; toIndex: number; page: ViewSliceStatePages; viewIndex: number }>
    ) => {
      const { page, viewIndex, fromIndex, toIndex } = actions.payload;
      const currentSortBy = state.pages[page].views[viewIndex].view.sortBy;
      if (currentSortBy) {
        const sortBy = currentSortBy[fromIndex];
        currentSortBy.splice(fromIndex, 1);
        currentSortBy.splice(toIndex, 0, sortBy);
        state.pages[page].views[viewIndex].view.sortBy = currentSortBy;
      }
    },
    // --------------------- Group by ---------------------
    updateGroupBy: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        viewIndex: number;
        groupByUpdate: "channels" | "peers";
      }>
    ) => {
      const { page, viewIndex, groupByUpdate } = actions.payload;
      state.pages[page].views[viewIndex].view.groupBy = groupByUpdate;
    },
  },
  // // The `extraReducers` field lets the slice handle actions defined elsewhere,
  // // including actions generated by createAsyncThunk or in other slices.
  extraReducers: (builder) => {
    //   builder.addMatcher(
    //     (action) => {
    //       return (
    //         ["payments/updatePaymentsFilters", "payments/updateColumns"].findIndex((item) => action.type === item) !== -1
    //       );
    //     },
    //     (state, _) => {
    //       state.paymentViews[state.selectedViewIndex].saved = false;
    //     }
    //   );
    //
    //   builder.addMatcher(torqApi.endpoints.createTableView.matchFulfilled, (state, { payload }) => {
    //     state.paymentViews[payload.index] = {
    //       ...payload.view.view,
    //       id: payload.view.id,
    //     };
    //     state.selectedViewIndex = payload.index;
    //   });
    //
    //   builder.addMatcher(torqApi.endpoints.deleteTableView.matchFulfilled, (state, { payload }) => {
    //     state.paymentViews = [
    //       ...state.paymentViews.slice(0, payload.index),
    //       ...state.paymentViews.slice(payload.index + 1, state.paymentViews.length),
    //     ];
    //     state.selectedViewIndex = 0;
    //   });
    // On initial load of the app, load the views from the backend
    builder.addMatcher(viewApi.endpoints.getTableViews.matchFulfilled, (state, { payload }) => {
      if (payload !== null && state.initiated === false) {
        // Only add the views to the viewSlice if it is not already populated.
        // This is because we want the views to be persisted first when the user chooses to save the view.
        if (payload.forwards) {
          state.pages["forwards"].views = payload.forwards;
        }

        if (payload.onChain) {
          state.pages["onChain"].views = payload.onChain;
        }

        if (payload.payments) {
          state.pages["payments"].views = payload.payments;
        }

        if (payload.invoices) {
          state.pages["invoices"].views = payload.invoices;
        }

        if (payload.channel) {
          state.pages["channel"].views = payload.channel;
        }

        if (payload.tags) {
          state.pages["tags"].views = payload.tags;
        }
        state.initiated = true;
      }
    });

    // Delete a view when it is deleted in the backend
    builder.addMatcher(viewApi.endpoints.deleteTableView.matchPending, (state, { meta }) => {
      const views = state.pages[meta.arg.originalArgs.page].views;
      const index = views.findIndex((view) => view.id === meta.arg.originalArgs.id);
      views.splice(index, 1);
      state.pages[meta.arg.originalArgs.page].views = views;
    });

    builder.addMatcher(viewApi.endpoints.updateTableView.matchFulfilled, (state, { payload }) => {
      const views = state.pages[payload.page].views;
      const index = views.findIndex((view) => view.id === payload.id);
      views[index] = payload;
      state.pages[payload.page].views = views;
    });

    builder.addMatcher(viewApi.endpoints.createTableView.matchFulfilled, (state, { payload }) => {
      const views = state.pages[payload.page].views;
      state.pages[payload.page].views = [...views, payload];
    });
  },
});

export const {
  addView,
  updateViewTitle,
  updateSelectedView,
  updateViewsOrder,
  deleteView,
  addColumn,
  updateColumn,
  updateColumnsOrder,
  deleteColumn,
  updateFilters,
  addSortBy,
  updateSortBy,
  deleteSortBy,
  updateSortByOrder,
  updateGroupBy,
} = viewsSlice.actions;

export const selectForwardsView = (state: RootState) => {
  const page = "forwards";
  const view = state.viewsSlice.pages[page].views[state.viewsSlice.pages[page].selected];
  return { viewResponse: view, selectedViewIndex: state.viewsSlice.pages[page].selected };
};

export const selectChannelView = (state: RootState) => {
  const page = "channel";
  const view = state.viewsSlice.pages[page].views[state.viewsSlice.pages[page].selected];
  return { viewResponse: view, selectedViewIndex: state.viewsSlice.pages[page].selected };
};

export const selectPaymentsView = (state: RootState) => {
  const page = "payments";
  const view = state.viewsSlice.pages[page].views[state.viewsSlice.pages[page].selected];
  return { viewResponse: view, selectedViewIndex: state.viewsSlice.pages[page].selected };
};

export const selectInvoicesView = (state: RootState) => {
  const page = "invoices";
  const view = state.viewsSlice.pages[page].views[state.viewsSlice.pages[page].selected];
  return { viewResponse: view, selectedViewIndex: state.viewsSlice.pages[page].selected };
};

export const selectOnChainView = (state: RootState) => {
  const page = "onChain";
  const view = state.viewsSlice.pages[page].views[state.viewsSlice.pages[page].selected];
  return { viewResponse: view, selectedViewIndex: state.viewsSlice.pages[page].selected };
};

export const selectViews = (state: RootState) => (page: ViewSliceStatePages) => {
  return state.viewsSlice.pages[page];
};

export const getSelectedView = (state: RootState) => (page: ViewSliceStatePages) => {
  return state.viewsSlice.pages[page].views[state.viewsSlice.pages[page].selected];
};

export default viewsSlice.reducer;
