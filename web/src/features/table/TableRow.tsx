import cellStyles from "components/table/cells/cell.module.scss";
import CheckboxCell from "components/table/cells/checkbox/CheckboxCell";
import classNames from "classnames";
import { ReactNode } from "react";
import { RowProp, ColumnMetaData } from "./types";

function Row<T>(props: RowProp<T>) {
  // const totalsRowRenderer = props.totalsRowRenderer ? props.totalsRowRenderer : defaultTotalsRowRenderer;

  // Adds empty cells at the start and end of each row. This is to give the table a buffer at each end.
  const rowContent: Array<ReactNode> = [];
  rowContent.push(
    <div
      className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.firstEmptyCell, cellStyles.locked, {
        [cellStyles.totalCell]: props.isTotalsRow,
      })}
      key={"first-cell-" + props.rowIndex}
    />
  );

  if (props.selectable) {
    rowContent.push(
      <CheckboxCell
        key={"checkbox-cell-" + props.rowIndex}
        checked={props.selected}
        className={classNames({ [cellStyles.totalCell]: props.isTotalsRow })}
      />
    );
  }

  rowContent.push(
    ...props.columns.map((columnMeta: ColumnMetaData<T>, columnIndex) => {
      return props.cellRenderer(props.row, props.rowIndex, columnMeta, columnIndex, props.isTotalsRow);
    })
  );

  rowContent.push(
    <div
      className={classNames(
        cellStyles.cell,
        cellStyles.empty,
        {
          [cellStyles.lastTotalCell]: props.isTotalsRow,
        },
        cellStyles.lastEmptyCell
      )}
      key={"last-cell-" + props.rowIndex}
    />
  );

  return (
    <div
      className={classNames(cellStyles.tableRow, "torq-row-" + props.rowIndex, {
        [cellStyles.totalsRow]: props.isTotalsRow,
      })}
      key={"torq-row-" + props.rowIndex}
    >
      {rowContent}
    </div>
  );
}

export default Row;
