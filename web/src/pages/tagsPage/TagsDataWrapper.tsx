import React, { useMemo } from "react";
import { useAppSelector } from "store/hooks";
import { selectFilters, selectSortBy } from "./tagsSlice";
import { orderBy } from "lodash";
import type { tag } from "./tagsTypes";
import { applyFilters, Clause, deserialiseQuery } from "features/sidebar/sections/filter/filter";
import clone from "clone";
import Table, { ColumnMetaData } from "features/table/Table";
import { useGetTagsQuery } from "apiSlice";

interface boxProps {
  activeColumns: ColumnMetaData[];
}

function TagsDataWrapper(props: boxProps) {
  const tagsResponse = useGetTagsQuery();

  const sortBy = useAppSelector(selectSortBy);
  const filters = useAppSelector(selectFilters);

  const [tags, columns] = useMemo(() => {
    if (tagsResponse.data?.length == 0) {
      return [];
    }

    let tags = clone<tag[]>(tagsResponse.data as tag[]) || [];
    const columns = clone<ColumnMetaData[]>(props.activeColumns) || [];

    if (filters) {
      const f = deserialiseQuery(clone<Clause>(filters));
      tags = applyFilters(f, tags);
    }
    tags = orderBy(
      tags,
      sortBy.map((s) => s.value),
      sortBy.map((s) => s.direction) as ["asc" | "desc"]
    );

    return [tags, columns];
  }, [props.activeColumns, tagsResponse.data, filters, sortBy]);

  return (
    <Table
      activeColumns={columns || []}
      data={tags as tag[]}
      isLoading={tagsResponse.isLoading || tagsResponse.isFetching || tagsResponse.isUninitialized}
    />
  );
}
const TagsDataWrapperMemo = React.memo(TagsDataWrapper);
export default TagsDataWrapperMemo;
