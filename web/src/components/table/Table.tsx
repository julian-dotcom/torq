import "./table.scss";
import HeaderCell from "./cells/HeaderCell";
import AliasCell from "./cells/AliasCell";
import NumericCell from "./cells/NumericCell";
import BarCell from "./cells/BarCell";
import { useAppSelector } from "../../store/hooks";
import { selectChannels, selectActiveColumns } from "./tableSlice";
import classNames from "classnames";

interface RowType {
  alias: string;
  amount_out: number;
  amount_in: number;
  amount_total: number;
  revenue_out: number;
  revenue_in: number;
  revenue_total: number;
  count_out: number;
  count_total: number;
  count_in: number;
  turnover_out: number;
  turnover_in: number;
  turnover_total: number;
  capacity: number;
}
interface TotalType {
  amount_out: number;
  amount_in: number;
  amount_total: number;
  revenue_out: number;
  revenue_in: number;
  revenue_total: number;
  count_out: number;
  count_total: number;
  count_in: number;
  turnover_out: number;
  turnover_in: number;
  turnover_total: number;
  capacity: number;
}

function Table() {
  let columns = useAppSelector(selectActiveColumns) || [];
  let channels = useAppSelector(selectChannels) || [];

  // TODO: Clean this up. Create a user selectable totals.
  let baseAcc: RowType = {
    alias: "Total",
    amount_out: 0,
    amount_in: 0,
    amount_total: 0,
    revenue_out: 0,
    revenue_in: 0,
    revenue_total: 0,
    count_out: 0,
    count_total: 0,
    count_in: 0,
    turnover_out: 0,
    turnover_in: 0,
    turnover_total: 0,
    capacity: 0
  };
  let total = Object.assign({}, baseAcc);
  let max = Object.assign({}, baseAcc);
  let average = Object.assign({}, baseAcc);

  // console.log(channelsSorted);
  if (channels.length > 0) {
    channels.forEach(row => {
      Object.keys(total).forEach(column => {
        // @ts-ignore
        total[column as keyof RowType] += row[column];
        // @ts-ignore
        max[column as keyof RowType] = Math.max(
          row[column],
          // @ts-ignore
          max[column as keyof RowType]
        );
      });
    });

    Object.keys(total).forEach(key => {
      // @ts-ignore
      average[key] = total[key] / channels.length
    })

    total.turnover_total = total.amount_total / total.capacity
    total.turnover_out = total.amount_out / total.capacity
    total.turnover_in = total.amount_in / total.capacity
  }

  const numColumns = Object.keys(columns).length;
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

  return (
    <div className="table-wrapper">
      <style>
        {".table-content {grid-template-columns: min-content repeat(" +
          numColumns +
          ",  minmax(min-content, auto)) min-content;" +
          rowGridStyle(numRows) +
          "}"}
      </style>
      <div className="table-content">
        {/*Empty header at the start*/}
        {<HeaderCell heading={""} className={"first-empty-header empty locked"} key={"first-empty-header"} />}

        {/* Header cells */}
        {columns.map((column, index) => {
          return <HeaderCell heading={column.heading} className={column.key} key={column.key + index} locked={column.locked} />
        })}

        {/*Empty header at the end*/}
        {<HeaderCell heading={""} className={"last-empty-header empty locked"} key={"last-empty-header"} />}

        {/* The main cells containing the data */}
        {channels.map((row, index) => {
          let returnedRow = columns.map((column, columnIndex) => {
            let key = column.key as keyof RowType;
            let past = channels[index][key];
            switch (column.type) {
              case "AliasCell":
                return <AliasCell current={row[key] as string} className={classNames(key, index, "locked")} key={key + index + columnIndex} />
              case "NumericCell":
                return <NumericCell current={row[key] as number} index={index} className={key} key={key + index + columnIndex} />;
              case "BarCell":
                return <BarCell current={row[key] as number} previous={row[key] as number} total={max[key] as number} index={index} className={key} key={key + index + columnIndex} />;
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
        {columns.map((column, index) => {
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
        {columns.map((column, index) => {
          let key = column.key as keyof TotalType;
          // const isTurnover = ['turnover_total', 'turnover_out', 'turnover_in'].find((item) => item === key)
          // let value = isTurnover ? average[key] as number :total[key] as number
          let value = total[key] as number;
          switch (column.type) {
            case "AliasCell":
              // @ts-ignore
              return <AliasCell current={"Total"} className={classNames(key, index, "total-cell locked")} key={key + index} />
            case "NumericCell":
              return <NumericCell current={value} index={index} className={key + " total-cell"} key={`total-${key}-${index}`} />;
            case "BarCell":
              return <BarCell current={value} previous={value} total={max[key] as number} index={index} className={key + " total-cell"} key={`total-${key}-${index}`} />;
            default:
              return <NumericCell current={value} index={index} className={key + " total-cell"} key={`total-${key}-${index}`} />;
          }
        })}
        {/*Empty cell at the end*/}
        {<div className={"cell empty total-cell"} />}
      </div>
    </div>
  );
}

export default Table;
