import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_page.module.scss";
import PageTitle from "features/templates/PageTitle";
import { MutableRefObject, useRef, useState } from "react";
import WorkflowNode from "./WorkflowNode";
import classNames from "classnames";

type WorkflowPageProps = {
  title?: string;
  breadcrumbs?: Array<any>;
};

function WorkflowPage<T>(props: WorkflowPageProps) {
  const { t } = useTranslations();
  const bradcrumbs = props.breadcrumbs || ["Workflow", "1"];

  const [position, setPosition] = useState({ x: 0, y: 0 });
  const wrapperRef = useRef() as MutableRefObject<HTMLDivElement>;
  const canvasRef = useRef() as MutableRefObject<HTMLDivElement>;
  const blankImgRef = useRef() as MutableRefObject<HTMLCanvasElement>;

  const [canvasPosition, setCanvasPositionBB] = useState({ left: 0, top: 0 });
  const [isDragging, setIsDragging] = useState(false);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    const canvasPosition = canvasRef.current.getBoundingClientRect();
    e.dataTransfer.effectAllowed = "move";
    const x = e.clientX - canvasPosition.left;
    const y = e.clientY - canvasPosition.top;
    e.dataTransfer.setDragImage(blankImgRef.current, x, y);
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
    <div className={styles.contentWrapper}>
      <PageTitle breadcrumbs={bradcrumbs} title={t.workflow}>
        {props.title}
      </PageTitle>

      <div className={styles.workspaceWrapper} ref={wrapperRef}>
        <div
          className={classNames(styles.workspaceCanvas, { [styles.dragging]: isDragging })}
          draggable="true"
          onDrag={handleDrag}
          onDragEnd={handleDragEnd}
          onDragStart={handleDragStart}
          onDragOver={(e) => e.preventDefault()}
        >
          <div style={{ transform: "translate(" + position.x + "px, " + position.y + "px)" }} ref={canvasRef}>
            <WorkflowNode canvasRef={canvasRef} heading={"Hello Heading"} blankImageRef={blankImgRef}>
              {"Hello"}
            </WorkflowNode>
            <WorkflowNode canvasRef={canvasRef} heading={"Hello Heading"} blankImageRef={blankImgRef}>
              {"Hello"}
            </WorkflowNode>
          </div>
          <canvas ref={blankImgRef} style={{ width: "1px", height: "1px" }} />
        </div>
      </div>
    </div>
  );
}

export default WorkflowPage;
