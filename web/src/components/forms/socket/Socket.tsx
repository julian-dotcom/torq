import useTranslations from "services/i18n/useTranslations";
import classNames from "classnames";
import { useContext, useEffect, useState } from "react";
import { WarningRegular as WarningIcon, ErrorCircleRegular as ErrorIcon } from "@fluentui/react-icons";
import styles from "./socket_input.module.scss";
import { GetColorClass, GetSizeClass, InputColorVaraint } from "components/forms/input/variants";
import { NodeContext } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { BasicInputType } from "components/forms/formTypes";
import { useAddNodeLinkMutation } from "pages/WorkflowPage/workflowApi";
import { WorkflowVersionNode } from "pages/WorkflowPage/workflowTypes";

export type SocketProps = BasicInputType & {
  id: string;
  workflowVersionId: number;
  workflowVersionNodeId: number;
  selectedNodes: Array<WorkflowVersionNode>;
  inputIndex: number;
  placeholder?: string;
};

function Socket<T>(props: SocketProps) {
  const { t } = useTranslations();
  const [addLink] = useAddNodeLinkMutation();

  let inputColorClass = GetColorClass(props.colorVariant);
  if (props.warningText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.warning);
  }
  if (props.errorText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.error);
  }

  const { nodeRef } = useContext(NodeContext);
  const [connectedNodeName, setConnectedNodeName] = useState<string>(
    props.selectedNodes.map((n) => n.name).toString() || ""
  );

  useEffect(() => {
    setConnectedNodeName(props.selectedNodes.map((n) => n.name).toString() || "");
  }, [props.selectedNodes]);

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    // Get the id of the nodes connector that was dropped
    const parentWorkflowVersionNodeId = parseInt(e.dataTransfer.getData("node/parentWorkflowVersionNodeId"));
    const parentOutputIndex = parseInt(e.dataTransfer.getData("node/parentOutputIndex"));
    const nodeName = e.dataTransfer.getData("node/name");

    if (nodeName) {
      // setConnectedNodeName(nodeName);

      addLink({
        workflowVersionId: props.workflowVersionId,
        childInputIndex: props.inputIndex,
        childWorkflowVersionNodeId: props.workflowVersionNodeId,
        parentOutputIndex: parentOutputIndex,
        parentWorkflowVersionNodeId: parentWorkflowVersionNodeId,
      });
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
