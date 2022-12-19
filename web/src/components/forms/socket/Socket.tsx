import useTranslations from "services/i18n/useTranslations";
import classNames from "classnames";
import { useContext, useState } from "react";
import { WarningRegular as WarningIcon, ErrorCircleRegular as ErrorIcon } from "@fluentui/react-icons";
import styles from "./socket_input.module.scss";
import { GetColorClass, GetSizeClass, InputColorVaraint } from "components/forms/input/variants";
import { NodeContext } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { BasicInputType } from "components/forms/formTypes";

export type SocketProps = BasicInputType & {
  id: string;
  connectedNodeName?: string;
  placeholder?: string;
};

function Socket<T>(props: SocketProps) {
  const { t } = useTranslations();

  let inputColorClass = GetColorClass(props.colorVariant);
  if (props.warningText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.warning);
  }
  if (props.errorText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.error);
  }

  const { nodeRef } = useContext(NodeContext);
  const [connectedNodeName, setConnectedNodeName] = useState<string>(props.connectedNodeName || "");

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
    <div className={classNames(styles.socketInputWrapper, inputColorClass)}>
      {props.label && (
        <div className={styles.labelWrapper}>
          <label className={styles.label}>{props.label}</label>
        </div>
      )}
      <div className={classNames(styles.socketInputContainer, GetSizeClass(props.sizeVariant))} onDrop={handleDrop}>
        <div className={classNames(styles.nodeSocket, styles.socket)}>
          <div className={styles.socketDot} />
        </div>
        <div className={styles.connectedNodeName}>{connectedNodeName || props.placeholder}</div>
      </div>
      {props.errorText && (
        <div className={classNames(styles.feedbackWrapper, styles.feedbackError)}>
          <div className={styles.feedbackIcon}>
            <ErrorIcon />
          </div>
          <div className={styles.feedbackText}>{props.errorText}</div>
        </div>
      )}
      {props.warningText && (
        <div className={classNames(styles.feedbackWrapper, styles.feedbackWarning)}>
          <div className={styles.feedbackIcon}>
            <WarningIcon />
          </div>
          <div className={styles.feedbackText}>{props.warningText}</div>
        </div>
      )}
    </div>
  );
}

export default Socket;
