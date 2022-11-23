import styles from "./table.module.scss";
import cellStyles from "components/table/cells/cell.module.scss";
import HeaderCell from "components/table/cells/header/HeaderCell";
import AliasCell from "components/table/cells/alias/AliasCell";
import NumericDoubleCell from "components/table/cells/numeric/NumericDoubleCell";
import NumericCell from "components/table/cells/numeric/NumericCell";
import BarCell from "components/table/cells/bar/BarCell";
import TextCell from "components/table/cells/text/TextCell";
import DurationCell from "components/table/cells/duration/DurationCell";
import BooleanCell from "components/table/cells/boolean/BooleanCell";
import classNames from "classnames";
import DateCell from "components/table/cells/date/DateCell";
import EnumCell from "components/table/cells/enum/EnumCell";
import LinkCell from "components/table/cells/link/LinkCell";
import BalanceCell from "components/table/cells/balance/BalanceCell";
import { SortByOptionType } from "features/sidebar/sections/sort/SortSectionOld";

export interface ColumnMetaData {
  heading: string;
  key: string;
  key2?: string;
  suffix?: string;
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

type TableProps = {
  activeColumns: Array<ColumnMetaData>;
  data: Array<any>;
  isLoading: boolean;
  showTotals?: boolean;
  rowRenderer?: (row: any, index: number, column: ColumnMetaData, columnIndex: number) => JSX.Element;
  totalsRowRenderer?: (column: ColumnMetaData, index: number) => JSX.Element;
};

function defaultRowRenderer(row: any, index: number, column: ColumnMetaData, columnIndex: number) {
  const key = column.key;
  const key2 = column?.key2 as string;
  const suffix = column?.suffix;
  const heading = column.heading;
  const percent = column.percent;
  switch (column.type) {
    case "AliasCell":
      return (
        <AliasCell
          current={row[key] as string}
          lndShortChannelId={row["lndShortChannelId"]}
          open={row["open"]}
          className={classNames(key, index, cellStyles.locked)}
          key={key + index + columnIndex}
        />
      );
    case "LinkCell":
      return (
        <LinkCell
          current={`Check in ${heading}`}
          link={row[key]}
          className={classNames(key, index)}
          key={key + index + columnIndex}
        />
      );
    case "NumericCell":
      return <NumericCell current={row[key] as number} className={key} key={key + index + columnIndex} />;
    case "BalanceCell":
      return <BalanceCell local={row[key] as number} remote={row[key2] as number} capacity={row?.capacity | 0} className={key} key={key + index + columnIndex} />;
    case "NumericDoubleCell":
      return <NumericDoubleCell local={row[key] as number} remote={row[key2] as number} suffix={suffix as string} className={key} key={key + index + columnIndex} />;
    case "DateCell":
      return <DateCell value={row[key] as string} className={key} key={key + index + columnIndex} />;
    case "BooleanCell":
      return (
        <BooleanCell
          falseTitle={"Failure"}
          trueTitle={"Success"}
          value={row[key] as boolean}
          className={classNames(key)}
          key={key + index + columnIndex}
        />
      );
    case "BarCell":
      return (
        <BarCell
          current={row[key] as number}
          total={column.max as number}
          className={key}
          showPercent={percent}
          key={key + index + columnIndex}
        />
      );
    case "TextCell":
      return (
        <TextCell current={row[key] as string} className={classNames(column.key, index)} key={column.key + index} />
      );
    case "DurationCell":
      return (
        <DurationCell seconds={row[key] as number} className={classNames(column.key, index)} key={column.key + index} />
      );
    case "EnumCell":
      return <EnumCell value={row[key] as string} className={key} key={key + index + columnIndex} />;
    default:
      return <NumericCell current={row[key] as number} className={key} key={key + index + columnIndex} />;
  }
}

function defaultTotalsRowRenderer(column: ColumnMetaData, index: number) {
  switch (column.type) {
    case "AliasCell":
      return (
        <AliasCell
          current={"Total"}
          lndShortChannelId={""}
          className={classNames(column.key, index, cellStyles.locked, cellStyles.totalCell)}
          key={column.key + index}
        />
      );
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
          className={classNames(column.key, index, styles.textCell, cellStyles.totalCell, cellStyles.firstTotalCell)}
          key={column.key + index}
        />
      );
    case "DurationCell":
      return (
        <TextCell
          current={" "}
          className={classNames(column.key, index, styles.textCell, cellStyles.totalCell, cellStyles.firstTotalCell)}
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

function Table(props: TableProps) {
  const numColumns = Object.keys(props.activeColumns).length;
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

  const rowRenderer = props.rowRenderer ? props.rowRenderer : defaultRowRenderer;
  const totalsRowRenderer = props.totalsRowRenderer ? props.totalsRowRenderer : defaultTotalsRowRenderer;

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
    <div className={styles.tableWrapper}>
      <style>{customStyle}</style>

      <div className={tableClass}>
        {/*/!*Empty header at the start*!/*/}
        {
          <HeaderCell
            heading={""}
            className={classNames(cellStyles.firstEmptyHeader, cellStyles.empty, cellStyles.locked)}
            key={"first-empty-header"}
          />
        }

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
          // Adds empty cells at the start and end of each row. This is to give the table a buffer at each end.
          return (
            <div className={classNames(styles.tableRow, "torq-row-" + index)} key={"torq-row-" + index}>
              {[
                <div
                  className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked)}
                  key={"first-cell-" + index}
                />,
                ...props.activeColumns.map((column, columnIndex) => {
                  return rowRenderer(row, index, column, columnIndex);
                }),
                <div
                  className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.lastEmptyCell)}
                  key={"last-cell-" + index}
                />,
              ]}
            </div>
          );
        })}

        {/* Empty filler cells to create an empty row that expands to push the last row down.
           It's ugly but seems to be the only way to do it */}
        {
          <div
            className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked, cellStyles.firstEmptyCell, {
              [cellStyles.noTotalsRow]: !props.showTotals,
            })}
          />
        }
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
        {
          <div
            className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.lastEmptyCell, {
              [cellStyles.noTotalsRow]: !props.showTotals,
            })}
          />
        }

        {/* Totals row */}
        {/* Empty cell at the start */}
        {props.showTotals && (
          <div className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked, cellStyles.totalCell)} />
        )}
        {props.showTotals && props.activeColumns.map(totalsRowRenderer)}
        {/*Empty cell at the end*/}
        {props.showTotals && (
          <div
            className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.totalCell, cellStyles.lastTotalCell)}
          />
        )}
      </div>
    </div>
  );
}

export default Table;
