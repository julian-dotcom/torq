import { useState } from "react";
import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_page.module.scss";
import PageTitle from "features/templates/PageTitle";
import { Link, useParams } from "react-router-dom";
import Sidebar from "features/sidebar/Sidebar";
import classNames from "classnames";
import { WORKFLOWS, MANAGE } from "constants/routes";
import { useStages, useWorkflowControls, useWorkflowData } from "./workflowHooks";
import { StageSelector } from "./WorkflowStageSelector";
import NodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { SectionContainer } from "features/section/SectionContainer";
import {
  Timer20Regular as TriggersIcon,
  Scales20Regular as EventTriggerIcon,
  ArrowRouting20Regular as ChannelOpenIcon,
} from "@fluentui/react-icons";
import { useUpdateWorkflowMutation } from "./workflowApi";

function WorkflowPage() {
  const { t } = useTranslations();

  // Fetch the workflow data
  const { workflowId, version } = useParams();
  const { workflow, workflowVersion, stages } = useWorkflowData(workflowId, version);

  const [selectedStage, setSelectedStage] = useState<number>(1);
  const stagedCanvases = useStages(workflowVersion?.workflowVersionId || 0, stages, selectedStage);

  // construct the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState<boolean>(false);
  const workflowControls = useWorkflowControls(sidebarExpanded, setSidebarExpanded);

  const [updateWorkflow] = useUpdateWorkflowMutation();

  function handleWorkflowNameChange(name: string) {
    updateWorkflow({ workflowId: parseInt(workflowId || "0"), name: name });
  }

  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
  };

  const [sectionState, setSectionState] = useState({
    triggers: true,
    actions: true,
  });

  const toggleSection = (section: keyof typeof sectionState) => {
    setSectionState({
      ...sectionState,
      [section]: !sectionState[section],
    });
  };

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
      {workflowControls}
      <div className={styles.tableWrapper}>
        <div className={styles.tableContainer}>
          <div className={styles.tableExpander}>
            {stagedCanvases}
            <StageSelector
              stages={stages}
              selectedStage={selectedStage}
              setSelectedStage={setSelectedStage}
              workflowVersionId={workflowVersion?.workflowVersionId || 0}
              workflowId={workflow?.workflowId || 0}
              version={workflowVersion?.version || 0}
            />
          </div>
        </div>
      </div>
      <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: sidebarExpanded })}>
        <Sidebar title={t.nodes} closeSidebarHandler={closeSidebarHandler}>
          <SectionContainer
            title={t.triggers}
            icon={TriggersIcon}
            expanded={sectionState.triggers}
            handleToggle={() => toggleSection("triggers")}
          >
            <NodeButtonWrapper title={"Interval"} nodeType={1} icon={<TriggersIcon />} />
            <NodeButtonWrapper title={"Channel Balance "} nodeType={2} icon={<EventTriggerIcon />} />
            <NodeButtonWrapper title={"Channel Opened"} nodeType={10} icon={<ChannelOpenIcon />} />
          </SectionContainer>
          <SectionContainer
            title={t.actions}
            icon={TriggersIcon}
            expanded={sectionState.actions}
            handleToggle={() => toggleSection("actions")}
          >
            <NodeButtonWrapper title={"Interval"} nodeType={1} icon={<TriggersIcon />} />
          </SectionContainer>
        </Sidebar>
      </div>
    </div>
  );
}

export default WorkflowPage;
