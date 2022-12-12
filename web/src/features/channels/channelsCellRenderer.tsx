import cellStyles from "components/table/cells/cell.module.scss";
import { ColumnMetaData } from "features/table/types";
import BalanceCell from "components/table/cells/balance/BalanceCell";
import { channel } from "./channelsTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import ChannelCell from "components/table/cells/channelCell/ChannelCell";

export default function channelsCellRenderer(
  row: channel,
  rowIndex: number,
  column: ColumnMetaData<channel>,
  columnIndex: number
): JSX.Element {
  if (column.key === "peerAlias") {
    return (
      <ChannelCell
        alias={row.peerAlias}
        open={true}
        channelId={row.channelId}
        nodeId={row.nodeId}
        className={cellStyles.locked}
        key={"channelsCell" + rowIndex}
      />
    );
  }

  if (column.type === "BalanceCell") {
    return (
      <BalanceCell
        capacity={row.capacity}
        local={row.localBalance}
        remote={row.remoteBalance}
        key={"balanceCell" + rowIndex}
      />
    );
  }

  // Use the defualt
  return DefaultCellRenderer(row, rowIndex, column, columnIndex);
}
