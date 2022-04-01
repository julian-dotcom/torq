// TODO: Create Documentation with examples

// available filter types that can be picked in the UI and a filter function implementation to achieve that
export const FilterFunctions = {
  number: {
    eq: (input: Array<any>, key: string, parameter: number) => input.filter(item => item[key] === parameter),
    neq: (input: Array<any>, key: string, parameter: number) => input.filter(item => item[key] !== parameter),
    gt: (input: Array<any>, key: string, parameter: number) => input.filter(item => item[key] > parameter),
    gte: (input: Array<any>, key: string, parameter: number) => input.filter(item => item[key] >= parameter),
    lt: (input: Array<any>, key: string, parameter: number) => input.filter(item => item[key] < parameter),
    lte: (input: Array<any>, key: string, parameter: number) => input.filter(item => item[key] <= parameter),
  },
  string: {
    include: (input: Array<any>, key: string, parameter: string) => input.filter(item => item[key].includes(parameter)),
    notInclude: (input: Array<any>, key: string, parameter: string) => input.filter(item => !item[key].includes(parameter)),
  }
}

type numberFilterType = 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte';
type stringFilterType = 'include' | 'notInclude';
type filterFuncNameType = numberFilterType | stringFilterType

// an interface for a user configured filter with the key to operate on and value to filter by
export interface FilterInterface {
  combiner: 'and' | 'or';
  category: 'number' | 'string';
  funcName: filterFuncNameType;
  key: string;
  parameter: number | string;
}

export function applyFilters(filters: Array<FilterInterface>, data: Array<any>): [] {
  // @ts-ignore
  return filters.reduce((prev: [], cur: FilterInterface) => {
    // @ts-ignore
    return FilterFunctions[cur.category as 'string' | 'number'][cur.funcName as filterNameType](prev, cur.key, cur.parameter);
  }, data)
}

