import { configureStore, ThunkAction, Action } from '@reduxjs/toolkit';
import navReducer from '../components/navigation/navSlice';
import tableReducer from '../components/table/tableSlice';
import timeIntervalReducer from '../components/timeIntervalSelect/timeIntervalSlice';

export const store = configureStore({
  reducer: {
    navigation: navReducer,
    table: tableReducer,
    timeInterval: timeIntervalReducer,
  },
});

export type AppDispatch = typeof store.dispatch;
export type RootState = ReturnType<typeof store.getState>;
export type AppThunk<ReturnType = void> = ThunkAction<
  ReturnType,
  RootState,
  unknown,
  Action<string>
>;
