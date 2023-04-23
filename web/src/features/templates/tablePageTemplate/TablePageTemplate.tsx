import React from "react";
import styles from "./table-page-template.module.scss";

import classNames from "classnames";
import PageTitle from "features/templates/PageTitle";

type TablePageTemplateProps = {
  title: string;
  titleContent?: React.ReactNode;
  sidebarExpanded?: boolean;
  sidebar?: React.ReactNode;
  pagination?: React.ReactNode;
  pageTotals?: React.ReactNode;
  tableControls?: React.ReactNode;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  breadcrumbs?: Array<any>;
  children?: React.ReactNode;
  onNameChange?: (title: string) => void;
  isDraft?: boolean;
};

export default function TablePageTemplate(props: TablePageTemplateProps) {
  return (
    <div className={classNames(styles.contentWrapper)}>
      <PageTitle
        breadcrumbs={props.breadcrumbs}
        title={props.title}
        onNameChange={props.onNameChange}
        isDraft={props.isDraft}
      >
        {props.titleContent}
      </PageTitle>

      {props.tableControls}

      <div className={styles.tableWrapper}>
        <div className={styles.tableContainer}>
          <div className={styles.tableExpander}>{props.children}</div>
        </div>
      </div>

      <div className={classNames(styles.paginationWrapper)}>{props.pagination}</div>

      <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: props.sidebarExpanded })}>
        {props.sidebar}
      </div>
    </div>
  );
}

export function TableControlSection(props: { intercomTarget: string; children?: React.ReactNode }) {
  return (
    <div className={classNames(styles.tableControlsSection)} data-intercom-target={props.intercomTarget}>
      {props.children}
    </div>
  );
}

export function TableControlsButtonGroup(props: { intercomTarget: string; children?: React.ReactNode }) {
  return (
    <div className={classNames(styles.tableControlsButtonGroup)} data-intercom-target={props.intercomTarget}>
      {props.children}
    </div>
  );
}
