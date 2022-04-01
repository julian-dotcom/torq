import DefaultButton from "../../../buttons/Button";
import classNames from "classnames";
import {
  Filter20Regular as FilterIcon,
  Dismiss20Regular as RemoveIcon,
  AddSquare20Regular as AddFilterIcon
} from "@fluentui/react-icons";
import React from "react";
import TorqSelect, {SelectOptionType} from "../../../inputs/Select";

import './filter_popover.scoped.scss';
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { selectAllColumns, selectFilters, updateFilters } from "../../tableSlice";
import { FilterFunctions, FilterInterface } from "../../filter";
import NumberFormat from "react-number-format";
import Popover from "../../../popover/Popover";

const combinerOptions = [
  { value: "and", label: "And" }
  // { value: "or", label: "Or" },
];

const ffLabels = {
  eq: "=",
  neq: "≠",
  gt: ">",
  gte: ">=",
  lt: "<",
  lte: "<=",
  include: "Include",
  notInclude: "Not include"
};

function getFilterFunctions(filterCategory: "number" | "string") {
  // @ts-ignore
  return Object.keys(FilterFunctions[filterCategory]).map((key: []) => {
    // @ts-ignore
    return { value: key, label: ffLabels[key] };
  });
}

interface filterRowInterface {
  index: number;
  rowValues: FilterInterface;
  handleUpdateFilter: Function;
  handleRemoveFilter: Function;
}


function FilterRow({index, rowValues, handleUpdateFilter, handleRemoveFilter}: filterRowInterface) {

  let columnsMeta = useAppSelector(selectAllColumns) || [];

  let columnOptions = columnsMeta
    .slice()
    .map((column: { key: string; heading: string; valueType: string }) => {
      return {
        value: column.key,
        label: column.heading,
        valueType: column.valueType
      };
    });

  columnOptions.sort((a: SelectOptionType, b: SelectOptionType) => {
    if(a.label < b.label) { return -1; }
    if(a.label > b.label) { return 1; }
    return 0;
  });

  let functionOptions = getFilterFunctions(rowValues.category);

  // @ts-ignore
  let combinerOption: SelectOptionType = combinerOptions.find(
    (item: SelectOptionType) => item.value === rowValues.combiner
  )
  // @ts-ignore
  let keyOption: SelectOptionType = columnOptions.find(
    (item: SelectOptionType) => item.value === rowValues.key
  )
  // @ts-ignore
  let funcOption: SelectOptionType = functionOptions.find(
    // @ts-ignore
    (item: SelectOptionType) => item.value === rowValues.funcName
  )

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
      {!index && <div className="filter-combiner-container">Where</div>}
      {!!index && (
        <div className="combiner-container">
          <TorqSelect
            options={combinerOptions}
            value={rowData.combiner}
            onChange={handleCombinerChange}
          />
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
      <div className="remove-filter" onClick={() => handleRemoveFilter(index)}>
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
      return filters.length + " filters";
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
