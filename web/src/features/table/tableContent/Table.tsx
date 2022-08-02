import styles from "./table.module.scss";
import cellStyles from "../cells/cell.module.scss";
import HeaderCell from "../cells/HeaderCell";
import AliasCell from "../cells/AliasCell";
import NumericCell from "../cells/NumericCell";
import BarCell from "../cells/BarCell";
import TextCell from "../cells/TextCell";
import { useAppSelector } from "../../../store/hooks";
import { ColumnMetaData, selectFilters, selectGroupBy, selectSortBy } from "../../forwards/forwardsSlice";
import classNames from "classnames";
import clone from "clone";

type TableProps = {
  activeColumns: Array<ColumnMetaData>;
  data: Array<any>;
  isLoading: boolean;
};

function Table(props: TableProps) {
  const numColumns = Object.keys(props.activeColumns).length;
  const numRows = (props.data || []).length;
  const rowGridStyle = (numRows: number): string => {
    if (numRows > 0) {
      return "grid-template-rows: min-content repeat(" + numRows + ",min-content) auto min-content;";
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
    ",  minmax(min-content, auto)) min-content;" +
    rowGridStyle(numRows) +
    "}";

  // if (props.isLoading == true) {
  //   return <div className={styles.tableWrapper}>No data</div>;
  // }

  return (
    <div className={styles.tableWrapper}>
      <style>{customStyle}</style>

      <div className={tableClass}>
        {/*Empty header at the start*/}
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
              className={classNames(column.key, column.type === "TextCell" ? cellStyles.textCell : "")}
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
          const returnedRow = props.activeColumns.map((column, columnIndex) => {
            const key = column.key;
            switch (column.type) {
              case "AliasCell":
                return (
                  <AliasCell
                    current={row[key] as string}
                    chanId={row["chan_id"]}
                    open={row["open"]}
                    className={classNames(key, index, cellStyles.locked)}
                    key={key + index + columnIndex}
                  />
                );
              case "NumericCell":
                return <NumericCell current={row[key] as number} className={key} key={key + index + columnIndex} />;
              case "BarCell":
                return (
                  <BarCell
                    current={row[key] as number}
                    previous={row[key] as number}
                    total={column.max as number}
                    className={key}
                    key={key + index + columnIndex}
                  />
                );
              case "TextCell":
                return (
                  <TextCell
                    current={row[key] as string}
                    className={classNames(column.key, index, styles.textCell)}
                    key={column.key + index}
                  />
                );
              default:
                return <NumericCell current={row[key] as number} className={key} key={key + index + columnIndex} />;
            }
          });
          // Adds empty cells at the start and end of each row. This is to give the table a buffer at each end.
          return (
            <div className={classNames(styles.tableRow, "torq-row-" + index)} key={"torq-row-" + index}>
              {[
                <div
                  className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked)}
                  key={"first-cell-" + index}
                />,
                ...returnedRow,
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
        {<div className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked)} />}
        {props.activeColumns.map((column, index) => {
          return (
            <div
              className={classNames(cellStyles.cell, cellStyles.empty, column.key)}
              key={`mid-cell-${column.key}-${index}`}
            />
          );
        })}

        {<div className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.lastEmptyCell)} />}

        {/* Totals row */}
        {/* Empty cell at the start */}
        {<div className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked, cellStyles.totalCell)} />}
        {props.activeColumns.map((column, index) => {
          switch (column.type) {
            case "AliasCell":
              return (
                <AliasCell
                  current={"Total"}
                  chanId={""}
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
            case "BarCell":
              return (
                <BarCell
                  current={column.total as number}
                  previous={column.total as number}
                  total={column.max as number}
                  className={classNames(column.key, index, cellStyles.totalCell)}
                  key={`total-${column.key}-${index}`}
                />
              );
            case "TextCell":
              return (
                <TextCell
                  current={" "}
                  className={classNames(
                    column.key,
                    index,
                    styles.textCell,
                    cellStyles.totalCell,
                    cellStyles.firstTotalCell
                  )}
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
        })}
        {/*Empty cell at the end*/}
        {
          <div
            className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.totalCell, cellStyles.lastTotalCell)}
          />
        }
      </div>
    </div>
  );
}

export default Table;
