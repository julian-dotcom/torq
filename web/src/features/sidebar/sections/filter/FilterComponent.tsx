import classNames from "classnames";
import { AddSquare20Regular as AddFilterIcon, AddSquareMultiple20Regular as AddGroupIcon } from "@fluentui/react-icons";
import FilterRow from "./FilterRow";
import { AndClause, OrClause, FilterClause, FilterInterface, Clause } from "./filter";
import { ColumnMetaData } from "features/table/types";
import styles from "./filter-section.module.scss";
import clone from "clone";

export interface FilterComponentProps<T> {
  columns: Array<ColumnMetaData<T>>;
  filters: AndClause | OrClause; //Base clause for this component needs to be And or Or clause
  defaultFilter: FilterInterface;
  onFilterUpdate: (filter: AndClause | OrClause) => void;
  onNoChildrenLeft?: () => void;
  child: boolean;
}

const combinerOptions = new Map<string, string>([
  ["$and", "And"],
  ["$or", "Or"],
]);

function FilterComponent<T>(props: FilterComponentProps<T>) {
  const generateUpdateFilterFunc = (index: number) => {
    const handleUpdateFilter = (updatedFilters: Clause) => {
      const filters = clone(props.filters);
      filters.childClauses[index] = updatedFilters;
      props.onFilterUpdate(filters);
    };
    return handleUpdateFilter;
  };

  const removeFilter = (index: number) => {
    // if it's the last child then let my parent remove me
    if (props.filters.childClauses.length === 1 && props.onNoChildrenLeft) {
      props.onNoChildrenLeft();
      return;
    }

    const filters = clone(props.filters);
    filters.childClauses.splice(index, 1);
    if (filters.childClauses.length === 1) {
      filters.prefix = "$and";
    }

    props.onFilterUpdate(filters);
  };

  const handleNoChildrenLeft = (index: number) => {
    return () => {
      console.log("changed it");
      removeFilter(index);
    };
  };

  const addFilter = () => {
    const filters = clone(props.filters);
    filters.addChildClause(new FilterClause(props.defaultFilter));
    props.onFilterUpdate(filters);
  };

  const addGroup = () => {
    const filters = clone(props.filters);
    filters.addChildClause(new AndClause([new FilterClause(props.defaultFilter)]));
    props.onFilterUpdate(filters);
  };

  const handleCombinerChange = () => {
    // this effectively changes the type of the object from AndClause to OrClause or vice versa
    const filters = clone(props.filters);
    filters.prefix = filters.prefix === "$and" ? "$or" : "$and";
    props.onFilterUpdate(filters);
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const filterOptions = props.columns.slice().map((column: any) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const columnOption = {
      value: column.key as string,
      label: column.heading as string,
      valueType: column.valueType as "string" | "number" | "boolean" | "date" | "array",
      selectOptions: column.selectOptions,
      arrayOptions: undefined as undefined | { value: boolean; label: string }[],
    };
    if (column.valueType === "array") {
      columnOption.arrayOptions = [
        { value: true, label: "True" },
        { value: false, label: "False" },
      ];
    }
    return columnOption;
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  filterOptions.sort((a: any, b: any) => {
    if (a.label < b.label) {
      return -1;
    }
    if (a.label > b.label) {
      return 1;
    }
    return 0;
  });

  return (
    <div className={classNames({ [styles.childFilter]: props.child })}>
      <div className={styles.filterRows}>
        {!props.filters && <div className={styles.noFilters}>No filters</div>}
        {((props.filters as AndClause | OrClause).childClauses || []).map((filter, index) => {
          if (filter.prefix === "$filter") {
            return (
              <div
                key={"filter-row-wrapper-" + index}
                className={classNames(styles.filterRow, { [styles.first]: !index })}
              >
                <FilterRow
                  child={props.child}
                  key={"filter-row-" + index}
                  filterClause={filter as FilterClause}
                  index={index}
                  filterOptions={filterOptions}
                  onUpdateFilter={generateUpdateFilterFunc(index)}
                  onRemoveFilter={removeFilter}
                  combiner={combinerOptions.get(props.filters.prefix)}
                  handleCombinerChange={handleCombinerChange}
                />
              </div>
            );
          } else {
            return (
              <div
                key={"filter-sub-group-" + index}
                className={classNames(styles.filterRow, { first: !index, [styles.childWrapper]: true })}
              >
                <div className={styles.filterKeyLabel} onClick={() => handleCombinerChange()}>
                  {combinerOptions.get(props.filters.prefix)}
                </div>
                <FilterComponent
                  child={true}
                  defaultFilter={props.defaultFilter}
                  columns={props.columns}
                  filters={filter as OrClause | AndClause}
                  onNoChildrenLeft={handleNoChildrenLeft(index)}
                  onFilterUpdate={generateUpdateFilterFunc(index)}
                />
              </div>
            );
          }
        })}
      </div>

      <div className={styles.buttonsRow}>
        <div className={styles.addFilterButton} onClick={addFilter}>
          <AddFilterIcon />
          <span className={styles.buttonText}>{"Add filter"}</span>
        </div>
        {!props.child && (
          <div className={styles.addFilterButton} onClick={addGroup}>
            <AddGroupIcon />
            <span className={styles.buttonText}>{"Add group"}</span>
          </div>
        )}
      </div>
    </div>
  );
}

export default FilterComponent;
