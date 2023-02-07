import { useState } from "react";
import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_page.module.scss";
import PageTitle from "features/templates/PageTitle";
import { Link, useParams } from "react-router-dom";
import { WORKFLOWS, MANAGE } from "constants/routes";
import { useWorkflowData } from "./workflowHooks";
import { useUpdateWorkflowMutation } from "./workflowApi";
import WorkflowSidebar from "components/workflow/sidebar/WorkflowSidebar";
import { WorkflowCanvases } from "components/workflow/canvas/WorkflowCanvasStages";
import { StageSelector } from "components/workflow/stages/WorkflowStageSelector";
import WorkflowControls from "./WorkflowControls";
import { Status } from "constants/backend";
import mixpanel from "mixpanel-browser";

function WorkflowPage() {
  const { t } = useTranslations();

  // Fetch the workflow data
  const { workflowId, version } = useParams();
  const { workflow, workflowVersion, stageNumbers } = useWorkflowData(workflowId, version);

  const [selectedStage, setSelectedStage] = useState<number>(1);
  const [sidebarExpanded, setSidebarExpanded] = useState<boolean>(false);

  function selectStage(stage: number) {
    mixpanel.track("Workflow Select Stage", {
      workflowId: workflowId,
      workflowVersion: version,
      workflowStage: stage,
    });
    setSelectedStage(stage);
  }

  const [updateWorkflow] = useUpdateWorkflowMutation();

  function handleWorkflowNameChange(name: string) {
    mixpanel.track("Workflow Update Name", {
      workflowId: workflow?.workflowId,
      workflowVersionId: workflowVersion?.workflowVersionId,
      workflowName: workflow?.name,
    });
    updateWorkflow({ workflowId: parseInt(workflowId || "0"), name: name });
  }

  const breadcrumbs = [
    <Link to={`/${MANAGE}/${WORKFLOWS}`} key={"workflowsLink"}>
      {t.workflows}
    </Link>,
    workflow?.name,
    workflowVersion?.name,
  ];

  return (
    <div className={styles.contentWrapper}>
      <PageTitle breadcrumbs={breadcrumbs} title={workflow?.name || ""} onNameChange={handleWorkflowNameChange} />
      <WorkflowControls
        sidebarExpanded={sidebarExpanded}
        setSidebarExpanded={setSidebarExpanded}
        workflowId={workflow?.workflowId || 0}
        status={workflow?.status || Status.Inactive}
      />
      <div className={styles.tableWrapper}>
        <div className={styles.tableContainer}>
          <div className={styles.tableExpander}>
            <WorkflowCanvases
              selectedStage={selectedStage}
              workflowVersionId={workflowVersion?.workflowVersionId || 0}
              workflowId={workflow?.workflowId || 0}
              version={workflowVersion?.version || 0}
            />
            <StageSelector
              stageNumbers={stageNumbers}
              selectedStage={selectedStage}
              setSelectedStage={selectStage}
              workflowVersionId={workflowVersion?.workflowVersionId || 0}
              workflowId={workflow?.workflowId || 0}
              version={workflowVersion?.version || 0}
              disabled={workflow?.status === Status.Active}
            />
          </div>
        </div>
      </div>
      <WorkflowSidebar expanded={sidebarExpanded} setExpanded={setSidebarExpanded} />
    </div>
  );
}

export default WorkflowPage;
