import { applyFilters, deserialiseQuery } from "../sidebar/sections/filter/filter";
import { OrderBy } from "../sidebar/sections/sort/SortSection";
import { orderBy } from "lodash";

export function useFilterData<T>(data: Array<T>, filters?: string): Array<T> {
  if (filters) {
    return applyFilters(deserialiseQuery(filters), data || []);
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
