import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_nodes.module.scss";
import { MutableRefObject, useContext, useRef, useState } from "react";
import classNames from "classnames";
import { CanvasContext } from "components/workflow/canvas/WorkflowCanvas";
import { NodeContext } from "./WorkflowNodeWrapper";

export enum connectorVariant {
  nodeInput = "nodeInput",
  nodeOutput = "nodeOutput",
}

const nodeConnectorVariantClass = new Map([
  [connectorVariant.nodeInput, styles.nodeInput],
  [connectorVariant.nodeOutput, styles.nodeOutput],
]);

export type NodeConnectorProps = {
  connectorVariant: connectorVariant;
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
    const nodeId = nodeRef?.current?.id ? nodeRef?.current?.id : "unknown";
    e.dataTransfer.setData("node/id", nodeId);
    e.dataTransfer.setData("node/connectorId", props.connectorVariant);
    const nodeBB = connectorRef.current.getBoundingClientRect();
    e.dataTransfer.effectAllowed = "move";
    const x = e.clientX - nodeBB.left;
    const y = e.clientY - nodeBB.top;
    e.dataTransfer.setDragImage(blankImgRef.current, x, y);
    setIsDragging(true);
    setNodeBB({ left: x, top: y });
  }

  function handleDrag(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    const bb = canvasRef?.current !== undefined || null ? canvasRef.current.getBoundingClientRect() : { x: 0, y: 0 };
    if (e.clientX !== 0 && e.clientY !== 0) {
      const newX = e.clientX - bb.x - nodeBB.left;
      const newY = e.clientY - bb.y - nodeBB.top;
      setPosition({ x: newX, y: newY });
    }
  }

  function handleDragEnd(e: React.DragEvent<HTMLDivElement>) {
    // e.preventDefault();
    // e.stopPropagation();
    setIsDragging(false);
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    // Get the id of the nodes connector that was dropped
    const nodeid = e.dataTransfer.getData("node/id");
    const connectorId = e.dataTransfer.getData("node/connectorId");
    // e.dataTransfer.clearData("node/id");
    console.log(
      "Dropped from node: " +
        nodeid +
        " with connector: " +
        connectorId +
        " on node: " +
        nodeRef?.current?.id +
        " with connector: " +
        props.connectorVariant
    );
  }

  return (
    <div
      className={classNames(styles.nodeConnector, nodeConnectorVariantClass.get(props.connectorVariant))}
      draggable="true"
      onDrag={handleDrag}
      ref={connectorRef}
      onDragEnd={handleDragEnd}
      onDragStart={handleDragStart}
      onDrop={handleDrop}
    >
      <div className={styles.nodeConnectorDot} />
      {/*<div*/}
      {/*  className={classNames(styles.nodeConnectorDragDot, { [styles.dragging]: isDragging })}*/}
      {/*  // style={{ left: position.x, top: position.y }}*/}
      {/*/>*/}
    </div>
  );
}

export default NodeConnector;
