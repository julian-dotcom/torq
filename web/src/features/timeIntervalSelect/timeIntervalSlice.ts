import { createSlice } from '@reduxjs/toolkit';
import { RootState } from '../../store/store';
import { defineds, getCompareRanges } from "./customRanges";

export interface TimeIntervalState {
  from: string;
  to: string;
  compareFrom: string;
  compareTo: string;
}

const initialState: TimeIntervalState = {
  to: defineds.endOfLast7Days.toString(),
  from: defineds.startOfLast7Days.toString(),
  compareTo: defineds.endOfLast7DaysCompare.toString(),
  compareFrom: defineds.startOfLast7DaysCompare.toString(),
};

export const timeIntervalSlice = createSlice({
  name: 'timeInterval',
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
});

export const { updateInterval } = timeIntervalSlice.actions;

// The function below is called a selector and allows us to select a value from
// the state. Selectors can also be defined inline where they're used instead of
// in the slice file. For example: `useSelector((state: RootState) => state.counter.value)`
export const selectTimeInterval = (state: RootState) => state.timeInterval;

export default timeIntervalSlice.reducer;
