import { useSelector } from "react-redux";
import { SelectWorkflowStages } from "pages/WorkflowPage/workflowApi";
import WorkflowCanvas from "./WorkflowCanvas";
import styles from "./workflow_canvas.module.scss";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import {
  ChannelPolicyNode,
  TimeTriggerNode,
  StageTriggerNode,
  ChannelFilterNode,
  ReBalanceChannelNode,
} from "components/workflow/nodes/nodes";
import { WorkflowVersionNode } from "pages/WorkflowPage/workflowTypes";
import classNames from "classnames";
import useTranslations from "services/i18n/useTranslations";
import NodeConnector from "components/workflow/nodeWrapper/NodeConnector";

type WorkflowCanvasStagesProps = {
  workflowId: number;
  version: number;
  workflowVersionId: number;
  selectedStage: number;
};

export function WorkflowCanvases(props: WorkflowCanvasStagesProps) {
  const { t } = useTranslations();
  const stages = useSelector(SelectWorkflowStages({ workflowId: props.workflowId, version: props.version }));

  const stageComponents = stages.map((stage, index) => {
    const triggerNodes = stage.triggers.map(getNodeComponent);
    const actionNodes = stage.actions.map(getNodeComponent);
    return (
      <WorkflowCanvas
        key={"stage-" + stage.stage}
        stageNumber={stage.stage}
        workflowVersionId={props.workflowVersionId}
        workflowId={props.workflowId}
        version={props.version}
        active={props.selectedStage === stage.stage}
      >
        <div className={styles.triggerNodeWrapper}>
          <div className={styles.triggerNodeContainer}>
            <div className={classNames(styles.triggerContainerHeading)}>
              <NodeConnector id={"ss"} name={t.triggers} workflowVersionNodeId={1} workflowVersionId={1} />
              {t.triggers}
            </div>
            <div className={styles.triggerBody}>{triggerNodes}</div>
          </div>
        </div>
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
    case WorkflowNodeType.StageTrigger:
      return <StageTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelPolicyConfigurator:
      return <ChannelPolicyNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelFilter:
      return <ChannelFilterNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.RebalanceParameters:
      return <ReBalanceChannelNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    default:
      return null;
  }
}
