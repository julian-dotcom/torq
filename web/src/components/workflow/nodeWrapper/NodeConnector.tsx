import styles from "./workflow_nodes.module.scss";
import { MutableRefObject, useContext, useEffect, useRef, useState } from "react";
import classNames from "classnames";
import { CanvasContext } from "components/workflow/canvas/WorkflowCanvas";
import { NodeContext } from "./WorkflowNodeWrapper";

export type NodeConnectorProps = {
  id: string;
  name: string;
  workflowVersionId: number;
  workflowVersionNodeId: number;
};

function NodeConnector(props: NodeConnectorProps) {
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const { blankImgRef } = useContext(CanvasContext);
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
    e.dataTransfer.setData("node/workflowVersionId", props.workflowVersionId.toString());
    e.dataTransfer.setData("node/parentWorkflowVersionNodeId", props.workflowVersionNodeId.toString());
    e.dataTransfer.setData("node/parentOutputIndex", "1");
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
    e.preventDefault();
    e.stopPropagation();
    nodeRef?.current?.classList.remove(styles.connecting);
    setPosition({ x: 0, y: 0 });
    setIsDragging(false);
  }

  function updater() {
    if (canvasRef !== null) {
      const connBB = connectorRef?.current?.getBoundingClientRect() || { left: 0, top: 0 };
      const canvasBB = canvasRef?.current?.getBoundingClientRect() || { left: 0, top: 0 };
      const x = connBB.x - canvasBB.x + connBB.width / 2 + -2; // -14 because of the 16 padding right on the connector and 4px line width
      const y = connBB.y - canvasBB.y + connBB.height / 2 - 15;
      const eventName = `parentLinkMove-${props.workflowVersionNodeId}-${1}`;
      const event = new CustomEvent(eventName, {
        detail: {
          x: x,
          y: y,
          nodeId: props.workflowVersionNodeId,
        },
      });
      window.dispatchEvent(event);
    }
  }

  function updatePosition(mutations: MutationRecord[]) {
    mutations.forEach(function (mutationRecord) {
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

  // run updater every second
  useEffect(() => {
    const interval = setInterval(() => {
      updater();
    }, 10);
    return () => clearInterval(interval);
  });

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
