import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RootState } from "store/store";
import { TableResponses, ViewResponse } from "./types";
import { viewApi } from "./viewsApiSlice";
import { DefaultForwardsView } from "features/forwards/forwardsDefaults";
import { DefaultOnChainView } from "features/transact/OnChain/onChainDefaults";
import { DefaultPaymentView } from "features/transact/Payments/paymentDefaults";
import { DefaultInvoiceView } from "features/transact/Invoices/invoiceDefaults";
import { DefaultChannelsView } from "features/channels/channelsDefaults";
import { DefaultClosedChannelsView } from "features/channelsClosed/channelsClosedDefaults";
import { DefaultPendingChannelsView } from "features/channelsPending/channelsPendingDefaults";
import { DefaultTagsView } from "pages/tags/tagsPage/tagsDefaults";
import { ColumnMetaData } from "features/table/types";
import { OrderBy } from "features/sidebar/sections/sort/SortSection";
import { deserialiseQuery, SerialisableFilterQuery } from "features/sidebar/sections/filter/filter";
import { DefaultPeersView } from "features/peers/peersDefaults";

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
    channelsClosed: {
      selected: 0,
      views: [DefaultClosedChannelsView],
    },
    channelsPending: {
      selected: 0,
      views: [DefaultPendingChannelsView],
    },
    tags: {
      selected: 0,
      views: [DefaultTagsView],
      // persistedViews: <Array<ViewResponse<tag>>>[],
    },
    peers: {
      selected: 0,
      views: [DefaultPeersView],
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
      if (state.pages[page].views[viewIndex].view.title !== title) {
        state.pages[page].views[viewIndex].dirty = true;
        state.pages[page].views[viewIndex].view.title = title;
      }
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
        const views = state.pages[page].views.slice();
        state.pages[page].views = <Array<ViewResponse<TableResponses>>>[...views];
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
      state.pages[page].views[viewIndex].dirty = true;
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
      state.pages[page].views[viewIndex].dirty = true;
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
      state.pages[page].views[viewIndex].dirty = true;
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
        columns.splice(columnIndex, 1);
        state.pages[page].views[viewIndex].view.columns = columns;
      }
      state.pages[page].views[viewIndex].dirty = true;
    },
    // --------------------- Filters ---------------------
    updateFilters: (
      state: ViewSliceState,
      actions: PayloadAction<{ page: ViewSliceStatePages; viewIndex: number; filterUpdate: SerialisableFilterQuery }>
    ) => {
      const { page, viewIndex, filterUpdate } = actions.payload;
      const q = deserialiseQuery(filterUpdate);
      if (q.length > 0) {
        state.pages[page].views[viewIndex].view.filters = filterUpdate;
      } else {
        state.pages[page].views[viewIndex].view.filters = undefined;
      }
      state.pages[page].views[viewIndex].dirty = true;
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
      state.pages[page].views[viewIndex].dirty = true;
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
      state.pages[page].views[viewIndex].dirty = true;
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
        currentSortBy.splice(sortByIndex, 1);
        state.pages[page].views[viewIndex].view.sortBy = currentSortBy;
        state.pages[page].views[viewIndex].dirty = true;
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
        state.pages[page].views[viewIndex].dirty = true;
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
      state.pages[page].views[viewIndex].dirty = true;
    },
  },
  // // The `extraReducers` field lets the slice handle actions defined elsewhere,
  // // including actions generated by createAsyncThunk or in other slices.
  extraReducers: (builder) => {
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

        if (payload.channelsClosed) {
          state.pages["channelsClosed"].views = payload.channelsClosed;
        }

        if (payload.channelsPending) {
          state.pages["channelsPending"].views = payload.channelsPending;
        }

        if (payload.peers) {
          state.pages["peers"].views = payload.peers;
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
      state.pages[meta.arg.originalArgs.page].selected = 0;
    });

    builder.addMatcher(viewApi.endpoints.updateTableView.matchFulfilled, (state, { payload }) => {
      const views = state.pages[payload.page].views;
      const index = views.findIndex((view) => view.id === payload.id);
      views[index] = payload;
      views[index].dirty = false;
      state.pages[payload.page].views = views;
    });

    builder.addMatcher(viewApi.endpoints.createTableView.matchFulfilled, (state, { meta, payload }) => {
      const view = payload;
      view.dirty = false;
      state.pages[payload.page].views[meta.arg.originalArgs.index] = view;
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

export const selectClosedChannelView = (state: RootState) => {
  const page = "channelsClosed";
  const view = state.viewsSlice.pages[page].views[state.viewsSlice.pages[page].selected];
  return { viewResponse: view, selectedViewIndex: state.viewsSlice.pages[page].selected };
};

export const selectPendingChannelView = (state: RootState) => {
  const page = "channelsPending";
  const view = state.viewsSlice.pages[page].views[state.viewsSlice.pages[page].selected];
  return { viewResponse: view, selectedViewIndex: state.viewsSlice.pages[page].selected };
};

export const selectPeersViews = (state: RootState) => {
  const page = "peers";
  const view = state.viewsSlice.pages[page].views[state.viewsSlice.pages[page].selected];
  return { viewResponse: view, selectedViewIndex: state.viewsSlice.pages[page].selected };
};

export default viewsSlice.reducer;
