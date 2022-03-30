import { configureStore, ThunkAction, Action } from '@reduxjs/toolkit';
import navReducer from '../components/navigation/navSlice';
import tableReducer from '../components/table/tableSlice';
import timeIntervalReducer from '../components/timeIntervalSelect/timeIntervalSlice';
import authreducer from '../pages/login/authSlice';

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
