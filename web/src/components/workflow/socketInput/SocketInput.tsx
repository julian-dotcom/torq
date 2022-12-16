import useTranslations from "services/i18n/useTranslations";
import { useContext, useState } from "react";
import styles from "./socket_input.module.scss";
import { NodeContext } from "../nodeWrapper/WorkflowNodeWrapper";
import classNames from "classnames";

export type SocketInputProps = {
  id: string;
  connectedNodeName: string;
};

function SocketInput<T>(props: SocketInputProps) {
  const { t } = useTranslations();
  const { nodeRef } = useContext(NodeContext);

  const [connectedNodeName, setConnectedNodeName] = useState<string>(props.connectedNodeName);

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    // Get the id of the nodes connector that was dropped
    const nodeid = e.dataTransfer.getData("node/id");
    const connectorId = e.dataTransfer.getData("node/connectorId");
    const nodeName = e.dataTransfer.getData("node/name");
    // e.dataTransfer.clearData("node/id");
    if (nodeName) {
      console.log(
        "Dropped from node: " +
          nodeid +
          " with connector: " +
          connectorId +
          " on node: " +
          nodeRef?.current?.id +
          " with socket: " +
          props.id
      );
      setConnectedNodeName(nodeName);
    }
  }

  return (
    <div className={styles.socketInputWrapper} onDrop={handleDrop}>
      <div className={classNames(styles.nodeSocket, styles.socket)}>
        <div className={styles.socketDot} />
      </div>
      <div className={styles.connectedNodeName}>{connectedNodeName}</div>
    </div>
  );
}

export default SocketInput;
