import styles from "./workflow_nodes.module.scss";
import { MutableRefObject, useContext, useEffect, useRef, useState } from "react";
import classNames from "classnames";
import { CanvasContext } from "components/workflow/canvas/WorkflowCanvas";
import { NodeContext } from "./WorkflowNodeWrapper";
import { useAppDispatch } from "store/hooks";
import { setDisplayPreviewLink } from "pages/WorkflowPage/WorkflowSlice";

export type NodeConnectorProps = {
  id: string;
  name: string;
  outputName: string;
  workflowVersionId: number;
  workflowVersionNodeId: number;
};

function NodeConnector(props: NodeConnectorProps) {
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const { canvasRef } = useContext(CanvasContext);
  const { nodeRef } = useContext(NodeContext);

  const connectorRef = useRef() as MutableRefObject<HTMLDivElement>;
  const [isDragging, setIsDragging] = useState(false);

  const [originalPosition, setOriginalPosition] = useState({ x: 0, y: 0 });
  const dispatch = useAppDispatch();

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    // Set the drag effect to link
    e.dataTransfer.effectAllowed = "link";

    // Set the information about which node and connector is being dragged.
    e.dataTransfer.setData("node/workflowVersionId", props.workflowVersionId.toString());
    e.dataTransfer.setData("node/parentWorkflowVersionNodeId", props.workflowVersionNodeId.toString());
    e.dataTransfer.setData("node/parentOutput", props.outputName);
    nodeRef?.current?.classList.add(styles.connecting);
    e.dataTransfer.setData("node/name", props.name);

    if (canvasRef !== null) {
      const { x, y } = getConnectorPosition();

      setIsDragging(true);

      setOriginalPosition({ x, y });
      dispatch(setDisplayPreviewLink(true));
    }

    e.stopPropagation();
  }

  function updatePreviewLink(connBB: DOMRect, canvasBB: DOMRect, newX: number, newY: number) {
    if (connBB !== null) {
      const x = connBB.x - canvasBB.x + connBB.width / 2 + -2; // -14 because of the 16 padding right on the connector and 4px line width
      const y = connBB.y - canvasBB.y + connBB.height / 2 - 15;

      const previewEvent = new CustomEvent("previewLinkEvent", {
        detail: {
          startX: originalPosition.x,
          startY: originalPosition.y,
          x: x + newX,
          y: y + newY,
        },
      });
      window.dispatchEvent(previewEvent);
    }
  }

  function handleDrag(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();

    const connBB = connectorRef?.current?.getBoundingClientRect() || { x: 0, y: 0 };

    if (canvasRef !== null && e.clientX !== 0 && e.clientY !== 0) {
      const canvasBB = canvasRef?.current?.getBoundingClientRect() || { left: 0, top: 0 };
      const newX = e.clientX - connBB.x - 14; // value 14 needed to compensate for the padding otherwise alignment issue
      const newY = e.clientY - connBB.y;

      setPosition({ x: newX, y: newY });
      updatePreviewLink(connBB, canvasBB, newX, newY);
    }
  }

  function handleDragEnd(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    nodeRef?.current?.classList.remove(styles.connecting);
    setPosition({ x: 0, y: 0 });
    setIsDragging(false);
    dispatch(setDisplayPreviewLink(false));
  }

  function updater() {
    if (canvasRef !== null) {
      const { x, y } = getConnectorPosition();
      const eventName = `parentLinkMove-${props.workflowVersionNodeId}-${props.outputName}`;

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
    mutations.forEach(function () {
      updater();
    });
  }

  function getConnectorPosition() {
    if (canvasRef === null) return { x: 0, y: 0 };

    const connBB = connectorRef?.current?.getBoundingClientRect() || { left: 0, top: 0 };
    const canvasBB = canvasRef?.current?.getBoundingClientRect() || { left: 0, top: 0 };
    const x = connBB.x - canvasBB.x + connBB.width / 2 + -2; // -14 because of the 16 padding right on the connector and 4px line width
    const y = connBB.y - canvasBB.y + connBB.height / 2 - 15;
    return { x, y };
  }

  useEffect(() => {
    updater();
  });

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
      {isDragging && <div className={classNames(styles.nodeConnectorDragDot, styles.selectedNonDraggable)} />}
    </div>
  );
}

export default NodeConnector;
