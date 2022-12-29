import useTranslations from "services/i18n/useTranslations";
import classNames from "classnames";
import styles from "./workflow_sidebar.module.scss";
import Sidebar from "features/sidebar/Sidebar";
import NodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { SectionContainer } from "features/section/SectionContainer";
import {
  Timer20Regular as TriggersIcon,
  Scales20Regular as EventTriggerIcon,
  ArrowRouting20Regular as ChannelOpenIcon,
} from "@fluentui/react-icons";
import { useState } from "react";

export type WorkflowSidebarProps = {
  expanded: boolean;
  setExpanded: (expanded: boolean) => void;
};

export default function WorkflowSidebar(props: WorkflowSidebarProps) {
  const { expanded, setExpanded } = props;

  const { t } = useTranslations();
  const closeSidebarHandler = () => {
    setExpanded(false);
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

  return (
    <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: expanded })}>
      <Sidebar title={t.nodes} closeSidebarHandler={closeSidebarHandler}>
        {" "}
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
  );
}
