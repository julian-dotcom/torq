import { ColumnMetaData } from "features/table/types";
import { Peer, PeerStatus } from "features/peers/peersTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import CellWrapper from "components/table/cells/cellWrapper/CellWrapper";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import mixpanel from "mixpanel-browser";
import useTranslations from "services/i18n/useTranslations";
import styles from "features/peers/peers.module.scss";
import { useDisconnectPeerMutation, useReconnectPeerMutation } from "features/peers/peersApi";
export default function peerCellRenderer(
  row: Peer,
  rowIndex: number,
  column: ColumnMetaData<Peer>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Peer
): JSX.Element {
  const { t } = useTranslations();
  const [disconnectNodeMutation] = useDisconnectPeerMutation();
  const [reconnectMutation] = useReconnectPeerMutation();

  if (column.key === "actions") {
    return (
      <CellWrapper cellWrapperClassName={styles.actionsWrapper} key={"connect-button-" + row.nodeId}>
        <Button
          disabled={row.status === PeerStatus.Active}
          buttonSize={SizeVariant.small}
          onClick={() => {
            mixpanel.track("Connect Peer", {
              tagId: row.nodeId,
            });
            reconnectMutation({ nodeId: row.nodeId, nodeConnectionDetailsNodeId: row.nodeConnectionDetailsNodeId });
          }}
          buttonColor={ColorVariant.success}
        >
          {t.peersPage.connect}
        </Button>
        <Button
          disabled={row.status === PeerStatus.Inactive}
          buttonSize={SizeVariant.small}
          onClick={() => {
            mixpanel.track("Disconnect Peer", {
              nodeId: row.nodeId,
            });
            disconnectNodeMutation({
              nodeId: row.nodeId,
              nodeConnectionDetailsNodeId: row.nodeConnectionDetailsNodeId,
            });
          }}
          buttonColor={ColorVariant.error}
        >
          {t.peersPage.disconnect}
        </Button>
      </CellWrapper>
    );
  }

  return DefaultCellRenderer(row, rowIndex, column, columnIndex, false, maxRow);
}
