import DefaultButton from "../../buttons/Button";
import classNames from "classnames";
import {
  Filter20Regular as FilterIcon,
  Dismiss20Regular as RemoveIcon,
  AddSquare20Regular as AddFilterIcon,
} from "@fluentui/react-icons";
import React, {SetStateAction, useEffect, useRef, useState} from "react";
import TorqSelect from "../../inputs/Select";

import './filter_popover.scoped.scss';
import {useAppDispatch, useAppSelector} from "../../../store/hooks";
import {selectColumns, selectFilters, updateFilters} from "../tableSlice";
import {FilterFunctions, FilterInterface} from "../filter";
import NumberFormat from "react-number-format";
import {log} from "util";

const combinerOptions = [
  { value: "and", label: "And" },
  // { value: "or", label: "Or" },
];

const ffLabels = {
  eq: '=',
  neq: '≠',
  gt: '>',
  gte: '>=',
  lt: '<',
  lte: '<=',
  includes: 'Include',
  notInclude: 'Not include',
}

function getFilterFunctions(filterCategory: 'number' | 'string') {
  // @ts-ignore
  return Object.keys(FilterFunctions[filterCategory]).map((key: []) => {
    // @ts-ignore
    return {value: key, label: ffLabels[key]}
  })
}

type optionType = {value: string, label:string}
interface filterRowInterface {
  index: number,
  rowValues: FilterInterface,
  handleUpdateFilter: Function,
  handleRemoveFilter: Function
}

function FilterRow({index, rowValues, handleUpdateFilter, handleRemoveFilter}: filterRowInterface) {

  let columnsMeta = useAppSelector(selectColumns) || [];

  let columnOptions = columnsMeta.slice().map((column: {key: string, heading: string}) => {
    return {value: column.key, label: column.heading}
  })

  columnOptions.sort((a: optionType, b: optionType) => {
    if(a.label < b.label) { return -1; }
    if(a.label > b.label) { return 1; }
    return 0;
  })

  let functionOptions = getFilterFunctions(rowValues.category)

  // @ts-ignore
  let combinerOption: optionType = combinerOptions.find((item: optionType) => item.value == rowValues.combiner)
  // @ts-ignore
  let keyOption: optionType = columnOptions.find((item: optionType) => item.value == rowValues.key)
  // @ts-ignore
  let funcOption: optionType = functionOptions.find((item: optionType) => item.value == rowValues.funcName)

  let rowData = {
    combiner: combinerOption,
    category: rowValues.category,
    func: funcOption,
    key: keyOption,
    param: rowValues.parameter,
  }

  const convertFilterData = (rowData: any): FilterInterface => {
    return {
      combiner: rowData.combiner.value,
      category: rowData.category,
      funcName: rowData.func.value,
      key: rowData.key.value,
      parameter: rowData.param,
    }
  }

  const handleCombinerChange = (item:any) => {
    handleUpdateFilter({
      ...convertFilterData(rowData),
      combiner: item.value
    }, index)
  }
  const handleKeyChange = (item:any) => {
    // TODO: Look up column and add column category (number or string)
    handleUpdateFilter({
      ...convertFilterData(rowData),
      key: item.value
    }, index)
  }
  const handleFunctionChange = (item:any) => {
    handleUpdateFilter({
      ...convertFilterData(rowData),
      funcName: item.value
    }, index)
  }
  const handleParamChange = (value: any) => {
    if (value.floatValue) {
      handleUpdateFilter({
        ...convertFilterData(rowData),
        parameter: value.floatValue
      }, index)
    }
  }

  return (
    <div className={classNames("filter-row", {first: !index})}>
      {(!index && (<div className="filter-combiner-container">Where</div>))}
      {(!!index && (
        <div className="combiner-container">
          <TorqSelect options={combinerOptions} value={rowData.combiner} onChange={handleCombinerChange}/>
        </div>)
      )}
      <div className="filter-key-container">
        <TorqSelect options={columnOptions} value={rowData.key} onChange={handleKeyChange}/>
      </div>
      <div className="filter-function-container">
        <TorqSelect options={functionOptions} value={rowData.func} onChange={handleFunctionChange} />
      </div>
      <div className="filter-parameter-container">
        <NumberFormat
          className={"torq-input-field"}
          thousandSeparator=',' value={rowData.param}
          onValueChange={handleParamChange}
        />
      </div>
      <div className="remove-filter" onClick={() => (handleRemoveFilter(index))}>
        <RemoveIcon/>
      </div>
    </div>
  )
}

function useOutsideClose(ref: any, setIsPopoverOpen: Function) {
  useEffect(() => {
    function handleClickOutside(event: any) {
      if (ref.current && !ref.current.contains(event.target)) {
        setIsPopoverOpen(false)
      }
    }
    // Bind the event listener
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      // Unbind the event listener on clean up
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [ref]);
}

const FilterPopover = () => {
  const wrapperRef = useRef(null)

  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  useOutsideClose(wrapperRef, setIsPopoverOpen)

  const filters = useAppSelector(selectFilters)
  const dispatch = useAppDispatch();

  const updateFilter = (filter: FilterInterface, index: number) => {
     const updatedFilters: FilterInterface[] = [
       ...filters.slice(0,index),
       filter,
        ...filters.slice(index+1, filters.length)
     ]
    dispatch(
      updateFilters( {filters: updatedFilters})
    )
  }

  const removeFilter = (index: number) => {
     const updatedFilters: FilterInterface[] = [
       ...filters.slice(0,index),
        ...filters.slice(index+1, filters.length)
     ]
    dispatch(
      updateFilters( {filters: updatedFilters})
    )
  }

  const addFilter = () => {
    const updatedFilters: FilterInterface[] = [
       ...filters.slice(),
       {combiner: 'and',
          funcName: 'gte',
          category: 'number',
          key: "capacity",
          parameter: 0
        },
     ]
    dispatch(
      updateFilters( {filters: updatedFilters})
    )
  }

  const buttonText = (): string => {
    if (filters.length > 0) {
      return filters.length + " filters"
    }
    return "Filter"
  }

  return (
    <div onClick={() => setIsPopoverOpen(!isPopoverOpen)}
         ref={wrapperRef}
         className={classNames("torq-popover-button-wrapper")} >
      <DefaultButton text={buttonText()} icon={<FilterIcon/>} className={"collapse-tablet"} isOpen={!!filters.length}/>
      <div className={classNames("popover-wrapper", {"popover-open": isPopoverOpen})}
           onClick={(e) =>{
             e.stopPropagation()
           }}>
        <div className="filter-rows">
          {!filters.length && (<div className={"no-filters"}>No filters</div>)}
          {filters.map((filter, index) => {
            return (<FilterRow
              key={'filter-row-'+index}
              rowValues={filter}
              index={index}
              handleUpdateFilter={updateFilter}
              handleRemoveFilter={removeFilter}
            />)
          })}
        </div>
        <div className="buttons-row">
          <DefaultButton text={"Add filter"} icon={<AddFilterIcon/>} onClick={addFilter} />
        </div>
      </div>
    </div>
  )
}

export default FilterPopover;
