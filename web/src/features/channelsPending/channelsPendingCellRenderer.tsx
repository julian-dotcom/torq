import cellStyles from "components/table/cells/cell.module.scss";
import { ColumnMetaData } from "features/table/types";
import { ChannelPending } from "features/channelsPending/channelsPendingTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import ChannelCell from "components/table/cells/channelCell/ChannelCell";
import TagsCell from "components/table/cells/tags/TagsCell";
import LongTextCell from "../../components/table/cells/longText/LongTextCell";

export default function channelsPendingCellRenderer(
  row: ChannelPending,
  rowIndex: number,
  column: ColumnMetaData<ChannelPending>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: ChannelPending
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
    case "tags":
      return <TagsCell tags={row.tags} key={"tagsCell" + rowIndex} channelId={row.channelId} nodeId={row.peerNodeId} />;
    case "fundingTransactionHash":
      if (column.type === "LongTextCell") {
        return (
          <LongTextCell
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
