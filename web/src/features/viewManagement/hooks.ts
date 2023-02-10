import { applyFilters, deserialiseQuery, SerialisableFilterQuery } from "features/sidebar/sections/filter/filter";
import { OrderBy } from "features/sidebar/sections/sort/SortSection";
import clone from "clone";

export function useFilterData<T extends Record<string, unknown>>(
  data: Array<T>,
  filters?: SerialisableFilterQuery
): Array<T> {
  if (filters && data) {
    return applyFilters(deserialiseQuery(filters), data);
  } else {
    return data;
  }
}

export function useSortData<T>(data: Array<T>, orderByList?: Array<OrderBy>): Array<T> {
  if (orderByList) {
    data = clone(data); // to allow editing read only data
    orderByList = clone(orderByList); // to allow reversing the readonly list
    for (const orderBy of orderByList.reverse()) {
      data = sortByFn(data, orderBy.key as keyof T, orderBy.direction);
    }
  }
  return data;
}

export const sortByFn = <T, U extends keyof T>(list: T[], index: U, direction: "asc" | "desc"): T[] => {
  if (!list) return [];
  return list.sort((a, b) => {
    if (a[index] < b[index]) return direction === "asc" ? -1 : 1;
    if (a[index] > b[index]) return direction === "asc" ? 1 : -1;
    return 0;
  });
};
