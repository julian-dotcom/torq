import classNames from "classnames";
import { Delete16Regular as RemoveIcon } from "@fluentui/react-icons";
import Select, { SelectOptionType } from "./FilterDropDown";

import { FilterClause } from "./filter";
import styles from "./filter-section.module.scss";
import { FilterFunctions } from "./filter";
import NumberFormat from "react-number-format";
import { useState } from "react";
import { format } from "d3";

const formatterDetailed = format(",.2f");
const formatter = format(",.0f");

function formatParameter(value: number) {
  return value % 1 != 0 ? value : formatter(value);
}

const ffLabels = new Map<string, string>([
  ["eq", "is equal to"],
  ["neq", "is not equal to"],
  ["gt", "is greater than"],
  ["gte", "is greater than or equal to"],
  ["lt", "is less than"],
  ["lte", "is less than or equal to"],
  ["include", "includes"],
  ["notInclude", "does not include"],
]);

//   mapSelectOptionType[] = [
//   { value: "and", label: "And" },
//   { value: "or", label: "Or" },
// ];

function getFilterFunctions(filterCategory: "number" | "string") {
  const filterFuncs = FilterFunctions.get(filterCategory)?.entries();
  if (!filterFuncs) {
    throw new Error("Filter category not found in list of filters");
  }
  return Array.from(filterFuncs, ([key, _]) => ({ value: key, label: ffLabels.get(key) ?? "Label not found" }));
}

interface filterRowInterface {
  index: number;
  child: boolean;
  filterClause: FilterClause;
  columnOptions: SelectOptionType[];
  onUpdateFilter: Function;
  onRemoveFilter: Function;
  handleCombinerChange: Function;
  combiner?: string;
}

function FilterRow({
  index,
  child,
  filterClause,
  columnOptions,
  onUpdateFilter,
  onRemoveFilter,
  handleCombinerChange,
  combiner,
}: filterRowInterface) {
  const [rowExpanded, setRowExpanded] = useState(false);

  const rowValues = filterClause.filter;

  let functionOptions = getFilterFunctions(rowValues.category);

  const keyOption = columnOptions.find((item) => item.value === rowValues.key);
  if (!keyOption) {
    throw new Error("key option not found");
  }
  const funcOption = functionOptions.find((item) => item.value === rowValues.funcName);
  if (!funcOption) {
    throw new Error("func option not found");
  }

  let selectData = {
    func: funcOption,
    key: keyOption,
  };

  const handleKeyChange = (item: any) => {
    rowValues.key = item.value;
    onUpdateFilter();
  };

  const handleFunctionChange = (item: any) => {
    rowValues.funcName = item.value;
    onUpdateFilter();
  };

  const handleParamChange = (e: any) => {
    if (rowValues.category === "number") {
      rowValues.parameter = e.floatValue;
    } else {
      rowValues.parameter = e.target.value ? e.target.value : "";
    }
    onUpdateFilter();
  };

  const label = columnOptions.find((item) => item.value === rowValues.key)?.label;

  return (
    <div className={classNames(styles.filter, { [styles.first]: !index })}>
      <div className={styles.filterKeyContainer}>
        {index !== 0 && (
          <div
            className={styles.combinerContainer}
            onClick={() => {
              handleCombinerChange();
            }}
          >
            {combiner}
          </div>
        )}
        <div className={styles.filterKeyLabel} onClick={() => setRowExpanded(!rowExpanded)}>
          {label}
          <span className={styles.filterFunctionLabel}> {funcOption.label} </span>
          {rowValues.category === "number" ? formatParameter(rowValues.parameter as number) : rowValues.parameter}
        </div>
        <div className={classNames(styles.removeFilter, styles.desktopRemove)} onClick={() => onRemoveFilter(index)}>
          <RemoveIcon />
        </div>
      </div>

      <div className={classNames(styles.filterOptions, { [styles.expanded]: rowExpanded })}>
        <Select
          selectProps={{ options: columnOptions, value: selectData.key, onChange: handleKeyChange }}
          child={child}
        />

        <div className="filter-function-container">
          <Select
            selectProps={{ options: functionOptions, value: selectData.func, onChange: handleFunctionChange }}
            child={child}
          />
        </div>

        <div className="filter-parameter-container">
          {rowValues.category === "number" ? (
            <NumberFormat
              className={classNames(styles.filterInputField, styles.small)}
              thousandSeparator=","
              value={rowValues.parameter}
              onValueChange={handleParamChange}
            />
          ) : (
            <input
              type="text"
              className={"torq-input-field"}
              value={rowValues.parameter}
              onChange={handleParamChange}
            />
          )}
        </div>
      </div>
    </div>
  );
}

export default FilterRow;
