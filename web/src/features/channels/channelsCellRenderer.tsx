import {ColumnMetaData} from "../table/types";
import BalanceCell from "components/table/cells/balance/BalanceCell";
import {channel} from "./channelsTypes";
import DefaultCellRenderer from "../table/DefaultCellRenderer";

export default function channelsCellRenderer(
  row: channel,
  rowIndex: number,
  column: ColumnMetaData<channel>,
  columnIndex: number,
): JSX.Element {

  if (column.type === "BalanceCell") {
    return <BalanceCell capacity={row.capacity} local={row.localBalance} remote={row.remoteBalance} key={"balanceCell" + rowIndex} />
  }

  // Use the defualt
  return DefaultCellRenderer(row, rowIndex, column, columnIndex);
}
