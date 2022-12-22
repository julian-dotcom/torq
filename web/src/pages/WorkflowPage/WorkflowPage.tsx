import { useState } from "react";
import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_page.module.scss";
import PageTitle from "features/templates/PageTitle";
import WorkflowCanvas from "components/workflow/canvas/WorkflowCanvas";
import { Link, useParams } from "react-router-dom";
import Sidebar from "features/sidebar/Sidebar";
import classNames from "classnames";
import { WORKFLOWS, MANAGE } from "constants/routes";
import { useExtractNodes, useWorkflowControls, useWorkflowData } from "./workflowHooks";

type WorkflowPageProps = {
  title?: string;
  workflowId?: string;
};

function WorkflowPage<T>(props: WorkflowPageProps) {
  const { t } = useTranslations();

  // Fetch the workflow data
  const { workflowId, version } = useParams();
  const { workflow, workflowVersion, stages } = useWorkflowData(workflowId, version);

  // Extract the nodes from the selected stage
  // TODO: Add stage selection buttons
  const [selectedStage, setSelectedStage] = useState<number>(1);
  const nodes = useExtractNodes(stages, selectedStage);

  // console.log("stage 1", stages[selectedStage]);

  // construct the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState<boolean>(false);
  const workflowControls = useWorkflowControls(sidebarExpanded, setSidebarExpanded);

  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
  };

  const bradcrumbs = [
    <Link to={`/${MANAGE}/${WORKFLOWS}`} key={"workflowsLink"}>
      {t.workflows}
    </Link>,
    workflow.name,
    workflowVersion.name,
  ];

  return (
    <div className={styles.contentWrapper}>
      <PageTitle breadcrumbs={bradcrumbs} title={workflow.name || ""} />
      {workflowControls}
      <div className={styles.tableWrapper}>
        <div className={styles.tableContainer}>
          <div className={styles.tableExpander}>
            <WorkflowCanvas>{nodes}</WorkflowCanvas>
          </div>
        </div>
      </div>
      <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: sidebarExpanded })}>
        <Sidebar title={t.nodes} closeSidebarHandler={closeSidebarHandler}>
          <div>{"nodes here"}</div>
        </Sidebar>
      </div>
    </div>
  );
}

export default WorkflowPage;
