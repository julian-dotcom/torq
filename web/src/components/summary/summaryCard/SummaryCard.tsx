import styles from "components/summary/summaryCard/summary-card.module.scss";

import { Eye20Regular as InspectIcon } from "@fluentui/react-icons";
import classNames from "classnames";
import React, { useState } from "react";
import { format } from "d3";

export type SummaryCardProps = {
  heading: string;
  value?: number;
  valueLabel: string;
  summaryClassOverride?: string;
  details?: React.ReactNode;
};

const formatter = format(",.0f");
export default function SummaryCard(props: SummaryCardProps) {
  const [showInspection, setShowInspection] = useState<boolean>(false);

  const value = props.value ? props.value : 0;
  return (
    <div className={props.summaryClassOverride ? props.summaryClassOverride : styles.summaryCard}>
      <div className={styles.headerContainer}>
        <div className={styles.heading}>{props.heading}</div>
        <div className={classNames(styles.heading, styles.icon)}>
          {props.details && (
            <InspectIcon
              // onMouseEnter={() => setShowInspection(true)}
              // onMouseLeave={() => setShowInspection(false)}
              onClick={() => setShowInspection(!showInspection)}
            />
          )}
        </div>
      </div>

      <div className={styles.valueContainer}>
        <div className={styles.value}>{formatter(value)}</div>
        <div className={styles.valueLabel}>{props.valueLabel}</div>
      </div>

      {showInspection && props.details && <div className={classNames(styles.inspectionContainer)}>{props.details}</div>}
    </div>
  );
}
