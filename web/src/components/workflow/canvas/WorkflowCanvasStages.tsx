import { useSelector } from "react-redux";
import { SelectWorkflowStages } from "pages/WorkflowPage/workflowApi";
import WorkflowCanvas from "./WorkflowCanvas";
import nodeStyles from "components/workflow/nodeWrapper/workflow_nodes.module.scss";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { ChannelPolicyNode } from "components/workflow/nodes/nodes";
import { WorkflowVersionNode } from "../../../pages/WorkflowPage/workflowTypes";
import { TimeTriggerNode } from "../nodes/timeTrigger/TimeTriggerNode";

type WorkflowCanvasStagesProps = {
  workflowId: number;
  version: number;
  workflowVersionId: number;
  selectedStage: number;
};

export function WorkflowCanvases(props: WorkflowCanvasStagesProps) {
  const stages = useSelector(SelectWorkflowStages({ workflowId: props.workflowId, version: props.version }));

  const stageComponents = stages.map((stage, index) => {
    const triggerNodes = stage.triggers.map(getNodeComponent);
    const actionNodes = stage.actions.map(getNodeComponent);
    return (
      <WorkflowCanvas
        key={"stage-" + stage.stage}
        stageNumber={stage.stage}
        workflowVersionId={props.workflowVersionId}
        active={props.selectedStage === stage.stage}
      >
        <div className={nodeStyles.triggerNodeWrapper}>
          <div className={nodeStyles.triggerNodeContainer}>{triggerNodes}</div>
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
      return <TimeTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelPolicyConfigurator:
      return <ChannelPolicyNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    default:
      return null;
  }
}
