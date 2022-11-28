import classNames from "classnames";
import { Delete16Regular as RemoveIcon } from "@fluentui/react-icons";
import Select from "./FilterDropDown";

import { FilterClause, FilterParameterType } from "./filter";
import styles from "./filter-section.module.scss";
import { FilterFunctions } from "./filter";
import NumberFormat from "react-number-format";
import { useState } from "react";
import { format } from "d3";
import { FilterCategoryType } from "./filter";

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
  ["like", "contains"],
  ["notLike", "does not contain"],
  ["any", "does not contain"],
  ["notAny", "does not contain"],
]);

function getFilterFunctions(filterCategory: FilterCategoryType) {
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
  filterOptions: Array<{ label: string; value: any; valueType?: FilterCategoryType; selectOptions?: Array<any> }>;
  onUpdateFilter: () => void;
  onRemoveFilter: (index: number) => void;
  handleCombinerChange: () => void;
  combiner?: string;
}

function FilterRow({
  index,
  child,
  filterClause,
  filterOptions,
  onUpdateFilter,
  onRemoveFilter,
  handleCombinerChange,
  combiner,
}: filterRowInterface) {
  const rowValues = filterClause.filter;

  const [rowExpanded, setRowExpanded] = useState(!rowValues.key);

  const functionOptions = getFilterFunctions(rowValues.category);

  const keyOption = filterOptions.find((item) => item.value === rowValues.key);
  const funcOption = functionOptions.find((item) => item.value === rowValues.funcName);

  const selectData = {
    func: funcOption,
    key: keyOption,
  };

  const handleKeyChange = (item: any) => {
    const newRow = { ...rowValues };
    newRow.key = item.value;
    const newCategory = filterOptions.find((item: any) => item.value === newRow.key)?.valueType || "number";
    switch (newCategory) {
      case "number":
        newRow.parameter = 0;
        newRow.funcName = "gte";
        break;
      case "boolean":
        if (newCategory !== newRow.category) {
          newRow.parameter = true;
        }
        newRow.funcName = "eq";
        break;
      case "date": {
        const nd = new Date().toISOString().slice(0, 10) + "T00:00:00";
        newRow.parameter = nd;
        newRow.funcName = "gte";
        break;
      }
      case "array":
        newRow.parameter = "";
        newRow.funcName = "eq";
        break;
      default:
        newRow.parameter = "";
        newRow.funcName = "like";
    }
    newRow.category = newCategory;
    filterClause.filter = newRow;
    onUpdateFilter();
  };

  const handleFunctionChange = (item: any) => {
    const newRow = { ...rowValues, funcName: item.value };
    filterClause.filter = newRow;
    onUpdateFilter();
  };

  const handleParamChange = (e: any) => {
    const newRow = { ...rowValues };
    switch (newRow.category) {
      case "number":
        newRow.parameter = e.floatValue || 0;
        break;
      case "boolean":
        newRow.parameter = e.value;
        break;
      case "array":
        newRow.parameter = String(e.value);
        break;
      default:
        newRow.parameter = e.target.value ? e.target.value : "";
    }
    filterClause.filter = newRow;
    onUpdateFilter();
  };

  const label = filterOptions.find((item) => item.value === rowValues.key)?.label;
  const options = filterOptions.find((item) => item.value === rowValues.key)?.selectOptions;

  const getInputField = (handleParamChange: any) => {
    switch (rowValues.category) {
      case "number":
        return (
          <NumberFormat
            className={classNames(styles.filterInputField, styles.small)}
            thousandSeparator=","
            value={rowValues.parameter as keyof FilterParameterType}
            onValueChange={handleParamChange}
          />
        );
      case "boolean":
        return (
          <Select
            selectProps={{
              options: [
                { label: "True", value: true },
                { label: "False", value: false },
              ],
              value: { label: !rowValues.parameter ? "False" : "True", value: rowValues.parameter },
              onChange: handleParamChange,
            }}
            child={child}
          />
        );
      case "array": {
        const label = options?.find((item) => {
          return item.value === rowValues.parameter ? item : "";
        })?.label;
        return (
          <Select
            selectProps={{
              options: options,
              value: { value: rowValues.parameter, label: label },
              onChange: handleParamChange,
            }}
            child={child}
          />
        );
      }
      case "date":
        return (
          <input
            type="datetime-local"
            className={"torq-input-field"}
            value={rowValues.parameter as string}
            onChange={handleParamChange}
          />
        );
      default:
        return (
          <input
            type="text"
            className={"torq-input-field"}
            value={rowValues.parameter as keyof FilterParameterType}
            onChange={handleParamChange}
          />
        );
    }
  };

  const getParameter = () => {
    switch (rowValues.category) {
      case "number":
        return formatParameter(rowValues.parameter as number);
      case "duration":
        return formatParameter(rowValues.parameter as number);
      case "boolean":
        return !rowValues.parameter ? "False" : "True";
      case "array":
        return options?.find((item) => {
          return item.value === rowValues.parameter ? item : "";
        })?.label;
      default:
        return rowValues.parameter;
    }
  };

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
          <span className={styles.filterFunctionLabel}> {funcOption?.label} </span>
          <span className={styles.parameterLabel}>{getParameter()}</span>
        </div>
        <div className={classNames(styles.removeFilter, styles.desktopRemove)} onClick={() => onRemoveFilter(index)}>
          <RemoveIcon />
        </div>
      </div>

      <div className={classNames(styles.filterOptions, { [styles.expanded]: rowExpanded })}>
        <Select
          selectProps={{ options: filterOptions, value: selectData.key, onChange: handleKeyChange }}
          child={child}
        />

        <div className="filter-function-container">
          <Select
            selectProps={{ options: functionOptions, value: selectData.func, onChange: handleFunctionChange }}
            child={child}
          />
        </div>

        <div className={styles.filterParameterContainer}>{getInputField(handleParamChange)}</div>
      </div>
    </div>
  );
}

export default FilterRow;
