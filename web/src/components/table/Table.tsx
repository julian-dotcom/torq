import "./table.scss";
import HeaderCell from "./cells/HeaderCell";
import AliasCell from "./cells/AliasCell";
import NumericCell from "./cells/NumericCell";
import BarCell from "./cells/BarCell";
import { useAppSelector } from "../../store/hooks";
import { selectChannels, selectActiveColumns, selectStatus, ColumnMetaData } from "./tableSlice";
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
      return (
        "grid-template-rows: min-content repeat(" +
        numRows +
        ",min-content) auto min-content;"
      );
    } else {
      return "grid-template-rows: min-content  auto min-content;";
    }
  };

  const tableClass = classNames("table-content", {
    'loading': status === 'loading',
    'idle': status === 'idle'
  });

  return (
    <div className="table-wrapper">
      <style>
        {".table-content {grid-template-columns: min-content repeat(" +
          numColumns +
          ",  minmax(min-content, auto)) min-content;" +
          rowGridStyle(numRows) +
          "}"}
      </style>

      <div className={tableClass}>
        {/*Empty header at the start*/}
        {<HeaderCell heading={""} className={"first-empty-header empty locked"} key={"first-empty-header"} />}

        {/* Header cells */}
        {activeColumns.map((column, index) => {
          return <HeaderCell heading={column.heading} className={column.key} key={column.key + index} locked={column.locked} />
        })}

        {/*Empty header at the end*/}
        {<HeaderCell heading={""} className={"last-empty-header empty locked"} key={"last-empty-header"} />}

        {/* The main cells containing the data */}
        {channels.map((row, index) => {
          const returnedRow = activeColumns.map((column, columnIndex) => {
            const key = column.key;
            switch (column.type) {
              case "AliasCell":
                return <AliasCell current={row[key] as string} className={classNames(key, index, "locked")} key={key + index + columnIndex} />
              case "NumericCell":
                return <NumericCell current={row[key] as number} index={index} className={key} key={key + index + columnIndex} />;
              case "BarCell":
                return <BarCell current={row[key] as number} previous={row[key] as number} total={column.max as number} index={index} className={key} key={key + index + columnIndex} />;
              default:
                return <NumericCell current={row[key] as number} index={index} className={key} key={key + index + columnIndex} />;
            }
          });
          // Adds empty cells at the start and end of each row. This is to give the table a buffer at each end.
          return [
            <div className={"cell empty locked"} key={"first-cell-" + index} />,
            ...returnedRow,
            <div className={"cell empty"} key={"last-cell-" + index} />
          ];
        })}

        {/* Empty filler cells to create an empty row that expands to push the last row down.
           It's ugly but seems to be the only way to do it */}
        {<div className={"cell empty locked"} />}
        {activeColumns.map((column, index) => {
          return (
            <div
              className={`cell empty ${column.key}`}
              key={`mid-cell-${column.key}-${index}`}
            />
          );
        })}


        {<div className={"cell empty "} />}

        {/* Totals row */}
        {/* Empty cell at the start */}
        {<div className={"cell empty total-cell locked"} />}
        {activeColumns.map((column, index) => {
          let value = column.total as number;
          switch (column.type) {
            case "AliasCell":
              return <AliasCell current={"Total"} className={classNames(column.key, index, "total-cell locked")} key={column.key + index} />
            case "NumericCell":
              return <NumericCell current={value} index={index} className={column.key + " total-cell"} key={`total-${column.key}-${index}`} />;
            case "BarCell":
              return <BarCell current={value} previous={value} total={column.max as number} index={index} className={column.key + " total-cell"} key={`total-${column.key}-${index}`} />;
            default:
              return <NumericCell current={value} index={index} className={column.key + " total-cell"} key={`total-${column.key}-${index}`} />;
          }
        })}
        {/*Empty cell at the end*/}
        {<div className={"cell empty total-cell"} />}
      </div>


    </div>
  );
}

export default Table;
