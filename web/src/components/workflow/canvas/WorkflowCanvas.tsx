import styles from "./workflow_canvas.module.scss";
import { MutableRefObject, ReactNode, useRef, useState } from "react";
import classNames from "classnames";

type WorkflowCanvasProps = {
  canvasRef: MutableRefObject<HTMLDivElement>;
  blankImgRef: MutableRefObject<HTMLCanvasElement>;
  children: ReactNode;
};

function WorkflowCanvas(props: WorkflowCanvasProps) {
  const [p, setPosition] = useState({ x: 0, y: 0 });
  const wrapperRef = useRef() as MutableRefObject<HTMLDivElement>;

  const [canvasPosition, setCanvasPositionBB] = useState({ left: 0, top: 0 });
  const [isDragging, setIsDragging] = useState(false);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    const canvasPosition = props.canvasRef.current.getBoundingClientRect();
    e.dataTransfer.effectAllowed = "move";
    const x = e.clientX - canvasPosition.left;
    const y = e.clientY - canvasPosition.top;
    e.dataTransfer.setDragImage(props.blankImgRef.current, x, y);
    setIsDragging(true);
    setCanvasPositionBB({ left: x, top: y });
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

  function handleDragEnd(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }

  return (
    <div className={styles.workspaceWrapper} ref={wrapperRef}>
      <div
        className={classNames(styles.workspaceCanvas, { [styles.dragging]: isDragging })}
        draggable="true"
        onDrag={handleDrag}
        onDragEnd={handleDragEnd}
        onDragStart={handleDragStart}
        onDragOver={(e) => e.preventDefault()}
        style={{ backgroundPosition: `${p.x}px ${p.y}px` }}
      >
        <div style={{ transform: "translate(" + p.x + "px, " + p.y + "px)" }} ref={props.canvasRef}>
          {props.children}
        </div>
        <canvas ref={props.blankImgRef} style={{ width: "1px", height: "1px" }} />
      </div>
    </div>
  );
}

export default WorkflowCanvas;
