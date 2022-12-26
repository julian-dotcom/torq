import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_nodes.module.scss";
import React, { createRef, MutableRefObject, useContext, useId, useRef, useState } from "react";
import classNames from "classnames";
import NodeConnector from "./NodeConnector";
import { CanvasContext } from "components/workflow/canvas/WorkflowCanvas";
import { ExpandUpRight16Regular as ExpandIcon, ContractDownLeft16Regular as CollapseIcon } from "@fluentui/react-icons";
import Collapse from "features/collapse/Collapse";
import { WorkflowNode } from "pages/WorkflowPage/workflowTypes";

type nodeRefType = { nodeRef: MutableRefObject<HTMLDivElement> | null; nodeName: string };
export const NodeContext = React.createContext<nodeRefType>({
  nodeRef: null,
  nodeName: "",
});

export type WorkflowNodeProps = WorkflowNode & {
  id: string;
  heading?: string;
  children?: React.ReactNode;
  x?: number;
  y?: number;
};

function WorkflowNodeWrapper<T>(props: WorkflowNodeProps) {
  const { t } = useTranslations();
  const [collapsed, setCollapsed] = useState(true);
  const [position, setPosition] = useState({
    x: props.visibilitySettings.xPosition || 100,
    y: props.visibilitySettings.yPosition || 100,
  });

  // Canvas and blankRef are used to calculate the position of the node. They are passed down from the canvas
  const { canvasRef, blankImgRef } = useContext(CanvasContext);

  // nodeRef is used by the NodeConnector to allow for drag and drop interaction between nodes.
  const nodeRef = createRef() as MutableRefObject<HTMLDivElement>;

  const headerRef = useRef() as MutableRefObject<HTMLDivElement>;
  const [nodeBB, setNodeBB] = useState({ left: 0, top: 0 });
  const [isDragging, setIsDragging] = useState(false);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    // Set the drag effect and remove the default drag image set by HTML5
    if (blankImgRef) {
      e.dataTransfer.setDragImage(blankImgRef.current, 0, 0);
    }
    e.dataTransfer.effectAllowed = "move";

    // Set the dragging state to true to allow for css changes
    setIsDragging(true);

    // This sets offsets the starting position of the node to the mouse position,
    // preventing the node from jumping to the mouse position when we drag it.
    const nodeBB = headerRef.current.getBoundingClientRect();
    const x = e.clientX - nodeBB.left;
    const y = e.clientY - nodeBB.top;
    setNodeBB({ left: x, top: y });
  }

  function handleDrag(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();

    // Get the position of the canvas
    const bb = canvasRef !== null ? canvasRef.current.getBoundingClientRect() : { x: 0, y: 0 };

    if (e.clientX !== 0 && e.clientY !== 0) {
      // Calculate the new position of the node based on the mouse position (e.clientX/Y),
      // the top left position of the canvas (bb.x/y) and the top left position of the node (nodeBB.x/y)
      const newX = e.clientX - bb.x - nodeBB.left;
      const newY = e.clientY - bb.y - nodeBB.top;
      setPosition({ x: newX, y: newY });
    }
  }

  function handleDragEnd(e: React.DragEvent<HTMLDivElement>) {
    setIsDragging(false);
  }

  const connectorId = useId();

  function handleCollapse() {
    // Toggle the collapsed state
    setCollapsed(!collapsed);
  }

  return (
    <NodeContext.Provider
      value={{
        nodeRef: nodeRef,
        nodeName: props.name,
      }}
    >
      <div
        id={props.id}
        className={classNames(styles.workflowNodeCard, {
          [styles.dragging]: isDragging,
        })}
        style={{ transform: "translate(" + position.x + "px, " + position.y + "px)" }}
        ref={nodeRef}
      >
        <div
          className={classNames(styles.workflowNodeHeader, { [styles.headerCollapsed]: collapsed })}
          draggable="true"
          onDrag={handleDrag}
          ref={headerRef}
          onDragEnd={handleDragEnd}
          onDragStart={handleDragStart}
          onDragOver={(e) => e.preventDefault()}
          onClick={handleCollapse}
        >
          <div>{props.heading + ": " + props.name}</div>
          {collapsed ? <ExpandIcon /> : <CollapseIcon />}
          <NodeConnector id={connectorId} name={props.name} />
        </div>
        <Collapse collapsed={collapsed} animate={true}>
          <div className={styles.workflowNodeBody}>{props.children}</div>
        </Collapse>
      </div>
    </NodeContext.Provider>
  );
}

export default WorkflowNodeWrapper;
