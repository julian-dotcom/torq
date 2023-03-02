import useTranslations from "services/i18n/useTranslations";
import classNames from "classnames";
import styles from "./workflow_sidebar.module.scss";
import Sidebar from "features/sidebar/Sidebar";
import { SectionContainer } from "features/section/SectionContainer";
import {
  Timer20Regular as TriggersIcon,
  Play20Regular as DataSourcesIcon,
  ArrowRouting20Regular as ChannelsIcon,
  Tag20Regular as TagsIcon,
} from "@fluentui/react-icons";
import { useState } from "react";
import {
  ChannelPolicyConfiguratorNodeButton,
  ChannelPolicyAutoRunNodeButton,
  ChannelPolicyRunNodeButton,
  IntervalTriggerNodeButton,
  CronTriggerNodeButton,
  ChannelFilterNodeButton,
  RebalanceConfiguratorNodeButton,
  RebalanceAutoRunNodeButton,
  RebalanceRunNodeButton,
  RemoveTagNodeButton,
  BalanceTriggerNodeButton,
  AddTagNodeButton,
  ChannelCloseTriggerNodeButton,
  ChannelOpenTriggerNodeButton,
  DataSourceTorqChannelsNodeButton,
  EventFilterNodeButton
} from "components/workflow/nodes/nodes";
import mixpanel from "mixpanel-browser";

export type WorkflowSidebarProps = {
  expanded: boolean;
  setExpanded: (expanded: boolean) => void;
};

export default function WorkflowSidebar(props: WorkflowSidebarProps) {
  const { expanded, setExpanded } = props;

  const { t } = useTranslations();
  const closeSidebarHandler = () => {
    mixpanel.track("Workflow Toggle Sidebar");
    setExpanded(false);
  };

  const [sectionState, setSectionState] = useState({
    triggers: true,
    dataSources: true,
    actions: true,
    advanced: true,
  });

  const toggleSection = (section: keyof typeof sectionState) => {
    setSectionState({
      ...sectionState,
      [section]: !sectionState[section],
    });
  };

  return (
    <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: expanded })}>
      <Sidebar title={t.actions} closeSidebarHandler={closeSidebarHandler}>
        {" "}
        <SectionContainer
          title={t.triggers}
          icon={TriggersIcon}
          expanded={sectionState.triggers}
          handleToggle={() => toggleSection("triggers")}
        >
          <IntervalTriggerNodeButton />
          <CronTriggerNodeButton />
          <BalanceTriggerNodeButton />
          <ChannelOpenTriggerNodeButton />
          <ChannelCloseTriggerNodeButton />
        </SectionContainer>
        <SectionContainer
          title={t.dataSources}
          icon={DataSourcesIcon}
          expanded={sectionState.dataSources}
          handleToggle={() => toggleSection("dataSources")}
        >
          <DataSourceTorqChannelsNodeButton />
        </SectionContainer>
        <SectionContainer
          title={t.actions}
          icon={ChannelsIcon}
          expanded={sectionState.actions}
          handleToggle={() => toggleSection("actions")}
        >
          <EventFilterNodeButton />
          <ChannelFilterNodeButton />
          <ChannelPolicyAutoRunNodeButton />
          <RebalanceAutoRunNodeButton />
          <AddTagNodeButton />
          <RemoveTagNodeButton />
        </SectionContainer>
        <SectionContainer
          title={t.AdvancedActions}
          icon={TagsIcon}
          expanded={sectionState.advanced}
          handleToggle={() => toggleSection("advanced")}
        >
          <ChannelPolicyRunNodeButton />
          <ChannelPolicyConfiguratorNodeButton />
          <RebalanceRunNodeButton />
          <RebalanceConfiguratorNodeButton />
        </SectionContainer>
      </Sidebar>
    </div>
  );
}
