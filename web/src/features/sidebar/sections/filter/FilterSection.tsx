import styles from "./filter-section.module.scss";
import FilterComponent from "./FilterComponent";
import { ColumnMetaData } from "features/table/types";
import { updateFilters } from "features/viewManagement/viewSlice";
import { AllViewsResponse } from "features/viewManagement/types";

type FilterSectionProps<T> = {
  page: keyof AllViewsResponse;
  viewIndex: number;
  filters: any;
  filterableColumns: Array<ColumnMetaData<T>>;
  defaultFilter: any;
};

function FilterSection<T>(props: FilterSectionProps<T>) {
  const handleFilterUpdate = () => {
    updateFilters({
      page: props.page,
      viewIndex: props.viewIndex,
      filterUpdate: props.filters,
    });
    // props.view.updateFilters(props.filters);
  };

  return (
    <div className={styles.filterPopoverContent}>
      <FilterComponent
        filters={props.filters}
        columns={props.filterableColumns}
        defaultFilter={props.defaultFilter}
        child={false}
        onFilterUpdate={handleFilterUpdate}
      />
    </div>
  );
}

export default FilterSection;
