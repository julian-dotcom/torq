import styles from "components/summary/summaryCard/summary-card.module.scss";
import { Eye20Regular as InspectIcon } from "@fluentui/react-icons";
import classNames from "classnames";
import React, { useState } from "react";
import useTouchDevice from "features/touch/useTouchDevice";

export type valueLabel = "" | "btc";

export type SummaryCardProps = {
  heading: string;
  value?: number;
  valueLabel: valueLabel;
  details?: React.ReactNode;
  children?: React.ReactNode;
};

function formatValue(value: number, valueLabel: valueLabel): string {
  if (valueLabel === "btc" && value > 0) {
    value = value / 100000000;
    return value.toFixed(8);
  }
  return value.toString();
}
export default function SummaryCard(props: SummaryCardProps) {
  const [showInspection, setShowInspection] = useState<boolean>(false);
  const { isTouchDevice } = useTouchDevice();
  const handleHover = (entering: boolean) => {
    if (!isTouchDevice) {
      setShowInspection(entering);
    }
  };

  const handleClick = () => {
    if (isTouchDevice) {
      setShowInspection(!showInspection);
    }
  };

  return (
    <div
      className={classNames({ [styles.expanded]: showInspection && props.details }, styles.summaryCard)}
      onMouseEnter={() => handleHover(true)}
      onMouseLeave={() => handleHover(false)}
      onClick={handleClick}
    >
      <div className={styles.headerContainer}>
        <div className={styles.heading}>{props.heading}</div>
        <div className={classNames(styles.heading, styles.icon)}>{props.details && <InspectIcon />}</div>
      </div>

      <div className={styles.valueContainer}>
        {props.children && (
          <>
            <div className={styles.value}>{props.children}</div>
            <div className={styles.valueLabel}>{props.valueLabel}</div>
          </>
        )}
        {!props.children && (
          <>
            <div className={styles.value}>{formatValue(props.value ?? 0, props.valueLabel)}</div>
            <div className={styles.valueLabel}>{props.valueLabel}</div>
          </>
        )}
      </div>
      {props.details && <div className={styles.detailsContainer}>{props.details}</div>}
    </div>
  );
}
