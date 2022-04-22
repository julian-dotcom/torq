import { configureStore, ThunkAction, Action } from '@reduxjs/toolkit';
import navReducer from '../features/navigation/navSlice';
import tableReducer from '../features/table/tableSlice';
import timeIntervalReducer from '../features/timeIntervalSelect/timeIntervalSlice';
import authreducer from '../features/auth/authSlice';

export const store = configureStore({
  reducer: {
    navigation: navReducer,
    table: tableReducer,
    timeInterval: timeIntervalReducer,
    auth: authreducer,
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
