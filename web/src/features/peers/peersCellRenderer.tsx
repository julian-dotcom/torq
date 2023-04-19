import { ColumnMetaData } from "features/table/types";
import { Peer } from "features/peers/peersTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import TextCell from "components/table/cells/text/TextCell";
import PeersAliasCell from "components/table/cells/peersCell/PeersAliasCell";

export default function peerCellRenderer(
  row: Peer,
  rowIndex: number,
  column: ColumnMetaData<Peer>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Peer
): JSX.Element {
  if (column.key === "peerAlias") {
    return (
      <PeersAliasCell
        key={column.key.toString() + rowIndex}
        alias={row.peerAlias}
        peerNodeId={row.nodeId}
        torqNodeId={row.torqNodeId}
        connectionStatus={row.connectionStatus}
      />
    );
  }

  if (column.key === "connectionStatus") {
    return <TextCell current={row.connectionStatus} key={column.key.toString() + rowIndex} totalCell={isTotalsRow} />;
  }

  return DefaultCellRenderer(row, rowIndex, column, columnIndex, false, maxRow);
}
