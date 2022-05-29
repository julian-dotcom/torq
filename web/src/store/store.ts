import { configureStore, ThunkAction, Action } from "@reduxjs/toolkit";
import navReducer from "../features/navigation/navSlice";
import channelReducer from "../features/channel/channelSlice";
import tableReducer from "../features/table/tableSlice";
import timeIntervalReducer from "../features/timeIntervalSelect/timeIntervalSlice";
import { torqApi } from "apiSlice";
import { setupListeners } from "@reduxjs/toolkit/query";

export const store = configureStore({
  reducer: {
    navigation: navReducer,
    table: tableReducer,
    timeInterval: timeIntervalReducer,
    channel: channelReducer,
    [torqApi.reducerPath]: torqApi.reducer,
  },
  middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(torqApi.middleware),
});

setupListeners(store.dispatch);

export type AppDispatch = typeof store.dispatch;
export type RootState = ReturnType<typeof store.getState>;
export type AppThunk<ReturnType = void> = ThunkAction<ReturnType, RootState, unknown, Action<string>>;
