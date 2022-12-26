import {
  PuzzlePiece20Regular as NodesIcon,
  Play20Regular as DeployIcon,
  Add16Regular as NewWorkflowIcon,
} from "@fluentui/react-icons";
import {
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import Button, { buttonColor } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import { useNavigate } from "react-router";
import { useGetWorkflowQuery, useNewWorkflowMutation } from "pages/WorkflowPage/workflowApi";
import { ReactNode } from "react";
import { Workflow, WorkflowStages, WorkflowVersion } from "./workflowTypes";
import ChannelPolicyNode from "components/workflow/nodes/channelPolicy/ChannelPolicy";
import WorkflowCanvas from "../../components/workflow/canvas/WorkflowCanvas";

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
      buttonColor={buttonColor.green}
      text={t.newWorkflow}
      className={"collapse-tablet"}
      icon={<NewWorkflowIcon />}
      onClick={newWorkflowHandler}
    />
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

export function useWorkflowControls(sidebarExpanded: boolean, setSidebarExpanded: (expanded: boolean) => void) {
  const { t } = useTranslations();
  return (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={buttonColor.green}
            text={t.deploy}
            className={"collapse-tablet"}
            icon={<DeployIcon />}
            onClick={() => {
              console.log("Not implemented yet");
            }}
          />
        </TableControlsTabsGroup>
        <Button
          buttonColor={buttonColor.subtle}
          text={t.nodes}
          className={"collapse-tablet"}
          id={"tableControlsButton"}
          icon={<NodesIcon />}
          onClick={() => {
            setSidebarExpanded(!sidebarExpanded);
          }}
        />
      </TableControlsButtonGroup>
    </TableControlSection>
  );
}
