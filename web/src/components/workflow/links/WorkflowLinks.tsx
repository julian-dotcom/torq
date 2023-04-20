import classNames from "classnames";
import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { useSelector } from "react-redux";
import { DismissCircle24Regular as DeleteLinkIcon } from "@fluentui/react-icons";
import { SelectWorkflowLinks, useDeleteNodeLinkMutation } from "pages/WorkflowPage/workflowApi";
import { selectDisplayPreviewLink } from "pages/WorkflowPage/WorkflowSlice";
import { WorkflowVersionNodeLink } from "pages/WorkflowPage/workflowTypes";
import styles from "./workflow_link.module.scss";
import mixpanel from "mixpanel-browser";

type WorkflowLinkProp = {
  link: WorkflowVersionNodeLink;
};

export type LinkPositionEventDetails = { x: number; y: number; nodeId: number };
export type PreviewLinkPositionEventDetails = { startX: number; startY: number; x: number; y: number; nodeId: number };

function PreviewWorkflowLink() {
  const linkPreviewRef = useRef<SVGLineElement>(null);
  function setPath(path: { x1: number; y1: number; x2: number; y2: number }) {
    if (linkPreviewRef === null) return;

    const { x1, y1, x2, y2 } = path;
    const t = 0.5;

    const controlPoint1X = x1 + t * (x2 - x1);
    const controlPoint2X = x2 - t * (x2 - x1);

    linkPreviewRef?.current &&
      linkPreviewRef.current.setAttribute(
        "d",
        `M ${x1} ${y1} C ${controlPoint1X} ${y1} ${controlPoint2X} ${y2} ${x2} ${y2}`
      );
  }

  function handlePreviewPositionUpdate(e: CustomEventInit<PreviewLinkPositionEventDetails>) {
    if (linkPreviewRef !== null && linkPreviewRef.current !== null) {
      const x1 = e.detail?.startX || 0;
      const y1 = e.detail?.startY || 0;

      const x2 = e.detail?.x || 0;
      const y2 = e.detail?.y || 0;

      setPath({ x1, y1, x2, y2 });
    }
  }

  useLayoutEffect(() => {
    window.addEventListener("previewLinkEvent", handlePreviewPositionUpdate);

    // cleanup event listeners
    return () => {
      window.removeEventListener("previewLinkEvent", handlePreviewPositionUpdate);
    };
  }, [linkPreviewRef]);

  return (
    <g className={styles.linkWrapper}>
      <path strokeDasharray="10,10" ref={linkPreviewRef} id={"previewLink"} className={styles.previewLink}>
        <animate attributeName="stroke-dashoffset" values="100;0" dur="2s" calcMode="linear" repeatCount="indefinite" />
      </path>
    </g>
  );
}

