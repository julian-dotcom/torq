import styles from "components/summary/summaryCard/summary-card.module.scss";
import { Eye20Regular as InspectIcon } from "@fluentui/react-icons";
import classNames from "classnames";
import React, { useState } from "react";
import { format } from "d3";
import useTouchDevice from "features/touch/useTouchDevice";

export type valueLabel = "" | "btc";

export type SummaryCardProps = {
  heading: string;
  value?: number;
  valueLabel: valueLabel;
  summaryClassOverride?: string;
  details?: React.ReactNode;
};

const formatter = format(",.2f");
export default function SummaryCard(props: SummaryCardProps) {
  const [showInspection, setShowInspection] = useState<boolean>(false);
  const { isTouchDevice } = useTouchDevice();

  let value = props.value ? props.value : 0;
  let useFormatter = false;
  if (props.valueLabel === "btc" && value > 0) {
    value = value / 100000000;
    useFormatter = true;
  }

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
      className={classNames(
        { [styles.expanded]: showInspection && props.details },
        styles.summaryCard,
        props.summaryClassOverride ? props.summaryClassOverride : ""
      )}
      onMouseEnter={() => handleHover(true)}
      onMouseLeave={() => handleHover(false)}
      onClick={handleClick}
    >
      <div className={styles.headerContainer}>
        <div className={styles.heading}>{props.heading}</div>
        <div className={classNames(styles.heading, styles.icon)}>{props.details && <InspectIcon />}</div>
      </div>

      <div className={styles.valueContainer}>
        <div className={styles.value}>{useFormatter ? formatter(value) : value}</div>
        <div className={styles.valueLabel}>{props.valueLabel}</div>
      </div>
      {props.details && <div>{props.details}</div>}
    </div>
  );
}
