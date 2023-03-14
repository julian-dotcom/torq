import clone from "clone";
import { SelectOption } from "features/forms/Select";
import { TagResponse } from "pages/tags/tagsTypes";
export type FilterCategoryType = "number" | "string" | "date" | "boolean" | "array" | "duration" | "enum" | "tag";
export type FilterParameterType = number | string | Date | boolean | Array<unknown>;
export type FilterFunc = <T extends Record<K, unknown>, K extends keyof T>(
  input: T,
  key: K,
  parameter: FilterParameterType
) => boolean;

// available filter types that can be picked in the UI and a filter function implementation to achieve that
export const FilterFunctions = new Map<string, Map<string, FilterFunc>>([
  [
    "number",
    new Map<string, FilterFunc>([
      ["eq", (input, key, parameter) => input[key] === parameter],
      ["neq", (input, key, parameter) => input[key] !== parameter],
      ["gt", (input, key, parameter) => input[key] > parameter],
      ["gte", (input, key, parameter) => input[key] >= parameter],
      ["lt", (input, key, parameter) => input[key] < parameter],
      ["lte", (input, key, parameter) => input[key] <= parameter],
    ]),
  ],
  [
    "tag",
    new Map<string, FilterFunc>([
      [
        "any",
        (input, key, parameter) => {
          return (
            ((input[key] as Array<TagResponse>) || []).filter((v) =>
              ((parameter as Array<number>) || []).includes(v.tagId || 0)
            )?.length >= 1
          );
        },
      ],
      [
        "notAny",
        (input, key, parameter) => {
          return (
            ((input[key] as Array<TagResponse>) || []).filter((v) =>
              ((parameter as Array<number>) || []).includes(v.tagId || 0)
            )?.length === 0
          );
        },
      ],
    ]),
  ],
  [
    "duration",
    new Map<string, FilterFunc>([
      ["eq", (input, key, parameter) => input[key] === parameter],
      ["neq", (input, key, parameter) => input[key] !== parameter],
      ["gt", (input, key, parameter) => input[key] > parameter],
      ["gte", (input, key, parameter) => input[key] >= parameter],
      ["lt", (input, key, parameter) => input[key] < parameter],
      ["lte", (input, key, parameter) => input[key] <= parameter],
    ]),
  ],
  [
    "string",
    new Map<string, FilterFunc>([
      ["like", (input, key, parameter) => (input[key] as string).toLowerCase().includes(parameter as string)],
      ["notLike", (input, key, parameter) => !(input[key] as string).toLowerCase().includes(parameter as string)],
    ]),
  ],
  [
    "link",
    new Map<string, FilterFunc>([
      ["like", (input, key, parameter) => (input[key] as string).toLowerCase().includes(parameter as string)],
      ["notLike", (input, key, parameter) => !(input[key] as string).toLowerCase().includes(parameter as string)],
    ]),
  ],
  [
    "enum",
    new Map<string, FilterFunc>([
      ["eq", (input, key, parameter) => (input[key] as string).toUpperCase() === (parameter as string).toUpperCase()],
      ["neq", (input, key, parameter) => (input[key] as string).toUpperCase() !== (parameter as string).toUpperCase()],
    ]),
  ],
  [
    "boolean",
    new Map<string, FilterFunc>([
      ["eq", (input, key, parameter) => !!input[key] === parameter],
      ["neq", (input, key, parameter) => !input[key] !== parameter],
    ]),
  ],
  [
    "date",
    new Map<string, FilterFunc>([
      ["eq", (input, key, parameter) => input[key] === parameter],
      ["neq", (input, key, parameter) => input[key] !== parameter],
      ["gt", (input, key, parameter) => input[key] > parameter],
      ["gte", (input, key, parameter) => input[key] >= parameter],
      ["lt", (input, key, parameter) => input[key] < parameter],
      ["lte", (input, key, parameter) => input[key] <= parameter],
    ]),
  ],
  [
    "array",
    new Map<string, FilterFunc>([
      [
        "eq",
        (input, key, parameter) =>
          !!(input[key] as Array<unknown>).filter((value: unknown) => (parameter as Array<unknown>).includes(value)),
      ],
      [
        "neq",
        (input, key, parameter) =>
          !(input[key] as Array<unknown>).filter((value: unknown) => (parameter as Array<unknown>).includes(value)),
      ],
    ]),
  ],
]);

// an interface for a user configured filter with the key to operate on and value to filter by
export type FilterInterface = {
  category: FilterCategoryType;
  funcName: string;
  parameter: FilterParameterType;
  key: string;
  selectOptions?: Array<SelectOption>;
  value?: unknown;
  label?: string;
};

