import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '../../store/store';
import { deserialiseQuery, applyFilters, AndClause } from './controls/filter/filter'
import { SortByOptionType } from "./controls/sort/SortControls";
import _, { cloneDeep } from "lodash";
import { torqApi } from 'apiSlice'
import {groupByReducer} from "./controls/group/groupBy";

export interface ColumnMetaData {
  heading: string;
  key: string;
  type?: string;
  width?: number;
  locked?: boolean;
  valueType: string;
  total?: number;
  max?: number;
}

export const availableColumns: ColumnMetaData[] = [
  { heading: "Name", type: "AliasCell", key: "alias", locked: true, valueType: "string" },
  { heading: "Revenue", type: "BarCell", key: "revenue_out", valueType: "number" },
  { heading: "Capacity", type: "NumericCell", key: "capacity", valueType: "number" },
  { heading: "Amount outbound", type: "BarCell", key: "amount_out", valueType: "number" },
  { heading: "Amount inbound", type: "BarCell", key: "amount_in", valueType: "number" },
  { heading: "Amount total", type: "BarCell", key: "amount_total", valueType: "number" },
  { heading: "Turnover outbound", type: "NumericCell", key: "turnover_out", valueType: "number" },
  { heading: "Turnover inbound", type: "NumericCell", key: "turnover_in", valueType: "number" },
  { heading: "Turnover total", type: "NumericCell", key: "turnover_total", valueType: "number" },
  { heading: "Successful outbound", type: "BarCell", key: "count_out", valueType: "number" },
  { heading: "Successful inbound", type: "BarCell", key: "count_in", valueType: "number" },
  { heading: "Successful total", type: "BarCell", key: "count_total", valueType: "number" },
  { heading: "Contributed revenue inbound", type: "BarCell", key: "revenue_in", valueType: "number" },
  { heading: "Contributed revenue total", type: "BarCell", key: "revenue_total", valueType: "number" },
  { heading: "Public key", type: "TextCell", key: "pub_key", valueType: "string" },
  { heading: "Channel point", type: "TextCell", key: "channel_point", valueType: "string" },
  { heading: "Channel short ID", type: "TextCell", key: "shortChannelId", valueType: "string" },
  { heading: "LND Channel short ID", type: "TextCell", key: "chan_id", valueType: "string" },
  { heading: "Open Channel", type: "TextCell", key: "open", valueType: "number" },

  { heading: "HTLC All failures in", type: "BarCell", key: "htlc_fail_all_in", valueType: "number" },
  { heading: "HTLC All failures out", type: "BarCell", key: "htlc_fail_all_out", valueType: "number" },
  { heading: "HTLC All failures total", type: "BarCell", key: "htlc_fail_all_total", valueType: "number" },
  { heading: "HTLC Unknown failures in", type: "BarCell", key: "htlc_forward_fail_in", valueType: "number" },
  { heading: "HTLC Unknown failures out", type: "BarCell", key: "htlc_forward_fail_out", valueType: "number" },
  { heading: "HTLC Unknown failures total", type: "BarCell", key: "htlc_forward_fail_total", valueType: "number" },
  { heading: "HTLC Link failures in", type: "BarCell", key: "htlc_link_fail_in", valueType: "number" },
  { heading: "HTLC Link failures out", type: "BarCell", key: "htlc_link_fail_out", valueType: "number" },
  { heading: "HTLC Link failures total", type: "BarCell", key: "htlc_link_fail_total", valueType: "number" },
]

export interface ViewInterface {
  title: string;
  id?: number;
  saved: boolean;
  filters?: any;
  columns: ColumnMetaData[];
  sortBy: SortByOptionType[],
  groupBy?: string,
}

export interface TableState {
  channels: [];
  modChannels: [];
  selectedViewIndex: number;
  views: ViewInterface[];
  status: 'idle' | 'loading' | 'failed';
}

export const DefaultView: ViewInterface = {
  title: "Untitled Table",
  saved: true,
  columns: availableColumns,
  filters: new AndClause().toJSON(),
  sortBy: [],
  groupBy: undefined,
}

const initialState: TableState = {
  channels: [],
  modChannels: [],
  selectedViewIndex: 0,
  views: [{
    ...DefaultView,
    title: 'Default table',
  }],
  status: 'idle',
};

export interface viewOrderInterface {
  id: number | undefined,
  view_order: number
}

