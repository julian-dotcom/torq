import classNames from "classnames";
import React, {MutableRefObject, useContext, useEffect, useRef, useState} from "react";
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
import mixpanel from "mixpanel-browser";
import NodeConnector from "components/workflow/nodeWrapper/NodeConnector";

export type SocketProps = BasicInputType & {
  workflowVersionId: number;
  workflowVersionNodeId: number;
  selectedNodes: Array<WorkflowVersionNode>;
  inputName: string;
  outputName?: string;
  collapsed: boolean;
  placeholder?: string;
  acceptMultiple?: boolean;
  editingDisabled: boolean;
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

  function addLinkFromDrop(parentOutput: string, parentWorkflowVersionNodeId: number) {
    mixpanel.track("Workflow Link Node", {
      workflowVersionId: props.workflowVersionId,
      workflowChildInput: props.inputName,
      workflowChildWorkflowVersionNodeId: props.workflowVersionNodeId,
      workflowParentOutput: parentOutput,
      workflowParentWorkflowVersionNodeId: parentWorkflowVersionNodeId,
    });
    addLink({
      workflowVersionId: props.workflowVersionId,
      childInput: props.inputName,
      childWorkflowVersionNodeId: props.workflowVersionNodeId,
      parentOutput: parentOutput,
      parentWorkflowVersionNodeId: parentWorkflowVersionNodeId,
    });
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    setIsDragover(false);

    if (props.editingDisabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }

    // Get the id of the nodes connector that was dropped
    const parentWorkflowVersionNodeId = parseInt(e.dataTransfer.getData("node/parentWorkflowVersionNodeId"));
    const parentOutput = e.dataTransfer.getData("node/parentOutput");
    if (props.selectedNodes?.length > 0 && !props.acceptMultiple) {
      toastRef?.current?.addToast(props.label + " " + t.socketCantAcceptMultipleInputs, toastCategory.error);
      return;
    }
    addLinkFromDrop(parentOutput, parentWorkflowVersionNodeId);
  }

  function updater() {
    if (canvasRef !== null) {
      const canvasBB = canvasRef?.current?.getBoundingClientRect() || { left: 0, top: 0 };
      const connBB = connectorRef?.current?.getBoundingClientRect() || { left: 0, top: 0 };
      const x = connBB.x - canvasBB.x + connBB.width / 2 - 8; // -14 because of the 16 padding right on the connector and 4px line width
      const y = connBB.y - canvasBB.y + connBB.height / 2 - 12.5;
      const eventNameInput = `childLinkMove-${props.workflowVersionNodeId}-${props.inputName}`;
      const eventInput = new CustomEvent(eventNameInput, {
        detail: {
          x: x,
          y: y,
          nodeId: props.workflowVersionNodeId.toString(),
        },
      });
      window.dispatchEvent(eventInput);
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
          {props.selectedNodes.length > 0 &&
            props.selectedNodes.map((n, index) => {
              return (
                <span className={styles.connectionTag} key={n.workflowVersionNodeId + "-" + index}>
                  {n.name}
                </span>
              );
            })}
          {props.selectedNodes.length === 0 && "Drop connection here"}
        </div>
        {props.outputName && (
          <NodeConnector
            id={`parentLinkConnector-${props.workflowVersionNodeId}-${props.outputName}`}
            name={props.label || ""}
            outputName={props.outputName || ""}
            workflowVersionNodeId={props.workflowVersionNodeId}
            workflowVersionId={props.workflowVersionId}
          />
        )}
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
