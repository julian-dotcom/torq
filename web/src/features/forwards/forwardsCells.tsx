import styles from "components/table/cells/cell.module.scss";
import { ColumnMetaData } from "../table/types";
import DefaultCellRenderer from "../table/DefaultCellRenderer";
import AliasCell from "../../components/table/cells/alias/AliasCell";
import { Forward } from "./forwardsTypes";

export default function channelsCellRenderer(
  row: Forward,
  rowIndex: number,
  column: ColumnMetaData<Forward>,
  columnIndex: number,
  isTotalsRow?: boolean
): JSX.Element {
  if (column.key === "alias") {
    return (
      <AliasCell
        current={row["alias"] as string}
        channelId={row["lndShortChannelId"]}
        open={row["open"]}
        key={"alias" + rowIndex + columnIndex}
        className={column.locked ? styles.locked : ""}
      />
    );
  }

  // Use the defualt
  return DefaultCellRenderer(row, rowIndex, column, columnIndex, isTotalsRow);
}
