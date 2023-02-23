import { createSlice, Dispatch, Middleware, MiddlewareAPI } from "@reduxjs/toolkit";
import { RootState } from "store/store";

export enum Network {
  MainNet,
  TestNet,
  RegTest,
  SigNet,
  SimNet,
}

export interface NetworkState {
  activeNetwork: Network;
}

const defaultNetwork = process.env.NODE_ENV === "development" ? Network.SimNet : Network.MainNet;

const getActiveNetworkFromLocalStorage = (): Network | undefined => {
  const lsActiveNetwork = localStorage.getItem("activeNetwork");
  if (!lsActiveNetwork) {
    return undefined;
  }
  return parseInt(lsActiveNetwork) as Network;
};

const initialState: NetworkState = {
  activeNetwork: getActiveNetworkFromLocalStorage() ?? defaultNetwork,
};

export const networkSlice = createSlice({
  name: "network",
  initialState,

  // The `reducers` field lets us define reducers and generate associated actions
  reducers: {
    setActiveNetwork: (state, action) => {
      state.activeNetwork = action.payload;
    },
  },
});

export const networkMiddleware: Middleware = (_: MiddlewareAPI) => (next: Dispatch) => (action) => {
  if (networkSlice.actions.setActiveNetwork.match(action)) {
    localStorage.setItem("activeNetwork", action.payload);
  }
  // Call the next dispatch method in the middleware chain.
  return next(action);
};

export const { setActiveNetwork } = networkSlice.actions;

// The function below is called a selector and allows us to select a value from
// the state. Selectors can also be defined inline where they're used instead of
// in the slice file. For example: `useSelector((state: RootState) => state.counter.value)`
export const selectActiveNetwork = (state: RootState) => state.network.activeNetwork;

export default networkSlice.reducer;
