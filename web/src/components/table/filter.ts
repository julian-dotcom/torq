// TODO: Create Documentation with example
// just some sample data. the below code makes the assumption that the data will be an array of objects
// export const data = [
//   {
//     name: "Alice",
//     age: 50,
//     tags: ["clever", "astute"]
//   },
//   {
//     name: "Bob",
//     age: 100,
//     tags: ["kind"]
//   }
// ]



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

type filterNames = 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'include' | 'notInclude'

// an interface for a user configured filter with the key to operate on and value to filter by
export interface FilterInterface {
  filterCategory: 'string' | 'number',
  filterName: filterNames,
  key: string,
  parameter: any
}

export function applyFilters(filters: Array<FilterInterface>, data: Array<any>): [] {
  // @ts-ignore
  return filters.reduce((prev: [], cur: FilterInterface) => {
    // @ts-ignore
    return FilterFunctions[cur.filterCategory as 'string' | 'number'][cur.filterName as filterNames](prev, cur.key, cur.parameter);
  }, data)
}

// these are the filters and the arguments that the user has picked in the UI
// let filters: Array<FilterInterfaceB> = [
//   {
//     filterFunc: FilterFunctions.number.eq,
//     key: "age",
//     parameter: 50
//   },
//   {
//     filterFunc: FilterFunctions.string.include,
//     key: "name",
//     parameter: "lice"
//   },
//   // {
//   //   filterFunc: filterFuncs[2],
//   //   key: "tags",
//   //   parameter: "astute"
//   // },
// ]

// console.log(data)
// // this takes all the collected filters and applies them one by one to the original data
// const new_data = filters.reduce((prev, cur) => cur.filterFunc(prev, cur.key, cur.parameter), data)
// console.log(new_data)
