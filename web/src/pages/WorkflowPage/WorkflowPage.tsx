import { useState } from "react";
import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_page.module.scss";
import PageTitle from "features/templates/PageTitle";
import { Link, useParams } from "react-router-dom";
import Sidebar from "features/sidebar/Sidebar";
import classNames from "classnames";
import { WORKFLOWS, MANAGE } from "constants/routes";
import { useStages, useWorkflowControls, useWorkflowData } from "./workflowHooks";
import { ReactComponent as StageArrowBack } from "./stageArrowBack.svg";
import { ReactComponent as StageArrowFront } from "./stageArrowFront.svg";
import { WorkflowStages } from "./workflowTypes";
import NodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { SectionContainer } from "features/section/SectionContainer";
import {
  Timer20Regular as TriggersIcon,
  Scales20Regular as EventTriggerIcon,
  ArrowRouting20Regular as ChannelOpenIcon,
} from "@fluentui/react-icons";

export function useStageButtons(
  stages: WorkflowStages,
  selectedStage: number,
  setSelectedStage: (stage: number) => void
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
  return <div className={styles.stagesWrapper}>{stageButtons}</div>;
}

function WorkflowPage() {
  const { t } = useTranslations();

  // Fetch the workflow data
  const { workflowId, version } = useParams();
  const { workflow, workflowVersion, stages } = useWorkflowData(workflowId, version);

  const [selectedStage, setSelectedStage] = useState<number>(1);
  const stageButtons = useStageButtons(stages, selectedStage, setSelectedStage);
  const stagedCanvases = useStages(workflowVersion?.workflowVersionId || 0, stages, selectedStage);

  // construct the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState<boolean>(false);
  const workflowControls = useWorkflowControls(sidebarExpanded, setSidebarExpanded);

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

  const bradcrumbs = [
    <Link to={`/${MANAGE}/${WORKFLOWS}`} key={"workflowsLink"}>
      {t.workflows}
    </Link>,
    workflow?.name,
    workflowVersion?.name,
  ];

  return (
    <div className={styles.contentWrapper}>
      <PageTitle breadcrumbs={bradcrumbs} title={workflow?.name || ""} />
      {workflowControls}
      <div className={styles.tableWrapper}>
        <div className={styles.tableContainer}>
          <div className={styles.tableExpander}>
            {stagedCanvases}
            {stageButtons}
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
            <NodeButtonWrapper title={"Channel Opened"} nodeType={3} icon={<ChannelOpenIcon />} />
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