export const tableSlice = createSlice({
  name: 'table',
  initialState,
  // The `reducers` field lets us define reducers and generate associated actions
  reducers: {
    updateFilters: (state, actions: PayloadAction<{ filters: any }>) => {
      state.views[state.selectedViewIndex].filters = actions.payload.filters
    },
    updateColumns: (state, actions: PayloadAction<{ columns: ColumnMetaData[] }>) => {
      state.views[state.selectedViewIndex].columns = actions.payload.columns
    },
    updateGroupBy: (state, actions: PayloadAction<{ groupBy: string }>) => {
      state.views[state.selectedViewIndex].groupBy = actions.payload.groupBy
    },
    updateViews: (state, actions: PayloadAction<{ views: ViewInterface[], index: number }>) => {
      state.views = actions.payload.views
      state.selectedViewIndex = actions.payload.index
    },
    updateViewsOrder: (state, actions: PayloadAction<{ views: ViewInterface[], index: number }>) => {
      state.views = actions.payload.views
      state.selectedViewIndex = actions.payload.index
    },
    deleteView: (state, actions: PayloadAction<{ view: ViewInterface, index: number }>) => {
      state.views = [
        ...state.views.slice(0, actions.payload.index),
        ...state.views.slice(actions.payload.index + 1, state.views.length),
      ]
      state.selectedViewIndex = 0
    },
    updateSelectedView: (state, actions: PayloadAction<{ index: number }>) => {
      state.selectedViewIndex = actions.payload.index
    },
    updateSortBy: (state, actions: PayloadAction<{ sortBy: SortByOptionType[] }>) => {
      state.views[state.selectedViewIndex].sortBy = actions.payload.sortBy
    },
  },
  // The `extraReducers` field lets the slice handle actions defined elsewhere,
  // including actions generated by createAsyncThunk or in other slices.
  extraReducers: (builder) => {

    builder.addMatcher((action) => {
      return ['table/updateFilters', 'table/updateSortBy', 'table/updateColumns', 'table/updateGroupBy']
        .findIndex((item) => action.type === item) !== -1
    }, (state, _) => {
      // TODO: create compare version to indicate it view is saved or not.
      state.views[state.selectedViewIndex].saved = false
    })

    builder.addMatcher(
      torqApi.endpoints.createTableView.matchFulfilled,
      (state, { payload }) => {
        state.views[payload.index] = { ...payload.view.view, id: payload.view.id }
        state.selectedViewIndex = payload.index
      }
    )

    builder.addMatcher(
      torqApi.endpoints.deleteTableView.matchFulfilled,
      (state, { payload }) => {
        state.views = [
          ...state.views.slice(0, payload.index),
          ...state.views.slice(payload.index + 1, state.views.length),
        ]
        state.selectedViewIndex = 0;
      }
    )

    builder.addMatcher(
      torqApi.endpoints.getTableViews.matchFulfilled,
      (state, { payload }) => {
        state.views = payload.map((view: { id: number, view: ViewInterface }) => { return { ...view.view, id: view.id } })
      }
    )

    builder.addMatcher(
      torqApi.endpoints.updateTableView.matchFulfilled,
      (state, { payload }) => {
        const view = state.views.find(v => v.id === payload.id)
        if (view) {
          view.saved = true
        }
      }
    )

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
  updateGroupBy,
} = tableSlice.actions;

export const selectChannels = (state: RootState) => {

  let channels = cloneDeep(state.table.channels ? state.table.channels : [] as any[])
  const filters = state.table.views[state.table.selectedViewIndex].filters
  const groupBy = state.table.views[state.table.selectedViewIndex].groupBy
  if (channels.length > 0) {
    channels = groupByReducer(channels, groupBy || 'channels')
  }

  if (filters) {
    const deserialisedFilters = deserialiseQuery(filters)
    channels = applyFilters(deserialisedFilters, channels)
  }
  const sorts = state.table.views[state.table.selectedViewIndex].sortBy || []
  return _.orderBy(channels, sorts.map((s) => s.value), sorts.map((s) => s.direction) as ['asc' | 'desc'])
};

export const selectActiveColumns = (state: RootState) => {
  return state.table.views[state.table.selectedViewIndex].columns || [];
}
export const selectAllColumns = (_: RootState) => availableColumns;
export const selectSortBy = (state: RootState) => state.table.views[state.table.selectedViewIndex].sortBy
export const selectGroupBy = (state: RootState) => state.table.views[state.table.selectedViewIndex].groupBy
export const selectFilters = (state: RootState) => { return state.table.views[state.table.selectedViewIndex].filters };
export const selectViews = (state: RootState) => state.table.views;
export const selectCurrentView = (state: RootState) => state.table.views[state.table.selectedViewIndex];
export const selectedViewIndex = (state: RootState) => state.table.selectedViewIndex;
export const selectStatus = (state: RootState) => state.table.status;

export default tableSlice.reducer;
