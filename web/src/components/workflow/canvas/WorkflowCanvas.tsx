import styles from "./workflow_canvas.module.scss";
import React, { createRef, MutableRefObject, ReactNode, useContext, useRef, useState } from "react";
import classNames from "classnames";
import { useAddNodeMutation } from "pages/WorkflowPage/workflowApi";
import WorkflowLinks from "../links/WorkflowLinks";
import { TriggerNodeTypes } from "pages/WorkflowPage/constants";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";
import useTranslations from "services/i18n/useTranslations";
import { userEvents } from "utils/userEvents";

type WorkflowCanvasProps = {
  workflowVersionId: number;
  workflowId: number;
  version: number;
  stageNumber: number;
  selected: boolean;
  children: ReactNode;
};

// Context provider is used to pass these references to the workflow nodes without having to pass them as props
export const CanvasContext = React.createContext<{
  canvasRef: MutableRefObject<HTMLDivElement> | null;
  svgRef: MutableRefObject<SVGSVGElement> | null;
  blankImgRef: MutableRefObject<HTMLCanvasElement> | null;
}>({
  canvasRef: null,
  svgRef: null,
  blankImgRef: null,
});

function WorkflowCanvas(props: WorkflowCanvasProps) {
  const { t } = useTranslations();
  const { track } = userEvents();
  const toastRef = useContext(ToastContext);

  const [addNode] = useAddNodeMutation();

  // p is used to store the current position of the canvas
  const [position, setPosition] = useState({ x: 50, y: 300 });

  // canvasRef is used to allow workflow nodes to use the canvas position as reference
  const canvasRef = createRef() as MutableRefObject<HTMLDivElement>;

  // svgRef is used to place connecting lines between workflow nodes
  const svgRef = createRef() as MutableRefObject<SVGSVGElement>;

  // blankImgRef is used to store a blank image to remove the default drag-and-drop cursor
  const blankImgRef = createRef() as MutableRefObject<HTMLCanvasElement>;

  // wrapperRef is used to refer to the wrapper element that surrounds the canvas
  const wrapperRef = useRef() as MutableRefObject<HTMLDivElement>;

  // canvasPosition is used to store the initial position of the canvas when a drag starts
  const [canvasPosition, setCanvasPositionBB] = useState({ left: 0, top: 0 });
  const [isDragging, setIsDragging] = useState(false);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    // Set the type of drag-and-drop operation that is allowed for the element being dragged
    e.dataTransfer.effectAllowed = "move";

    // Calculate the x and y coordinates of the mouse cursor relative to the top-left corner of the canvas
    const canvasPosition = canvasRef.current.getBoundingClientRect();
    const x = e.clientX - canvasPosition.left;
    const y = e.clientY - canvasPosition.top;
    // Set the image that is shown as the element is being dragged
    e.dataTransfer.setDragImage(blankImgRef.current, x, y);

    // Set the isDragging and canvasPositionBB state variables
    setIsDragging(true);
    setCanvasPositionBB({ left: x, top: y });
    track("Workflow Drag Canvas", {
      workflowId: props.workflowId,
      workflowVersionId: props.workflowVersionId,
      workflowVersion: props.version,
      workflowStageNumber: props.stageNumber,
    });
  }

  function handleDrag(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    const bb = wrapperRef.current.getBoundingClientRect();
    if (e.clientX !== 0 && e.clientY !== 0) {
      const newX = e.clientX - bb.x - canvasPosition.left;
      const newY = e.clientY - bb.y - canvasPosition.top;
      setPosition({ x: newX, y: newY });
    }
  }

  function handleDragEnd() {
    setIsDragging(false);
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    if (e.dataTransfer.getData("node/event") === "add") {
      const bb = wrapperRef.current.getBoundingClientRect();

      // Calculate the x and y coordinates of the mouse cursor (e.client) relative to the top-left corner of the canvas (bb.left/top), then
      // add the current position of the canvas (position.x/y).
      const newX = e.clientX - bb.left - position.x;
      const newY = e.clientY - bb.top - position.y;

      const nodeType = parseInt(e.dataTransfer.getData("node/type"));
      const nodeName = e.dataTransfer.getData("node/name");
      const nodeParameters = e.dataTransfer.getData("node/parameters");
      let jsonParam = {};
      if (nodeParameters) {
        jsonParam = JSON.parse(nodeParameters);
      }

      if (TriggerNodeTypes.includes(nodeType) && props.stageNumber !== 1) {
        e.dataTransfer.effectAllowed = "none";
        // Add a toast to inform the user that trigger nodes can only be added to the first stage
        toastRef?.current &&
          toastRef.current.addToast(t.workflowDetails.cantAddTriggerNodeToStages, toastCategory.error);
        return;
      }

      track("Workflow Add New Node", {
        workflowId: props.workflowId,
        workflowVersionId: props.workflowVersionId,
        workflowVersion: props.version,
        workflowStageNumber: props.stageNumber,
        workflowNodeType: nodeType,
        workflowNodeName: nodeName,
        workflowNodeMethod: "canvas drop",
      });

      addNode({
        type: nodeType,
        name: nodeName,
        visibilitySettings: {
          xPosition: Math.floor(newX),
          yPosition: Math.floor(newY),
          collapsed: false,
        },
        workflowVersionId: props.workflowVersionId,
        stage: props.stageNumber,
        parameters: jsonParam,
      });
    }
  }

  return (
    <CanvasContext.Provider
      value={{
        canvasRef: canvasRef,
        svgRef: svgRef,
        blankImgRef: blankImgRef,
      }}
    >
      <div
        className={classNames(styles.workflowWrapper, {
          [styles.selectedStage]: props.selected,
        })}
        ref={wrapperRef}
      >
        <div
          className={classNames(styles.workspaceCanvas, { [styles.dragging]: isDragging })}
          onDragOver={(e) => e.preventDefault()}
          style={{ backgroundPosition: `${position.x}px ${position.y}px` }}
        >
          <div
            className={styles.canvasDragSurface}
            draggable="true"
            onDrop={handleDrop}
            onDrag={handleDrag}
            onDragEnd={handleDragEnd}
            onDragStart={handleDragStart}
            onDragOver={(e) => e.preventDefault()}
          />
          <div
            style={{
              transform: "translate(" + position.x + "px, " + position.y + "px)",
              zIndex: 100,
              position: "relative",
            }}
            ref={canvasRef}
          >
            {props.children}
          </div>
          <WorkflowLinks
            style={{ transform: "translate(" + position.x + "px, " + position.y + "px)" }}
            stage={props.stageNumber}
            selectedStage={props.selected}
            workflowId={props.workflowId}
            version={props.version}
            workflowVersionId={props.workflowVersionId}
          />
          <canvas
            ref={blankImgRef}
            style={{ width: "1px", height: "1px", border: "1px solid transaparent" }}
            width={"1px"}
            height={"1px"}
          />
        </div>
      </div>
    </CanvasContext.Provider>
  );
}

export default WorkflowCanvas;
