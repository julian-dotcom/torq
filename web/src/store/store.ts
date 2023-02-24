import { configureStore, ThunkAction, Action } from "@reduxjs/toolkit";
import navReducer from "features/navigation/navSlice";
import channelReducer from "features/channel/channelSlice";
import networkReducer, { networkMiddleware } from "features/network/networkSlice";
// import channelsReducer from "features/channels/ChannelsSlice";
// import tableReducer from "features/forwards/forwardsSlice";
import timeIntervalReducer from "features/timeIntervalSelect/timeIntervalSlice";
import { torqApi } from "apiSlice";
import { setupListeners } from "@reduxjs/toolkit/query";
import { viewsSlice } from "features/viewManagement/viewSlice";
// import paymentsReducer from "features/transact/Payments/paymentsSlice";
// import invoicesReducer from "features/transact/Invoices/invoicesSlice";
// import onChainReducer from "features/transact/OnChain/onChainSlice";
import i18nReducer from "services/i18n/i18nSlice";
import workflowReducer from "pages/WorkflowPage/WorkflowSlice";
export const store = configureStore({
  reducer: {
    viewsSlice: viewsSlice.reducer,
    navigation: navReducer,
    network: networkReducer,
    // table: tableReducer,
    // payments: paymentsReducer,
    // invoices: invoicesReducer,
    // onChain: onChainReducer,
    timeInterval: timeIntervalReducer,
    channel: channelReducer,
    // channels: channelsReducer,
    [torqApi.reducerPath]: torqApi.reducer,
    i18n: i18nReducer,
    workflow: workflowReducer,
  },
  middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(torqApi.middleware).concat(networkMiddleware),
});

setupListeners(store.dispatch);

export type AppDispatch = typeof store.dispatch;
export type RootState = ReturnType<typeof store.getState>;
export type AppThunk<ReturnType = void> = ThunkAction<ReturnType, RootState, unknown, Action<string>>;
