import classNames from "classnames";
import { AddSquare20Regular as AddFilterIcon, AddSquareMultiple20Regular as AddGroupIcon } from "@fluentui/react-icons";
import FilterRow from "./FilterRow";
import { AndClause, OrClause, FilterClause, FilterInterface } from "./filter";
import { ColumnMetaData } from "features/table/types";
import styles from "./filter-section.module.scss";
import useTranslations from "services/i18n/useTranslations";

export interface FilterComponentProps<T> {
  columns: Array<ColumnMetaData<T>>;
  filters: AndClause | OrClause | FilterClause;
  defaultFilter: FilterInterface;
  onFilterUpdate: () => void;
  onNoChildrenLeft?: () => void;
  child: boolean;
}

const combinerOptions = new Map<string, string>([
  ["$and", "And"],
  ["$or", "Or"],
]);

function FilterComponent<T>(props: FilterComponentProps<T>) {
  const { t } = useTranslations();
  const updateFilter = () => {
    props.onFilterUpdate();
  };

  const removeFilter = (index: number) => {
    const filters = props.filters as AndClause | OrClause;
    filters.childClauses.splice(index, 1);
    if (filters.childClauses.length === 1) {
      filters.prefix = "$and";
    }
    if (filters.childClauses.length === 0) {
      if (props.onNoChildrenLeft) {
        props.onNoChildrenLeft();
      }
    }
    props.onFilterUpdate();
  };

  const handleNoChildrenLeft = (index: number) => {
    return () => {
      removeFilter(index);
    };
  };

  const addFilter = () => {
    if (!Object.hasOwn(props.filters, "childClauses")) {
      throw new Error("Cannot add filter to a filter");
    } else if (Object.hasOwn(props.filters, "childClauses")) {
      console.log("adding filter", props.filters, Object.hasOwn(props.filters, "childClauses"));
      ((props.filters as AndClause) || OrClause).addChildClause(new FilterClause(props.defaultFilter));
      props.onFilterUpdate();
    }
  };

  const addGroup = () => {
    if (!props.filters) {
      props.filters = new AndClause();
    }
    if (props.filters instanceof AndClause || OrClause) {
      ((props.filters as AndClause) || OrClause).addChildClause(new AndClause([new FilterClause(props.defaultFilter)]));
      props.onFilterUpdate();
    }
  };

  const handleCombinerChange = () => {
    if (props.filters?.prefix) {
      // this effectively changes the type of the object from AndClause to OrClause or vice versa
      props.filters.prefix = props.filters.prefix === "$and" ? "$or" : "$and";
      props.onFilterUpdate();
    }
  };

  const filterOptions = props.columns.slice().map((column: any) => {
    const columnOption: any = {
      value: column.key,
      label: column.heading,
      valueType: column.valueType as "string" | "number" | "boolean" | "date" | "array",
      selectOptions: column.selectOptions,
    };
    if (column.valueType === "array") {
      columnOption.arrayOptions = [
        { value: true, label: "True" },
        { value: false, label: "False" },
      ];
    }
    return columnOption;
  });

  filterOptions.sort((a: any, b: any) => {
    if (a.label < b.label) {
      return -1;
    }
    if (a.label > b.label) {
      return 1;
    }
    return 0;
  });

  // const renderNoFiltersMessage = () => {
  //   return (
  //     <div className={classNames({ [styles.childFilter]: props.child })}>
  //       <div className={styles.filterRows}>
  //         <div className={styles.noFilters}>No filters</div>
  //       </div>
  //     </div>
  //   );
  // };
  //
  // const returned: Array<JSX.Element> = [];
  //
  // const renderFilterClause = (index: number, filter: FilterClause) => {
  //   return (
  //     <div key={"filter-row-wrapper-" + index} className={classNames(styles.filterRow, { [styles.first]: !index })}>
  //       <FilterRow
  //         child={props.child}
  //         key={"filter-row-" + index}
  //         filterClause={filter}
  //         index={index}
  //         filterOptions={filterOptions}
  //         onUpdateFilter={updateFilter}
  //         onRemoveFilter={removeFilter}
  //         combiner={combinerOptions.get(props.filters?.prefix || "$and")}
  //         handleCombinerChange={handleCombinerChange}
  //       />
  //     </div>
  //   );
  // };

  // const renderAndOrClause = (index: number, filter: Clause) => {
  //   return (
  //     <div
  //       key={"filter-sub-group-" + index}
  //       className={classNames(styles.filterRow, { first: !index, [styles.childWrapper]: true })}
  //     >
  //       <div className={styles.filterKeyLabel} onClick={() => handleCombinerChange()}>
  //         {combinerOptions.get(props.filters.prefix || "$and")}
  //       </div>
  //       <FilterComponent
  //         child={true}
  //         defaultFilter={props.defaultFilter}
  //         columns={props.columns}
  //         filters={((props.filters as AndClause) || OrClause).childClauses[index]}
  //         onNoChildrenLeft={handleNoChildrenLeft(index)}
  //         onFilterUpdate={props.onFilterUpdate}
  //       />
  //     </div>
  //   );
  // };

  // if (!props.filters?.length) {
  //   return (
  //     <div className={classNames({ [styles.childFilter]: props.child })}>
  //       {renderNoFiltersMessage()}
  //       <div className={styles.buttonsRow}>
  //         <div className={styles.addFilterButton} onClick={addFilter}>
  //           <AddFilterIcon />
  //           <span className={styles.buttonText}>{t.addFilter}</span>
  //         </div>
  //         <div className={styles.addFilterButton} onClick={addGroup}>
  //           <AddGroupIcon />
  //           <span className={styles.buttonText}>{t.addFilterGroup}</span>
  //         </div>
  //       </div>
  //     </div>
  //   );
  // }

  // if (props.filters.prefix === "$filter") {
  //   return renderFilterClause(0, props.filters as FilterClause);
  // }
  //
  // if (props.filters?.length && (props.filters.prefix === "$and" || props.filters.prefix === "$or")) {
  //   return (
  //     <div className={classNames({ [styles.childFilter]: props.child })}>
  //       <div className={styles.buttonsRow}>
  //         <div className={styles.addFilterButton} onClick={addFilter}>
  //           <AddFilterIcon />
  //           <span className={styles.buttonText}>{t.addFilter}</span>
  //         </div>
  //         {!props.child && (
  //           <div className={styles.addFilterButton} onClick={addGroup}>
  //             <AddGroupIcon />
  //             <span className={styles.buttonText}>{t.addFilterGroup}</span>
  //           </div>
  //         )}
  //       </div>
  //     </div>
  //   );
  // }
  //
  // if (Object.hasOwn(props.filters, "childClauses")) {
  //   return ((props.filters as AndClause) || OrClause).childClauses.map((filter, index) => {
  //     return (
  //       <div className={classNames({ [styles.childFilter]: props.child })} key={"filter-group-index" + index}>
  //         {renderAndOrClause(index, filter as AndClause | OrClause)}
  //         <div className={styles.buttonsRow}>
  //           <div className={styles.addFilterButton} onClick={addFilter}>
  //             <AddFilterIcon />
  //             <span className={styles.buttonText}>{t.addFilter}</span>
  //           </div>
  //           {!props.child && (
  //             <div className={styles.addFilterButton} onClick={addGroup}>
  //               <AddGroupIcon />
  //               <span className={styles.buttonText}>{t.addFilterGroup}</span>
  //             </div>
  //           )}
  //         </div>
  //       </div>
  //     );
  //   });
  // } else if (props.filters.prefix === "$filter") {
  //   returned.push(renderFilterClause(0, props.filters as FilterClause));
  // }
  //
  // return <div>{returned}</div>;
  return (
    <div className={classNames({ [styles.childFilter]: props.child })}>
      <div className={styles.filterRows}>
        {!props.filters.length && <div className={styles.noFilters}>No filters</div>}

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
                  onUpdateFilter={updateFilter}
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
                  filters={filter}
                  onNoChildrenLeft={handleNoChildrenLeft(index)}
                  onFilterUpdate={props.onFilterUpdate}
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
