import styles from "./filter-section.module.scss";
import { FilterInterface } from "./filter";
import FilterComponent from "./FilterComponent";
import { ColumnMetaData } from "features/table/types";
import View from "features/viewManagement/View";

type FilterSectionProps<T> = {
  columns: Array<ColumnMetaData<T>>;
  view: View<T>;
  defaultFilter: FilterInterface;
};

function FilterSection<T>(props: FilterSectionProps<T>) {
  const filters = props.view.filters;

  const handleFilterUpdate = () => {
    props.view.updateFilters(filters);
  };

  return (
    <div className={styles.filterPopoverContent}>
      <FilterComponent<T>
        filters={filters}
        columns={props.columns}
        defaultFilter={props.defaultFilter}
        child={false}
        onFilterUpdate={handleFilterUpdate}
      />
    </div>
  );
}

export default FilterSection;
