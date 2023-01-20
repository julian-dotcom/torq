import classNames from "classnames";
import { MutableRefObject, useContext, useEffect, useRef, useState } from "react";
import { WarningRegular as WarningIcon, ErrorCircleRegular as ErrorIcon } from "@fluentui/react-icons";
import styles from "./socket_input.module.scss";
import { GetColorClass, GetSizeClass, InputColorVaraint } from "components/forms/input/variants";
import { BasicInputType } from "components/forms/formTypes";
import { useAddNodeLinkMutation } from "pages/WorkflowPage/workflowApi";
import { WorkflowVersionNode } from "pages/WorkflowPage/workflowTypes";
import { CanvasContext } from "components/workflow/canvas/WorkflowCanvas";
import { NodeContext } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";
import useTranslations from "services/i18n/useTranslations";

export type SocketProps = BasicInputType & {
  workflowVersionId: number;
  workflowVersionNodeId: number;
  selectedNodes: Array<WorkflowVersionNode>;
  inputIndex: number;
  collapsed: boolean;
  placeholder?: string;
  acceptMultiple?: boolean;
};

function Socket(props: SocketProps) {
  const { t } = useTranslations();
  const [addLink] = useAddNodeLinkMutation();
  const toastRef = useContext(ToastContext);

  let inputColorClass = GetColorClass(props.colorVariant);
  if (props.warningText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.warning);
  }
  if (props.errorText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.error);
  }

  const { nodeRef } = useContext(NodeContext);
  const { canvasRef } = useContext(CanvasContext);
  const connectorRef = useRef() as MutableRefObject<HTMLDivElement>;

  const [isDragover, setIsDragover] = useState(false);

  // Handle drag enter event on the socket by setting setIsDragover to true
  function handleDragEnter(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    setIsDragover(true);
  }

  // Handle drag leave event on the socket by setting setIsDragover to false
  function handleDragLeave(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    setIsDragover(false);
  }

  function addLinkFromDrop(parentOutputIndex: number, parentWorkflowVersionNodeId: number) {
    addLink({
      workflowVersionId: props.workflowVersionId,
      childInputIndex: props.inputIndex,
      childWorkflowVersionNodeId: props.workflowVersionNodeId,
      parentOutputIndex: parentOutputIndex,
      parentWorkflowVersionNodeId: parentWorkflowVersionNodeId,
    });
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    setIsDragover(false);
    // Get the id of the nodes connector that was dropped
    const parentWorkflowVersionNodeId = parseInt(e.dataTransfer.getData("node/parentWorkflowVersionNodeId"));
    const parentOutputIndex = parseInt(e.dataTransfer.getData("node/parentOutputIndex"));
    if (props.selectedNodes?.length > 0 && !props.acceptMultiple) {
      toastRef?.current?.addToast(props.label + " " + t.socketCantAcceptMultipleInputs, toastCategory.error);
      return;
    }
    addLinkFromDrop(parentOutputIndex, parentWorkflowVersionNodeId);
  }

  function updater() {
    if (canvasRef !== null) {
      const connBB = connectorRef?.current?.getBoundingClientRect() || { left: 0, top: 0 };
      const canvasBB = canvasRef?.current?.getBoundingClientRect() || { left: 0, top: 0 };
      const x = connBB.x - canvasBB.x + connBB.width / 2 - 8; // -14 because of the 16 padding right on the connector and 4px line width
      const y = connBB.y - canvasBB.y + connBB.height / 2 - 12.5;
      const eventName = `childLinkMove-${props.workflowVersionNodeId}-${props.inputIndex}`;
      const event = new CustomEvent(eventName, {
        detail: {
          x: x,
          y: y,
          nodeId: props.workflowVersionNodeId.toString(),
        },
      });
      window.dispatchEvent(event);
    }
  }

  function updatePosition(mutations: MutationRecord[]) {
    mutations.forEach(function () {
      updater();
    });
  }

  // Add a listener to the node card to update the position of the connector when the node is moved.
  useEffect(() => {
    const observer = new MutationObserver(updatePosition);

    if (nodeRef !== null && nodeRef.current !== undefined) {
      observer.observe(nodeRef.current, { attributes: true, attributeFilter: ["style"] });
    }

    return () => {
      if (nodeRef !== null && nodeRef.current !== undefined) {
        observer.disconnect();
      }
    };
  }, [canvasRef?.current, nodeRef?.current, connectorRef?.current]);

  useEffect(() => {
    const interval = setInterval(() => {
      updater();
    }, 10);
    return () => clearInterval(interval);
  });

  return (
    <div
      className={classNames(styles.socketInputWrapper, inputColorClass, { [styles.dragOver]: isDragover })}
      onDragOver={handleDragEnter}
      onDragLeave={handleDragLeave}
    >
      {props.label && (
        <div className={styles.labelWrapper}>
          <label className={styles.label}>{props.label}</label>
        </div>
      )}
      <div
        className={classNames(styles.socketInputContainer, GetSizeClass(props.sizeVariant), {
          [styles.collapsed]: props.collapsed,
        })}
        onDrop={handleDrop}
      >
        <div className={classNames(styles.nodeSocket, styles.socket)} ref={connectorRef}>
          <div className={styles.socketDot} />
        </div>
        <div className={styles.connectedNodeNames}>
          {props.selectedNodes.map((n, index) => {
            return (
              <span className={styles.connectionTag} key={n.workflowVersionNodeId + "-" + index}>
                {n.name}
              </span>
            );
          }) || "Drag connection here"}
        </div>
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
