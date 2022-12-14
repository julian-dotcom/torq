import { createSlice } from "@reduxjs/toolkit";
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

const initialState: NetworkState = {
  activeNetwork: Network.MainNet,
};

export const networkSlice = createSlice({
  name: "network",
  initialState,

  // The `reducers` field lets us define reducers and generate associated actions
  reducers: {
    setActiveNetwork: (state, action) => {
      // Redux Toolkit allows us to write "mutating" logic in reducers. It
      // doesn't actually mutate the state because it uses the Immer library,
      // which detects changes to a "draft state" and produces a brand new
      // immutable state based off those changes
      state.activeNetwork = action.payload;
    },
  },
});

export const { setActiveNetwork } = networkSlice.actions;

// The function below is called a selector and allows us to select a value from
// the state. Selectors can also be defined inline where they're used instead of
// in the slice file. For example: `useSelector((state: RootState) => state.counter.value)`
export const selectActiveNetwork = (state: RootState) => state.network.activeNetwork;

export default networkSlice.reducer;
