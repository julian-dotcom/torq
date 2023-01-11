import styles from "./filter-section.module.scss";
import FilterComponent from "./FilterComponent";
import { ColumnMetaData } from "features/table/types";
import { updateFilters } from "features/viewManagement/viewSlice";
import { AllViewsResponse } from "features/viewManagement/types";
import { AndClause, OrClause, FilterInterface } from "./filter";
import { useAppDispatch } from "store/hooks";

type FilterSectionProps<T> = {
  page: keyof AllViewsResponse;
  viewIndex: number;
  filters: OrClause | AndClause;
  filterableColumns: Array<ColumnMetaData<T>>;
  defaultFilter: FilterInterface;
};

function FilterSection<T>(props: FilterSectionProps<T>) {
  const dispatch = useAppDispatch();
  const handleFilterUpdate = (filters: AndClause | OrClause) => {
    dispatch(
      updateFilters({
        page: props.page,
        viewIndex: props.viewIndex,
        filterUpdate: filters.toJSON(),
      })
    );
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
