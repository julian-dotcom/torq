import { ColumnMetaData } from "features/table/types";
import { Peer, PeerStatus } from "features/peers/peersTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import CellWrapper from "components/table/cells/cellWrapper/CellWrapper";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import mixpanel from "mixpanel-browser";
import useTranslations from "services/i18n/useTranslations";
import styles from "features/peers/peers.module.scss";
import { useDisconnectPeerMutation, useReconnectPeerMutation } from "features/peers/peersApi";
import React from "react";
import ToastContext from "features/toast/context";
import { toastCategory } from "../toast/Toasts";
import { mergeServerError, ServerErrorType } from "components/errors/errors";

export default function peerCellRenderer(
  row: Peer,
  rowIndex: number,
  column: ColumnMetaData<Peer>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Peer
): JSX.Element {
  const { t } = useTranslations();
  const [disconnectNodeMutation, { error: disconnectError }] = useDisconnectPeerMutation();
  const [reconnectMutation, { error: reconnectError }] = useReconnectPeerMutation();
  const toastRef = React.useContext(ToastContext);

  React.useEffect(() => {
    if (disconnectError && "data" in disconnectError && disconnectError.data) {
      const err = mergeServerError(disconnectError.data as ServerErrorType, {});
      if (err?.server && err.server.length > 0) {
        toastRef?.current?.addToast(err.server[0].description || "", toastCategory.error);
      }
    }
  }, [disconnectError]);

  React.useEffect(() => {
    if (reconnectError && "data" in reconnectError && reconnectError.data) {
      const err = mergeServerError(reconnectError.data as ServerErrorType, {});
      if (err?.server && err.server.length > 0) {
        toastRef?.current?.addToast(err.server[0].description || "", toastCategory.error);
      }
    }
  }, [reconnectError]);

  if (column.key === "actions") {
    return (
      <CellWrapper cellWrapperClassName={styles.actionsWrapper} key={"connect-button-" + row.nodeId}>
        <Button
          disabled={row.connectionStatus === PeerStatus.Active}
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
          disabled={row.connectionStatus === PeerStatus.Inactive}
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
