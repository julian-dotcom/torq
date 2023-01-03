import { applyFilters, deserialiseQuery, FilterQueryObject } from "features/sidebar/sections/filter/filter";
import { OrderBy } from "features/sidebar/sections/sort/SortSection";
import { orderBy } from "lodash";

export function useFilterData<T extends Record<string, unknown>>(
  data: Array<T>,
  filters?: FilterQueryObject
): Array<T> {
  if (filters && data) {
    return applyFilters(deserialiseQuery(filters), data);
  } else {
    return data;
  }
}

export function useSortData<T>(data: Array<T>, sortBy?: Array<OrderBy>): Array<T> {
  if (sortBy) {
    const keys = sortBy.map((s) => s.key);
    const directions = sortBy.map((s) => s.direction);
    return orderBy(data, keys, directions);
  } else {
    return data;
  }
}
