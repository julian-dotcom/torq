import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { TableResponses, ViewResponse } from "./types";
import { uuid } from "uuidv4";
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

// type ViewSliceState = {
//   pages: {
//     [key in keyof AllViewsResponse]: { selectedView: number; views: Array<ViewResponse<keyof TableResponses>> };
//   };
// };

const initialState = {
  pages: {
    forwards: {
      selected: DefaultForwardsView.uuid,
      views: [DefaultForwardsView],
      // persistedViews: <Array<ViewResponse<Forward>>>[],
    },
    onChain: {
      selected: DefaultOnChainView.uuid,
      views: [DefaultOnChainView],
      // persistedViews: <Array<ViewResponse<OnChainTx>>>[],
    },
    payments: {
      selected: DefaultPaymentView.uuid,
      views: [DefaultPaymentView],
      // persistedViews: <Array<ViewResponse<Payment>>>[],
    },
    invoices: {
      selected: DefaultInvoiceView.uuid,
      views: [DefaultInvoiceView],
      // persistedViews: <Array<ViewResponse<Invoice>>>[],
    },
    channel: {
      selected: DefaultChannelsView.uuid,
      views: [DefaultChannelsView],
      // persistedViews: <Array<ViewResponse<channel>>>[],
    },
    tags: {
      selected: DefaultTagsView.uuid,
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
      // Change the UUID to a new one
      view.uuid = uuid();
      const views = state.pages[view.page].views;
      state.pages[view.page].views = <Array<ViewResponse<TableResponses>>>[...views, view];
    },
    updateViewTitle: (
      state: ViewSliceState,
      action: PayloadAction<{ page: ViewSliceStatePages; uuid: string; title: string }>
    ) => {
      const { page, uuid, title } = action.payload;
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
      state.pages[page].views[viewIndex].view.title = title;
    },
    updateSelectedView: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        uuid: string;
      }>
    ) => {
      const { page, uuid } = actions.payload;
      state.pages[page].selected = uuid;
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
        uuid: string;
      }>
    ) => {
      const { page, uuid } = actions.payload;
      if (state.pages[page].views.length !== 1) {
        const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
        state.pages[page].views = state.pages[page].views.splice(viewIndex, 1);
        if (uuid === state.pages[page].selected) {
          state.pages[page].selected = state.pages[page].views[0].uuid;
        }
      }
    },
    // --------------------- Columns ---------------------
    addColumn: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        uuid: string;
        newColumn: ColumnMetaData<TableResponses>;
      }>
    ) => {
      const { page, uuid, newColumn } = actions.payload;
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
      const columns = state.pages[page].views[viewIndex].view.columns;
      if (!columns) {
        state.pages[page].views[viewIndex].view.columns = [...columns, newColumn];
      } else {
        state.pages[page].views[viewIndex].view.columns = [newColumn];
      }
    },
    updateColumn: (
      state: ViewSliceState,
      actions: PayloadAction<{
        page: ViewSliceStatePages;
        uuid: string;
        columnIndex: number;
        columnUpdate: Partial<ColumnMetaData<TableResponses>>;
      }>
    ) => {
      const { page, uuid, columnIndex, columnUpdate } = actions.payload;
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
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
        uuid: string;
        fromIndex: number;
        toIndex: number;
      }>
    ) => {
      const { page, uuid, fromIndex, toIndex } = actions.payload;
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
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
        uuid: string;
        columnIndex: number;
      }>
    ) => {
      const { page, uuid, columnIndex } = actions.payload;
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
      const columns = state.pages[page].views[viewIndex].view.columns;
      if (columns) {
        state.pages[page].views[viewIndex].view.columns = columns.splice(columnIndex, 1);
      }
    },
    // --------------------- Filters ---------------------
    updateFilters: (
      state: ViewSliceState,
      actions: PayloadAction<{ page: ViewSliceStatePages; uuid: string; filterUpdate: any }>
    ) => {
      const { page, uuid, filterUpdate } = actions.payload;
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
      state.pages[page].views[viewIndex].view.filters = actions.payload.filterUpdate;
    },
    // --------------------- Sort ---------------------
    addSortBy: (
      state: ViewSliceState,
      actions: PayloadAction<{ page: ViewSliceStatePages; uuid: string; sortBy: OrderBy }>
    ) => {
      const { page, uuid, sortBy } = actions.payload;
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
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
        uuid: string;
        sortByUpdate: OrderBy;
        sortByIndex: number;
      }>
    ) => {
      const { page, uuid, sortByUpdate, sortByIndex } = actions.payload;
      // Find the view
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
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
        uuid: string;
        sortByIndex: number;
      }>
    ) => {
      const { page, uuid, sortByIndex } = actions.payload;
      // Find the view
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
      // Find the current array of sort by
      const currentSortBy = state.pages[page].views[viewIndex].view.sortBy;
      // Delete the sort by if it exists
      if (currentSortBy) {
        state.pages[page].views[viewIndex].view.sortBy = currentSortBy.splice(sortByIndex, 1);
      }
    },
    updateSortByOrder: (
      state,
      actions: PayloadAction<{ fromIndex: number; toIndex: number; page: ViewSliceStatePages; uuid: string }>
    ) => {
      const { page, uuid, fromIndex, toIndex } = actions.payload;
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
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
        uuid: string;
        groupByUpdate: "channels" | "peers";
      }>
    ) => {
      const { page, uuid, groupByUpdate } = actions.payload;
      const viewIndex = state.pages[page].views.findIndex((view) => view.uuid === uuid);
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
      if (payload !== null) {
        // Only add the views to the viewSlice if it is not already populated.
        // This is because we want the views to be persisted first when the user chooses to save the view.
        if (
          payload.forwards &&
          state.pages["forwards"].views[0].uuid === DefaultForwardsView.uuid &&
          state.pages["forwards"].views[0].id === undefined
        ) {
          state.pages["forwards"].views = payload.forwards;
          state.pages["forwards"].selected = payload.forwards[0].uuid;
        }

        if (
          payload.onChain &&
          state.pages["onChain"].views[0].uuid === DefaultOnChainView.uuid &&
          state.pages["onChain"].views[0].id === undefined
        ) {
          state.pages["onChain"].views = payload.onChain;
          state.pages["onChain"].selected = payload.onChain[0].uuid;
        }

        if (
          payload.payments &&
          state.pages["payments"].views[0].uuid === DefaultPaymentView.uuid &&
          state.pages["payments"].views[0].id === undefined
        ) {
          state.pages["payments"].views = payload.payments;
          state.pages["payments"].selected = payload.payments[0].uuid;
        }

        if (
          payload.invoices &&
          state.pages["invoices"].views[0].uuid === DefaultInvoiceView.uuid &&
          state.pages["invoices"].views[0].id === undefined
        ) {
          state.pages["invoices"].views = payload.invoices;
          state.pages["invoices"].selected = payload.invoices[0].uuid;
        }

        if (
          payload.channel &&
          state.pages["channel"].views[0].uuid === DefaultChannelsView.uuid &&
          state.pages["channel"].views[0].id === undefined
        ) {
          state.pages["channel"].views = payload.channel;
          state.pages["channel"].selected = payload.channel[0].uuid;
        }

        if (
          payload.tags &&
          state.pages["tags"].views[0].uuid === DefaultTagsView.uuid &&
          state.pages["tags"].views[0].id === undefined
        ) {
          state.pages["tags"].views = payload.tags;
          state.pages["tags"].selected = payload.tags[0].uuid;
        }
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
      const index = views.findIndex((view) => view.uuid === payload.uuid);
      views[index] = payload;
      state.pages[payload.page].views = views;
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
  return (
    state.viewsSlice.pages[page].views.find((view) => view.uuid === state.viewsSlice.pages[page].selected) ||
    DefaultForwardsView
  );
};

export const selectChannelView = (state: RootState) => {
  const page = "channel";
  return (
    state.viewsSlice.pages[page].views.find((view) => view.uuid === state.viewsSlice.pages[page].selected) ||
    DefaultChannelsView
  );
};

export const selectPaymentsView = (state: RootState) => {
  const page = "payments";
  return (
    state.viewsSlice.pages[page].views.find((view) => view.uuid === state.viewsSlice.pages[page].selected) ||
    DefaultPaymentView
  );
};

export const selectInvoicesView = (state: RootState) => {
  const page = "invoices";
  return (
    state.viewsSlice.pages[page].views.find((view) => view.uuid === state.viewsSlice.pages[page].selected) ||
    DefaultInvoiceView
  );
};

export const selectOnChainView = (state: RootState) => {
  const page = "onChain";
  return (
    state.viewsSlice.pages[page].views.find((view) => view.uuid === state.viewsSlice.pages[page].selected) ||
    DefaultOnChainView
  );
};

export const selectViews = (state: RootState) => (page: ViewSliceStatePages) => state.viewsSlice.pages[page];

export const getSelectedView =
  (state: RootState) =>
  <T>(page: ViewSliceStatePages) => {
    const vieweIndex = state.viewsSlice.pages[page].views.findIndex(
      (view) => view.uuid === state.viewsSlice.pages[page].selected
    );
    return state.viewsSlice.pages[page].views[vieweIndex] as ViewResponse<T>;
  };

export default viewsSlice.reducer;
