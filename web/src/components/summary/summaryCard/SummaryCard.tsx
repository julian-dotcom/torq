import styles from "components/summary/summaryCard/summary-card.module.scss";

import { Eye20Regular as InspectIcon } from "@fluentui/react-icons";
import classNames from "classnames";
import { useState } from "react";

export type SummaryCardProps = {
  heading: string;
  value: string;
  valueLabel: string;

  canInspect?: boolean;
  summaryClassOverride?: string;
};

export default function SummaryCard(props: SummaryCardProps) {
  const { canInspect } = props;
  const [showInspection, setShowInspection] = useState<boolean>(false);
  return (
    <div className={props.summaryClassOverride ? props.summaryClassOverride : styles.summaryCard}>
      <div className={styles.headerContainer}>
        <div className={styles.heading}>{props.heading}</div>
        <div className={classNames(styles.heading, styles.icon)}>
          {canInspect && (
            <InspectIcon
              // onMouseEnter={() => setShowInspection(true)}
              // onMouseLeave={() => setShowInspection(false)}
              onClick={() => setShowInspection(!showInspection)}
            />
          )}
        </div>
      </div>

      <div className={styles.valueContainer}>
        <div className={styles.value}>{props.value}</div>
        <div className={styles.valueLabel}>{props.valueLabel}</div>
      </div>

      {showInspection && (
        <div className={classNames(styles.inspectionContainer)}>
          <div>
            <dl>
              <dt>Confirmed</dt>
              <dd>600,512,313</dd>
              <dt>Unconfirmed</dt>
              <dd>100,123,543</dd>
              <dt>Locked Balance</dt>
              <dd>0</dd>
            </dl>
          </div>
        </div>
      )}
    </div>
  );
}
