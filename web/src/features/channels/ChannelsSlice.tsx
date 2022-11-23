import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RootState } from "store/store";
import { AndClause, FilterClause } from "features/sidebar/sections/filter/filter";
import { SortByOptionType } from "features/sidebar/sections/sort/SortSectionOld";
import { ColumnMetaData, ViewInterface } from "features/table/Table";
import { torqApi } from "apiSlice";

export const channelsColumns: ColumnMetaData[] = [
  {
    heading: "Active",
    type: "BooleanCell",
    key: "active",
    valueType: "boolean",
    locked: false,
  },
  {
    heading: "Peer Alias",
    type: "TextCell",
    key: "peerAlias",
    valueType: "string",
  },
  {
    heading: "Short Channel ID",
    type: "TextCell",
    key: "shortChannelId",
    valueType: "string",
  },
  {
    heading: "Remote Balance",
    type: "NumericCell",
    key: "remoteBalance",
    valueType: "number",
  },
  {
    heading: "Balance",
    type: "BalanceCell",
    key: "localBalance",
    key2: "remoteBalance",
    valueType: "number",
    percent: false,
  },
  {
    heading: "Local Balance",
    type: "NumericCell",
    key: "localBalance",
    valueType: "number",
  },
  {
    heading: "Capacity",
    type: "NumericCell",
    key: "capacity",
    valueType: "number",
  },
  {
    heading: "Fee rate (PPM)",
    type: "NumericDoubleCell",
    key: "feeRatePpm",
    key2: "remoteFeeRatePpm",
    suffix: "ppm",
    valueType: "number",
  },
  {
    heading: "Base Fee Msat",
    type: "NumericDoubleCell",
    key: "baseFeeMsat",
    key2: "remoteBaseFeeMsat",
    suffix: "sat",
    valueType: "number",
  },
  {
    heading: "Minimum HTLC",
    type: "NumericDoubleCell",
    key: "minHtlc",
    key2: "remoteMinHtlc",
    suffix: "sat",
    valueType: "number",
  },
  {
    heading: "Maximum HTLC Amount",
    type: "NumericDoubleCell",
    key: "maxHtlcMsat",
    key2: "remoteMaxHtlcMsat",
    suffix: "sat",
    valueType: "number",
  },
  {
    heading: "Time Lock Delta",
    type: "NumericCell",
    key: "timeLockDelta",
    valueType: "number",
  },
  {
    heading: "LND Short Channel ID",
    type: "TextCell",
    key: "lndShortChannelId",
    valueType: "string",
  },
  {
    heading: "Funding Transaction",
    type: "TextCell",
    key: "fundingTransactionHash",
    valueType: "string",
  },
  {
    heading: "Unsettled Balance",
    type: "NumericCell",
    key: "unsettledBalance",
    valueType: "number",
  },
  {
    heading: "Satoshis Sent",
    type: "NumericCell",
    key: "totalSatoshisSent",
    valueType: "number",
  },
  {
    heading: "Satoshis Received",
    type: "NumericCell",
    key: "totalSatoshisReceived",
    valueType: "number",
  },
  {
    heading: "Pending Forwarding HTLCs count",
    type: "NumericCell",
    key: "pendingForwardingHTLCsCount",
    valueType: "number",
  },
  {
    heading: "Pending Forwarding HTLCs",
    type: "NumericCell",
    key: "pendingForwardingHTLCsAmount",
    valueType: "number",
  },
  {
    heading: "Pending Forwarding HTLCs count",
    type: "NumericCell",
    key: "pendingLocalHTLCsCount",
    valueType: "number",
  },
  {
    heading: "Pending Forwarding HTLCs",
    type: "NumericCell",
    key: "pendingLocalHTLCsAmount",
    valueType: "number",
  },
  {
    heading: "Total Pending Forwarding HTLCs count",
    type: "NumericCell",
    key: "pendingTotalHTLCsCount",
    valueType: "number",
  },
  {
    heading: "Total Pending Forwarding HTLCs",
    type: "NumericCell",
    key: "pendingTotalHTLCsAmount",
    valueType: "number",
  },
  {
    heading: "Commit Fee",
    type: "NumericCell",
    key: "commitFee",
    valueType: "number",
  },
  {
    heading: "Node ID",
    type: "NumericCell",
    key: "nodeId",
    valueType: "number",
  },
  {
    heading: "Node Name",
    type: "TextCell",
    key: "nodeName",
    valueType: "string",
  },
  {
    heading: "Mempool",
    type: "LinkCell",
    key: "mempoolSpace",
    valueType: "string",
  },
  {
    heading: "Amboss",
    type: "LinkCell",
    key: "ambossSpace",
    valueType: "string",
  },
  {
    heading: "1ML",
    type: "LinkCell",
    key: "1ml",
    valueType: "string",
  },
];
export interface PolicyInterface {
  feeRatePpm: number;
  timeLockDelta: number;
  maxHtlcMsat: number;
  minHtlcMsat: number;
  baseFeeMsat: number;
  fundingTransactionHash: string;
  fundingOutputIndex: number;
  nodeId: number;
}
export interface TableChannelsState {
  channels: [];
  selectedViewIndex: number;
  views: ViewInterface[];
  status: "idle" | "loading" | "failed";
}

const defaultFilter = new AndClause();
defaultFilter.addChildClause(
  new FilterClause({
    funcName: "eq",
    category: "boolean",
    key: "active",
    parameter: true,
  })
);

