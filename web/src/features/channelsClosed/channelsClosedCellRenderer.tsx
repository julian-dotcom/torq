import cellStyles from "components/table/cells/cell.module.scss";
import { ColumnMetaData } from "features/table/types";
import { ChannelClosed } from "features/channelsClosed/channelsClosedTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import ChannelCell from "components/table/cells/channelCell/ChannelCell";
import LongTextCell from "components/table/cells/longText/LongTextCell";
export default function channelsClosedCellRenderer(
  row: ChannelClosed,
  rowIndex: number,
  column: ColumnMetaData<ChannelClosed>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: ChannelClosed
): JSX.Element {
  switch (column.key) {
    case "peerAlias":
      return (
        <ChannelCell
          alias={row.peerAlias}
          open={true}
          channelId={row.channelId}
          nodeId={row.nodeId}
          className={cellStyles.locked}
          key={"channelsCell" + rowIndex}
          hideActionButtons
        />
      );
    case "fundingTransactionHash":
      if (column.type === "LongTextCell") {
        return (
          <LongTextCell
            key={"fundingTransactionHashCell" + rowIndex}
            current={row.fundingTransactionHash}
            link={"https://mempool.space/tx/" + row.fundingTransactionHash}
            copyText={row.fundingTransactionHash}
          />
        );
      }
      break;
    case "closingTransactionHash":
      if (column.type === "LongTextCell") {
        return (
          <LongTextCell
            key={"closingTransactionHashCell" + rowIndex}
            current={row.closingTransactionHash}
            link={"https://mempool.space/tx/" + row.closingTransactionHash}
            copyText={row.closingTransactionHash}
          />
        );
      }
      break;
  }

  return DefaultCellRenderer(row, rowIndex, column, columnIndex, false, maxRow);
}
