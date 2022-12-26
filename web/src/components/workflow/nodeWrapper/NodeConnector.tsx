import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_nodes.module.scss";
import { MutableRefObject, useContext, useRef, useState } from "react";
import classNames from "classnames";
import { CanvasContext } from "components/workflow/canvas/WorkflowCanvas";
import { NodeContext } from "./WorkflowNodeWrapper";

export type NodeConnectorProps = {
  id: string;
  name: string;
};

function NodeConnector<T>(props: NodeConnectorProps) {
  const { t } = useTranslations();
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const { canvasRef, blankImgRef } = useContext(CanvasContext);
  const { nodeRef } = useContext(NodeContext);

  const connectorRef = useRef() as MutableRefObject<HTMLDivElement>;
  const [nodeBB, setNodeBB] = useState({ left: 0, top: 0 });
  const [isDragging, setIsDragging] = useState(false);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    // Hide the drag image (silly HTML5 default)
    if (blankImgRef) {
      e.dataTransfer.setDragImage(blankImgRef.current, 0, 0);
    }

    // Set the drag effect to link
    e.dataTransfer.effectAllowed = "link";

    // Set the information about which node and connector is being dragged.
    const nodeId = nodeRef?.current?.id ? nodeRef?.current?.id : "unknown";
    e.dataTransfer.setData("node/id", nodeId);
    e.dataTransfer.setData("node/connectorId", props.id);
    nodeRef?.current?.classList.add(styles.connecting);
    e.dataTransfer.setData("node/name", props.name);

    // Get the current position of the dragged connector.
    const nodeBB = connectorRef.current.getBoundingClientRect();

    const x = e.clientX - nodeBB.left;
    const y = e.clientY - nodeBB.top;

    setIsDragging(true);
    setNodeBB({ left: x, top: y });
    e.stopPropagation();
  }

  function handleDrag(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();

    const bb = connectorRef?.current?.getBoundingClientRect() || { x: 0, y: 0 };
    if (e.clientX !== 0 && e.clientY !== 0) {
      const newX = e.clientX - bb.x - nodeBB.left;
      const newY = e.clientY - bb.y - nodeBB.top;
      setPosition({ x: newX, y: newY });
    }
  }

  function handleDragEnd(e: React.DragEvent<HTMLDivElement>) {
    nodeRef?.current?.classList.remove(styles.connecting);
    setPosition({ x: 0, y: 0 });
    setIsDragging(false);
  }

  return (
    <div
      className={classNames(styles.nodeConnector)}
      draggable="true"
      onDrag={handleDrag}
      ref={connectorRef}
      onDragEnd={handleDragEnd}
      onDragStart={handleDragStart}
    >
      <div
        className={classNames(styles.nodeConnectorDragDot, { [styles.dragging]: isDragging })}
        style={{ transform: "translate(" + position.x + "px, " + position.y + "px)" }}
      />
      <div className={styles.nodeConnectorDot} />
    </div>
  );
}

export default NodeConnector;