function WorkflowLink(props: WorkflowLinkProp) {
  const { link } = props;
  const shadowLinkRef = useRef<SVGLineElement>(null);
  const linkRef = useRef<SVGLineElement>(null);
  const circleRef = useRef<SVGCircleElement>(null);
  const iconRef = useRef<SVGGElement>(null);
  const parentEventName = `parentLinkMove-${props.link.parentWorkflowVersionNodeId.toString()}-${
    props.link.parentOutput
  }`;
  const childEventName = `childLinkMove-${props.link.childWorkflowVersionNodeId.toString()}-${props.link.childInput}`;
  const initialPath = "M 0 0 C 1 1 1 1 2 2";
  const [deleteLink] = useDeleteNodeLinkMutation();

  function handleDeleteLink() {
    mixpanel.track("Workflow Delete link", {
      workflowVersionId: link.workflowVersionId,
      workflowStageNumber: link.stage,
      workflowParentOutput: link.parentOutput,
      workflowChildInput: link.childInput,
    });
    deleteLink({ linkId: link.workflowVersionNodeLinkId });
  }
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
    shadowLinkRef?.current &&
      shadowLinkRef.current.setAttribute(
        "d",
        `M ${x1} ${y1} C ${controlPoint1X} ${y1} ${controlPoint2X} ${y2} ${x2} ${y2}`
      );
    // Place the icon in the middle of the link path and offset it py it's height
    iconRef?.current &&
      iconRef.current.setAttribute("transform", `translate(${(x1 + x2) / 2 - 24 / 2}, ${(y1 + y2) / 2 - 24 / 2})`);
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
    const x2 = e.detail?.x || 0;
    const y2 = e.detail?.y || 0;
    if (linkRef !== null && linkRef.current !== null) {
      const pathCommands = (linkRef?.current?.getAttribute("d") || initialPath).split(" ");
      const x1 = parseFloat(pathCommands[1]);
      const y1 = parseFloat(pathCommands[2]);

      setPath({ x1, y1, x2, y2 });
    }
    if (circleRef !== null && circleRef.current !== null) {
      circleRef.current.setAttribute("cx", x2.toString());
      circleRef.current.setAttribute("cy", y2.toString());
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
  }, [circleRef, linkRef, shadowLinkRef, iconRef]);

  return (
    <g className={styles.linkWrapper}>
      <path
        ref={linkRef}
        id={"link-" + link.workflowVersionNodeLinkId}
        key={"link-" + link.workflowVersionNodeLinkId}
        className={styles.link}
      />
      <path
        ref={shadowLinkRef}
        id={"shadow-link-" + link.workflowVersionNodeLinkId}
        key={"shadow-link-" + link.workflowVersionNodeLinkId}
        className={styles.shadowLink}
      />
      <circle r="7" ref={circleRef} />
      <g ref={iconRef} className={styles.deleteLinksWrapper} onClick={() => handleDeleteLink()}>
        <circle r={24} cx={12} cy={12} strokeWidth={1} stroke={"transparent"} className={styles.shadowDeleteLink} />
        <circle r={10} cx={12} cy={12} strokeWidth={1} stroke={"transparent"} />
        <DeleteLinkIcon />
      </g>
    </g>
  );
}

type WorkflowLinkProps = {
  workflowId: number;
  version: number;
  workflowVersionId: number;
  stage: number;
  selectedStage: boolean;
  style?: React.CSSProperties;
};

function WorkflowLinks(props: WorkflowLinkProps) {
  const links = useSelector(
    SelectWorkflowLinks({ workflowId: props.workflowId, version: props.version, stage: props.stage })
  );

  const displayPreviewLink = useSelector(selectDisplayPreviewLink);
  const [scale, setScale] = useState(1 / window.devicePixelRatio || 1);

  const handleResize = () => {
    const scale = window.devicePixelRatio || 1;

    if (scale < 1.5) setScale(1 / scale);

    // TODO: Due to retina this if buggy on some zoom levels. Not sure what is happening.
    //   Could be caused by the preserveAspectRatio.
    //   the problem is that when we manipulate the initial scale
    if (scale >= 1.5 && scale < 2) setScale(2 / scale);
    if (scale == 2) setScale(2 / scale);
    if (scale >= 2.5 && scale < 3) setScale(2 / scale);
    if (scale == 3) setScale(3 / scale);
    if (scale == 4) setScale(4 / scale);
  };

  useEffect(() => {
    window.addEventListener("resize", handleResize, false);
  }, []);

  useEffect(() => {
    handleResize();
  }, []);

  return (
    <div
      style={{
        position: "absolute",
        top: 0,
        left: 0,
        bottom: 0,
        right: 0,
        overflow: "visible",
        pointerEvents: "none",
        zIndex: 900,
      }}
    >
      <svg
        className={classNames(styles.workflowLinks, { [styles.active]: props.selectedStage })}
        viewBox={`0 0 ${scale} ${scale}`}
        overflow={"visible"}
        style={props.style}
      >
        {displayPreviewLink && <PreviewWorkflowLink key="preview" />}

        {links.map((link) => {
          return <WorkflowLink key={"link-" + link.workflowVersionNodeLinkId} link={link} />;
        })}
      </svg>
    </div>
  );
}

export default WorkflowLinks;
