import { createSlice } from "@reduxjs/toolkit";
import { RootState } from "store/store";

export interface NavState {
  hidden: boolean;
}

const initialState: NavState = {
  hidden: false,
};

export const navSlice = createSlice({
  name: "navigation",
  initialState,

  // The `reducers` field lets us define reducers and generate associated actions
  reducers: {
    toggleNav: (state) => {
      // Redux Toolkit allows us to write "mutating" logic in reducers. It
      // doesn't actually mutate the state because it uses the Immer library,
      // which detects changes to a "draft state" and produces a brand new
      // immutable state based off those changes
      state.hidden = !state.hidden;
    },
  },
});

export const { toggleNav } = navSlice.actions;

// The function below is called a selector and allows us to select a value from
// the state. Selectors can also be defined inline where they're used instead of
// in the slice file. For example: `useSelector((state: RootState) => state.counter.value)`
export const selectHidden = (state: RootState) => state.navigation.hidden;

export default navSlice.reducer;
