import { ColumnMetaData } from "features/table/types";
import { ConnectionStatus, NodeConnectionSetting, Peer } from "features/peers/peersTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import CellWrapper from "components/table/cells/cellWrapper/CellWrapper";
import Button, { ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import mixpanel from "mixpanel-browser";
import useTranslations from "services/i18n/useTranslations";
import styles from "features/peers/peers.module.scss";
import { useDisconnectPeerMutation, useReconnectPeerMutation } from "features/peers/peersApi";
import React from "react";
import ToastContext from "features/toast/context";
import { toastCategory } from "../toast/Toasts";
import { mergeServerError, ServerErrorType } from "components/errors/errors";
import TextCell from "components/table/cells/text/TextCell";
import * as Routes from "constants/routes";
import { useLocation } from "react-router";

export default function peerCellRenderer(
  row: Peer,
  rowIndex: number,
  column: ColumnMetaData<Peer>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Peer
): JSX.Element {
  const { t } = useTranslations();
  const [disconnectNodeMutation, { error: disconnectError, isLoading: disconnectIsLoading }] =
    useDisconnectPeerMutation();
  const [reconnectMutation, { error: reconnectError, isLoading: reconnectIsLoading }] = useReconnectPeerMutation();
  const toastRef = React.useContext(ToastContext);
  const location = useLocation();

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

  if (column.key === "connectionStatus") {
    return (
      <TextCell
        current={t.peersPage[ConnectionStatus[row.connectionStatus]]}
        key={column.key.toString() + rowIndex}
        totalCell={isTotalsRow}
      />
    );
  }

  if (column.key === "setting") {
    return (
      <TextCell
        current={t.peersPage[NodeConnectionSetting[row.setting]]}
        key={column.key.toString() + rowIndex}
        totalCell={isTotalsRow}
      />
    );
  }

  if (column.key === "actions") {
    return (
      <CellWrapper cellWrapperClassName={styles.actionsWrapper} key={"connect-button-" + row.nodeId}>
        <Button
          disabled={row.connectionStatus === ConnectionStatus.Connected || reconnectIsLoading || disconnectIsLoading}
          buttonSize={SizeVariant.small}
          onClick={() => {
            mixpanel.track("Connect Peer", {
              tagId: row.nodeId,
            });
            reconnectMutation({ nodeId: row.nodeId, torqNodeId: row.torqNodeId });
          }}
          buttonColor={ColorVariant.success}
        >
          {t.peersPage.connect}
        </Button>
        <Button
          disabled={row.connectionStatus === ConnectionStatus.Disconnected || reconnectIsLoading || disconnectIsLoading}
          buttonSize={SizeVariant.small}
          onClick={() => {
            mixpanel.track("Disconnect Peer", {
              nodeId: row.nodeId,
            });
            disconnectNodeMutation({
              nodeId: row.nodeId,
              torqNodeId: row.torqNodeId,
            });
          }}
          buttonColor={ColorVariant.error}
        >
          {t.peersPage.disconnect}
        </Button>

        <LinkButton
          to={`${Routes.UPDATE_PEER}?torqNodeId=${row.torqNodeId}&peerNodeId=${row.nodeId}`}
          state={{ background: location }}
          hideMobileText={true}
          buttonColor={ColorVariant.primary}
          buttonSize={SizeVariant.small}
          onClick={() => {
            mixpanel.track("Navigate to Update Peer", {
              nodeId: row.nodeId,
            });
          }}
        >
          {t.update}
        </LinkButton>
      </CellWrapper>
    );
  }

  return DefaultCellRenderer(row, rowIndex, column, columnIndex, false, maxRow);
}
