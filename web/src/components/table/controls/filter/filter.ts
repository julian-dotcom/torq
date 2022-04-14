import clone from "clone"

// TODO: Create Documentation with examples

// available filter types that can be picked in the UI and a filter function implementation to achieve that
export const FilterFunctions = new Map<string, Map<string, Function>>([
  ["number", new Map<string, Function>([
    ["eq", (input: any, key: string, parameter: number) => input[key] === parameter],
    ["neq", (input: any, key: string, parameter: number) => input[key] !== parameter],
    ["gt", (input: any, key: string, parameter: number) => input[key] > parameter],
    ["gte", (input: any, key: string, parameter: number) => input[key] >= parameter],
    ["lt", (input: any, key: string, parameter: number) => input[key] < parameter],
    ["lte", (input: any, key: string, parameter: number) => input[key] <= parameter],
  ])],
  ["string", new Map<string, Function>([
    ["include", (input: any, key: string, parameter: string) => input[key].includes(parameter)],
    ["notInclude", (input: any, key: string, parameter: string) => !input[key].includes(parameter)]
  ])]
])

// an interface for a user configured filter with the key to operate on and value to filter by
export interface FilterInterface {
  combiner: 'and' | 'or';
  category: 'number' | 'string';
  funcName: string;
  key: string;
  parameter: number | string;
}

export function applyFilters(filters: Clause, data: Array<any>): any[] {
  return data.filter(item => processQuery(filters, item))
}

class FilterClause {
  prefix: string = "$filter"
  constructor(public filter: FilterInterface) { }
  toJSON(): object {
    return { [this.prefix]: this.filter }
  }
}

class AndClause {
  prefix: string = "$and"
  childClauses: Clause[] = []
  constructor(childClauses?: Clause[]) {
    if (childClauses) {
      this.childClauses = childClauses
    }
  }
  addChildClause(clause: Clause): void {
    this.childClauses.push(clause)
  }
  toJSON(): object {
    return { [this.prefix]: this.childClauses }
  }
}

class OrClause extends AndClause {
  prefix: string = "$or"
}

type Clause = FilterClause | OrClause | AndClause

type ClauseWithResult = Clause & {
  result?: boolean
}

const parseClause = (clause: ClauseWithResult, data: any) => {
  typeSwitch: switch (clause.prefix) {
    case "$filter": {
      const filterClause = clause as FilterClause
      const filterFunc = FilterFunctions.get(filterClause.filter.category)?.get(filterClause.filter.funcName)
      if (!filterFunc) {
        throw new Error("Filter function is not yet defined")
      }
      clause.result = filterFunc(data, filterClause.filter.key, filterClause.filter.parameter)
      break;
    }
    case "$and": {
      for (const childClause of (clause as AndClause).childClauses) {
        // recursive call processing each child clause
        parseClause(childClause, data)
        // if any of the sibling filters are false then the AND fails, no need to process the rest
        if ((childClause as ClauseWithResult).result === false) {
          clause.result = false
          break typeSwitch
        }
      }
      // check that every filter is true so satisfy the AND
      if ((clause as AndClause).childClauses.every(sc => (sc as ClauseWithResult).result === true)) {
        clause.result = true
      }
      break;
    }
    case "$or": {
      for (const childClause of (clause as OrClause).childClauses) {
        // recursive call processing each child clause
        parseClause(childClause, data)
        // if any of the sibling filters are true then the OR succeeds, no need to process the rest
        if ((childClause as ClauseWithResult).result === true) {
          clause.result = true
          break typeSwitch
        }
      }
      // if we made it here all of the previous filters must have returned false so whole OR fails
      clause.result = false
      break;
    }
  }
}

const processQuery = (query: any, data: any): boolean => {
  // clone query to modify it and leave original untouched
  const clonedQuery = clone<ClauseWithResult>(query)
  parseClause(clonedQuery, data)
  if (clonedQuery.result === undefined) {
    throw new Error("Query result must be true or false")
  }
  return clonedQuery.result
}

const deserialiseQuery = (query: any): Clause => {
  if (Object.keys(query)[0] === "$filter") {
    return new FilterClause(query.$filter)
  }
  if (Object.keys(query)[0] === "$and") {
    return new AndClause(query.$and.map((subclause: Clause) => deserialiseQuery(subclause)))
  }
  if (Object.keys(query)[0] === "$or") {
    return new OrClause(query.$or.map((subclause: Clause) => deserialiseQuery(subclause)))
  }
  throw new Error("Expected JSON to contain $filter, $or or $and")
}

const deserialiseQueryJSON = (queryJSON: string): Clause => {
  return deserialiseQuery(JSON.parse(queryJSON))
}

export { FilterClause, OrClause, AndClause, processQuery, deserialiseQueryJSON }
export type { Clause }
