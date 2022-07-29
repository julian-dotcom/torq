import React from "react";

import styles from "./sidebar.module.scss";
import { FluentIconsProps } from "@fluentui/react-icons";
import { ChevronDown20Regular as CollapsedIcon, LineHorizontal120Regular as ExpandedIcon } from "@fluentui/react-icons";

type SidebarSectionProps = {
  title: string;
  icon: React.FC<FluentIconsProps>;
  children: React.ReactNode;
  collapsed?: boolean;
};

function SidebarSection(props: SidebarSectionProps) {
  return (
    <div className={styles.SectionContainer}>
      <div className={styles.SidebarTitleContainer} onClick={() => console.log("hello")}>
        <div className={styles.sidebarIcon}>
          <props.icon />
        </div>
        <div className={styles.SidebarTitle}>{props.title}</div>
        <div className={styles.sidebarIcon}>{props.collapsed ? <CollapsedIcon /> : <ExpandedIcon />}</div>
      </div>
      {!props.collapsed && <div className={styles.SidebarSectionContent}>{props.children}</div>}
    </div>
  );
}

export default SidebarSection;
