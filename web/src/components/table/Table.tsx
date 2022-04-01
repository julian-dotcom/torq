import "./table.scss";
import HeaderCell from "./cells/HeaderCell";
import AliasCell from "./cells/AliasCell";
import NumericCell from "./cells/NumericCell";
import BarCell from "./cells/BarCell";
import EmptyCell from "./cells/EmptyCell";
import { FilterInterface } from "./filter";
import fieldSorter from "./sort";
import { useAppSelector } from "../../store/hooks";
import { selectChannels, selectColumns } from "./tableSlice";

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
  let columns = useAppSelector(selectColumns) || [];
  let channels = useAppSelector(selectChannels) || [];

  let total: RowType = {
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
  let max: RowType = {
    alias: "Max",
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

  // console.log(channels);
  // console.log(columns);

  // const newCha =
  // const channelsSorted = [...channels].sort(
  //   fieldSorter(["alias", "-capacity"])
  // );

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
  }

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
        {HeaderCell("", "first-empty-header", "empty locked")}

        {/* Header cells */}
        {columns.map(column => {
          return HeaderCell(column.heading, column.key, "", column.locked);
        })}

        {/*Empty header at the end*/}
        {HeaderCell("", "last-empty-header", "empty")}

        {/* The main cells containing the data */}
        {channels.map((row, index) => {
          let returnedRow = columns.map(column => {
            let key = column.key as keyof RowType;
            let past = channels[index][key];
            switch (column.type) {
              case "AliasCell":
                return AliasCell(row[key] as string, key, index);
              case "NumericCell":
                return NumericCell(
                  row[key] as number,
                  past as number,
                  key,
                  index
                );
              case "BarCell":
                return BarCell(
                  row[key] as number,
                  max[key] as number,
                  past as number,
                  key,
                  index
                );
              default:
                return NumericCell(
                  row[key] as number,
                  past as number,
                  key,
                  index
                );
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
        {columns.map(column => {
          return (
            <div
              className={"cell empty " + column.key}
              key={"mid-cell-" + column.key}
            />
          );
        })}
        {<div className={"cell empty "} />}

        {/* Totals row */}
        {/* Empty cell at the start */}
        {<div className={"cell empty total-cell locked"} />}
        {columns.map(column => {
          let key = column.key as keyof TotalType;
          switch (column.type) {
            case "AliasCell":
              return AliasCell("Total", "alias", "totals", "total-cell");
            case "NumericCell":
              return NumericCell(
                total[key] as number,
                total[key] as number,
                key,
                "totals",
                "total-cell"
              );
            case "BarCell":
              return BarCell(
                total[key] as number,
                total[key] as number,
                total[key] as number,
                key,
                "totals",
                "total-cell"
              );
            case "EmptyCell":
              return EmptyCell(key, "totals", "total-cell");
            default:
              return NumericCell(
                total[key] as number,
                total[key] as number,
                key,
                "totals",
                "total-cell"
              );
          }
        })}
        {/*Empty cell at the end*/}
        {<div className={"cell empty total-cell"} />}
      </div>
    </div>
  );
}

export default Table;