export function applyFilters<T extends Record<string, unknown>>(filters: Clause, data: Array<T>): Array<T> {
  return data.filter((item) => processQuery(filters, item));
}

class FilterClause {
  prefix = "$filter";
  constructor(public filter: FilterInterface) {}
  get length() {
    return 1;
  }
  toJSON(): object {
    return { [this.prefix]: this.filter };
  }
}

class AndClause {
  prefix = "$and";
  childClauses: Array<Clause> = [];
  constructor(childClauses?: Clause[]) {
    if (childClauses) {
      this.childClauses = childClauses;
    }
  }
  get length() {
    let length = 0;
    for (const clause of this.childClauses) {
      length += clause.length;
    }
    return length;
  }
  addChildClause(clause: Clause): void {
    this.childClauses.push(clause);
  }
  toJSON(): object {
    return { [this.prefix]: this.childClauses.map((child) => child.toJSON()) };
  }
}

class OrClause extends AndClause {
  prefix = "$or";
}

export type Clause = FilterClause | OrClause | AndClause;

type ClauseWithResult = Clause & {
  result?: boolean;
};

export interface SerialisableFilterQuery {
  $filter?: FilterInterface;
  $or?: Array<SerialisableFilterQuery>;
  $and?: Array<SerialisableFilterQuery>;
}

const parseClause = <T extends Record<string, unknown>>(clause: ClauseWithResult, data: T) => {
  // const toastRef = React.useContext(ToastContext);
  typeSwitch: switch (clause.prefix) {
    case "$filter": {
      const filterClause = clause as FilterClause;
      const filterFunc = FilterFunctions.get(filterClause?.filter?.category)?.get(filterClause?.filter?.funcName);
      if (!filterFunc) {
        throw new Error("Filter function is not yet defined");
      }
      if (data[filterClause?.filter?.key] !== undefined) {
        clause.result = filterFunc(data, filterClause?.filter?.key ?? "", filterClause?.filter?.parameter);
      } else {
        // DISABLED BECAUSE IT CAUSED RENDERING ISSUES
        // Let the user know that the filter key is not valid (most likely because the view filter data is out of date)
        // Then return true so that the filter does not remove all data
        // toastRef?.current?.addToast(`${filterClause.filter.key} is not found in the dataset.`, toastCategory.error);
        clause.result = true;
      }
      break;
    }
    case "$and": {
      for (const childClause of (clause as AndClause).childClauses) {
        // recursive call processing each child clause
        parseClause(childClause, data);
        // if any of the sibling filters are false then the AND fails, no need to process the rest
        if ((childClause as ClauseWithResult).result === false) {
          clause.result = false;
          break typeSwitch;
        }
      }
      // check that every filter is true so satisfy the AND
      if ((clause as AndClause).childClauses.every((sc) => (sc as ClauseWithResult).result === true)) {
        clause.result = true;
      }
      break;
    }
    case "$or": {
      for (const childClause of (clause as OrClause).childClauses) {
        // recursive call processing each child clause
        parseClause(childClause, data);
        // if any of the sibling filters are true then the OR succeeds, no need to process the rest
        if ((childClause as ClauseWithResult).result === true) {
          clause.result = true;
          break typeSwitch;
        }
      }
      // if we made it here all of the previous filters must have returned false so whole OR fails
      clause.result = false;
      break;
    }
  }
};

const processQuery = <T extends Record<string, unknown>>(query: ClauseWithResult, data: T): boolean => {
  // clone query to modify it and leave original untouched
  const clonedQuery = clone<ClauseWithResult>(query);
  parseClause(clonedQuery, data);
  if (clonedQuery.result === undefined) {
    throw new Error("Query result must be true or false");
  }
  return clonedQuery.result;
};

const deserialiseQuery = (query: SerialisableFilterQuery | undefined): Clause => {
  if (!query) {
    return new AndClause();
  }
  if (query.$filter) {
    return new FilterClause(query.$filter);
  }
  if (query.$and) {
    return new AndClause(query.$and.map((subclause) => deserialiseQuery(subclause)));
  }
  if (query.$or) {
    return new OrClause(query.$or.map((subclause) => deserialiseQuery(subclause)));
  }
  throw new Error("Expected JSON to contain $filter, $or or $and");
};

const deserialiseQueryFromString = (query: string): Clause => {
  return deserialiseQuery(JSON.parse(query));
};

export { FilterClause, OrClause, AndClause, processQuery, deserialiseQuery, deserialiseQueryFromString };
