import { ColumnMetaData } from "features/table/types";
import AliasCell from "components/table/cells/alias/AliasCell";
import NumericCell from "components/table/cells/numeric/NumericCell";
import BarCell from "components/table/cells/bar/BarCell";
import TextCell from "components/table/cells/text/TextCell";
import DurationCell from "components/table/cells/duration/DurationCell";
import BooleanCell from "components/table/cells/boolean/BooleanCell";
import DateCell from "components/table/cells/date/DateCell";
import EnumCell from "components/table/cells/enum/EnumCell";
// import LinkCell from "components/table/cells/link/LinkCell";
import { Forward } from "./forwardsTypes";

export function forwardsCellRenderer(
  row: Forward,
  rowIndex: number,
  column: ColumnMetaData<Forward>,
  columnIndex: number
): JSX.Element {
  const dataKey = column.key as keyof Forward;
  const heading = column.heading;
  const percent = column.percent;

  switch (typeof row[dataKey]) {
    case "string":
      switch (column.type) {
        case "AliasCell":
          return (
            <AliasCell
              current={row["alias"] as string}
              lndShortChannelId={row["lndShortChannelId"]}
              open={row["open"]}
              key={dataKey + rowIndex + columnIndex}
            />
          );
        case "TextCell":
          return <TextCell current={row[dataKey] as string} key={column.key + rowIndex} />;
        case "DurationCell":
          return <DurationCell seconds={row[dataKey] as number} key={column.key + rowIndex} />;
        case "EnumCell":
          return <EnumCell value={row[dataKey] as string} key={dataKey + rowIndex + columnIndex} />;
      }
      break;
    case "boolean":
      switch (column.type) {
        case "BooleanCell":
          return (
            <BooleanCell
              falseTitle={"Failure"}
              trueTitle={"Success"}
              value={row[dataKey] as boolean}
              key={dataKey + rowIndex + columnIndex}
            />
          );
      }
      break;
    case "number":
      switch (column.type) {
        case "NumericCell":
          return <NumericCell current={row[dataKey] as number} key={dataKey + rowIndex + columnIndex} />;
        case "DateCell":
          return <DateCell value={row[dataKey] as string} key={dataKey + rowIndex + columnIndex} />;

        case "BarCell":
          return (
            <BarCell
              current={row[dataKey] as number}
              total={column.max as number}
              showPercent={percent}
              key={dataKey + rowIndex + columnIndex}
            />
          );
      }
      break;
    default:
      return <TextCell current={row[dataKey] as string} key={column.key + rowIndex} />;
  }
  return <TextCell current={row[dataKey] as string} key={column.key + rowIndex} />;
}
