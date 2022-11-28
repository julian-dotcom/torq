import { ColumnMetaData } from "features/table/types";
import NumericCell from "components/table/cells/numeric/NumericCell";
import BarCell from "components/table/cells/bar/BarCell";
import TextCell from "components/table/cells/text/TextCell";
import DurationCell from "components/table/cells/duration/DurationCell";
import BooleanCell from "components/table/cells/boolean/BooleanCell";
import DateCell from "components/table/cells/date/DateCell";
import EnumCell from "components/table/cells/enum/EnumCell";

export default function DefaultCellRenderer<T>(
  row: T,
  rowIndex: number,
  column: ColumnMetaData<T>,
  columnIndex: number
): JSX.Element {
  const dataKey = column.key as keyof T;
  // const heading = column.heading;
  const percent = column.percent;

  switch (typeof row[dataKey]) {
    case "string":
      switch (column.type) {
        case "TextCell":
          return <TextCell current={row[dataKey] as string} key={dataKey.toString() + rowIndex} />;
        case "DurationCell":
          return <DurationCell seconds={row[dataKey] as number} key={dataKey.toString() + rowIndex} />;
        case "EnumCell":
          return <EnumCell value={row[dataKey] as string} key={dataKey.toString() + rowIndex + columnIndex} />;
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
              key={dataKey.toString() + rowIndex + columnIndex}
            />
          );
      }
      break;
    case "number":
      switch (column.type) {
        case "NumericCell":
          return <NumericCell current={row[dataKey] as number} key={dataKey.toString() + rowIndex + columnIndex} />;
        case "DateCell":
          return <DateCell value={row[dataKey] as string} key={dataKey.toString() + rowIndex + columnIndex} />;

        case "BarCell":
          return (
            <BarCell
              current={row[dataKey] as number}
              total={column.max as number}
              showPercent={percent}
              key={dataKey.toString() + rowIndex + columnIndex}
            />
          );
      }
  }
  return <TextCell current={row[dataKey] as string} key={dataKey.toString() + rowIndex} />;
}
