import styles from "components/table/cells/cell.module.scss";
import { ColumnMetaData } from "features/table/types";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import AliasCell from "components/table/cells/alias/AliasCell";
import { Htlc } from "features/htlcs/htlcsTypes";
import TagsCell from "components/table/cells/tags/TagsCell";

export default function channelsCellRenderer(
  row: Htlc,
  rowIndex: number,
  column: ColumnMetaData<Htlc>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Htlc
): JSX.Element {
  if (column.key === "incomingAlias") {
    return (
      <AliasCell
        current={row["incomingAlias"] as string}
        channelId={row.incomingChannelId}
        nodeIds={[row.incomingNodeId]}
        key={"incomingAlias" + rowIndex + columnIndex}
        className={column.locked ? styles.locked : ""}
        isTotalsRow={isTotalsRow}
      />
    );
  }
  if (column.key === "outgoingAlias") {
    return (
      <AliasCell
        current={row["outgoingAlias"] as string}
        channelId={row.outgoingChannelId}
        nodeIds={[row.outgoingNodeId]}
        key={"outgoingAlias" + rowIndex + columnIndex}
        className={column.locked ? styles.locked : ""}
        isTotalsRow={isTotalsRow}
      />
    );
  }

  if (column.key === "incomingChannelTags") {
    return (
      <TagsCell
        tags={row.incomingChannelTags || []}
        key={"incomingChannelTags" + rowIndex + columnIndex}
        channelId={row.incomingChannelId}
        nodeId={row.incomingNodeId}
        totalCell={isTotalsRow}
      />
    );
  }
  if (column.key === "outgoingChannelTags") {
    return (
      <TagsCell
        tags={row.outgoingChannelTags || []}
        key={"outgoingChannelTags" + rowIndex + columnIndex}
        channelId={row.outgoingChannelId}
        nodeId={row.outgoingNodeId}
        totalCell={isTotalsRow}
      />
    );
  }

  // Use the defualt
  return DefaultCellRenderer(row, rowIndex, column, columnIndex, isTotalsRow, maxRow);
}
