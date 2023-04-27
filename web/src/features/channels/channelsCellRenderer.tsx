import cellStyles from "components/table/cells/cell.module.scss";
import { ColumnMetaData } from "features/table/types";
import BalanceCell from "components/table/cells/balance/BalanceCell";
import { channel } from "./channelsTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import ChannelCell from "components/table/cells/channelCell/ChannelCell";
import TagsCell from "components/table/cells/tags/TagsCell";
import LinkCell from "components/table/cells/link/LinkCell";
import TextCell from "components/table/cells/text/TextCell";

const MEMPOOL_SPACE = "mempoolSpace";
const AMBOSS_SPACE = "ambossSpace";
const ONE_ML = "oneMl";

const links = new Map([
  [MEMPOOL_SPACE, "Mempool"],
  [AMBOSS_SPACE, "Amboss"],
  [ONE_ML, "1ML"],
]);

export default function channelsCellRenderer(
  row: channel,
  rowIndex: number,
  column: ColumnMetaData<channel>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: channel
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
        hideActionButtons={false}
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

  if (column.key === "tags") {
    return <TagsCell tags={row.tags} key={"tagsCell" + rowIndex} channelId={row.channelId} nodeId={row.peerNodeId} />;
  }

  if ([MEMPOOL_SPACE, AMBOSS_SPACE, ONE_ML].includes(column.key)) {
    if (row.private)
      return (
        <TextCell
          key={column.key + rowIndex}
          className={cellStyles.cell}
          current={"Channel private, links unavailable"}
        />
      );

    return (
      <LinkCell
        text={links.get(column.key) || ""}
        link={row[column.key] as string}
        key={column.key + rowIndex}
        totalCell={isTotalsRow}
      />
    );
  }

  // Use the defualt
  return DefaultCellRenderer(row, rowIndex, column, columnIndex, false, maxRow);
}
