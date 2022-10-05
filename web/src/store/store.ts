import { configureStore, ThunkAction, Action } from "@reduxjs/toolkit";
import navReducer from "../features/navigation/navSlice";
import channelReducer from "../features/channel/channelSlice";
import tableReducer from "../features/forwards/forwardsSlice";
import timeIntervalReducer from "../features/timeIntervalSelect/timeIntervalSlice";
import { torqApi } from "apiSlice";
import { setupListeners } from "@reduxjs/toolkit/query";
import paymentsReducer from "features/transact/Payments/paymentsSlice";
import invoicesReducer from "features/transact/Invoices/invoicesSlice";
import onChainReducer from "features/transact/OnChain/onChainSlice";
import i18nReducer from "services/i18n/i18nSlice";

export const store = configureStore({
  reducer: {
    navigation: navReducer,
    table: tableReducer,
    payments: paymentsReducer,
    invoices: invoicesReducer,
    onChain: onChainReducer,
    timeInterval: timeIntervalReducer,
    channel: channelReducer,
    [torqApi.reducerPath]: torqApi.reducer,
    i18n: i18nReducer,
  },
  middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(torqApi.middleware),
});

setupListeners(store.dispatch);

export type AppDispatch = typeof store.dispatch;
export type RootState = ReturnType<typeof store.getState>;
export type AppThunk<ReturnType = void> = ThunkAction<ReturnType, RootState, unknown, Action<string>>;
