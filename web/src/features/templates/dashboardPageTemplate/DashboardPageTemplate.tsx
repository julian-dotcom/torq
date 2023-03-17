import React from "react";
import styles from "./dashboard-page-template.module.scss";
import classNames from "classnames";
import PageTitle from "features/templates/PageTitle";

type DashboardPageProps = {
  welcomeMessage: string;
  title: string;
  titleContent?: React.ReactNode;
  sidebarExpanded?: boolean;
  sidebar?: React.ReactNode;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  breadcrumbs?: Array<any>;
  children?: React.ReactNode;
};

function DashboardPage(props: DashboardPageProps) {
  return (
    <div className={styles.contentWrapper}>
      <PageTitle breadcrumbs={props.breadcrumbs} title={props.title} className={styles.dashboardPageTitle}>
        {props.titleContent}
      </PageTitle>

      <div className={styles.dashboardPageContent}>{props.children}</div>

      <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: props.sidebarExpanded })}>
        {props.sidebar}
      </div>
    </div>
  );
}

const memoizedDashboardPage = React.memo(DashboardPage);
export default memoizedDashboardPage;
