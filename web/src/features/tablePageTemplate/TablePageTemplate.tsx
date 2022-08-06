import React from "react";
import styles from "./table-page-template.module.scss";
import Breadcrumbs from "../breadcrumbs/Breadcrumbs";

import classNames from "classnames";
import TimeIntervalSelect from "features/timeIntervalSelect/TimeIntervalSelect";
import { FluentIconsProps } from "@fluentui/react-icons";
import { NavLink } from "react-router-dom";

type TablePageTemplateProps = {
  title: string;
  sidebarExpanded?: boolean;
  sidebar?: React.ReactNode;
  pageTotals?: React.ReactNode;
  tableControls?: React.ReactNode;
  breadcrumbs?: Array<any>;
  children?: React.ReactNode;
};

export default function TablePageTemplate(props: TablePageTemplateProps) {
  return (
    <div className={styles.contentWrapper}>
      <div className={styles.pageControlsWrapper}>
        <div className={styles.leftWrapper}>
          <Breadcrumbs breadcrumbs={props.breadcrumbs || []} />
          <h1 className={styles.titleContainer}>{props.title}</h1>
        </div>
        <div className={styles.rightWrapper}>
          <TimeIntervalSelect />
        </div>
      </div>

      <div className={classNames(styles.totalWrapper)}>{props.pageTotals}</div>

      {props.tableControls}

      <div className={styles.tableWrapper}>
        <div className={styles.tableContainer}>
          <div className={styles.tableExpander}>{props.children}</div>
        </div>
      </div>

      <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: props.sidebarExpanded })}>
        {props.sidebar}
      </div>
    </div>
  );
}

type TableControlsButtonProps = {
  icon: React.FC<FluentIconsProps>;
  onClickHandler?: (event: React.MouseEvent<HTMLDivElement, MouseEvent>) => void;
  active?: boolean;
};

export function TableControlsButton(props: TableControlsButtonProps) {
  return (
    <div
      className={classNames(styles.tableControlsButtonWrapper, { [styles.active]: props.active })}
      onClick={props.onClickHandler}
    >
      <div className={styles.tableControlsButtonIcon}>
        <props.icon />
      </div>
      {/*<div className={styles.title}>{props.title}</div>*/}
    </div>
  );
}

export function TableControlSection(props: { children?: React.ReactNode }) {
  return <div className={classNames(styles.tableControlsSection)}>{props.children}</div>;
}

export function TableControlsButtonGroup(props: { children?: React.ReactNode }) {
  return <div className={classNames(styles.tableControlsButtonGroup)}>{props.children}</div>;
}

export function TableControlsTabsGroup(props: { children?: React.ReactNode }) {
  return <div className={classNames(styles.tableControlsTabsGroup)}>{props.children}</div>;
}
