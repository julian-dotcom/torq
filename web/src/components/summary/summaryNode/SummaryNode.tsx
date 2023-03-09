import styles from "components/summary/summaryNode/summary-node.module.scss";
import React from "react";
export type SummaryNodeProps = {
  nodeName: string;
  children?: React.ReactNode;
};

export default function SummaryNode(props: SummaryNodeProps) {
  return (
    <div className={styles.nodeSummaryContainer}>
      <div className={styles.header}>{props.nodeName}</div>
      {props.children}
    </div>
  );
}
