import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';
import { RootState, AppThunk } from '../../store/store';
import { addDays, format } from 'date-fns';

import fieldSorter from './controls/sort/sort';
import { FilterInterface, FilterFunctions, applyFilters } from './controls/filter/filter'

export interface ColumnMetaData {
  heading: string;
  key: string;
  type?: string;
  width?: number;
  locked?: boolean;
  valueType: string;
}

export const columns: ColumnMetaData[] = [
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
]

export interface ViewInterface {
  title: string;
  saved: boolean;
  filters: Array<FilterInterface>;
  columns: ColumnMetaData[];
}

export interface TableState {
  channels: [];
  modChannels: [];
  selectedViewIndex: number;
  views: ViewInterface[];
  sortBy: []; // Feilds to sort by
  sorts: []; // Fields added
  sortOptions: ColumnMetaData[]; // Fields remaining to be added
  status: 'idle' | 'loading' | 'failed';
}

export const DefaultView: ViewInterface = {
    title: "New Table",
    saved: true,
    filters: [],
    columns: columns,
  }

const initialState: TableState = {
  channels: [],
  modChannels: [],
  selectedViewIndex: 0,
  views: loadTableState() || [DefaultView],  //
  status: 'idle',
  sortBy: [],
  sorts: [],
  sortOptions: columns,

};
const init: RequestInit = {
  credentials: 'include',
  headers: { 'Content-Type': 'application/json' },
  mode: 'cors',
};

function fetchChannels(from: string, to: string) {
  to = format(addDays(new Date(to), 1), "yyyy-MM-dd")
  const body = fetch(`http://localhost:8080/api/channels?from=${from}&to=${to}`, init)
    .then(response => {
      return response.json()
    })
  return body
}

// The function below is called a thunk and allows us to perform async logic. It
// can be dispatched like a regular action: `dispatch(incrementAsync(10))`. This
// will call the thunk with the `dispatch` function as the first argument. Async
// code can then be executed and other actions can be dispatched. Thunks are
// typically used to make async requests.
export const fetchChannelsAsync = createAsyncThunk(
  'table/fetchChannels',
  async (data: { from: string, to: string }) => {
    const response = await fetchChannels(data.from, data.to);
    return response
  }
);


function getDifference(array1: any[], array2: { key: any }[]) {
  return array1.filter((object1: { key: any }) => {
    return !array2.some((object2: { key: any }) => {
      return object1.key === object2.key;
    });
  })
};


export function loadTableState() {
  try {
    const serializedState = localStorage.getItem("torq_temp_view");
    if (!serializedState) return undefined;
    return JSON.parse(serializedState);
  } catch (e) {
    return undefined;
  }
}

export async function saveTempView(state: any) {
  try {
    const serializedState = JSON.stringify(state);
    localStorage.setItem("torq_temp_view", serializedState);
  } catch (e) {
    console.log(e)
  }
}

export const tableSlice = createSlice({
  name: 'table',
  initialState,
  // The `reducers` field lets us define reducers and generate associated actions
  reducers: {
    updateFilters: (state, actions: PayloadAction<{ filters: FilterInterface[] }>) => {
      state.views[state.selectedViewIndex].filters = actions.payload.filters
      // TODO: Skip localstorage, save on server when the user chooses too.
      saveTempView(state.views)
    },
    updateColumns: (state, actions: PayloadAction<{columns: ColumnMetaData[]}>) => {
      state.views[state.selectedViewIndex].columns = actions.payload.columns
      // TODO: Skip localstorage, save on server when the user chooses too.
      saveTempView(state.views)
    },
    updateViews: (state, actions: PayloadAction<{ views: ViewInterface[] }>) => {
      state.views = actions.payload.views
      saveTempView(state.views)
    },
    updateSelectedView: (state, actions: PayloadAction<{ index: number }>) => {
      state.selectedViewIndex = actions.payload.index
    },
    updateSortOptions: (state, actions: PayloadAction<any[]>) => {
      //@ts-ignore
      state.sorts = actions.payload[0]
      state.sortOptions = actions.payload[1]
    },
    updateSort: (state, actions: PayloadAction<string[]>) => {
      // @ts-ignore
      state.sortBy = actions.payload
      // @ts-ignore
      state.channels = [...state.channels].sort(fieldSorter(state.sortBy))
    },
  },
  // The `extraReducers` field lets the slice handle actions defined elsewhere,
  // including actions generated by createAsyncThunk or in other slices.
  extraReducers: (builder) => {
    builder
      .addCase(fetchChannelsAsync.pending, (state) => {
        state.status = 'loading';
      })
      .addCase(fetchChannelsAsync.fulfilled, (state, action) => {
        state.status = 'idle';
        state.channels = action.payload
      });
  },
});

export const { updateFilters, updateViews, updateSelectedView, updateSortOptions, updateSort, updateColumns} = tableSlice.actions;

export const selectChannels = (state: RootState) => {
  const filters = state.table.views[state.table.selectedViewIndex].filters || []
  return applyFilters(filters, state.table.channels)
};

export const selectActiveColumns = (state: RootState) => {
  return state.table.views[state.table.selectedViewIndex].columns || [];
}
export const selectAllColumns = (state: RootState) => columns;
export const selectSorts = (state: RootState) => state.table.sorts;
export const selectSortByOptions = (state: RootState) => state.table.sortOptions;
export const selectSortBy = (state: RootState) => state.table.sortBy
export const selectFilters = (state: RootState) => {
  return state.table.views[state.table.selectedViewIndex].filters || []
};
export const selectViews = (state: RootState) => state.table.views;
export const selectedViewindex = (state: RootState) => state.table.selectedViewIndex;

export default tableSlice.reducer;
