import styles from "./table.module.scss";
import cellStyles from "components/table/cells/cell.module.scss";
import AliasCell from "components/table/cells/alias/AliasCell";
import NumericCell from "components/table/cells/numeric/NumericCell";
import BarCell from "components/table/cells/bar/BarCell";
import TextCell from "components/table/cells/text/TextCell";
import DurationCell from "components/table/cells/duration/DurationCell";
import BooleanCell from "components/table/cells/boolean/BooleanCell";
import CheckboxCell from "components/table/cells/checkbox/CheckboxCell";
import classNames from "classnames";
import DateCell from "components/table/cells/date/DateCell";
import EnumCell from "components/table/cells/enum/EnumCell";
import LinkCell from "components/table/cells/link/LinkCell";
import { ReactNode } from "react";
import { ColumnMetaData } from "./Table";

type rowRendererProps = {
  row: any;
  rowIndex: number;
  column: ColumnMetaData;
  columnIndex: number;
};

function defaultRowRenderer(row: any, rowIndex: number, columnIndex: number, column: ColumnMetaData) {
  const dataKey = column.key;
  const heading = column.heading;
  const percent = column.percent;
  switch (column.type) {
    case "AliasCell":
      return (
        <AliasCell
          current={row[dataKey] as string}
          lndShortChannelId={row["lndShortChannelId"]}
          open={row["lndShortChannelId"]}
          className={classNames(dataKey, rowIndex, cellStyles.locked)}
          key={dataKey + rowIndex + columnIndex}
        />
      );
    case "LinkCell":
      return (
        <LinkCell
          current={`Check in ${heading}`}
          link={row[dataKey]}
          className={classNames(dataKey, rowIndex)}
          key={dataKey + rowIndex + columnIndex}
        />
      );
    case "NumericCell":
      return (
        <NumericCell
          current={row[dataKey] as number}
          className={classNames(dataKey)}
          key={dataKey + rowIndex + columnIndex}
        />
      );
    case "DateCell":
      return (
        <DateCell
          value={row[dataKey] as string}
          className={classNames(dataKey)}
          key={dataKey + rowIndex + columnIndex}
        />
      );
    case "BooleanCell":
      return (
        <BooleanCell
          falseTitle={"Failure"}
          trueTitle={"Success"}
          value={row[dataKey] as boolean}
          className={classNames(dataKey)}
          key={dataKey + rowIndex + columnIndex}
        />
      );
    case "BarCell":
      return (
        <BarCell
          current={row[dataKey] as number}
          total={column.max as number}
          className={classNames(dataKey)}
          showPercent={percent}
          key={dataKey + rowIndex + columnIndex}
        />
      );
    case "TextCell":
      return (
        <TextCell
          current={row[dataKey] as string}
          className={classNames(column.key, rowIndex)}
          key={column.key + rowIndex}
        />
      );
    case "DurationCell":
      return (
        <DurationCell
          seconds={row[dataKey] as number}
          className={classNames(column.key, rowIndex)}
          key={column.key + rowIndex}
        />
      );
    case "EnumCell":
      return (
        <EnumCell
          value={row[dataKey] as string}
          className={classNames(dataKey)}
          key={dataKey + rowIndex + columnIndex}
        />
      );
    default:
      return (
        <NumericCell
          current={row[dataKey] as number}
          className={classNames(dataKey)}
          key={dataKey + rowIndex + columnIndex}
        />
      );
  }
}

function defaultTotalsRowRenderer(column: ColumnMetaData, index: number) {
  switch (column.type) {
    case "NumericCell":
      return (
        <NumericCell
          current={column.total as number}
          className={classNames(column.key, index, cellStyles.totalCell)}
          key={`total-${column.key}-${index}`}
        />
      );
    case "BooleanCell":
      return (
        <BooleanCell
          value={false}
          className={classNames(column.key, index, cellStyles.totalCell)}
          key={`total-${column.key}-${index}`}
        />
      );
    case "BarCell":
      return (
        <BarCell
          current={column.total as number}
          total={column.max as number}
          className={classNames(column.key, index, cellStyles.totalCell)}
          key={`total-${column.key}-${index}`}
        />
      );
    case "TextCell":
      return (
        <TextCell
          current={" "}
          className={classNames(column.key, index, styles.textCell, cellStyles.totalCell)}
          key={column.key + index}
        />
      );
    case "DurationCell":
      return (
        <TextCell
          current={" "}
          className={classNames(column.key, index, styles.textCell, cellStyles.totalCell)}
          key={column.key + index}
        />
      );
    default:
      return (
        <NumericCell
          current={column.total as number}
          className={classNames(column.key, index, cellStyles.totalCell)}
          key={`total-${column.key}-${index}`}
        />
      );
  }
}

type RowPropes = {
  row: any;
  rowIndex: number;
  columns: Array<ColumnMetaData>;
  selectable?: boolean;
  selected: boolean;
  // rowRenderer?: (row: any, index: number, column: ColumnMetaData, columnIndex: number) => JSX.Element;
  totalsRowRenderer?: (column: ColumnMetaData, index: number) => JSX.Element;
  isTotalsRow?: boolean;
};

function Row(props: RowPropes) {
  // console.log(props);
  // const rowRenderer = props.rowRenderer ? props.rowRenderer : defaultRowRenderer;
  const totalsRowRenderer = props.totalsRowRenderer ? props.totalsRowRenderer : defaultTotalsRowRenderer;

  // Adds empty cells at the start and end of each row. This is to give the table a buffer at each end.
  const rowContent: Array<ReactNode> = [];
  rowContent.push(
    <div
      className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked, {
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
    ...props.columns.map((column: ColumnMetaData, columnIndex) => {
      return defaultRowRenderer(props.row, props.rowIndex, columnIndex, column);
    })
  );

  rowContent.push(
    <div
      className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.lastEmptyCell, {
        [cellStyles.totalCell]: props.isTotalsRow,
      })}
      key={"last-cell-" + props.rowIndex}
    />
  );

  return (
    <div
      className={classNames(cellStyles.tableRow, "torq-row-" + props.rowIndex, {
        [cellStyles.totalCell]: props.isTotalsRow,
      })}
      key={"torq-row-" + props.rowIndex}
    >
      {rowContent}
    </div>
  );
}

export default Row;
