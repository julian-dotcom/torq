import React from "react";
import { Options24Regular as SideBarIcon, DismissCircle24Regular as CloseIcon } from "@fluentui/react-icons";

import styles from "./sidebar.module.scss";

type SidebarProps = {
  title: string;
  children?: React.ReactNode;
};

function Sidebar(props: SidebarProps) {
  return (
    <div className={styles.sidebarContainer}>
      <div className={styles.titleContainer}>
        <div className={styles.icon}>
          <SideBarIcon />
        </div>
        <div className={styles.title}>{props.title}</div>
        <div className={styles.close}>
          <CloseIcon />
        </div>
      </div>
      <div className={styles.sidebarContent}>{props.children}</div>
    </div>
  );
}

export default Sidebar;
