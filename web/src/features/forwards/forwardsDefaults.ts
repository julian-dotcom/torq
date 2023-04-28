import { AndClause, FilterInterface } from "features/sidebar/sections/filter/filter";
import { ViewResponse } from "features/viewManagement/types";
import { OrderBy } from "features/sidebar/sections/sort/SortSection";
import { Forward } from "features/forwards/forwardsTypes";
import { AllForwardsColumns } from "features/forwards/forwardsColumns.generated";

export const ForwardsFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number",
  parameter: 0,
  key: "amountTotal",
};

const defaultColumns: Array<keyof Forward> = [
  "alias",
  "revenueOut",
  "countTotal",
  "amountOut",
  "amountIn",
  "amountTotal",
  "turnoverTotal",
  "capacity",
];

export const ForwardsSortByTemplate: Array<OrderBy> = [{ key: "revenueOut", direction: "desc" }];

export const DefaultForwardsColumns = AllForwardsColumns.filter((c) => defaultColumns.includes(c.key));

export const DefaultForwardsView: ViewResponse<Forward> = {
  page: "forwards",
  dirty: true,
  view: {
    title: "Draft View",
    filters: new AndClause().toJSON(),
    columns: DefaultForwardsColumns,
    sortBy: ForwardsSortByTemplate,
    groupBy: "channels",
  },
};
