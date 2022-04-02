import {createAsyncThunk, createSlice, current, PayloadAction} from '@reduxjs/toolkit';
import {RootState} from '../../store/store';
import { addDays, format } from 'date-fns';
import { FilterInterface, FilterFunctions, applyFilters } from './controls/filter/filter'
import {SortByOptionType} from "./controls/sort/SortControls";
import _ from "lodash";

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
  id?: number;
  saved: boolean;
  filters: Array<FilterInterface>;
  columns: ColumnMetaData[];
  sortBy: SortByOptionType[],
}

export interface TableState {
  channels: [];
  modChannels: [];
  selectedViewIndex: number;
  views: ViewInterface[];
  status: 'idle' | 'loading' | 'failed';
}

export const DefaultView: ViewInterface = {
    title: "New Table",
    saved: false,
    filters: [],
    columns: columns,
    sortBy: [],
  }

const initialState: TableState = {
  channels: [],
  modChannels: [],
  selectedViewIndex: 0,
  views: [DefaultView],
  status: 'idle',
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

export const fetchChannelsAsync = createAsyncThunk(
  'table/fetchChannels',
  async (data: { from: string, to: string }) => {
    const response = await fetchChannels(data.from, data.to);
    return response
  }
);


function fetchTableViews() {
  const body = fetch(`http://localhost:8080/api/table-views`, init)
    .then(response => {
      return response.json()
    })
  return body
}

export const fetchTableViewsAsync = createAsyncThunk(
  'table/fetchTableViews',
  async () => {
    const response = await fetchTableViews();
    return response
  }
);


function updateTableView(view: ViewInterface) {
  const init: RequestInit = {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    mode: 'cors',
    method: 'PUT',
    body: JSON.stringify({id: view.id, view: view}),
  };
  const body = fetch(`http://localhost:8080/api/table-views`, init)
    .then(response => {
      return response.json()
    })
  return body
}

export const updateTableViewAsync = createAsyncThunk(
  'table/updateTableView',
  async (data: { view: ViewInterface, index: number }) => {

    let body = await updateTableView(data.view)
    return data.index
})

function createTableView(view: ViewInterface) {
  const init: RequestInit = {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    mode: 'cors',
    method: 'POST',
    body: JSON.stringify({id: null, view: view}),
  };
  const body = fetch(`http://localhost:8080/api/table-views`, init)
    .then(response => {
      return response.json()
    })
  return body
}

export const createTableViewAsync = createAsyncThunk(
  'table/createTableView',
  async (data: { view: ViewInterface, index: number }) => {

    let body = await createTableView(data.view)

    return {view: body, index: data.index}
})


// export function loadTableState() {
//   try {
//     const serializedState = localStorage.getItem("torq_temp_view");
//     if (!serializedState) return undefined;
//     return JSON.parse(serializedState);
//   } catch (e) {
//     return undefined;
//   }
// }
//
// export async function saveTempView(state: any) {
//   try {
//     const serializedState = JSON.stringify(state);
//     localStorage.setItem("torq_temp_view", serializedState);
//   } catch (e) {
//     console.log(e)
//   }
// }

export const tableSlice = createSlice({
  name: 'table',
  initialState,
  // The `reducers` field lets us define reducers and generate associated actions
  reducers: {
    updateFilters: (state, actions: PayloadAction<{ filters: FilterInterface[] }>) => {
      state.views[state.selectedViewIndex].filters = actions.payload.filters
    },
    updateColumns: (state, actions: PayloadAction<{columns: ColumnMetaData[]}>) => {
      state.views[state.selectedViewIndex].columns = actions.payload.columns
    },
    updateViews: (state, actions: PayloadAction<{ views: ViewInterface[] }>) => {
      state.views = actions.payload.views
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

    builder
      .addCase(fetchTableViewsAsync.pending, (state) => {
        state.status = 'loading';
      })
      .addCase(fetchTableViewsAsync.fulfilled, (state, action) => {
        state.status = 'idle';
        if(action.payload) {
          state.views = action.payload.map((view: {id: number, view: ViewInterface}) => {return {...view.view, id: view.id}})
        }
      });

    builder
      .addCase(createTableViewAsync.pending, (state) => {
        state.status = 'loading';
      })
      .addCase(createTableViewAsync.fulfilled, (state, action) => {
        state.status = 'idle';
        state.views[action.payload.index] = action.payload.view.view
        state.views[action.payload.index].saved = true
      });

    builder
      .addCase(updateTableViewAsync.pending, (state) => {
        state.status = 'loading';
      })
      .addCase(updateTableViewAsync.fulfilled, (state, action) => {
        state.status = 'idle';
        state.views[action.payload].saved = true
      });

    builder
      .addCase(fetchChannelsAsync.pending, (state) => {
        state.status = 'loading';
      })
      .addCase(fetchChannelsAsync.fulfilled, (state, action) => {
        state.status = 'idle';
        state.channels = action.payload
      });

    builder.addMatcher((action) => {
      return ['table/updateFilters', 'table/updateSortBy', 'table/updateViews', 'table/updateColumns']
        .findIndex((item) => action.type === item) !== -1
    }, (state, actions) => {
      // TODO: create compare version to indicate it view is saved or not.
      state.views[state.selectedViewIndex].saved = false
    })

    // Store the new name view name in the backend
    builder.addMatcher((action) => action.type === 'table/updateViews', (state, actions) => {
      const uView = {...current(state.views[state.selectedViewIndex]), saved: true}
      // TODO: the index here is not correct when creating new view
      console.log(uView)
        if (!uView.id) {
          createTableView(uView).then(() => console.log('finished'))
        }
        updateTableView(uView).then(() => console.log('updated'))
      })

  },
});

export const { updateFilters, updateViews, updateSelectedView, updateSortBy, updateColumns} = tableSlice.actions;


export const selectChannels = (state: RootState) => {
  const filters = state.table.views[state.table.selectedViewIndex].filters || []
  const sorts = state.table.views[state.table.selectedViewIndex].sortBy || []

  // TODO: Clean up
  let channels = state.table.channels ? state.table.channels.slice() : []
  return _.orderBy(applyFilters(filters, channels), sorts.map((s) => s.value), sorts.map((s) => s.direction) as ['asc'|'desc'])
};

export const selectActiveColumns = (state: RootState) => {
  return state.table.views[state.table.selectedViewIndex].columns || [];
}
export const selectAllColumns = (state: RootState) => columns;
export const selectSortBy = (state: RootState) => state.table.views[state.table.selectedViewIndex].sortBy
export const selectFilters = (state: RootState) => {
  return state.table.views[state.table.selectedViewIndex].filters || []
};
export const selectViews = (state: RootState) => state.table.views;
export const selectCurrentView = (state: RootState) => state.table.views[state.table.selectedViewIndex];
export const selectedViewIndex = (state: RootState) => state.table.selectedViewIndex;

export default tableSlice.reducer;
