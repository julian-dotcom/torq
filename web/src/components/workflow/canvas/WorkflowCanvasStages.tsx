import { useSelector } from "react-redux";
import { SelectWorkflowMainTriggerNode, SelectWorkflowStages } from "pages/WorkflowPage/workflowApi";
import WorkflowCanvas from "./WorkflowCanvas";
import styles from "./workflow_canvas.module.scss";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import {
  ChannelPolicyNode,
  TimeTriggerNode,
  StageTriggerNode,
  ChannelFilterNode,
  ReBalanceChannelNode,
  AddTagNode,
  BalanceTriggerNode,
  ChannelOpenTriggerNode,
  ChannelCloseTriggerNode,
} from "components/workflow/nodes/nodes";
import { WorkflowVersionNode } from "pages/WorkflowPage/workflowTypes";
import classNames from "classnames";
import useTranslations from "services/i18n/useTranslations";
import NodeConnector from "components/workflow/nodeWrapper/NodeConnector";
import { RemoveTagNode } from "../nodes/tags/RemoveTagNode";

type WorkflowCanvasStagesProps = {
  workflowId: number;
  version: number;
  workflowVersionId: number;
  selectedStage: number;
};

function FirstStageTrigger(props: {
  workflowVersionId: number;
  version: number;
  workflowId: number;
  stage: number;
  triggers: WorkflowVersionNode[];
}) {
  const { t } = useTranslations();
  const triggerNode = useSelector(
    SelectWorkflowMainTriggerNode({ workflowId: props.workflowId, version: props.version })
  );

  const triggerNodes = props.triggers.map(getNodeComponent);

  if (props.stage === 1) {
    return (
      <div className={classNames(styles.triggerNodeWrapper)}>
        <div className={styles.triggerNodeContainer}>
          <NodeConnector
            id={"ss"}
            name={t.triggers}
            outputName={"all"}
            workflowVersionNodeId={triggerNode?.workflowVersionNodeId || 0}
            workflowVersionId={props.workflowVersionId}
          />
          <div className={classNames(styles.triggerContainerHeading)}>{t.triggers}</div>
          <div className={styles.triggerBody}>{triggerNodes}</div>
        </div>
      </div>
    );
  } else {
    return <div className={styles.stageTriggerContainer}>{triggerNodes}</div>;
  }
}

export function WorkflowCanvases(props: WorkflowCanvasStagesProps) {
  const stages = useSelector(SelectWorkflowStages({ workflowId: props.workflowId, version: props.version }));

  const stageComponents = stages.map((stage) => {
    const actionNodes = stage.actions.map(getNodeComponent);
    return (
      <WorkflowCanvas
        key={"stage-" + stage.stage}
        stageNumber={stage.stage}
        workflowVersionId={props.workflowVersionId}
        workflowId={props.workflowId}
        version={props.version}
        selected={props.selectedStage === stage.stage}
      >
        <FirstStageTrigger
          workflowVersionId={props.workflowVersionId}
          version={props.version}
          workflowId={props.workflowId}
          triggers={stage.triggers}
          stage={stage.stage}
        />
        {actionNodes}
      </WorkflowCanvas>
    );
  });

  return <>{stageComponents}</>;
}

// A function that returns the correct node component based on the node type
function getNodeComponent(node: WorkflowVersionNode) {
  switch (node.type) {
    case WorkflowNodeType.TimeTrigger:
      return <TimeTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelBalanceEventTrigger:
      return <BalanceTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelOpenEventTrigger:
      return <ChannelOpenTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelCloseEventTrigger:
      return <ChannelCloseTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.StageTrigger:
      return <StageTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelPolicyConfigurator:
      return <ChannelPolicyNode {...node} key={"node-id-" + node.workflowVersionNodeId} outputName={"channels"} />;
    case WorkflowNodeType.ChannelFilter:
      return <ChannelFilterNode {...node} key={"node-id-" + node.workflowVersionNodeId} outputName={"channels"} />;
    case WorkflowNodeType.RebalanceParameters:
      return <ReBalanceChannelNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.AddTag:
      return <AddTagNode {...node} key={"node-id-" + node.workflowVersionNodeId} outputName={"channels"} />;
    case WorkflowNodeType.RemoveTag:
      return <RemoveTagNode {...node} key={"node-id-" + node.workflowVersionNodeId} outputName={"channels"} />;
    default:
      return null;
  }
}
