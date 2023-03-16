import styles from "components/summary/summaryNode/summary-node.module.scss";
import React from "react";
import { FlashCheckmark20Regular as Indicator } from "@fluentui/react-icons";
import classNames from "classnames";
import useTranslations from "services/i18n/useTranslations";

enum nodeStatus {
  inactive = 0,
  active = 1,
}
export type SummaryNodeProps = {
  nodeName: string;
  status: nodeStatus;
  children?: React.ReactNode;
};

export default function SummaryNode(props: SummaryNodeProps) {
  const { t } = useTranslations();
  return (
    <div className={styles.nodeSummaryContainer}>
      <div className={styles.headerContainer}>
        <div className={styles.header}>{props.nodeName}</div>
        <div
          className={classNames(
            styles.statusContainer,
            props.status === nodeStatus.active ? styles.online : styles.offline
          )}
        >
          <div className={styles.statusIcon}>
            <Indicator />
          </div>
          <div
            className={classNames(
              styles.statusText,
              props.status === nodeStatus.active ? styles.online : styles.offline
            )}
          >
            {props.status === nodeStatus.active ? t.summaryNode.online : t.summaryNode.offline}
          </div>
        </div>
      </div>

      {props.children}
    </div>
  );
}