export const activeChannelsColumns = channelsColumns.filter((c) =>
  [
    "active",
    "peerAlias",
    "shortChannelId",
    "feeRatePpm",
    "remoteBalance",
    "gauge",
    "localBalance",
    "capacity",
    "totalSatoshisSent",
    "totalSatoshisReceived",
    "unsettledBalance",
    "commitFee",
    "baseFeeMsat",
    "minHtlc",
    "maxHtlcMsat",
    "nodeName",
  ].includes(c.key)
);


export const DefaultView: ViewInterface = {
  title: "Untitled View",
  saved: true,
  columns: activeChannelsColumns,
  filters: defaultFilter.toJSON(),
  sortBy: [],
  groupBy: undefined,
  page: 'channels'
};

const initialState: TableChannelsState = {
  channels: [],
  selectedViewIndex: 0,
  views: [
    {
      ...DefaultView,
      title: "Default View",
    },
  ],
  status: "idle",
};

export const channelsSlice = createSlice({
  name: "channels",
  initialState,
  // The `reducers` field lets us define reducers and generate associated actions
  reducers: {
    updateFilters: (state, actions: PayloadAction<{ filters: any }>) => {
      state.views[state.selectedViewIndex].filters = actions.payload.filters;
    },
    updateColumns: (state, actions: PayloadAction<{ columns: Array<ColumnMetaData> }>) => {
      state.views[state.selectedViewIndex].columns = actions.payload.columns;
    },
    updateGroupBy: (state, actions: PayloadAction<{ groupBy: string }>) => {
      state.views[state.selectedViewIndex].groupBy = actions.payload.groupBy;
    },
    updateSortBy: (state, actions: PayloadAction<{ sortBy: SortByOptionType[] }>) => {
      state.views[state.selectedViewIndex].sortBy = actions.payload.sortBy;
    },
    updateViews: (state, actions: PayloadAction<{ views: ViewInterface[]; index: number }>) => {
      state.views = actions.payload.views;
      state.selectedViewIndex = actions.payload.index;
    },
    updateViewsOrder: (state, actions: PayloadAction<{ views: ViewInterface[]; index: number }>) => {
      state.views = actions.payload.views;
      state.selectedViewIndex = actions.payload.index;
    },
    deleteView: (state, actions: PayloadAction<{ view: ViewInterface; index: number }>) => {
      state.views = [
        ...state.views.slice(0, actions.payload.index),
        ...state.views.slice(actions.payload.index + 1, state.views.length),
      ];
      state.selectedViewIndex = 0;
    },
    updateSelectedView: (state, actions: PayloadAction<{ index: number }>) => {
      state.selectedViewIndex = actions.payload.index;
    },
  },
  // The `extraReducers` field lets the slice handle actions defined elsewhere,
  // including actions generated by createAsyncThunk or in other slices.
  extraReducers: (builder) => {
    builder.addMatcher(
      (action) => {
        return (
          ["channels/updateFilters", "channels/updateSortBy", "channels/updateColumns", "channels/updateGroupBy"].findIndex(
            (item) => action.type === item
          ) !== -1
        );
      },
      (state, _) => {
        state.views[state.selectedViewIndex].saved = false;
      }
    );

    builder.addMatcher(torqApi.endpoints.createTableView.matchFulfilled, (state, { payload }) => {
      state.views[payload.index] = {
        ...payload.view.view,
        id: payload.view.id,
      };
      state.selectedViewIndex = payload.index;
    });

    builder.addMatcher(torqApi.endpoints.deleteTableView.matchFulfilled, (state, { payload }) => {
      state.views = [
        ...state.views.slice(0, payload.index),
        ...state.views.slice(payload.index + 1, state.views.length),
      ];
      state.selectedViewIndex = 0;
    });

    builder.addMatcher(torqApi.endpoints.getTableViews.matchFulfilled, (state, { payload }) => {
      if (payload !== null) {
        state.views = payload.map((view: { id: number; view: ViewInterface }) => {
          return { ...view.view, id: view.id };
        });
      }
    });

    builder.addMatcher(torqApi.endpoints.updateTableView.matchFulfilled, (state, { payload }) => {
      const view = state.views.find((v) => v.id === payload.id);
      if (view) {
        view.saved = true;
      }
    });
  },
});

export const {
  updateFilters,
  updateViews,
  updateViewsOrder,
  deleteView,
  updateSelectedView,
  updateSortBy,
  updateColumns,
  updateGroupBy
} = channelsSlice.actions;

export const selectActiveColumns = (state: RootState) => {
  return state.channels.views[state.channels.selectedViewIndex].columns || [];
};
export const selectAllColumns = (_: RootState) => channelsColumns;
export const selectSortBy = (state: RootState) => state.channels.views[state.channels.selectedViewIndex].sortBy;
export const selectGroupBy = (state: RootState) => state.channels.views[state.channels.selectedViewIndex].groupBy;
export const selectFilters = (state: RootState) => {
  return state.channels.views[state.channels.selectedViewIndex].filters;
};
export const selectViews = (state: RootState) => state.channels.views;
export const selectCurrentView = (state: RootState) => state.channels.views[state.channels.selectedViewIndex];
export const selectedViewIndex = (state: RootState) => state.channels.selectedViewIndex;

export default channelsSlice.reducer;
