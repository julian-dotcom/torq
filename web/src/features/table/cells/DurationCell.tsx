import styles from "./cell.module.scss";
import React from "react";
import { format } from "d3";
import { formatDuration, intervalToDuration } from "date-fns";
const classNames = require("classnames");

interface DurationCellProps {
  seconds: number;
  className?: string;
}

const subSecFormat = format("0.2f");

function DurationCell({ seconds, className }: DurationCellProps) {
  let pif = "Unknown";
  if (seconds >= 1) {
    const d = intervalToDuration({ start: 0, end: seconds * 1000 });
    pif = formatDuration({
      years: d.years,
      months: d.months,
      days: d.days,
      hours: d.hours,
      minutes: d.minutes,
      seconds: d.seconds,
    });
  } else if (seconds < 1 && seconds > 0) {
    pif = `${subSecFormat(seconds)} seconds`;
  }
  return (
    <div className={classNames(styles.cell, styles.alignLeft, styles.DurationCell, className)}>
      <div className={styles.current}>{pif}</div>
    </div>
  );
}

const DurationCellMemo = React.memo(DurationCell);
export default DurationCellMemo;
