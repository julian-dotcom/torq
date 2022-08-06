import React from "react";
import clone from "clone";
import styles from "./filter-section.module.scss";
import { deserialiseQuery, Clause } from "./filter";
import FilterComponent from "./FilterComponent";

type FilterSectionProps = {
  filters: Clause;
  filterUpdateHandler: Function;
};

const FilterSection = (props: FilterSectionProps) => {
  const filtersFromStore = clone<Clause>(props.filters);
  const filters = filtersFromStore ? deserialiseQuery(filtersFromStore) : undefined;

  const handleFilterUpdate = () => {
    props.filterUpdateHandler(filters);
  };

  return (
    <div className={styles.filterPopoverContent}>
      {filters && <FilterComponent filters={filters} onFilterUpdate={handleFilterUpdate} child={false} />}
    </div>
  );
};

export default FilterSection;
