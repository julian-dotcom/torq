import DefaultButton from "../../../buttons/Button";
import classNames from "classnames";
import {
  Filter20Regular as FilterIcon,
  Dismiss20Regular as RemoveIcon,
  AddSquare20Regular as AddFilterIcon
} from "@fluentui/react-icons";
import React from "react";
import TorqSelect, { SelectOptionType } from "../../../inputs/Select";

import './filter_popover.scoped.scss';
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { selectAllColumns, selectFilters, updateFilters } from "../../tableSlice";
import { FilterFunctions, FilterInterface } from "./filter";
import NumberFormat from "react-number-format";
import Popover from "../../../popover/Popover";

const combinerOptions: SelectOptionType[] = [
  { value: "and", label: "And" }
  // { value: "or", label: "Or" },
];

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
  rowValues: FilterInterface;
  columnOptions: SelectOptionType[];
  handleUpdateFilter: Function;
  handleRemoveFilter: Function;
}

interface filterOptionsInterface {
  key: string;
  heading: string;
  valueType: string;
}

function FilterRow({ index, rowValues, columnOptions, handleUpdateFilter, handleRemoveFilter }: filterRowInterface) {

  let functionOptions = getFilterFunctions(rowValues.category);
  console.log(functionOptions)

  const combinerOption = combinerOptions.find(
    item => item.value === rowValues.combiner
  )
  if (!combinerOption) {
    throw new Error("combiner not found")
  }
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

  let rowData = {
    combiner: combinerOption,
    category: rowValues.category,
    func: funcOption,
    key: keyOption,
    param: rowValues.parameter
  };

  const convertFilterData = (rowData: any): FilterInterface => {
    return {
      combiner: rowData.combiner.value,
      category: rowData.category,
      funcName: rowData.func.value,
      key: rowData.key.value,
      parameter: rowData.param
    };
  };

  const handleCombinerChange = (item: any) => {
    handleUpdateFilter(
      {
        ...convertFilterData(rowData),
        combiner: item.value
      },
      index
    );
  };

  const handleKeyChange = (item: any) => {
    let newFilter = {
      ...convertFilterData(rowData),
      key: item.value,
      category: item.valueType
    };
    if (item.valueType !== rowData.category) {
      if (item.valueType === "string") {
        newFilter.funcName = "include";
        newFilter.parameter = "";
      } else {
        newFilter.funcName = "gte";
        newFilter.parameter = 0;
      }
    }

    handleUpdateFilter(newFilter, index);
  };

  const handleFunctionChange = (item: any) => {
    handleUpdateFilter(
      {
        ...convertFilterData(rowData),
        funcName: item.value
      },
      index
    );
  };

  const handleParamChange = (e: any) => {
    if (rowData.category === "number") {
      handleUpdateFilter(
        {
          ...convertFilterData(rowData),
          parameter: e.floatValue
        },
        index
      );
      return;
    }

    handleUpdateFilter(
      {
        ...convertFilterData(rowData),
        parameter: e.target.value ? e.target.value : ""
      },
      index
    );
  };

  return (
    <div className={classNames("filter-row", { first: !index })}>
      {!index && <div className="combiner-container">
        <div>Where</div>
        {/*Clean this up using grid named cells*/}
        <div className="remove-filter mobile-remove" onClick={() => handleRemoveFilter(index)}>
          <RemoveIcon />
        </div>
      </div>}
      {!!index && (
        <div className="combiner-container">
          <TorqSelect
            options={combinerOptions}
            value={rowData.combiner}
            onChange={handleCombinerChange}
          />
          <div className="remove-filter mobile-remove" onClick={() => handleRemoveFilter(index)}>
            <RemoveIcon />
          </div>
        </div>
      )}
      <div className="filter-key-container">
        <TorqSelect
          options={columnOptions}
          value={rowData.key}
          onChange={handleKeyChange}
        />
      </div>
      <div className="filter-function-container">
        <TorqSelect
          options={functionOptions}
          value={rowData.func}
          onChange={handleFunctionChange}
        />
      </div>
      <div className="filter-parameter-container">
        {rowData.category === "number" ? (
          <NumberFormat
            className={"torq-input-field"}
            thousandSeparator=","
            value={rowData.param}
            onValueChange={handleParamChange}
          />
        ) : (
          <input
            type="text"
            className={"torq-input-field"}
            value={rowData.param}
            onChange={handleParamChange}
          />
        )}
      </div>
      <div className="remove-filter desktop-remove" onClick={() => handleRemoveFilter(index)}>
        <RemoveIcon />
      </div>
    </div>
  );
}

const FilterPopover = () => {
  const filters = useAppSelector(selectFilters);
  const dispatch = useAppDispatch();

  const updateFilter = (filter: FilterInterface, index: number) => {
    const updatedFilters: FilterInterface[] = [
      ...filters.slice(0, index),
      filter,
      ...filters.slice(index + 1, filters.length)
    ];
    dispatch(updateFilters({ filters: updatedFilters }));
  };

  const removeFilter = (index: number) => {
    const updatedFilters: FilterInterface[] = [
      ...filters.slice(0, index),
      ...filters.slice(index + 1, filters.length)
    ];
    dispatch(updateFilters({ filters: updatedFilters }));
  };

  const addFilter = () => {
    const updatedFilters: FilterInterface[] = [
      ...filters.slice(),
      {
        combiner: "and",
        funcName: "gte",
        category: "number",
        key: "capacity",
        parameter: 0
      }
    ];
    dispatch(updateFilters({ filters: updatedFilters }));
  };

  const buttonText = (): string => {
    if (filters.length > 0) {
      return filters.length + " Filter" + (filters.length > 1 ? "s" : "")
    }
    return "Filter";
  };

  let popOverButton = (
    <DefaultButton
      text={buttonText()}
      icon={<FilterIcon />}
      className={"collapse-tablet"}
      isOpen={!!filters.length}
    />
  );

  const columnsMeta = useAppSelector(selectAllColumns) || [];

  let columnOptions = columnsMeta
    .slice()
    .map((column: filterOptionsInterface) => {
      return {
        value: column.key,
        label: column.heading,
        valueType: column.valueType
      };
    });

  columnOptions.sort((a: SelectOptionType, b: SelectOptionType) => {
    if (a.label < b.label) { return -1; }
    if (a.label > b.label) { return 1; }
    return 0;
  });

  return (
    <Popover button={popOverButton}>
      <div className={"filter-popover-content"}>
        <div className="filter-rows">
          {!filters.length && <div className={"no-filters"}>No filters</div>}
          {filters.map((filter, index) => {
            return (
              <FilterRow
                key={"filter-row-" + index}
                rowValues={filter}
                index={index}
                columnOptions={columnOptions}
                handleUpdateFilter={updateFilter}
                handleRemoveFilter={removeFilter}
              />
            );
          })}
        </div>
        <div className="buttons-row">
          <DefaultButton
            text={"Add filter"}
            icon={<AddFilterIcon />}
            onClick={addFilter}
          />
        </div>
      </div>
    </Popover>
  );
};

export default FilterPopover;
