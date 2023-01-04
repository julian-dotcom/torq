import React from "react";
import styles from "./details-page-template.module.scss";
import PageTitle from "features/templates/PageTitle";
import classNames from "classnames";

type DetailsPageProps = {
  title: string;
  titleContent?: React.ReactNode;
  sidebarExpanded?: boolean;
  sidebar?: React.ReactNode;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  breadcrumbs?: Array<any>;
  children?: React.ReactNode;
};

function DetailsPage(props: DetailsPageProps) {
  return (
    <div className={styles.contentWrapper}>
      <PageTitle breadcrumbs={props.breadcrumbs} title={props.title} className={styles.detailsPageTitle}>
        {props.titleContent}
      </PageTitle>

      <div className={styles.detailsPageContent}>{props.children}</div>

      <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: props.sidebarExpanded })}>
        {props.sidebar}
      </div>
    </div>
  );
}

const memoizedDetailsPage = React.memo(DetailsPage);
export default memoizedDetailsPage;
