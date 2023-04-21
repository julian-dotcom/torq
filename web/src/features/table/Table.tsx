import styles from "./table.module.scss";
import cellStyles from "components/table/cells/cell.module.scss";
import HeaderCell from "components/table/cells/header/HeaderCell";
import classNames from "classnames";
import TableRow from "./TableRow";
import { TableProps } from "./types";

function Table<T>(props: TableProps<T>) {
  const numColumns = (props.selectable ? 1 : 0) + props.activeColumns.length;
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
  return (
    <div className={styles.tableWrapper} data-intercom-target={props.intercomTarget}>
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
              className={classNames(column.key as string, cellStyles[(column.type as string) || "NumericCell"])}
              key={(column.key as string) + index}
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
        {props.data.map((row: T, index: number) => {
          return (
            <TableRow
              row={row}
              rowIndex={index}
              cellRenderer={props.cellRenderer}
              columns={props.activeColumns}
              key={"table-row-" + index}
              selectable={props.selectable}
              selected={false}
              isTotalsRow={false}
              maxRow={props.maxRow}
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
              className={classNames(cellStyles.cell, cellStyles.empty, column.key as string, {
                [cellStyles.noTotalsRow]: !props.showTotals,
              })}
              key={`mid-cell-${column.key as string}-${index}`}
            />
          );
        })}
        {/* Render a filler row to fill all available space vertically */}
        <div
          className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.lastEmptyCell, {
            [cellStyles.noTotalsRow]: !props.showTotals,
          })}
        />
        {props.showTotals && props.totalRow && (
          <TableRow
            row={props.totalRow}
            cellRenderer={props.cellRenderer}
            rowIndex={props.activeColumns.length + 1}
            columns={props.activeColumns}
            key={"row-" + "total" + 1}
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
