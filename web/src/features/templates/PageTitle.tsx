import React from "react";
import styles from "./templates.module.scss";
import Breadcrumbs from "features/breadcrumbs/Breadcrumbs";
import classNames from "classnames";

type PageTitleProps = {
  title: string;
  breadcrumbs?: Array<any>;
  className?: string;
  children?: React.ReactNode;
};

function PageTitle(props: PageTitleProps) {
  return (
    <div className={classNames(styles.pageTitleWrapper, props.className)}>
      <div className={styles.leftWrapper}>
        <Breadcrumbs breadcrumbs={props.breadcrumbs || []} />
        <h1 className={styles.titleContainer}>{props.title}</h1>
      </div>
      <div className={styles.rightWrapper}>{props.children}</div>
    </div>
  );
}

const memoizedPageTitle = React.memo(PageTitle);
export default memoizedPageTitle;
