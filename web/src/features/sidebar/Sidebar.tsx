import React, { useLayoutEffect, useRef, useState } from "react";
import {
  Options24Regular as SideBarIcon,
  DismissCircle24Regular as CloseIcon,
  FluentIconsProps,
  ChevronDown20Regular as CollapsedIcon,
  LineHorizontal120Regular as ExpandedIcon,
} from "@fluentui/react-icons";

import styles from "./sidebar.module.scss";
import classNames from "classnames";

type SidebarProps = {
  title: string;
  children?: React.ReactNode;
  closeSidebarHandler: (event: React.MouseEvent<HTMLDivElement, MouseEvent>) => void;
};

function Sidebar(props: SidebarProps) {
  return (
    <div className={styles.sidebarWrapper}>
      <div className={styles.titleContainer}>
        <div className={styles.icon}>
          <SideBarIcon />
        </div>
        <div className={styles.title}>{props.title}</div>
        <div className={styles.close} onClick={props.closeSidebarHandler}>
          <CloseIcon />
        </div>
      </div>
      <div className={styles.sidebarContent}>{props.children}</div>
    </div>
  );
}

type SidebarSectionProps = {
  title: string;
  icon: React.FC<FluentIconsProps>;
  children: React.ReactNode;
  expanded?: boolean;
  handleToggle?: (event: React.MouseEvent<HTMLDivElement, MouseEvent>) => void;
};

export function SidebarSection(props: SidebarSectionProps) {
  const ref = useRef<HTMLDivElement>(null);
  const [styleState, setStyleState] = useState({});

  useLayoutEffect(() => {
    if (ref.current) {
      setStyleState({ height: props.expanded ? ref.current.scrollHeight + "px" : "0" });
    }
  }, [props]);

  return (
    <div className={styles.sectionContainer}>
      <div className={styles.sectionTitleContainer} onClick={props.handleToggle}>
        <div className={styles.sidebarIcon}>
          <props.icon />
        </div>
        <div className={styles.sidebarTitle}>{props.title}</div>
        <div className={styles.sidebarIcon}>{props.expanded ? <ExpandedIcon /> : <CollapsedIcon />}</div>
      </div>
      <div className={styles.sidebarSectionContent} ref={ref} style={styleState}>
        {props.children}
      </div>
    </div>
  );
}

export default Sidebar;
