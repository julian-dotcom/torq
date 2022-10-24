import { createSlice } from "@reduxjs/toolkit";
import { RootState } from "store/store";
import { defineds, getCompareRanges, defaultStaticRangesFn } from "./customRanges";
import { torqApi } from "apiSlice";
import { Draft } from "immer";

export interface TimeIntervalState {
  from: string;
  to: string;
  compareFrom: string;
  compareTo: string;
  defaultDateRange: string;
  weekStartsOn: number;
}

const startOfWeekTable = {
  saturday: 6,
  sunday: 0,
  monday: 1,
};

const initialState: TimeIntervalState = {
  to: defineds.endOfLast7Days.toString(),
  from: defineds.startOfLast7Days.toString(),
  compareTo: defineds.endOfLast7DaysCompare.toString(),
  compareFrom: defineds.startOfLast7DaysCompare.toString(),
  defaultDateRange: "",
  weekStartsOn: 1, // aka Sunday
};

export const timeIntervalSlice = createSlice({
  name: "timeInterval",
  initialState,

  // The `reducers` field lets us define reducers and generate associated actions
  reducers: {
    updateInterval: (state, action) => {
      state.from = action.payload.from;
      state.to = action.payload.to;
      const compareRange = getCompareRanges(new Date(state.from), new Date(state.to));
      state.compareFrom = compareRange[1].toString();
      state.compareTo = compareRange[0].toString();
    },
  },
  extraReducers: (builder) => {
    builder.addMatcher(torqApi.endpoints.getSettings.matchFulfilled, (state, { payload }) => {
      const weekStartsOn = startOfWeekTable[payload.weekStartsOn];
      state.weekStartsOn = weekStartsOn;
      if (!state.defaultDateRange) {
        setRangeToDefault(state, payload.defaultDateRange);
      }
      if (state.defaultDateRange !== payload.defaultDateRange) {
        state.defaultDateRange = payload.defaultDateRange;
      }
    });

    builder.addMatcher(torqApi.endpoints.logout.matchFulfilled, (state) => {
      setRangeToDefault(state, state.defaultDateRange);
    });
  },
});

function setRangeToDefault(state: Draft<TimeIntervalState>, defaultDateRange: string) {
  const defaultStaticRanges = defaultStaticRangesFn(state.weekStartsOn);
  const staticRange = defaultStaticRanges.find((sr: any) => sr.code === defaultDateRange);
  state.to = staticRange.range().endDate.toString();
  state.from = staticRange.range().startDate.toString();
  state.compareTo = staticRange.rangeCompare().endDate.toString();
  state.compareFrom = staticRange.rangeCompare().startDate.toString();
}

export const { updateInterval } = timeIntervalSlice.actions;

// The function below is called a selector and allows us to select a value from
// the state. Selectors can also be defined inline where they're used instead of
// in the slice file. For example: `useSelector((state: RootState) => state.counter.value)`
export const selectTimeInterval = (state: RootState) => state.timeInterval;

export default timeIntervalSlice.reducer;
