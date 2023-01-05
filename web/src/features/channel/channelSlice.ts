import { createSlice } from "@reduxjs/toolkit";
import { RootState } from "store/store";

type keyOption = { value: "revenue" | "amount" | "count"; label: string };

export type ChannelState = {
  flowKey: keyOption;
  eventChartKey: keyOption;
  profitChartKey: keyOption;
  balanceChanID: { value: number; label: string };
};

const initialState: ChannelState = {
  profitChartKey: { value: "revenue", label: "Revenue" },
  eventChartKey: { value: "amount", label: "Amount" },
  flowKey: { value: "amount", label: "Amount" },
  balanceChanID: { value: 0, label: "" },
};

export const channelSlice = createSlice({
  name: "channel",
  initialState,

  // The `reducers` field lets us define reducers and generate associated actions
  reducers: {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    updateFlowKey: (state: any, action: any) => {
      state.flowKey = action.payload.flowKey;
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    updateProfitChartKey: (state: any, action: any) => {
      state.profitChartKey = action.payload.key;
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    updateEventChartKey: (state: any, action: any) => {
      state.eventChartKey = action.payload.key;
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    updateBalanceChanID: (state: any, action: any) => {
      state.balanceChanID = action.payload.key;
    },
  },
});

export const { updateFlowKey, updateProfitChartKey, updateEventChartKey, updateBalanceChanID } = channelSlice.actions;

// The function below is called a selector and allows us to select a value from
// the state. Selectors can also be defined inline where they're used instead of
// in the slice file. For example: `useSelector((state: RootState) => state.counter.value)`
export const selectFlowKeys = (state: RootState) => state.channel.flowKey;
export const selectProfitChartKey = (state: RootState) => state.channel.profitChartKey;
export const selectEventChartKey = (state: RootState) => state.channel.eventChartKey;
export const selectBalanceChanID = (state: RootState) => state.channel.balanceChanID;

export default channelSlice.reducer;
