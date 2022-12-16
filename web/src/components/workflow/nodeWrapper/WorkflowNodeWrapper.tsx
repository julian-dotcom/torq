import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_nodes.module.scss";
import React, { createRef, MutableRefObject, useContext, useRef, useState } from "react";
import classNames from "classnames";
// import * as d3 from "d3";
import NodeConnector, { connectorVariant } from "./NodeConnector";
import { CanvasContext } from "../canvas/WorkflowCanvas";

type nodeRefType = { nodeRef: MutableRefObject<HTMLDivElement> | null };
export const NodeContext = React.createContext<nodeRefType>({
  nodeRef: null,
});

export type WorkflowNodeProps = {
  id: string;
  heading?: string;
  children?: React.ReactNode;
  nextNodeRef?: MutableRefObject<HTMLCanvasElement>;
};

function WorkflowNodeWrapper<T>(props: WorkflowNodeProps) {
  const { t } = useTranslations();
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const { canvasRef, blankImgRef } = useContext(CanvasContext);
  // nodeRef is used by the NodeConnector to allow for drag and drop interaction between nodes.
  const nodeRef = createRef() as MutableRefObject<HTMLDivElement>;

  const itemRef = useRef() as MutableRefObject<HTMLDivElement>;
  const [nodeBB, setNodeBB] = useState({ left: 0, top: 0 });
  const [isDragging, setIsDragging] = useState(false);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    const nodeBB = itemRef.current.getBoundingClientRect();
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
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }

  return (
    <NodeContext.Provider
      value={{
        nodeRef: nodeRef,
      }}
    >
      <div
        id={props.id}
        className={classNames(styles.workflowNodeCard, { [styles.dragging]: isDragging })}
        style={{ transform: "translate(" + position.x + "px, " + position.y + "px)" }}
        ref={nodeRef}
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
          <NodeConnector connectorVariant={connectorVariant.nodeInput} />
          <div>{props.heading}</div>
          <NodeConnector connectorVariant={connectorVariant.nodeOutput} />
        </div>
        <div className={styles.workflowNodeBodyWrapper}>{props.children}</div>
      </div>
    </NodeContext.Provider>
  );
}

export default WorkflowNodeWrapper;
