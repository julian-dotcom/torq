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
  Pause16Regular as DeactivateIcon,
  Play16Regular as ActivateIcon,
} from "@fluentui/react-icons";
import Collapse from "features/collapse/Collapse";
import { WorkflowVersionNode } from "pages/WorkflowPage/workflowTypes";
import NodeName from "./NodeNameInput";
import {
  useDeleteNodeMutation,
  useUpdateNodeMutation,
  useUpdateNodeVisibilitySettingsMutation,
} from "pages/WorkflowPage/workflowApi";
import PopoverButton from "features/popover/Popover";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { TriggerNodeTypes } from "pages/WorkflowPage/constants";
import { NodeColorVariant, GetColorClass } from "components/workflow/nodes/nodeVariants";
import { Status } from "constants/backend";
import { useClickOutside } from "utils/hooks";
import useTranslations from "services/i18n/useTranslations";
import mixpanel from "mixpanel-browser";

type nodeRefType = { nodeRef: MutableRefObject<HTMLDivElement> | null; nodeName: string };
export const NodeContext = React.createContext<nodeRefType>({
  nodeRef: null,
  nodeName: "",
});

export type WorkflowNodeProps = WorkflowVersionNode & {
  heading?: string;
  headerIcon?: JSX.Element;
  colorVariant: NodeColorVariant;
  noDeletion?: boolean;
  noDeactivation?: boolean;
  noOptions?: boolean;
  outputName?: string;
  children?: React.ReactNode;
};

function WorkflowNodeWrapper(props: WorkflowNodeProps) {
  const { t } = useTranslations();
  const [nodeIsSelected, setNodeIsSelected] = useState<boolean>(false);

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
  const [updateNodeVisibilitySettings] = useUpdateNodeVisibilitySettingsMutation();
  const [deleteNode] = useDeleteNodeMutation();

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    setNodeIsSelected(true);
    // Don't initiate dragging with the user is editing the node name
    if (nameInputVisible) {
      e.preventDefault();
      return;
    }

    // Don't drag trigger nodes
    if (TriggerNodeTypes.includes(props.type)) {
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
    mixpanel.track("Workflow Drag Node", {
      nodeType: props.type,
      workflowId: props.workflowId,
      workflowVersionId: props.workflowVersionId,
      stageNumber: props.stage,
      nodeId: props.workflowVersionNodeId,
      status: props.status,
      version: props.version,
    });
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

  function handleDragEnd() {
    if (nameInputVisible) {
      return;
    }
    setIsDragging(false);

    updateNodeVisibilitySettings({
      workflowVersionNodeId: props.workflowVersionNodeId,
      visibilitySettings: { xPosition: position.x, yPosition: position.y, collapsed: collapsed },
    });
  }

  const connectorId = useId();

  function handleCollapse() {
    mixpanel.track("Workflow Collapse Node", {
      nodeType: props.type,
      workflowId: props.workflowId,
      workflowVersionId: props.workflowVersionId,
      stageNumber: props.stage,
      nodeId: props.workflowVersionNodeId,
      status: props.status,
      version: props.version,
    });
    // Toggle the collapsed state
    updateNodeVisibilitySettings({
      workflowVersionNodeId: props.workflowVersionNodeId,
      visibilitySettings: { xPosition: position.x, yPosition: position.y, collapsed: !collapsed },
    });
    setCollapsed(!collapsed);
  }

  const transform = TriggerNodeTypes.includes(props.type)
    ? `translate(0px, 0px)`
    : `translate(${position.x}px, ${position.y}px)`;

  // TODO: Add selected status to a node, so that it can be styled differently and z index can be increased to ensure it is on top of other nodes

  useClickOutside(nodeRef, () => {
    setNodeIsSelected(false);
  });

  return (
    <NodeContext.Provider
      value={{
        nodeRef: nodeRef,
        nodeName: props.name,
      }}
    >
      <div
        onClick={() => setNodeIsSelected(true)}
        className={classNames(styles.workflowNodeCard, GetColorClass(props.colorVariant), {
          [styles.dragging]: isDragging,
          [styles.triggerNode]: TriggerNodeTypes.includes(props.type),
          [styles.selected]: nodeIsSelected,
          [styles.inactive]: props.status === Status.Inactive,
        })}
        style={{ transform: `${transform}` }}
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
            name={props.name + (props.status === Status.Inactive ? " - [ INACTIVE ]" : "")}
            isVisible={nameInputVisible}
            setVisible={setNameInputVisible}
          />
          <div className={classNames(styles.icon, styles.collapseIcon)} onClick={handleCollapse}>
            {collapsed ? <ExpandIcon /> : <CollapseIcon />}
          </div>
          {!props.noOptions && (
            <PopoverButton
              button={
                <div className={classNames(styles.icon, styles.optionsIcon)}>
                  <OptionsIcon />
                </div>
              }
            >
              <div className={styles.buttonGroup}>
                {!props.noDeactivation && (
                  <Button
                    icon={props.status === Status.Active ? <DeactivateIcon /> : <ActivateIcon />}
                    buttonColor={ColorVariant.primary}
                    buttonSize={SizeVariant.small}
                    onClick={() => {
                      mixpanel.track("Workflow Node Status", {
                        nodeType: props.type,
                        workflowId: props.workflowId,
                        workflowVersionId: props.workflowVersionId,
                        stageNumber: props.stage,
                        nodeId: props.workflowVersionNodeId,
                        version: props.version,
                        status: props.status === 0 ? t.active : t.deactivate,
                      });
                      updateNode({
                        workflowVersionNodeId: props.workflowVersionNodeId,
                        // TODO: Switch to enum here
                        status: props.status === 0 ? 1 : 0,
                      });
                    }}
                  >
                    {props.status === 0 ? t.activate : t.deactivate}
                  </Button>
                )}
                {!props.noDeletion && (
                  <Button
                    icon={<DeleteIcon />}
                    buttonColor={ColorVariant.error}
                    buttonSize={SizeVariant.small}
                    onClick={() => {
                      mixpanel.track("Workflow Delete Node", {
                        nodeType: props.type,
                        workflowId: props.workflowId,
                        workflowVersionId: props.workflowVersionId,
                        stageNumber: props.stage,
                        nodeId: props.workflowVersionNodeId,
                        version: props.version,
                        status: props.status,
                      });
                      deleteNode({ nodeId: props.workflowVersionNodeId });
                    }}
                  >
                    {t.delete}
                  </Button>
                )}
              </div>
            </PopoverButton>
          )}
          {props.outputName && (
            <NodeConnector
              id={connectorId}
              name={props.name}
              outputName={props.outputName || ""}
              workflowVersionNodeId={props.workflowVersionNodeId}
              workflowVersionId={props.workflowVersionId}
            />
          )}
        </div>
        <Collapse collapsed={collapsed} animate={true}>
          <div className={styles.workflowNodeBody}>{props.children}</div>
        </Collapse>
      </div>
    </NodeContext.Provider>
  );
}

export default WorkflowNodeWrapper;
