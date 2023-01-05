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
    const { x, y, nodeId } = e.detail ? e.detail : { x: 0, y: 0, nodeId: 0 };
    if (nodeId == 0) {
      return;
    }
    if (linkRef.current) {
      linkRef.current.setAttribute("x1", x.toString());
      linkRef.current.setAttribute("y1", y.toString());
    }
  }

  function handleChildPositionUpdate(e: CustomEventInit<LinkPositionEventDetails>) {
    const { x, y, nodeId } = e.detail ? e.detail : { x: 0, y: 0, nodeId: 0 };
    if (nodeId == 0) {
      return;
    }
    if (linkRef.current) {
      linkRef.current.setAttribute("x2", x.toString());
      linkRef.current.setAttribute("y2", y.toString());
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
};

function WorkflowLinks(props: WorkflowLinkProps) {
  const links = useSelector(
    SelectWorkflowLinks({ workflowId: props.workflowId, version: props.version, stage: props.stage })
  );

  return (
    <div style={{ height: 0, width: 0, position: "absolute", top: 0 }}>
      <svg
        className={classNames(styles.workflowLinks, { [styles.active]: props.active })}
        width={"1px"}
        height={"1px"}
        overflow={props.active ? "visible" : "hidden"}
      >
        {links.map((link) => {
          return <WorkflowLink key={"link-" + link.workflowVersionNodeLinkId} link={link} />;
        })}
      </svg>
    </div>
  );
}

export default WorkflowLinks;
