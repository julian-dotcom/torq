import { createSlice } from "@reduxjs/toolkit";
import { RootState } from "store/store";
import mixpanel from "mixpanel-browser";

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

  reducers: {
    updateFlowKey: (state, action) => {
      mixpanel.track("Update FlowChart Key", { key: action.payload.flowKey });
      state.flowKey = action.payload.flowKey;
    },
    updateProfitChartKey: (state, action) => {
      mixpanel.track("Update ProfitChart Key", { key: action.payload.key });
      state.profitChartKey = action.payload.key;
    },
    updateEventChartKey: (state, action) => {
      mixpanel.track("Update EventChart Key", { key: action.payload.key });
      state.eventChartKey = action.payload.key;
    },
    updateBalanceChanID: (state, action) => {
      mixpanel.track("Update Balance Channel ID", {
        channel_id: action.payload.balanceChanID,
      });
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
