import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_nodes.module.scss";
import React, { createRef, MutableRefObject, useContext, useId, useRef, useState } from "react";
import classNames from "classnames";
import NodeConnector from "./NodeConnector";
import { CanvasContext } from "components/workflow/canvas/WorkflowCanvas";
import {
  ContractDownLeft20Regular as CollapseIcon,
  ExpandUpRight20Regular as ExpandIcon,
  MoreVertical20Regular as OptionsIcon,
  Delete16Regular as DeleteIcon,
} from "@fluentui/react-icons";
import Collapse from "features/collapse/Collapse";
import { WorkflowNode } from "pages/WorkflowPage/workflowTypes";
import NodeName from "./NodeNameInput";
import { useDeleteNodeMutation, useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import PopoverButton from "features/popover/Popover";
import Button, { buttonColor, buttonSize } from "components/buttons/Button";

type nodeRefType = { nodeRef: MutableRefObject<HTMLDivElement> | null; nodeName: string };
export const NodeContext = React.createContext<nodeRefType>({
  nodeRef: null,
  nodeName: "",
});

export type WorkflowNodeProps = WorkflowNode & {
  id: string;
  heading?: string;
  headerIcon?: JSX.Element;
  children?: React.ReactNode;
};

function WorkflowNodeWrapper<T>(props: WorkflowNodeProps) {
  const { t } = useTranslations();
  const [collapsed, setCollapsed] = useState(props.visibilitySettings.collapsed);
  const [position, setPosition] = useState({
    x: props.visibilitySettings.xPosition || 100,
    y: props.visibilitySettings.yPosition || 100,
  });

  // Canvas and blankRef are used to calculate the position of the node. They are passed down from the canvas
  const { canvasRef, blankImgRef } = useContext(CanvasContext);

  // nodeRef is used by the NodeConnector to allow for drag and drop interaction between nodes.
  const nodeRef = createRef() as MutableRefObject<HTMLDivElement>;

  const headerRef = useRef() as MutableRefObject<HTMLDivElement>;
  const [nodeBB, setNodeBB] = useState({ left: 0, top: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [nameInputVisible, setNameInputVisible] = useState(false);
  const [updateNode] = useUpdateNodeMutation();
  const [deleteNode] = useDeleteNodeMutation();

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    if (nameInputVisible) {
      e.preventDefault();
      return;
    }
    // Set the drag effect and remove the default drag image set by HTML5
    if (blankImgRef) {
      e.dataTransfer.setDragImage(blankImgRef.current, 0, 0);
    }
    e.dataTransfer.effectAllowed = "move";

    // Set the dragging state to true to allow for css changes
    setIsDragging(true);

    // This sets offsets the starting position of the node to the mouse position,
    // preventing the node from jumping to the mouse position when we drag it.
    const nodeBB = headerRef.current.getBoundingClientRect();
    const x = e.clientX - nodeBB.left;
    const y = e.clientY - nodeBB.top;
    setNodeBB({ left: x, top: y });
  }

  function handleDrag(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    if (nameInputVisible) {
      return;
    }

    // Get the position of the canvas
    const bb = canvasRef !== null ? canvasRef.current.getBoundingClientRect() : { x: 0, y: 0 };

    if (e.clientX !== 0 && e.clientY !== 0) {
      // Calculate the new position of the node based on the mouse position (e.clientX/Y),
      // the top left position of the canvas (bb.x/y) and the top left position of the node (nodeBB.x/y)
      const newX = e.clientX - bb.x - nodeBB.left;
      const newY = e.clientY - bb.y - nodeBB.top;
      setPosition({ x: newX, y: newY });
    }
  }

  function handleDragEnd(e: React.DragEvent<HTMLDivElement>) {
    if (nameInputVisible) {
      return;
    }
    setIsDragging(false);

    updateNode({
      workflowVersionNodeId: props.workflowVersionNodeId,
      visibilitySettings: { xPosition: position.x, yPosition: position.y, collapsed: collapsed },
    });
  }

  const connectorId = useId();

  function handleCollapse() {
    // Toggle the collapsed state
    updateNode({
      workflowVersionNodeId: props.workflowVersionNodeId,
      visibilitySettings: { xPosition: position.x, yPosition: position.y, collapsed: !collapsed },
    });
    setCollapsed(!collapsed);
  }

  return (
    <NodeContext.Provider
      value={{
        nodeRef: nodeRef,
        nodeName: props.name,
      }}
    >
      <div
        id={props.id}
        className={classNames(styles.workflowNodeCard, {
          [styles.dragging]: isDragging,
        })}
        style={{ transform: "translate(" + position.x + "px, " + position.y + "px)" }}
        ref={nodeRef}
      >
        <div
          className={classNames(styles.workflowNodeHeader, { [styles.headerCollapsed]: collapsed })}
          draggable="true"
          onDrag={handleDrag}
          ref={headerRef}
          onDragEnd={handleDragEnd}
          onDragStart={handleDragStart}
          onDragOver={(e) => e.preventDefault()}
        >
          <div className={styles.icon}>{props.headerIcon}</div>
          <NodeName
            nodeId={props.workflowVersionNodeId}
            name={props.name}
            isVisible={nameInputVisible}
            setVisible={setNameInputVisible}
          />
          <div className={classNames(styles.icon, styles.collapseIcon)} onClick={handleCollapse}>
            {collapsed ? <ExpandIcon /> : <CollapseIcon />}
          </div>
          <PopoverButton
            button={
              <div className={classNames(styles.icon, styles.optionsIcon)}>
                <OptionsIcon />
              </div>
            }
          >
            <Button
              text={"Delete"}
              icon={<DeleteIcon />}
              buttonColor={buttonColor.warning}
              buttonSize={buttonSize.small}
              onClick={() => {
                deleteNode({ nodeId: props.workflowVersionNodeId });
              }}
            />
          </PopoverButton>
          <NodeConnector id={connectorId} name={props.name} />
        </div>
        <Collapse collapsed={collapsed} animate={true}>
          <div className={styles.workflowNodeBody}>{props.children}</div>
        </Collapse>
      </div>
    </NodeContext.Provider>
  );
}

export default WorkflowNodeWrapper;
