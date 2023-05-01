import React, { MutableRefObject, useContext, useRef, useState } from "react";
import styles from "./node_button_wrapper.module.scss";
import classNames from "classnames";
import { GetColorClass, NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { WorkflowContext } from "components/workflow/WorkflowContext";
import { Status } from "constants/backend";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";
import useTranslations from "services/i18n/useTranslations";

export type WorkflowNodeButtonProps = {
  title: string;
  icon?: React.ReactNode;
  nodeType: number;
  colorVariant: NodeColorVariant;
  parameters?: string;
  disabled?: boolean;
  intercomTarget?: string;
};

function WorkflowNodeButtonWrapper(props: WorkflowNodeButtonProps) {
  const { t } = useTranslations();
  const buttonRef = useRef() as MutableRefObject<HTMLDivElement>;

  const { workflowStatus } = useContext(WorkflowContext);
  const editingDisabled = workflowStatus === Status.Active;
  const toastRef = React.useContext(ToastContext);

  const [isDragging, setIsDragging] = useState(false);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    if (props.disabled) {
      return;
    }

    if (editingDisabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }

    setIsDragging(true);

    // Set the drag image to the button itself with the position of the mouse relative to the button
    e.dataTransfer.setDragImage(buttonRef.current, 0, 0);
    e.dataTransfer.effectAllowed = "move";

    e.dataTransfer.setData("node/type", props.nodeType.toString());
    e.dataTransfer.setData("node/event", "add");
    e.dataTransfer.setData("node/name", props.title);
    e.dataTransfer.setData("node/parameters", props.parameters || "{}");
    e.dataTransfer.setData("item", "nodeButton");
  }

  function handleDragEnd() {
    setIsDragging(false);
  }
  return (
    <div
      data-intercom-target={props.intercomTarget}
      className={classNames(
        styles.nodeButtonWrapper,
        { [styles.dragging]: isDragging },
        GetColorClass(props.colorVariant),
        { [styles.disabled]: props.disabled }
      )}
      draggable={true}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      ref={buttonRef}
    >
      <div className={styles.iconContainer}>{props.icon}</div>
      <div className={styles.titleContainer}>{props.title}</div>
    </div>
  );
}

export default WorkflowNodeButtonWrapper;
