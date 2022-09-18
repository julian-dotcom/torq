import React, { useMemo } from "react";

import { cloneDeep, orderBy } from "lodash";
import { useAppSelector } from "store/hooks";
import { selectTimeInterval } from "../timeIntervalSelect/timeIntervalSlice";
import { addDays, format } from "date-fns";
import { useGetForwardsQuery } from "../../apiSlice";
import { selectFilters, selectGroupBy, selectSortBy } from "./forwardsSlice";
import { applyFilters, Clause, deserialiseQuery } from "../sidebar/sections/filter/filter";
import { groupByFn } from "../sidebar/sections/group/groupBy";
import clone from "../../clone";
import Table, { ColumnMetaData } from "../table/Table";

interface boxProps {
  activeColumns: ColumnMetaData[];
}

function ForwardsDataWrapper(props: boxProps) {
  const currentPeriod = useAppSelector(selectTimeInterval);
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd");

  const chanResponse = useGetForwardsQuery({ from: from, to: to });

  // const columns = useAppSelector(selectAllColumns);
  const sortBy = useAppSelector(selectSortBy);
  const groupBy = useAppSelector(selectGroupBy) || "channels";
  const filters = useAppSelector(selectFilters);

  // const data = chanResponse.data || [];

  const [channels, columns] = useMemo(() => {
    if (chanResponse.data?.length == 0) {
      return [];
    }
    let channels = cloneDeep(chanResponse.data ? chanResponse.data : ([] as any[]));
    const columns = clone<Array<ColumnMetaData>>(props.activeColumns) || [];

    if (channels.length > 0) {
      channels = groupByFn(channels, groupBy || "channels");
    }
    if (filters) {
      const f = deserialiseQuery(clone<Clause>(filters));
      channels = applyFilters(f, channels);
    }
    channels = orderBy(
      channels,
      sortBy.map((s) => s.value),
      sortBy.map((s) => s.direction) as ["asc" | "desc"]
    );

    /* const channels = chanResponse.data || []; */
    if (channels.length > 0) {
      for (const channel of channels) {
        for (const column of columns) {
          column.total = (column.total ?? 0) + channel[column.key];
          column.max = Math.max(column.max ?? 0, channel[column.key] ?? 0);
        }
      }

      const turnover_total_col = columns.find((col) => col.key === "turnover_total");
      const turnover_out_col = columns.find((col) => col.key === "turnover_out");
      const turnover_in_col = columns.find((col) => col.key === "turnover_in");
      const amount_total_col = columns.find((col) => col.key === "amount_total");
      const amount_out_col = columns.find((col) => col.key === "amount_out");
      const amount_in_col = columns.find((col) => col.key === "amount_in");
      const capacity_col = columns.find((col) => col.key === "capacity");

      if (turnover_total_col) {
        turnover_total_col.total = (amount_total_col?.total ?? 0) / (capacity_col?.total ?? 1);
      }

      if (turnover_out_col) {
        turnover_out_col.total = (amount_out_col?.total ?? 0) / (capacity_col?.total ?? 1);
      }

      if (turnover_in_col) {
        turnover_in_col.total = (amount_in_col?.total ?? 0) / (capacity_col?.total ?? 1);
      }
    }
    return [channels, columns];
  }, [props.activeColumns, chanResponse.data, filters, groupBy, sortBy]);

  return (
    <Table
      activeColumns={columns || []}
      data={channels}
      isLoading={chanResponse.isLoading || chanResponse.isFetching || chanResponse.isUninitialized}
      showTotals={true}
    />
  );
}
const ForwardsDataWrapperMemo = React.memo(ForwardsDataWrapper);
export default ForwardsDataWrapperMemo;
