import styles from "./table.module.scss";
import cellStyles from "../cells/cell.module.scss";
import HeaderCell from "../cells/HeaderCell";
import AliasCell from "../cells/AliasCell";
import NumericCell from "../cells/NumericCell";
import BarCell from "../cells/BarCell";
import TextCell from "../cells/TextCell";
import { useAppSelector } from "../../../store/hooks";
import { selectChannels, selectActiveColumns, selectStatus, ColumnMetaData } from "../tableSlice";
import classNames from "classnames";
import clone from "clone"

function Table() {
  const activeColumns = clone<ColumnMetaData[]>(useAppSelector(selectActiveColumns)) || [];
  const channels = useAppSelector(selectChannels) || [];
  const status = useAppSelector(selectStatus)

  if (channels.length > 0) {

    for (const channel of channels) {
      for (const column of activeColumns) {
        column.total = (column.total ?? 0) + channel[column.key]
        column.max = Math.max(column.max ?? 0, channel[column.key] ?? 0)
      }
    }

    const turnover_total_col = activeColumns.find(col => col.key === "turnover_total")
    const turnover_out_col = activeColumns.find(col => col.key === "turnover_out")
    const turnover_in_col = activeColumns.find(col => col.key === "turnover_in")
    const amount_total_col = activeColumns.find(col => col.key === "amount_total")
    const amount_out_col = activeColumns.find(col => col.key === "amount_out")
    const amount_in_col = activeColumns.find(col => col.key === "amount_in")
    const capacity_col = activeColumns.find(col => col.key === "capacity")
    if (turnover_total_col) {
      turnover_total_col.total = (amount_total_col?.total ?? 0) / (capacity_col?.total ?? 1)
    }

    if (turnover_out_col) {
      turnover_out_col.total = (amount_out_col?.total ?? 0) / (capacity_col?.total ?? 1)
    }

    if (turnover_in_col) {
      turnover_in_col.total = (amount_in_col?.total ?? 0) / (capacity_col?.total ?? 1)
    }
  }

  const numColumns = Object.keys(activeColumns).length;
  const numRows = channels.length;
  const rowGridStyle = (numRows: number): string => {
    if (numRows > 0) {
      return ("grid-template-rows: min-content repeat(" + numRows + ",min-content) auto min-content;");
    } else {
      return "grid-template-rows: min-content  auto min-content;";
    }
  };

  const tableClass = classNames(styles.tableContent, {
    [styles.loading]: status === 'loading',
    [styles.idle]: status === 'idle'
  });

  const customStyle = "." + styles.tableContent + " {" +
      "grid-template-columns: min-content repeat(" +numColumns +",  minmax(min-content, auto)) min-content;"+
      rowGridStyle(numRows) +
    "}"


  return (
    <div className={styles.tableWrapper}>
      <style>
        {customStyle}
      </style>

      <div className={tableClass}>
        {/*Empty header at the start*/}
        {<HeaderCell heading={""} className={classNames(cellStyles.firstEmptyHeader, cellStyles.empty, cellStyles.locked)} key={"first-empty-header"} />}

        {/* Header cells */}
        {activeColumns.map((column, index) => {
          return <HeaderCell heading={column.heading} className={column.key} key={column.key + index} locked={column.locked} />
        })}

        {/*Empty header at the end*/}
        {<HeaderCell heading={""} className={classNames(cellStyles.lastEmptyHeader, cellStyles.empty)} key={"last-empty-header"} />}

        {/* The main cells containing the data */}
        {channels.map((row, index) => {
          const returnedRow = activeColumns.map((column, columnIndex) => {
            const key = column.key;
            switch (column.type) {
              case "AliasCell":
                return <AliasCell current={row[key] as string} className={classNames(key, index, cellStyles.locked)} key={key + index + columnIndex} />
              case "NumericCell":
                return <NumericCell current={row[key] as number} index={index} className={key} key={key + index + columnIndex} />;
              case "BarCell":
                return <BarCell current={row[key] as number} previous={row[key] as number} total={column.max as number} index={index} className={key} key={key + index + columnIndex} />;
            case "TextCell":
              return <TextCell current={row[key] as string} className={classNames(column.key, index)} key={column.key + index} />
              default:
                return <NumericCell current={row[key] as number} index={index} className={key} key={key + index + columnIndex} />;
            }
          });
          // Adds empty cells at the start and end of each row. This is to give the table a buffer at each end.
          return (<div className={classNames(styles.tableRow, "torq-row-"+index)} key={"torq-row-"+index}>{[
            <div className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked)} key={"first-cell-" + index}/>,
            ...returnedRow,
            <div className={classNames(cellStyles.cell, cellStyles.empty)} key={"last-cell-" + index}/>
          ]}</div>);
        })}

        {/* Empty filler cells to create an empty row that expands to push the last row down.
           It's ugly but seems to be the only way to do it */}
        {<div className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked)} />}
        {activeColumns.map((column, index) => {
          return (
            <div
              className={classNames(cellStyles.cell, cellStyles.empty, column.key)}
              key={`mid-cell-${column.key}-${index}`}
            />
          );
        })}


        {<div className={classNames(cellStyles.cell, cellStyles.empty)} />}

        {/* Totals row */}
        {/* Empty cell at the start */}
        {<div className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.locked, cellStyles.totalCell)} />}
        {activeColumns.map((column, index) => {
          switch (column.type) {
            case "AliasCell":
              return <AliasCell current={"Total"} className={classNames(column.key, index, cellStyles.locked, cellStyles.totalCell)} key={column.key + index} />
            case "NumericCell":
              return <NumericCell current={column.total as number} index={index} className={classNames(column.key, index, cellStyles.totalCell)} key={`total-${column.key}-${index}`} />;
            case "BarCell":
              return <BarCell current={column.total as number} previous={column.total as number} total={column.max as number} index={index} className={classNames(column.key, index, cellStyles.totalCell)} key={`total-${column.key}-${index}`} />;
            case "TextCell":
              return <TextCell current={' '} className={classNames(column.key, index, cellStyles.totalCell)} key={column.key + index} />
            default:
              return <NumericCell current={column.total as number} index={index} className={classNames(column.key, index, cellStyles.totalCell)} key={`total-${column.key}-${index}`} />;
          }
        })}
        {/*Empty cell at the end*/}
        {<div className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.totalCell)} />}
      </div>


    </div>
  );
}

export default Table;
