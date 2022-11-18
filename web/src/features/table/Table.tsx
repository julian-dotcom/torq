import styles from "./table.module.scss";
import cellStyles from "components/table/cells/cell.module.scss";
import HeaderCell from "components/table/cells/header/HeaderCell";
import classNames from "classnames";
import { SortByOptionType } from "features/sidebar/sections/sort/SortSectionOld";
import TableRow from "./TableRow";

export interface ColumnMetaData {
  heading: string;
  key: string;
  type?: string;
  width?: number;
  locked?: boolean;
  valueType: string;
  total?: number;
  max?: number;
  percent?: boolean;
}

export interface ViewInterface {
  title: string;
  id?: number;
  saved: boolean;
  filters?: any;
  columns: ColumnMetaData[];
  sortBy: SortByOptionType[];
  groupBy?: string;
  page: string;
}

export interface viewOrderInterface {
  id: number | undefined;
  view_order: number;
}

export type TableProps = {
  activeColumns: Array<ColumnMetaData>;
  data: Array<any>;
  isLoading: boolean;
  showTotals?: boolean;
  rowRenderer?: (row: any, index: number, column: ColumnMetaData, columnIndex: number) => JSX.Element;
  totalsRowRenderer?: (column: ColumnMetaData, index: number) => JSX.Element;
  selectable?: boolean;
  selectedRowIds?: Array<number>;
};

function Table(props: TableProps) {
  const numColumns = Object.keys(props.activeColumns).length + (props.selectable ? 1 : 0);
  const numRows = (props.data || []).length;
  const rowGridStyle = (numRows: number): string => {
    if (numRows > 0) {
      return `grid-template-rows: min-content repeat(${numRows}, min-content) auto ${
        props.showTotals ? "min-content" : ""
      }`;
    } else {
      return "grid-template-rows: min-content  auto min-content;";
    }
  };

  const tableClass = classNames(styles.tableContent, { [styles.loading]: props.isLoading });
  const customStyle =
    "." +
    styles.tableContent +
    " {" +
    "grid-template-columns: min-content repeat(" +
    numColumns +
    ",  min-content) auto;" +
    rowGridStyle(numRows) +
    "}";
  console.log(props.data);
  return (
    <div className={styles.tableWrapper}>
      <style>{customStyle}</style>

      <div className={tableClass}>
        {/*/!*Empty header at the start*!/*/}
        <HeaderCell
          heading={""}
          className={classNames(cellStyles.firstEmptyHeader, cellStyles.empty, cellStyles.locked)}
          key={"first-empty-header"}
        />
        {props.selectable && (
          <HeaderCell heading={""} className={classNames(cellStyles.empty)} key={"checkbox-header"} />
        )}
        {/* Header cells */}
        {props.activeColumns.map((column, index) => {
          return (
            <HeaderCell
              heading={column.heading}
              className={classNames(column.key, cellStyles[column.type || "NumericCell"])}
              key={column.key + index}
              locked={column.locked}
            />
          );
        })}
        {/*Empty header at the end*/}
        {
          <HeaderCell
            heading={""}
            className={classNames(cellStyles.lastEmptyHeader, cellStyles.empty)}
            key={"last-empty-header"}
          />
        }

        {/* The main cells containing the data */}
        {props.data.map((row: any, index: any) => {
          console.log(row);
          return (
            <TableRow
              row={row}
              rowIndex={index}
              columns={props.activeColumns}
              key={"row-" + index}
              selectable={props.selectable}
              selected={false}
              isTotalsRow={false}
            />
          );
        })}

        {/* Empty filler cells to create an empty row that expands to push the last row down.
           It's ugly but seems to be the only way to do it */}
        <div
          className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked, cellStyles.firstEmptyCell, {
            [cellStyles.noTotalsRow]: !props.showTotals,
          })}
        />

        {props.selectable && (
          <div
            className={classNames(cellStyles.cell, cellStyles.empty, { [cellStyles.noTotalsRow]: !props.showTotals })}
          />
        )}

        {props.activeColumns.map((column, index) => {
          return (
            <div
              className={classNames(cellStyles.cell, cellStyles.empty, column.key, {
                [cellStyles.noTotalsRow]: !props.showTotals,
              })}
              key={`mid-cell-${column.key}-${index}`}
            />
          );
        })}
        {/* Render a filler row to fill all available space vertically */}
        <div
          className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.lastEmptyCell, {
            [cellStyles.noTotalsRow]: !props.showTotals,
          })}
        />
        {props.showTotals && props.data.length && (
          <TableRow
            row={props.data[0]}
            rowIndex={props.activeColumns.length + 1}
            columns={props.activeColumns}
            key={"row-" + "total"}
            selectable={props.selectable}
            selected={false}
            isTotalsRow={true}
          />
        )}
      </div>
    </div>
  );
}

export default Table;
