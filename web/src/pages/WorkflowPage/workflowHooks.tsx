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
import Button, { ColorVariant } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import { useNavigate } from "react-router";
import { useGetWorkflowQuery, useNewWorkflowMutation } from "pages/WorkflowPage/workflowApi";
import { ReactNode } from "react";
import { Workflow, WorkflowStages, WorkflowVersion, WorkflowVersionNode } from "./workflowTypes";

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

  // reduce the workflow nodes to an object of stages containing an array of nodes
  const stages: WorkflowStages =
    (data?.nodes || []).reduce((acc: WorkflowStages, node: WorkflowVersionNode) => {
      const stage = acc[node.stage] || [];
      stage.push(node);
      acc[node.stage] = stage;
      return acc;
    }, {}) || {};

  return { workflow, workflowVersion, stages };
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
