import styles from "components/summary/summaryCard/summary-card.module.scss";

import { Eye20Regular as InspectIcon } from "@fluentui/react-icons";
import classNames from "classnames";
export type SummaryCardProps = {
  heading: string;
  value: string;
  valueLabel: string;

  canInspect?: boolean;
};

export default function SummaryCard(props: SummaryCardProps) {
  const { canInspect } = props;
  return (
    <div className={styles.summaryCard}>
      <div className={styles.headerContainer}>
        <div className={styles.heading}>{props.heading}</div>
        <div className={classNames(styles.heading, styles.icon)}>{canInspect && <InspectIcon />}</div>
      </div>

      <div className={styles.valueContainer}>
        <div className={styles.value}>{props.value}</div>
        <div className={styles.valueLabel}>{props.valueLabel}</div>
      </div>
    </div>
  );
}
