import React, { MutableRefObject, useRef, useState } from "react";
import styles from "./node_button_wrapper.module.scss";
import classNames from "classnames";
import { GetColorClass, NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import mixpanel from "mixpanel-browser";

export type WorkflowNodeButtonProps = {
  title: string;
  icon?: React.ReactNode;
  nodeType: number;
  colorVariant: NodeColorVariant;
  parameters?: string;
  disabled?: boolean;
};

function WorkflowNodeButtonWrapper(props: WorkflowNodeButtonProps) {
  const buttonRef = useRef() as MutableRefObject<HTMLDivElement>;

  const [isDragging, setIsDragging] = useState(false);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    if (props.disabled) {
      return;
    }
    setIsDragging(true);

    // Set the drag image to the button itself with the position of the mouse relative to the button
    e.dataTransfer.setDragImage(buttonRef.current, 0, 0);
    e.dataTransfer.effectAllowed = "copy";

    e.dataTransfer.setData("node/type", props.nodeType.toString());
    e.dataTransfer.setData("node/event", "add");
    e.dataTransfer.setData("node/name", props.title);
    e.dataTransfer.setData("node/parameters", props.parameters || "{}");
  }

  function handleDragEnd() {
    setIsDragging(false);
    mixpanel.track("Workflow Node added", {
      nodeType: props.nodeType,
      nodeTitle: props.title,
    });
  }
  return (
    <div
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
