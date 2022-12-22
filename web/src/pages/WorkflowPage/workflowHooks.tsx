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
import { WorkflowStages } from "./workflowTypes";
import ChannelPolicyNode from "components/workflow/nodes/ChannelPolicy";

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

  const workflow = data?.workflow || { name: "Workflow" };
  const workflowVersion = data?.version || { name: "Draft" };

  const stages: WorkflowStages = data?.workflowForest?.sortedStageTrees || {}; //.map((s) => parseInt(s));

  return { workflow, workflowVersion, stages };
}

export function useExtractNodes(stages: WorkflowStages, selectedStage: number) {
  return (stages[selectedStage] || []).map((s) => {
    const nodeId = s.workflowVersionNodeId;
    return <ChannelPolicyNode key={`node-${nodeId}`} id={`node-${nodeId}`} nodeName={s.name} />;
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
