import classNames from "classnames";
import {
  PuzzlePiece20Regular as NodesIcon,
  Play20Regular as DeployIcon,
  Add16Regular as NewWorkflowIcon,
  Add16Regular as NewStageIcon,
} from "@fluentui/react-icons";
import {
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import Button, { ColorVariant } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import { useNavigate } from "react-router";
import { useGetWorkflowQuery, useNewWorkflowMutation } from "pages/WorkflowPage/workflowApi";
import { ReactNode } from "react";
import { Workflow, WorkflowStages, WorkflowVersion } from "./workflowTypes";
import ChannelPolicyNode from "components/workflow/nodes/channelPolicy/ChannelPolicy";
import WorkflowCanvas from "components/workflow/canvas/WorkflowCanvas";
import styles from "./workflow_page.module.scss";
import { ReactComponent as StageArrowBack } from "./stageArrowBack.svg";
import { ReactComponent as StageArrowFront } from "./stageArrowFront.svg";
import { useAddNodeMutation } from "./workflowApi";
import { WorkflowNodeType } from "./constants";

export function useNewWorkflowButton(): ReactNode {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const [newWorkflow] = useNewWorkflowMutation();

  function newWorkflowHandler() {
    const response = newWorkflow();
    response
      .then((res) => {
        console.log(res);
        const data = (res as { data: { workflowId: number; version: number } }).data;
        navigate(`/manage/workflows/${data.workflowId}/versions/${data.version}`);
      })
      .catch((err) => {
        // TODO: Handle error and show a toast
        console.log(err);
      });
  }

  return (
    <Button
      buttonColor={ColorVariant.success}
      className={"collapse-tablet"}
      icon={<NewWorkflowIcon />}
      onClick={newWorkflowHandler}
    >
      {t.newWorkflow}
    </Button>
  );
}

export function useWorkflowData(workflowId?: string, version?: string) {
  const { data } = useGetWorkflowQuery(
    {
      workflowId: parseInt(workflowId || ""),
      version: parseInt(version || ""),
    },
    { skip: !workflowId || !version }
  );

  const workflow: Workflow | undefined = data?.workflow;
  const workflowVersion: WorkflowVersion | undefined = data?.version;

  const stages: WorkflowStages = data?.workflowForest?.sortedStageTrees || {}; //.map((s) => parseInt(s));

  return { workflow, workflowVersion, stages };
}

export function useNodes(stages: WorkflowStages, stageNumber: number) {
  return (stages[stageNumber] || []).map((node) => {
    const nodeId = node.workflowVersionNodeId;
    return <ChannelPolicyNode {...node} key={`node-${nodeId}`} id={`node-${nodeId}`} name={node.name} />;
  });
}

export function useStages(workflowVersionId: number, stages: WorkflowStages, selectedStage: number) {
  return Object.entries(stages).map((stage) => {
    const stageNumber = parseInt(stage[0]);
    const nodes = useNodes(stages, stageNumber);
    return (
      <WorkflowCanvas
        active={selectedStage === stageNumber}
        key={`stage-${stageNumber}`}
        workflowVersionId={workflowVersionId}
        stageNumber={stageNumber}
      >
        {nodes}
      </WorkflowCanvas>
    );
  });
}

export function useStageButtons(
  stages: WorkflowStages,
  selectedStage: number,
  setSelectedStage: (stage: number) => void,
  workflowVersionId: number
) {
  const stageButtons = Object.keys(stages).map((stage, index) => {
    return (
      <button
        key={`stage-${stage}`}
        className={classNames(styles.stageContainer, { [styles.selected]: parseInt(stage) === selectedStage })}
        onClick={() => setSelectedStage(parseInt(stage))}
      >
        {index !== 0 && <StageArrowBack />}
        <div className={styles.stage}>{`Stage ${stage}`}</div>
        <StageArrowFront />
      </button>
    );
  });

  const [addNode] = useAddNodeMutation();

  function addStage() {
    addNode({
      type: WorkflowNodeType.StageTrigger,
      visibilitySettings: {
        xPosition: 0,
        yPosition: 0,
        collapsed: false,
      },
      workflowVersionId: workflowVersionId,
      stage: Math.max(...Object.keys(stages).map((stage) => parseInt(stage))) + 1,
    });
  }

  const addStageButton = (
    <button key={`stage-add-stage`} className={classNames(styles.stageContainer)} onClick={addStage}>
      <StageArrowBack />
      <div className={styles.stage}>
        <NewStageIcon />
      </div>
      <StageArrowFront />
    </button>
  );

  return (
    <div className={styles.stagesWrapper}>
      {stageButtons}
      {addStageButton}
    </div>
  );
}

export function useWorkflowControls(sidebarExpanded: boolean, setSidebarExpanded: (expanded: boolean) => void) {
  const { t } = useTranslations();
  return (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={ColorVariant.success}
            className={"collapse-tablet"}
            icon={<DeployIcon />}
            onClick={() => {
              console.log("Not implemented yet");
            }}
          >
            {t.deploy}
          </Button>
        </TableControlsTabsGroup>
        <Button
          buttonColor={ColorVariant.primary}
          className={"collapse-tablet"}
          id={"tableControlsButton"}
          icon={<NodesIcon />}
          onClick={() => {
            setSidebarExpanded(!sidebarExpanded);
          }}
        >
          {t.nodes}
        </Button>
      </TableControlsButtonGroup>
    </TableControlSection>
  );
}
