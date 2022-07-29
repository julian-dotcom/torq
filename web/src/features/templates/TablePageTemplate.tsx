import React from "react";
import styles from "./table-page-template.module.scss";
import TableControls from "../table/controls/TableControls";
import Breadcrumbs from "../breadcrumbs/Breadcrumbs";
import Sidebar from "../sidebar/Sidebar";

import classNames from "classnames";

type TablePageTemplateProps = {
  title: string;
  sidebarExpanded?: boolean;
  sidebarChildren?: React.ReactNode;
  breadcrumbs?: Array<any>;
  children?: React.ReactNode;
};

function TablePageTemplate(props: TablePageTemplateProps) {
  return (
    <div className={styles.contentWrapper}>
      <div className={styles.pageControlsWrapper}>
        <Breadcrumbs breadcrumbs={props.breadcrumbs || []} />
        <h1 className={styles.titleContainer}>{props.title}</h1>
        <TableControls />
      </div>
      <div className={styles.tableWrapper}>
        <div className={styles.tableContainer}>{props.children}</div>
      </div>

      <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: props.sidebarExpanded })}>
        <Sidebar title={"Table settings"}>{props.sidebarChildren}</Sidebar>
      </div>
    </div>
  );
}

export default TablePageTemplate;
