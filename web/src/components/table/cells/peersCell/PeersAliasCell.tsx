import { ConnectionStatus } from "features/peers/peersTypes";
import * as Routes from "constants/routes";
import Spinny from "features/spinny/Spinny";
import { Link16Regular as ConnectedIcon } from "@fluentui/react-icons";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";
import { mergeServerError, ServerErrorType } from "components/errors/errors";
import { useDisconnectPeerMutation, useReconnectPeerMutation } from "features/peers/peersApi";
import styles from "components/table/cells/cell.module.scss";
import { useLocation } from "react-router-dom";
import classNames from "classnames";
import useTranslations from "services/i18n/useTranslations";
import { useContext, useEffect } from "react";
import Button, { ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import { userEvents } from "utils/userEvents";

interface ChannelCell {
  alias: string;
  peerNodeId: number;
  torqNodeId: number;
  connectionStatus: ConnectionStatus;
  className?: string;
}

function ChannelCell(props: ChannelCell) {
  const { t } = useTranslations();
  const { track } = userEvents();
  const location = useLocation();
  const [disconnectNodeMutation, { error: disconnectError, isLoading: disconnectIsLoading }] =
    useDisconnectPeerMutation();
  const [reconnectMutation, { error: reconnectError, isLoading: reconnectIsLoading }] = useReconnectPeerMutation();

  const toastRef = useContext(ToastContext);

  useEffect(() => {
    if (disconnectError && "data" in disconnectError && disconnectError.data) {
      const err = mergeServerError(disconnectError.data as ServerErrorType, {});
      if (err?.server && err.server.length > 0) {
        toastRef?.current?.addToast(err.server[0].description || "", toastCategory.error);
      }
    }
  }, [disconnectError]);

  useEffect(() => {
    if (reconnectError && "data" in reconnectError && reconnectError.data) {
      const err = mergeServerError(reconnectError.data as ServerErrorType, {});
      if (err?.server && err.server.length > 0) {
        toastRef?.current?.addToast(err.server[0].description || "", toastCategory.error);
      }
    }
  }, [reconnectError]);

  const content = (
    <>
      <div className={classNames(styles.current, styles.text)}>{props.alias}</div>
      <div className={styles.actionButtons}>
        {props.connectionStatus.toString() === ConnectionStatus.Disconnected && (
          <Button
            disabled={reconnectIsLoading || disconnectIsLoading}
            buttonSize={SizeVariant.tiny}
            onClick={() => {
              track("Connect Peer", {
                peerNodeId: props.peerNodeId,
                torqNodeId: props.torqNodeId,
              });
              reconnectMutation({ nodeId: props.peerNodeId, torqNodeId: props.torqNodeId });
            }}
            icon={reconnectIsLoading ? <Spinny /> : <ConnectedIcon />}
            buttonColor={ColorVariant.success}
          >
            {t.peersPage.connect}
          </Button>
        )}
        {props.connectionStatus.toString() === ConnectionStatus.Connected && (
          <Button
            disabled={reconnectIsLoading || disconnectIsLoading}
            buttonSize={SizeVariant.tiny}
            onClick={() => {
              track("Disconnect Peer", {
                peerNodeId: props.peerNodeId,
                torqNodeId: props.torqNodeId,
              });
              disconnectNodeMutation({
                nodeId: props.peerNodeId,
                torqNodeId: props.torqNodeId,
              });
            }}
            icon={disconnectIsLoading ? <Spinny /> : <ConnectedIcon />}
            buttonColor={ColorVariant.error}
          >
            {t.peersPage.disconnect}
          </Button>
        )}

        <LinkButton
          to={`${Routes.UPDATE_PEER}?torqNodeId=${props.torqNodeId}&peerNodeId=${props.peerNodeId}`}
          state={{ background: location }}
          hideMobileText={true}
          buttonColor={ColorVariant.accent1}
          buttonSize={SizeVariant.tiny}
          onClick={() => {
            track("Navigate to Update Peer", {
              peerNodeId: props.peerNodeId,
              torqNodeId: props.torqNodeId,
            });
          }}
        >
          {t.update}
        </LinkButton>
      </div>
    </>
  );

  return (
    <div className={classNames(styles.cell, styles.alignLeft, props.className, styles.channelCellWrapper)}>
      {content}
    </div>
  );
}
export default ChannelCell;
