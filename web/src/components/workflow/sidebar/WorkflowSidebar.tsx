import useTranslations from "services/i18n/useTranslations";
import classNames from "classnames";
import styles from "./workflow_sidebar.module.scss";
import Sidebar from "features/sidebar/Sidebar";
import { SectionContainer } from "features/section/SectionContainer";
import {
  Timer20Regular as TriggersIcon,
  ArrowRouting20Regular as ChannelsIcon,
  Tag20Regular as TagsIcon,
} from "@fluentui/react-icons";
import { useState } from "react";
import {
  ChannelPolicyConfigurationNodeButton,
  TimeTriggerNodeButton,
  CronTriggerNodeButton,
  ChannelFilterNodeButton,
  RebalanceConfiguratorNodeButton,
  RemoveTagNodeButton,
  BalanceTriggerNodeButton,
  AddTagNodeButton,
  ChannelCloseTriggerNodeButton,
  ChannelOpenTriggerNodeButton,
} from "components/workflow/nodes/nodes";

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
      <Sidebar title={t.actions} closeSidebarHandler={closeSidebarHandler}>
        {" "}
        <SectionContainer
          title={t.triggers}
          icon={TriggersIcon}
          expanded={sectionState.triggers}
          handleToggle={() => toggleSection("triggers")}
        >
          <TimeTriggerNodeButton />
          <CronTriggerNodeButton />
          <BalanceTriggerNodeButton />
          <ChannelOpenTriggerNodeButton />
          <ChannelCloseTriggerNodeButton />
        </SectionContainer>
        <SectionContainer
          title={t.channels}
          icon={ChannelsIcon}
          expanded={sectionState.actions}
          handleToggle={() => toggleSection("actions")}
        >
          <ChannelPolicyConfigurationNodeButton />
          <ChannelFilterNodeButton />
          <RebalanceConfiguratorNodeButton />
          <SectionContainer
            title={t.tags}
            icon={TagsIcon}
            expanded={sectionState.actions}
            handleToggle={() => toggleSection("actions")}
          >
            <AddTagNodeButton />
            <RemoveTagNodeButton />
          </SectionContainer>
        </SectionContainer>
      </Sidebar>
    </div>
  );
}
