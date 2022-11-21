export const a = 1;
// import { createSlice, PayloadAction } from "@reduxjs/toolkit";
// import { RootState } from "store/store";
// import { SortByOptionType } from "features/sidebar/sections/sort/SortSectionOld";
// import { ColumnMetaData } from "features/table/types";
// import { PageViewType, ViewInterface } from "features/viewManagement/types";
// import { torqApi } from "apiSlice";
// import { tag } from "pages/tagsPage/tagsTypes";
// import { channel } from "features/channels/channelsTypes";
// import { ForwardResponse } from "types/api";
// import { Payment } from "features/transact/Payments/types";
// import { OnchainTx } from "features/transact/OnChain/types";
// import { Invoice } from "types/invoice";
//
// export type ViewsState = {
//   tagsViews: PageViewType<tag>;
//   channelsViews: PageViewType<channel>;
//   forwardsViews: PageViewType<ForwardResponse>;
//   paymentViews: PageViewType<Payment>;
//   onChainViews: PageViewType<OnchainTx>;
//   invoiceViews: PageViewType<Invoice>;
//   // status: "idle" | "loading" | "failed";
// };
//
// // export const DefaultView: ViewInterface<T> = {
// //   title: "Untitled View",
// //   saved: true,
// //   columns: activeChannelsColumns,
// //   filters: defaultFilter.toJSON(),
// //   sortBy: [],
// //   groupBy: undefined,
// //   page: "channels",
// // };
//
// const initialState: ViewsState = {
//   tagsViews: {
//     selectedViewIndex: 0,
//     views: [
//       {
//         page: "tags",
//         title: "Untitled View",
//         saved: true,
//         columns: [],
//         filters: {},
//         sortBy: [],
//         groupBy: undefined,
//       },
//     ],
//   },
//   channelsViews: {
//     selectedViewIndex: 0,
//     views: [
//       {
//         page: "tags",
//         title: "Untitled View",
//         saved: true,
//         columns: [],
//         filters: {},
//         sortBy: [],
//         groupBy: undefined,
//       },
//     ],
//   },
//   forwardsViews: {
//     selectedViewIndex: 0,
//     views: [
//       {
//         page: "tags",
//         title: "Untitled View",
//         saved: true,
//         columns: [],
//         filters: {},
//         sortBy: [],
//         groupBy: undefined,
//       },
//     ],
//   },
//   paymentViews: {
//     selectedViewIndex: 0,
//     views: [
//       {
//         page: "tags",
//         title: "Untitled View",
//         saved: true,
//         columns: [],
//         filters: {},
//         sortBy: [],
//         groupBy: undefined,
//       },
//     ],
//   },
//   onChainViews: {
//     selectedViewIndex: 0,
//     views: [
//       {
//         page: "tags",
//         title: "Untitled View",
//         saved: true,
//         columns: [],
//         filters: {},
//         sortBy: [],
//         groupBy: undefined,
//       },
//     ],
//   },
//   invoiceViews: {
//     selectedViewIndex: 0,
//     views: [
//       {
//         page: "tags",
//         title: "Untitled View",
//         saved: true,
//         columns: [],
//         filters: {},
//         sortBy: [],
//         groupBy: undefined,
//       },
//     ],
//   },
// };
//
// export const viewsSlice = createSlice({
//   name: "views",
//   initialState,
//   // The `reducers` field lets us define reducers and generate associated actions
//   reducers: {
//     updateFilters: (state: ViewsState, actions: PayloadAction<{ filters: FilterInterface }>) => {
//       // state.views[state.selectedViewIndex].filters = actions.payload.filters;
//     },
//     updateColumns: (state: ViewsState, actions: PayloadAction<{ columns: Array<ColumnMetaData<channel>> }>) => {
//       // state.views[state.selectedViewIndex].columns = actions.payload.columns;
//     },
//     updateGroupBy: (state: ViewsState, actions: PayloadAction<{ groupBy: string }>) => {
//       // state.views[state.selectedViewIndex].groupBy = actions.payload.groupBy;
//     },
//     updateSortBy: (state: ViewsState, actions: PayloadAction<{ sortBy: SortByOptionType[] }>) => {
//       // state.views[state.selectedViewIndex].sortBy = actions.payload.sortBy;
//     },
//     updateViews: (state: ViewsState, actions: PayloadAction<{ views: ViewInterface<channel>[]; index: number }>) => {
//       // state.views = actions.payload.views;
//       // state.selectedViewIndex = actions.payload.index;
//     },
//     updateViewsOrder: (
//       state: ViewsState,
//       actions: PayloadAction<{ views: ViewInterface<channel>[]; index: number }>
//     ) => {
//       // state.views = actions.payload.views;
//       // state.selectedViewIndex = actions.payload.index;
//     },
//     deleteView: (state: ViewsState, actions: PayloadAction<{ view: ViewInterface<channel>; index: number }>) => {
//       // state.views = [
//       //   ...state.views.slice(0, actions.payload.index),
//       //   ...state.views.slice(actions.payload.index + 1, state.views.length),
//       // ];
//       // state.selectedViewIndex = 0;
//     },
//     updateSelectedView: (state: ViewsState, actions: PayloadAction<{ index: number }>) => {
//       // state.selectedViewIndex = actions.payload.index;
//     },
//   },
//   // The `extraReducers` field lets the slice handle actions defined elsewhere,
//   // including actions generated by createAsyncThunk or in other slices.
//   extraReducers: (builder) => {
//     builder.addMatcher(
//       (action) => {
//         return (
//           [
//             "channels/updateFilters",
//             "channels/updateSortBy",
//             "channels/updateColumns",
//             "channels/updateGroupBy",
//           ].findIndex((item) => action.type === item) !== -1
//         );
//       },
//       (state, _) => {
//         state.views[state.selectedViewIndex].saved = false;
//       }
//     );
//
//     builder.addMatcher(torqApi.endpoints.createTableView.matchFulfilled, (state, { payload }) => {
//       state.views[payload.index] = {
//         ...payload.view.view,
//         id: payload.view.id,
//       };
//       state.selectedViewIndex = payload.index;
//     });
//
//     builder.addMatcher(torqApi.endpoints.deleteTableView.matchFulfilled, (state, { payload }) => {
//       state.views = [
//         ...state.views.slice(0, payload.index),
//         ...state.views.slice(payload.index + 1, state.views.length),
//       ];
//       state.selectedViewIndex = 0;
//     });
//
//     builder.addMatcher(torqApi.endpoints.getTableViews.matchFulfilled, (state, { payload }) => {
//       if (payload !== null) {
//         state.views = payload.map((view: { id: number; view: ViewInterface<channel> }) => {
//           return { ...view.view, id: view.id };
//         });
//       }
//     });
//
//     builder.addMatcher(torqApi.endpoints.updateTableView.matchFulfilled, (state, { payload }) => {
//       const view = state.views.find((v) => v.id === payload.id);
//       if (view) {
//         view.saved = true;
//       }
//     });
//   },
// });
//
// export const {
//   updateFilters,
//   updateViews,
//   updateViewsOrder,
//   deleteView,
//   updateSelectedView,
//   updateSortBy,
//   updateColumns,
//   updateGroupBy,
// } = channelsSlice.actions;
//
// export const selectActiveColumns = (state: RootState) => {
//   return state.channels.views[state.channels.selectedViewIndex].columns || [];
// };
// export const selectAllColumns = (_: RootState) => channelsColumns;
// export const selectSortBy = (state: RootState) => state.channels.views[state.channels.selectedViewIndex].sortBy;
// export const selectGroupBy = (state: RootState) => state.channels.views[state.channels.selectedViewIndex].groupBy;
// export const selectFilters = (state: RootState) => {
//   return state.channels.views[state.channels.selectedViewIndex].filters;
// };
// export const selectViews = (state: RootState) => state.channels.views;
// export const selectCurrentView = (state: RootState) => state.channels.views[state.channels.selectedViewIndex];
// export const selectedViewIndex = (state: RootState) => state.channels.selectedViewIndex;
//
// export default channelsSlice.reducer;
