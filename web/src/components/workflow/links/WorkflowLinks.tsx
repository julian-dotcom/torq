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

  function handleParentPositionUpdate(e: CustomEventInit<LinkPositionEventDetails>) {
    if (linkRef !== null && linkRef.current !== null) {
      linkRef.current.setAttribute("x1", (e.detail?.x || 0).toString());
      linkRef.current.setAttribute("y1", (e.detail?.y || 0).toString());
    }
  }

  function handleChildPositionUpdate(e: CustomEventInit<LinkPositionEventDetails>) {
    if (linkRef !== null && linkRef.current !== null) {
      linkRef.current.setAttribute("x2", (e.detail?.x || 0).toString());
      linkRef.current.setAttribute("y2", (e.detail?.y || 0).toString());
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
    <line ref={linkRef} id={"link-" + link.workflowVersionNodeLinkId} key={"link-" + link.workflowVersionNodeLinkId} />
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
