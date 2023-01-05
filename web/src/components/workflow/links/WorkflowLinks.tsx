import styles from "./workflow_link.module.scss";
import { useLayoutEffect, useRef } from "react";
import { useSelector } from "react-redux";
import { SelectWorkflowLinks } from "pages/WorkflowPage/workflowApi";
import { WorkflowVersionNodeLink } from "pages/WorkflowPage/workflowTypes";
import classNames from "classnames";

type WorkflowLinkProp = {
  link: WorkflowVersionNodeLink;
};

export type LinkPositionEventDetails = { x: number; y: number; nodeId: number };

function WorkflowLink(props: WorkflowLinkProp) {
  const { link } = props;
  const linkRef = useRef<SVGLineElement>(null);
  const parentEventName = `parentLinkMove-${props.link.parentWorkflowVersionNodeId.toString()}-${props.link.parentOutputIndex.toString()}`;
  const childEventName = `childLinkMove-${props.link.childWorkflowVersionNodeId.toString()}-${props.link.childInputIndex.toString()}`;
  const initialPath = "M 0 0 C 1 1 1 1 2 2";

  function setPath(path: { x1: number; y1: number; x2: number; y2: number }) {
    if (linkRef === null) {
      return;
    }
    const { x1, y1, x2, y2 } = path;
    const t = 0.5;

    const controlPoint1X = x1 + t * (x2 - x1);
    const controlPoint2X = x2 - t * (x2 - x1);

    linkRef?.current &&
      linkRef.current.setAttribute("d", `M ${x1} ${y1} C ${controlPoint1X} ${y1} ${controlPoint2X} ${y2} ${x2} ${y2}`);
  }

  function handleParentPositionUpdate(e: CustomEventInit<LinkPositionEventDetails>) {
    if (linkRef !== null && linkRef.current !== null) {
      const x1 = e.detail?.x || 0;
      const y1 = e.detail?.y || 0;

      const pathCommands = (linkRef?.current?.getAttribute("d") || initialPath).split(" ");
      const x2 = parseFloat(pathCommands[8]);
      const y2 = parseFloat(pathCommands[9]);

      setPath({ x1, y1, x2, y2 });
    }
  }

  function handleChildPositionUpdate(e: CustomEventInit<LinkPositionEventDetails>) {
    if (linkRef !== null && linkRef.current !== null) {
      const x2 = e.detail?.x || 0;
      const y2 = e.detail?.y || 0;

      const pathCommands = (linkRef?.current?.getAttribute("d") || initialPath).split(" ");
      const x1 = parseFloat(pathCommands[1]);
      const y1 = parseFloat(pathCommands[2]);

      setPath({ x1, y1, x2, y2 });
    }
  }

  useLayoutEffect(() => {
    window.addEventListener(parentEventName, handleParentPositionUpdate);
    window.addEventListener(childEventName, handleChildPositionUpdate);

    // cleanup event listeners
    return () => {
      window.removeEventListener(parentEventName, handleParentPositionUpdate);
      window.removeEventListener(childEventName, handleChildPositionUpdate);
    };
  }, []);

  return (
    <path ref={linkRef} id={"link-" + link.workflowVersionNodeLinkId} key={"link-" + link.workflowVersionNodeLinkId} />
  );
}

type WorkflowLinkProps = {
  workflowId: number;
  version: number;
  workflowVersionId: number;
  stage: number;
  active: boolean;
  style?: React.CSSProperties;
};

function WorkflowLinks(props: WorkflowLinkProps) {
  const links = useSelector(
    SelectWorkflowLinks({ workflowId: props.workflowId, version: props.version, stage: props.stage })
  );

  return (
    <div style={{ height: 0, width: 0, position: "absolute", top: 0, ...props.style }}>
      <svg
        className={classNames(styles.workflowLinks, { [styles.active]: props.active })}
        width={"1px"}
        height={"1px"}
        overflow={"visible"}
      >
        {links.map((link) => {
          return <WorkflowLink key={"link-" + link.workflowVersionNodeLinkId} link={link} />;
        })}
      </svg>
    </div>
  );
}

export default WorkflowLinks;
