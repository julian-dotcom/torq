import DefaultButton from "../../../buttons/Button";
import {
  Filter20Regular as FilterIcon,
} from "@fluentui/react-icons";
import React from "react";
import clone from "clone"

import styles from './filter_popover.module.scss';
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { selectFilters, updateFilters } from "../../tableSlice";
import { deserialiseQuery, Clause } from "./filter";
import Popover from "../../../popover/Popover";
import FilterComponent from "./FilterComponent"

const FilterPopover = () => {
  const dispatch = useAppDispatch();
  const filtersFromStore = clone<Clause>(useAppSelector(selectFilters));
  const filters = filtersFromStore ? deserialiseQuery(filtersFromStore) : undefined;

  const handleFilterUpdate = () => {
    if (filters) {
      dispatch(updateFilters({ filters: filters.toJSON() }));
    }
  }

  const buttonText = (): string => {
    if ((filters?.length || 0) > 0) {
      return filters?.length + " Filter" + ((filters?.length || 0) > 1 ? "s" : "")
    }
    return "Filter";
  };

  let popOverButton = (
    <DefaultButton
      text={buttonText()}
      icon={<FilterIcon />}
      className={"collapse-tablet"}
      isOpen={!!(filters?.length || 0)}
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
