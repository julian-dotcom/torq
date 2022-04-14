import DefaultButton from "../../../buttons/Button";
import classNames from "classnames";
import {
  Filter20Regular as FilterIcon,
  Dismiss20Regular as RemoveIcon,
  AddSquare20Regular as AddFilterIcon
} from "@fluentui/react-icons";
import React from "react";
import TorqSelect, { SelectOptionType } from "../../../inputs/Select";

import styles from './filter_popover.module.scss';
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { selectAllColumns, selectFilters, updateFilters } from "../../tableSlice";
import { FilterFunctions, FilterInterface, deserialiseQueryJSON } from "./filter";
import NumberFormat from "react-number-format";
import Popover from "../../../popover/Popover";
import FilterComponent from "./FilterComponent"

import { Clause } from './filter'

const FilterPopover = () => {
  const dispatch = useAppDispatch();
  const filtersString = useAppSelector(selectFilters);
  const filters = filtersString ? deserialiseQueryJSON(filtersString) : undefined;

  function handleClick() {
  }
  const handleFilterUpdate = () => {
    if (filters) {
      /* console.log(filters) */
      dispatch(updateFilters({ filters: JSON.stringify(filters) }));
    }
  }

  const buttonText = (): string => {
    /* if (filters.length > 0) {
*   return filters.length + " Filter" + (filters.length > 1 ? "s" : "")
* } */
    return "Filter";
  };

  let popOverButton = (
    <DefaultButton
      text={buttonText()}
      icon={<FilterIcon />}
      className={"collapse-tablet"}
      isOpen={false}
    />
  );

  return (
    <Popover button={popOverButton}>
      <div className={styles.filterPopoverContent}>
        {filters &&
          <FilterComponent filters={filters} onFilterUpdate={handleFilterUpdate} />
        }
      </div>
    </Popover>
  );
};

export default FilterPopover;
