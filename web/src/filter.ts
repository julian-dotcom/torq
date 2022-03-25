// just some sample data. the below code makes the assumption that the data will be an array of objects
export const data = [
  {
    name: "Alice",
    age: 50,
    tags: ["clever", "astute"]
  },
  {
    name: "Bob",
    age: 20,
    tags: ["kind"]
  }
]

// interface for a function that will filter the data
interface FilterFunc {
  (input: Array<any>, key: string, parameter: any): Array<any>;
}

// interface for some metadata about a filter function such as what datatype it operates on and its name
interface FilterFuncObj {
  name: string,
  type: string,
  func: FilterFunc,
}

// available filter types that can be picked in the UI and a filter function implementation to achieve that
const filterFuncs: Array<FilterFuncObj> = [
  {
    name: "equals",
    type: "number",
    func: (input: Array<any>, key: string, parameter: number) => input.filter(item => item[key] === parameter)
  },
  {
    name: "matches",
    type: "string",
    func: (input: Array<any>, key: string, parameter: number) => input.filter(item => item[key].includes(parameter))
  },
  {
    name: "includes",
    type: "tag",
    func: (input: Array<any>, key: string, parameter: number) => input.filter(item => item[key].some((tag: any) => tag === parameter))
  },
]

// an interface for a user configured filter with the key to operate on and value to filter by
interface filter {
  filterFunc: FilterFuncObj,
  key: string,
  paramter: any
}

// these are the filters and the arguments that the user has picked in the UI
let filters: Array<filter> = [
  {
    filterFunc: filterFuncs[0],
    key: "age",
    paramter: 50
  },
  {
    filterFunc: filterFuncs[1],
    key: "name",
    paramter: "lice"
  },
  {
    filterFunc: filterFuncs[2],
    key: "tags",
    paramter: "astute"
  },
]

// this takes all the collected filters and applies them one by one to the original data
const new_data = filters.reduce((prev, cur) => cur.filterFunc.func(prev, cur.key, cur.paramter), data)

console.log(new_data)
