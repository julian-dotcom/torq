import { createSlice } from "@reduxjs/toolkit";
import { RootState } from "../../store/store";
import { FlowData } from "../charts/flowChartCanvas";

export interface NavState {
  flowKey: { value: "revenue" | "amount" | "count"; label: string };
}

const initialState: NavState = {
  flowKey: { value: "amount", label: "Amount" },
};

export const flowChartSlice = createSlice({
  name: "flowChart",
  initialState,

  // The `reducers` field lets us define reducers and generate associated actions
  reducers: {
    updateFlowKey: (state: any, action: any) => {
      state.flowKey = action.payload.flowKey;
    },
  },
});

export const { updateFlowKey } = flowChartSlice.actions;

// The function below is called a selector and allows us to select a value from
// the state. Selectors can also be defined inline where they're used instead of
// in the slice file. For example: `useSelector((state: RootState) => state.counter.value)`
export const selectFlowKeys = (state: RootState) => state.flowChart.flowKey;

export default flowChartSlice.reducer;
