import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_nodes.module.scss";
import { MutableRefObject, useRef, useState } from "react";
import classNames from "classnames";

type WorkflowNodeProps = {
  heading: string;
  canvasRef: MutableRefObject<HTMLDivElement>;
  blankImageRef: MutableRefObject<HTMLCanvasElement>;
  children?: React.ReactNode;
};

function WorkflowNodeWrapper<T>(props: WorkflowNodeProps) {
  const { t } = useTranslations();
  const [position, setPosition] = useState({ x: 0, y: 0 });

  const itemRef = useRef() as MutableRefObject<HTMLDivElement>;
  const [nodeBB, setNodeBB] = useState({ left: 0, top: 0 });
  const [isDragging, setIsDragging] = useState(false);

  // Add a blank canvas to the DOM in order to hide the html5 drag and drop image functionality.

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    const nodeBB = itemRef.current.getBoundingClientRect();
    e.dataTransfer.effectAllowed = "move";
    const x = e.clientX - nodeBB.left;
    const y = e.clientY - nodeBB.top;
    e.dataTransfer.setDragImage(props.blankImageRef.current, x, y);
    setIsDragging(true);
    setNodeBB({ left: x, top: y });
  }

  function handleDrag(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    const bb = props.canvasRef.current.getBoundingClientRect();
    if (e.clientX !== 0 && e.clientY !== 0) {
      const newX = e.clientX - bb.x - nodeBB.left;
      const newY = e.clientY - bb.y - nodeBB.top;
      setPosition({ x: newX, y: newY });
    }
  }

  function handleDragEnd(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }

  return (
    <div
      className={classNames(styles.workflowNodeCard, { [styles.dragging]: isDragging })}
      style={{ transform: "translate(" + position.x + "px, " + position.y + "px)" }}
    >
      <div
        className={styles.workflowNodeHeader}
        draggable="true"
        onDrag={handleDrag}
        ref={itemRef}
        onDragEnd={handleDragEnd}
        onDragStart={handleDragStart}
        onDragOver={(e) => e.preventDefault()}
      >
        {props.heading}
      </div>
      <div className={styles.workflowNodeBodyWrapper}>{props.children}</div>
    </div>
  );
}

export default WorkflowNodeWrapper;
