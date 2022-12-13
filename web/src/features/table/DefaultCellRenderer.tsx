import { ColumnMetaData } from "features/table/types";
import NumericCell from "components/table/cells/numeric/NumericCell";
import BarCell from "components/table/cells/bar/BarCell";
import TextCell from "components/table/cells/text/TextCell";
import DurationCell from "components/table/cells/duration/DurationCell";
import BooleanCell from "components/table/cells/boolean/BooleanCell";
import DateCell from "components/table/cells/date/DateCell";
import EnumCell from "components/table/cells/enum/EnumCell";
import NumericDoubleCell from "components/table/cells/numeric/NumericDoubleCell";
import AliasCell from "components/table/cells/alias/AliasCell";
import styles from "components/table/cells/cell.module.scss";
import LongTextCell from "components/table/cells/longText/LongTextCell";
import LinkCell from "components/table/cells/link/LinkCell";

export default function DefaultCellRenderer<T>(
  row: T,
  rowIndex: number,
  column: ColumnMetaData<T>,
  columnIndex: number,
  totalsRow?: boolean
): JSX.Element {
  const dataKey = column.key as keyof T;
  const dataKey2 = column.key2 as keyof T;
  const suffix = column.suffix as string;
  // const heading = column.heading;
  const percent = column.percent;

  switch (column.valueType) {
    case "string":
      switch (column.type) {
        case "AliasCell":
          return (
            <AliasCell
              current={row[dataKey] as string}
              key={dataKey.toString() + rowIndex}
              className={column.locked ? styles.locked : ""}
            />
          );
        case "LongTextCell":
          return (
            <LongTextCell
              current={row[dataKey] as string}
              key={dataKey.toString() + rowIndex}
              copyText={row[dataKey] as string}
              totalCell={totalsRow}
            />
          );
        case "TextCell":
          return (
            <TextCell current={row[dataKey] as string} key={dataKey.toString() + rowIndex} totalCell={totalsRow} />
          );
        case "DurationCell":
          return (
            <DurationCell seconds={row[dataKey] as number} key={dataKey.toString() + rowIndex} totalCell={totalsRow} />
          );
        case "EnumCell":
          return (
            <EnumCell
              value={row[dataKey] as string}
              key={dataKey.toString() + rowIndex + columnIndex}
              totalCell={totalsRow}
            />
          );
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
              totalCell={totalsRow}
            />
          );
      }
      break;
    case "date":
      return <DateCell value={row[dataKey] as Date} key={dataKey.toString() + rowIndex} totalCell={totalsRow} />;
      break;
    case "duration":
      return (
        <DurationCell seconds={row[dataKey] as number} key={dataKey.toString() + rowIndex} totalCell={totalsRow} />
      );
      break;
    case "link":
      return (
        <LinkCell
          text={row[dataKey] as string}
          link={row[dataKey] as string}
          key={dataKey.toString() + rowIndex}
          totalCell={totalsRow}
        />
      );
      break;
    case "number":
      switch (column.type) {
        case "NumericCell":
          return <NumericCell current={row[dataKey] as number} key={dataKey.toString() + rowIndex + columnIndex} />;
        case "BarCell":
          return (
            <BarCell
              current={row[dataKey] as number}
              total={column.max as number}
              showPercent={percent}
              key={dataKey.toString() + rowIndex + columnIndex}
            />
          );
        case "NumericDoubleCell":
          return (
            <NumericDoubleCell
              local={row[dataKey] as number}
              remote={row[dataKey2] as number}
              suffix={suffix as string}
              className={dataKey.toString()}
              key={dataKey.toString() + rowIndex + columnIndex}
              totalCell={totalsRow}
            />
          );
      }
  }
  return (
    <TextCell current={row[dataKey] as string} key={dataKey.toString() + rowIndex} copyText={row[dataKey] as string} />
  );
}
