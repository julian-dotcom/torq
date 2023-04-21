import { useSelector } from "react-redux";
import {
  SelectWorkflowMainTriggerNode,
  SelectWorkflowStages,
  useAddManualWorkflowTriggerMutation,
  useAddNodeMutation,
} from "pages/WorkflowPage/workflowApi";
import WorkflowCanvas from "components/workflow/canvas/WorkflowCanvas";
import styles from "components/workflow/canvas/workflow_canvas.module.scss";
import { TriggerNodeTypes, WorkflowNodeType } from "pages/WorkflowPage/constants";
import {
  ChannelPolicyAutoRunNode,
  ChannelPolicyConfiguratorNode,
  ChannelPolicyRunNode,
  IntervalTriggerNode,
  CronTriggerNode,
  StageTriggerNode,
  ChannelFilterNode,
  RebalanceAutoRunNode,
  RebalanceConfiguratorNode,
  RebalanceRunNode,
  AddTagNode,
  BalanceTriggerNode,
  ChannelOpenTriggerNode,
  ChannelCloseTriggerNode,
  DataSourceTorqChannelsNode,
  ChannelBalanceEventFilterNode,
} from "components/workflow/nodes/nodes";
import { WorkflowVersionNode } from "pages/WorkflowPage/workflowTypes";
import classNames from "classnames";
import useTranslations from "services/i18n/useTranslations";
import { RemoveTagNode } from "components/workflow/nodes/tags/RemoveTagNode";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { Play12Regular as PlayIcon } from "@fluentui/react-icons";
import { useContext } from "react";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";
import Spinny from "features/spinny/Spinny";
import { userEvents } from "utils/userEvents";

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
  const { track } = userEvents();
  const triggerNode = useSelector(
    SelectWorkflowMainTriggerNode({ workflowId: props.workflowId, version: props.version })
  );
  const toastRef = useContext(ToastContext);
  const [addNode] = useAddNodeMutation();

  const triggerNodes = props.triggers.map(getNodeComponent);

  const [triggerWorkflow, { isLoading: runIsProcessing }] = useAddManualWorkflowTriggerMutation();

  function handleManualTrigger() {
    track("Workflow Manually Triggered", {
      workflowId: props.workflowId,
      workflowVersionId: props.workflowVersionId,
      workflowVersion: props.version,
      workflowStageNumber: props.stage,
    });

    triggerWorkflow({
      type: WorkflowNodeType.StageTrigger,
      workflowVersionId: props.workflowVersionId,
      workflowId: props.workflowId,
      workflowVersionNodeId: triggerNode?.workflowVersionNodeId || 0,
    })
      .then(() => {
        toastRef?.current && toastRef.current.addToast(t.workflowDetails.manualTriggerSuccess, toastCategory.success);
      })
      .catch(() => {
        toastRef?.current && toastRef.current.addToast(t.workflowDetails.manualTriggerError, toastCategory.error);
      });
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    e.stopPropagation();
    if (e.dataTransfer.getData("node/event") === "add") {
      const nodeType = parseInt(e.dataTransfer.getData("node/type"));
      const nodeName = e.dataTransfer.getData("node/name");
      const nodeParameters = e.dataTransfer.getData("node/parameters");
      let jsonParam = {};
      if (nodeParameters) {
        jsonParam = JSON.parse(nodeParameters);
      }

      if (!TriggerNodeTypes.includes(nodeType)) {
        e.dataTransfer.effectAllowed = "none";
        // Add a toast to inform the user that trigger nodes can only be added to the first stage
        toastRef?.current &&
          toastRef.current.addToast(t.workflowDetails.cantDropActionOnDriggerContainer, toastCategory.error);
        return;
      }

      track("Workflow Add New Node", {
        workflowId: props.workflowId,
        workflowVersionId: props.workflowVersionId,
        workflowVersion: props.version,
        workflowStageNumber: props.stage,
        workflowNodeType: nodeType,
        workflowNodeName: nodeName,
        workflowNodeMethod: "trigger drop",
      });
      addNode({
        type: nodeType,
        name: nodeName,
        visibilitySettings: {
          xPosition: 0,
          yPosition: 0,
          collapsed: false,
        },
        workflowVersionId: props.workflowVersionId,
        stage: props.stage,
        parameters: jsonParam,
      });
    }
  }

  if (props.stage === 1) {
    return (
      <div className={classNames(styles.triggerNodeWrapper)} onDrop={handleDrop}>
        <div className={styles.triggerNodeContainer} data-intercom-target={"workflow-triggers-container"}>
          <div className={classNames(styles.triggerContainerHeading)}>
            <div>{t.triggers}</div>
            <Button
              icon={runIsProcessing ? <Spinny /> : <PlayIcon />}
              disabled={runIsProcessing}
              buttonSize={SizeVariant.tiny}
              buttonColor={ColorVariant.success}
              hideMobileText={true}
              onClick={handleManualTrigger}
            >
              {t.workflowNodes.run}
            </Button>
          </div>
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
    case WorkflowNodeType.IntervalTrigger:
      return <IntervalTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.CronTrigger:
      return <CronTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelBalanceEventTrigger:
      return <BalanceTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelOpenEventTrigger:
      return <ChannelOpenTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelCloseEventTrigger:
      return <ChannelCloseTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.StageTrigger:
      return <StageTriggerNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelPolicyAutoRun:
      return <ChannelPolicyAutoRunNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelPolicyConfigurator:
      return <ChannelPolicyConfiguratorNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelPolicyRun:
      return <ChannelPolicyRunNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelFilter:
      return <ChannelFilterNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.RebalanceAutoRun:
      return <RebalanceAutoRunNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.RebalanceConfigurator:
      return <RebalanceConfiguratorNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.RebalanceRun:
      return <RebalanceRunNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.AddTag:
      return <AddTagNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.RemoveTag:
      return <RemoveTagNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.DataSourceTorqChannels:
      return <DataSourceTorqChannelsNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    case WorkflowNodeType.ChannelBalanceEventFilter:
      return <ChannelBalanceEventFilterNode {...node} key={"node-id-" + node.workflowVersionNodeId} />;
    default:
      return null;
  }
}
