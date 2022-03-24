const data = [
  {
    name: "Alice",
    age: "50",
    tags: ["clever", "astute"]
  },
  {
    name: "Bob",
    age: "20",
    tags: ["kind"]
  }
]

interface FilterFunc {
  (input: Array, key: string, parameter: any): Array;
}

interface FilterFuncObj {
  name: string,
  type: string,
  func: FilterFunc,
}

const filterFuncs: Array<FilterFuncObj> = [
  {
    name: "equals",
    type: "number",
    func: (input: Array<any>, key: string, parameter: number) => input.filter(item => item["key"] === parameter)
  },
]

interface filter {
  filterFunc: FilterFuncObj,
  key: string,
  paramter: any
}

let filters: Array<filter> = [
  {
    filterFunc: filterFuncs[0],
    key: "age",
    paramter: 50
  }
]

const new_data = filters.reduce((prev, cur) => cur.filterFunc.func(prev, cur.key, cur.paramter), data)

console.log(new_data)
