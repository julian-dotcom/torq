import styles from "./workflow_link.module.scss";
import React, { createContext, useImperativeHandle } from "react";
import { useSelector } from "react-redux";
import { SelectWorkflowLinks } from "pages/WorkflowPage/workflowApi";
import { WorkflowVersionNodeLink } from "pages/WorkflowPage/workflowTypes";

type WorkflowLinkProps = {
  workflowId: number;
  version: number;
  workflowVersionId: number;
};

export type WorkflowLinkContextType = {
  updateParentPosition: (linkId: number, x: number, y: number) => void;
  updateChildPosition: (linkId: number, x: number, y: number) => void;
  svgRef: React.RefObject<SVGSVGElement>;
};

export const WorkflowLinksContext = createContext<React.MutableRefObject<WorkflowLinkContextType | undefined> | null>(
  null
);

// function WorkflowLink() {
//   const links = useSelector(SelectWorkflowLinks({ workflowId: 1, version: 1 }));
//
//   return (
//     <line
//       id={"link-" + link.workflowVersionNodeLinkId}
//       key={"link-" + link.workflowVersionNodeLinkId}
//       x={link.visibilitySettings?.parent?.xPosition}
//       y={link.visibilitySettings?.parent?.yPosition}
//       x2={link.visibilitySettings?.child?.xPosition}
//       y2={link.visibilitySettings?.child?.yPosition}
//       stroke="black"
//     />
//   );
// }

const WorkflowLinks = React.forwardRef(function WorkflowLink(props: WorkflowLinkProps, ref) {
  const links = useSelector(SelectWorkflowLinks({ workflowId: props.workflowId, version: props.version }));

  const svgRef = React.useRef<SVGSVGElement>(null);

  useImperativeHandle(ref, () => {
    return {
      svgRef: svgRef,
      updateChildPosition: (linkId: number, x: number, y: number) => {
        const link =
          links.find((link) => link.parentWorkflowVersionNodeId === linkId) || ({} as WorkflowVersionNodeLink);
        if (link?.visibilitySettings?.child) {
          link.visibilitySettings.child = {
            xPosition: x,
            yPosition: y,
          };
        }
      },
      updateParentPosition: (linkId: number, x: number, y: number) => {
        const link =
          links.find((link) => link.childWorkflowVersionNodeId === linkId) || ({} as WorkflowVersionNodeLink);
        if (link?.visibilitySettings?.parent) {
          link.visibilitySettings.parent = {
            xPosition: x,
            yPosition: y,
          };
        }
      },
    };
  });

  return (
    <div style={{ height: 0, width: 0, position: "absolute", top: 0 }}>
      <svg className={styles.workflowLink} width={"1px"} height={"1px"} overflow={"visible"} ref={svgRef}>
        {links.map((link) => {
          return (
            <line
              id={"link-" + link.workflowVersionNodeLinkId}
              key={"link-" + link.workflowVersionNodeLinkId}
              x1={link.visibilitySettings?.parent?.xPosition}
              y1={link.visibilitySettings?.parent?.yPosition}
              x2={link.visibilitySettings?.child?.xPosition}
              y2={link.visibilitySettings?.child?.yPosition}
              className={styles.workflowLink}
            />
          );
        })}
      </svg>
    </div>
  );
});

export default WorkflowLinks;
