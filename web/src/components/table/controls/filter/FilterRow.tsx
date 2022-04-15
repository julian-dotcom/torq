import classNames from "classnames";
import {
  Dismiss20Regular as RemoveIcon,
} from "@fluentui/react-icons";
import TorqSelect, { SelectOptionType } from "../../../inputs/Select";

import { FilterClause } from "./filter";
import styles from './filter_popover.module.scss';
import { FilterFunctions } from "./filter";
import NumberFormat from "react-number-format";

const ffLabels = new Map<string, string>([
  ["eq", "="],
  ["neq", "≠"],
  ["gt", ">"],
  ["gte", ">="],
  ["lt", "<"],
  ["lte", "<="],
  ["include", "Include"],
  ["notInclude", "Not include"]
]);

function getFilterFunctions(filterCategory: "number" | "string") {
  const filterFuncs = FilterFunctions.get(filterCategory)?.entries()
  if (!filterFuncs) {
    throw new Error("Filter category not found in list of filters")
  }
  return Array.from(filterFuncs, ([key, _]) => (
    { value: key, label: ffLabels.get(key) ?? "Label not found" }))
}

interface filterRowInterface {
  index: number;
  filterClause: FilterClause;
  columnOptions: SelectOptionType[];
  onUpdateFilter: Function;
  onRemoveFilter: Function;
}

function FilterRow({ index, filterClause, columnOptions, onUpdateFilter, onRemoveFilter }: filterRowInterface) {

  const rowValues = filterClause.filter

  let functionOptions = getFilterFunctions(rowValues.category);

  const keyOption = columnOptions.find(
    item => item.value === rowValues.key
  )
  if (!keyOption) {
    throw new Error("key option not found")
  }
  const funcOption = functionOptions.find(
    item => item.value === rowValues.funcName
  )
  if (!funcOption) {
    throw new Error("func option not found")
  }

  let selectData = {
    func: funcOption,
    key: keyOption,
  };

  const handleKeyChange = (item: any) => {
    rowValues.key = item.value
    onUpdateFilter()
  };

  const handleFunctionChange = (item: any) => {
    rowValues.funcName = item.value
    onUpdateFilter()
  };

  const handleParamChange = (e: any) => {
    if (rowValues.category === "number") {
      rowValues.parameter = e.floatValue
    } else {
      rowValues.parameter = e.target.value ? e.target.value : ""
    }
    onUpdateFilter()
  };

  return (
    <div className={classNames(styles.filter, { first: !index })}>
      <div className="filter-key-container">
        <TorqSelect
          options={columnOptions}
          value={selectData.key}
          onChange={handleKeyChange}
        />
      </div>
      <div className="filter-function-container">
        <TorqSelect
          options={functionOptions}
          value={selectData.func}
          onChange={handleFunctionChange}
        />
      </div>
      <div className="filter-parameter-container">
        {rowValues.category === "number" ? (
          <NumberFormat
            className={"torq-input-field"}
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
      <div className={classNames(styles.removeFilter, styles.desktopRemove)} onClick={() => onRemoveFilter(index)}>
        <RemoveIcon />
      </div>
    </div >
  );
}

export default FilterRow;
