import React, { useMemo } from "react";

import { orderBy } from "lodash";
import { useAppSelector } from "store/hooks";
import { useGetChannelsQuery } from "apiSlice";
import { selectFilters, selectGroupBy, selectSortBy } from "./ChannelsSlice";
import type { channel } from "apiTypes";
import { applyFilters, Clause, deserialiseQuery } from "features/sidebar/sections/filter/filter";
import { groupByFn } from "features/sidebar/sections/group/groupBy";
import clone from "clone";
import Table, { ColumnMetaData } from "features/table/Table";

interface boxProps {
  activeColumns: ColumnMetaData[];
}

function ChannelsDataWrapper(props: boxProps) {
  const chanResponse = useGetChannelsQuery();

  const sortBy = useAppSelector(selectSortBy);
  const groupBy = useAppSelector(selectGroupBy) || "channels";
  const filters = useAppSelector(selectFilters);

  const [channels, columns] = useMemo(() => {
    if (chanResponse.data?.length == 0) {
      return [];
    }

    let channels = clone<channel[]>(chanResponse.data as channel[]) || [];
    const columns = clone<ColumnMetaData[]>(props.activeColumns) || [];

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

    if (channels.length > 0) {
      for (const channel of channels) {
        for (const column of columns) {
          column.total = (column.total ?? 0) + channel.gauge;
          column.max = 100;
        }
      }
    }
    return [channels, columns];
  }, [props.activeColumns, chanResponse.data, filters, groupBy, sortBy]);

  return (
    <Table
      activeColumns={columns || []}
      data={channels as channel[]}
      isLoading={chanResponse.isLoading || chanResponse.isFetching || chanResponse.isUninitialized}
    />
  );
}
const ChannelsDataWrapperMemo = React.memo(ChannelsDataWrapper);
export default ChannelsDataWrapperMemo;
