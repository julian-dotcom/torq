import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RootState } from "store/store";

const initialState = {
  displayPreviewLink: false,
};

export const workflowSlice = createSlice({
  name: "workflowpage",
  initialState,
  reducers: {
    setDisplayPreviewLink: (state, action: PayloadAction<boolean>) => {
      state.displayPreviewLink = action.payload;
    },
  },
});

export default workflowSlice.reducer;

export const selectDisplayPreviewLink = (state: RootState) => {
  return state.workflow.displayPreviewLink;
};

export const { setDisplayPreviewLink } = workflowSlice.actions;
