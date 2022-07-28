import React from "react";
import styles from "./table-page-template.module.scss";
import TableControls from "../table/controls/TableControls";

type TablePageTemplateProps = {
  title: string;
  children: React.ReactNode;
};

function TablePageTemplate(props: TablePageTemplateProps) {
  return (
    <div className={styles.contentWrapper}>
      <div className={styles.contentControlsWrapper}>
        <h1 className={styles.titleContainer}>{props.title}</h1>
        <TableControls />
      </div>
      <div className={styles.tableWrapper}>{props.children}</div>
    </div>
  );
}

export default TablePageTemplate;
