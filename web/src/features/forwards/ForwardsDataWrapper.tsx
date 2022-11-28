export const aaa = 1;
// import React from "react";
// // import { orderBy } from "lodash";
// import { useAppSelector } from "store/hooks";
// import { selectTimeInterval } from "features/timeIntervalSelect/timeIntervalSlice";
// import { addDays, format } from "date-fns";
// import { useGetForwardsQuery } from "apiSlice";
// // import { selectFilters, selectGroupBy, selectSortBy } from "./forwardsSlice";
// // import { applyFilters, Clause, deserialiseQuery } from "features/sidebar/sections/filter/filter";
// // import { groupByFn } from "features/sidebar/sections/group/groupBy";
// // import clone from "clone";
// import Table from "features/table/Table";
// // import { ColumnMetaData } from "features/table/types";
// // import { Forward } from "types/api";
// import { Forward } from "types/api";
// // import { SortByOptionType } from "../sidebar/sections/sort/SortSectionOld";
//
// interface ForwardsDataWrapperProps {
//   selectedView: number;
// }
//
// function ForwardsDataWrapper(props: ForwardsDataWrapperProps) {
//   const currentPeriod = useAppSelector(selectTimeInterval);
//   const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
//   const to = format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd");
//
//   const chanResponse = useGetForwardsQuery<{
//     data: Array<Forward>;
//     isLoading: boolean;
//     isFetching: boolean;
//     isUninitialized: boolean;
//     isSuccess: boolean;
//   }>({ from: from, to: to });
//
//   // const sortBy = useAppSelector(selectSortBy);
//   // const groupBy = useAppSelector(selectGroupBy) || "channels";
//   // const filters = useAppSelector(selectFilters);
//
//   // const [channels, columns] = useMemo(() => {
//   //   if (chanResponse.data?.length == 0) {
//   //     return [];
//   //   }
//   //   let channels = clone<ForwardResponse[]>(chanResponse.data ? chanResponse.data : []);
//   //   const columns = clone<ColumnMetaData<ForwardResponse>[]>(props.activeColumns) || [];
//   //
//   //   if (channels.length > 0) {
//   //     channels = groupByFn(channels, groupBy || "channels");
//   //   }
//   //
//   //   if (filters) {
//   //     const f = deserialiseQuery(clone<Clause>(filters));
//   //     channels = applyFilters(f, channels);
//   //   }
//   //
//   //   // channels = orderBy(
//   //   //   channels,
//   //   //   sortBy.map((value: SortByOptionType) => (value ? value.value : {}))
//   //   //   sortBy.map((value: SortByOptionType) => (value ? value.direction : "asc"))
//   //   // );
//   //
//   //   if (channels.length > 0) {
//   //     for (const channel of channels) {
//   //       for (const column of columns) {
//   //         if (typeof channel[column.key as keyof ForwardResponse] == "number") {
//   //           if (!column.total) column.total = 0;
//   //           column.total += (channel[column.key as keyof ForwardResponse] as number) || 0;
//   //           column.max = Math.max(column.max ?? 0, (channel[column.key as keyof ForwardResponse] as number) || 0);
//   //         } else {
//   //           column.total = 0;
//   //           column.max = 0;
//   //         }
//   //       }
//   //     }
//   //
//   //     const turnover_total_col = columns.find((col) => col.key === "turnoverTotal");
//   //     const turnover_out_col = columns.find((col) => col.key === "turnoverOut");
//   //     const turnover_in_col = columns.find((col) => col.key === "turnoverIn");
//   //     const amount_total_col = columns.find((col) => col.key === "amountTotal");
//   //     const amount_out_col = columns.find((col) => col.key === "amountOut");
//   //     const amount_in_col = columns.find((col) => col.key === "amountIn");
//   //     const capacity_col = columns.find((col) => col.key === "capacity");
//   //
//   //     if (turnover_total_col) {
//   //       turnover_total_col.total = (amount_total_col?.total ?? 0) / (capacity_col?.total ?? 1);
//   //     }
//   //
//   //     if (turnover_out_col) {
//   //       turnover_out_col.total = (amount_out_col?.total ?? 0) / (capacity_col?.total ?? 1);
//   //     }
//   //
//   //     if (turnover_in_col) {
//   //       turnover_in_col.total = (amount_in_col?.total ?? 0) / (capacity_col?.total ?? 1);
//   //     }
//   //   }
//   //   return [channels, columns];
//   // }, [props.activeColumns, chanResponse.data, filters, groupBy, sortBy]);
//
//   return (
//     <Table
//       activeColumns={activeForwardsColumns}
//       data={chanResponse?.data || []}
//       cellRenderer={forwardsCellRenderer}
//       isLoading={chanResponse.isLoading || chanResponse.isFetching || chanResponse.isUninitialized}
//       showTotals={true}
//     />
//   );
// }
// const ForwardsDataWrapperMemo = React.memo(ForwardsDataWrapper);
// export default ForwardsDataWrapperMemo;
