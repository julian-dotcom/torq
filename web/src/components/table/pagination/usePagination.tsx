import React, { useState } from "react";
import Pagination from "./Pagination";
import useLocalStorage from "features/helpers/useLocalStorage";

export function usePagination<T>(page: string): [(total: number) => React.ReactNode, number, number] {
  const [limit, setLimit] = useLocalStorage(`${page.toString()}Limit`, 100);
  const [offset, setOffset] = useState(0);

  const getPaginator = (total: number) => (
    <Pagination limit={limit} offset={offset} total={total} perPageHandler={setLimit} offsetHandler={setOffset} />
  );

  return [getPaginator, limit, offset];
}
