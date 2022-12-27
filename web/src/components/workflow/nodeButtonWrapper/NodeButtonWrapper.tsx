import React, { MutableRefObject, useRef, useState } from "react";
import styles from "./node_button_wrapper.module.scss";
import classNames from "classnames";

export type WorkflowNodeButtonProps = {
  icon?: React.ReactNode;
  title: string;
  nodeType: number;
};

function WorkflowNodeButtonWrapper<T>(props: WorkflowNodeButtonProps) {
  const buttonRef = useRef() as MutableRefObject<HTMLDivElement>;

  const [isDragging, setIsDragging] = useState(false);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    setIsDragging(true);

    // Set the drag image to the button itself with the position of the mouse relative to the button
    e.dataTransfer.setDragImage(buttonRef.current, 0, 0);
    e.dataTransfer.effectAllowed = "copy";

    e.dataTransfer.setData("node/type", props.nodeType.toString());
    e.dataTransfer.setData("node/event", "add");
  }

  function handleDragEnd(e: React.DragEvent<HTMLDivElement>) {
    setIsDragging(false);
  }
  return (
    <div
      className={classNames(styles.nodeButtonWrapper, { [styles.dragging]: isDragging })}
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
