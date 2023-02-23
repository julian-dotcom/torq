import cellStyles from "components/table/cells/cell.module.scss";
import { ColumnMetaData } from "features/table/types";
import { ChannelClosed } from "features/channelsClosed/channelsClosedTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import ChannelCell from "components/table/cells/channelCell/ChannelCell";
import TagsCell from "components/table/cells/tags/TagsCell";
import LinkCell from "components/table/cells/link/LinkCell";
import TextCell from "components/table/cells/text/TextCell";
import { ChannelStatus } from "constants/backend";
import { GetFormattedEnumLabelByValue } from "utils/stringUtils";

const links = new Map([
  ["mempoolSpace", "Mempool"],
  ["ambossSpace", "Amboss"],
  ["oneMl", "1ML"],
]);

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
    case "tags":
      return <TagsCell tags={row.tags} key={"tagsCell" + rowIndex} channelId={row.channelId} nodeId={row.peerNodeId} />;
    case "mempoolSpace":
    case "ambossSpace":
    case "oneMl":
      return (
        <LinkCell
          text={links.get(column.key) || ""}
          link={row[column.key] as string}
          key={column.key + rowIndex}
          totalCell={isTotalsRow}
        />
      );
    case "status": {
      const enumLabel = GetFormattedEnumLabelByValue(ChannelStatus, row[column.key]);
      return <TextCell current={enumLabel} key={column.key + rowIndex} copyText={enumLabel} />;
    }
    default:
      return DefaultCellRenderer(row, rowIndex, column, columnIndex, false, maxRow);
  }
}
