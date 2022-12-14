import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_page.module.scss";
import PageTitle from "features/templates/PageTitle";
import { MutableRefObject, useRef, useState } from "react";
import classNames from "classnames";

type WorkflowPageProps = {
  title?: string;
  breadcrumbs?: Array<any>;
};

// function useDragging() {
//   const [isDragging, setIsDragging] = useState(false);
//   const [pos, setPos] = useState({ x: 0, y: 0 });
//   const ref = useRef() as MutableRefObject<HTMLDivElement>;
//   let offsetX = 0;
//   let offsetY = 0;
//   console.log(offsetY, offsetX);
//
//   function onMouseMove(e: any) {
//     if (!isDragging) return;
//     setPos({
//       x: e.x - offsetX * 2,
//       y: e.y - offsetY * 2,
//     });
//     e.stopPropagation();
//     e.preventDefault();
//   }
//
//   function onMouseUp(e: any) {
//     setIsDragging(false);
//     e.stopPropagation();
//     e.preventDefault();
//   }
//
//   function onMouseDown(e: any) {
//     if (e.button !== 0) return;
//     setIsDragging(true);
//     // console.log("onMouseDown", e);
//     // setPos({
//     //   x: e.x - ref.current.offsetWidth,
//     //   y: e.y - ref.current.offsetTop,
//     // });
//     offsetX = e.x;
//     offsetY = e.y;
//
//     e.stopPropagation();
//     e.preventDefault();
//   }
//
//   // When the element mounts, attach an mousedown listener
//   useEffect(() => {
//     ref.current.addEventListener("mousedown", onMouseDown);
//
//     return () => {
//       ref.current.removeEventListener("mousedown", onMouseDown);
//     };
//   }, [ref.current]);
//
//   // Everytime the isDragging state changes, assign or remove
//   // the corresponding mousemove and mouseup handlers
//   useEffect(() => {
//     if (isDragging) {
//       document.addEventListener("mouseup", onMouseUp);
//       document.addEventListener("mousemove", onMouseMove);
//     } else {
//       document.removeEventListener("mouseup", onMouseUp);
//       document.removeEventListener("mousemove", onMouseMove);
//     }
//     return () => {
//       document.removeEventListener("mouseup", onMouseUp);
//       document.removeEventListener("mousemove", onMouseMove);
//     };
//   }, [isDragging]);
//
//   return [ref, pos.x, pos.y, isDragging];
// }

function WorkflowPage<T>(props: WorkflowPageProps) {
  const { t } = useTranslations();
  const bradcrumbs = props.breadcrumbs || ["Workflow", "1"];

  const [position, setPosition] = useState({ x: 0, y: 0 });

  const canvasRef = useRef() as MutableRefObject<HTMLDivElement>;
  const itemRef = useRef() as MutableRefObject<HTMLDivElement>;
  const [itemBB, setItemBB] = useState({ left: 0, top: 0 });

  const [isDragging, setIsDragging] = useState(false);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    const itemBB = itemRef.current.getBoundingClientRect();
    const x = e.clientX - itemBB.left;
    const y = e.clientY - itemBB.top;
    e.dataTransfer.setDragImage(new Image(), x, y);
    setIsDragging(true);
    setItemBB({ left: x, top: y });
  }

  function handleDrag(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    const bb = canvasRef.current.getBoundingClientRect();
    if (e.clientX !== 0 && e.clientY !== 0) {
      const newX = e.clientX - bb.x - itemBB.left;
      const newY = e.clientY - bb.y - itemBB.top;
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

      <div className={styles.workspaceWrapper}>
        <div className={styles.workspaceContainer} ref={canvasRef}>
          <div
            className={classNames(styles.workflowNodeCard, { [styles.dragging]: isDragging })}
            draggable="true"
            onDrag={handleDrag}
            ref={itemRef}
            onDragEnd={handleDragEnd}
            onDragStart={handleDragStart}
            onDragOver={(e) => e.preventDefault()}
            style={{ transform: "translate(" + position.x + "px, " + position.y + "px)" }}
          >
            dsfasdfasdf
          </div>
        </div>
      </div>
    </div>
  );
}

export default WorkflowPage;
